package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ModernSelfHost/vertexos/internal/wrappers"
)

// Drive describes a detected block device.
type Drive struct {
	Name       string              `json:"name"`
	Device     string              `json:"device"`
	StableID   string              `json:"stable_id"`
	Model      string              `json:"model"`
	Serial     string              `json:"serial"`
	SizeBytes  uint64              `json:"size_bytes"`
	Type       string              `json:"type"`
	Mountpoint string              `json:"mountpoint"`
	Rotational bool                `json:"rotational"`
	Smart      *wrappers.DriveSmart `json:"smart,omitempty"`
	Health     string              `json:"health"`
}

type lsblkOut struct {
	Blockdevices []lsblkDev `json:"blockdevices"`
}
type lsblkDev struct {
	Name       string     `json:"name"`
	Size      any        `json:"size"`
	Type      string     `json:"type"`
	Model     string     `json:"model"`
	Serial    string     `json:"serial"`
	Mountpoint string    `json:"mountpoint"`
	Rota      any        `json:"rota"`
	Children  []lsblkDev `json:"children"`
}

// ListDrives enumerates top-level disk devices and enriches with stable IDs + SMART.
func ListDrives(ctx context.Context) ([]Drive, error) {
	out, err := wrappers.RunCombined(ctx, 30*time.Second, "lsblk", "-J", "-b", "-o", "NAME,SIZE,TYPE,MODEL,SERIAL,MOUNTPOINT,ROTA")
	if err != nil {
		return nil, err
	}
	var parsed lsblkOut
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		return nil, fmt.Errorf("parse lsblk: %w", err)
	}

	stableMap := mapByStableID()

	var drives []Drive
	var walk func(devs []lsblkDev)
	walk = func(devs []lsblkDev) {
		for _, d := range devs {
			if d.Type == "disk" {
				size := uint64(0)
				switch v := d.Size.(type) {
				case float64:
					size = uint64(v)
				case string:
					size, _ = strconv.ParseUint(strings.TrimSpace(v), 10, 64)
				case json.Number:
					size, _ = strconv.ParseUint(v.String(), 10, 64)
				}
				rot := false
				switch v := d.Rota.(type) {
				case bool:
					rot = v
				case string:
					rot = v == "1" || v == "true"
				case float64:
					rot = v != 0
				}
				drive := Drive{
					Name:       d.Name,
					Device:     "/dev/" + d.Name,
					StableID:   stableMap[d.Name],
					Model:      strings.TrimSpace(d.Model),
					Serial:     strings.TrimSpace(d.Serial),
					SizeBytes:  size,
					Type:       d.Type,
					Mountpoint: d.Mountpoint,
					Rotational: rot,
					Health:     "unknown",
				}
				// SMART (best effort)
				if sm, err := wrappers.SmartctlJSON(ctx, drive.Device); err == nil {
					smCopy := sm
					drive.Smart = &smCopy
					drive.Health = classifyHealth(sm)
					if drive.Model == "" {
						drive.Model = sm.Model
					}
					if drive.Serial == "" {
						drive.Serial = sm.Serial
					}
				}
				drives = append(drives, drive)
			}
			if len(d.Children) > 0 {
				walk(d.Children)
			}
		}
	}
	walk(parsed.Blockdevices)
	return drives, nil
}

func classifyHealth(sm wrappers.DriveSmart) string {
	if sm.Health == "FAILED" {
		return "failed"
	}
	if sm.Health == "PASSED" {
		return "healthy"
	}
	return "unknown"
}

// mapByStableID builds a map from kernel name (sda) to /dev/disk/by-id entry.
func mapByStableID() map[string]string {
	result := map[string]string{}
	entries, err := os.ReadDir("/dev/disk/by-id")
	if err != nil {
		return result
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// skip partitions? keep all
		link, err := os.Readlink(filepath.Join("/dev/disk/by-id", name))
		if err != nil {
			continue
		}
		// link like ../../sda
		base := filepath.Base(link)
		if _, ok := result[base]; !ok {
			result[base] = "/dev/disk/by-id/" + name
		}
	}
	return result
}
