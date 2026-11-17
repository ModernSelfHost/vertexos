package wrappers

import (
	"context"
	"encoding/json"
	"time"
)

// Pool is a simplified ZFS pool view.
type Pool struct {
	Name   string `json:"name"`
	Health string `json:"health"`
	Size   string `json:"size"`
	Alloc  string `json:"alloc"`
	Free   string `json:"free"`
}

// ListPools runs `zpool list -j` and parses the result.
func ListPools(ctx context.Context) ([]Pool, error) {
	out, err := RunCombined(ctx, 30*time.Second, "zpool", "list", "-j")
	if err != nil {
		return nil, err
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return nil, err
	}
	poolsRaw, _ := raw["pools"].([]any)
	var pools []Pool
	for _, p := range poolsRaw {
		m, _ := p.(map[string]any)
		pool := Pool{
			Name:   strField(m, "name"),
			Health: strField(m, "health"),
			Size:   strField(m, "size"),
			Alloc:  strField(m, "alloc"),
			Free:   strField(m, "free"),
		}
		// zpool list -j nests properties sometimes; try top-level and "properties".
		if pool.Health == "" {
			pool.Health = strNested(m, "health")
		}
		pools = append(pools, pool)
	}
	return pools, nil
}

// Dataset is a simplified ZFS dataset view.
type Dataset struct {
	Name       string `json:"name"`
	Used       string `json:"used"`
	Avail      string `json:"avail"`
	Refer      string `json:"refer"`
	Mountpoint string `json:"mountpoint"`
}

// ListDatasets runs `zfs list -j`.
func ListDatasets(ctx context.Context) ([]Dataset, error) {
	out, err := RunCombined(ctx, 30*time.Second, "zfs", "list", "-j")
	if err != nil {
		return nil, err
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return nil, err
	}
	dsRaw, _ := raw["datasets"].([]any)
	var out2 []Dataset
	for _, d := range dsRaw {
		m, _ := d.(map[string]any)
		out2 = append(out2, Dataset{
			Name:       strField(m, "name"),
			Used:       strField(m, "used"),
			Avail:      strField(m, "avail"),
			Refer:      strField(m, "refer"),
			Mountpoint: strField(m, "mountpoint"),
		})
	}
	return out2, nil
}

func strField(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key]; ok {
		switch t := v.(type) {
		case string:
			return t
		case map[string]any:
			// property form: {"value":"..."}
			if val, ok := t["value"]; ok {
				if s, ok := val.(string); ok {
					return s
				}
			}
		}
	}
	return ""
}

func strNested(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key]; ok {
		if mm, ok := v.(map[string]any); ok {
			if val, ok := mm["value"]; ok {
				if s, ok := val.(string); ok {
					return s
				}
			}
		}
	}
	return ""
}
