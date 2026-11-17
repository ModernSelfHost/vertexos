package wrappers

import (
	"context"
	"encoding/json"
	"time"
)

// DriveSmart holds a subset of SMART data.
type DriveSmart struct {
	Device     string `json:"device"`
	Model      string `json:"model_name"`
	Serial     string `json:"serial_number"`
	Health     string `json:"smart_status"`
	TempC      int    `json:"temperature_celsius"`
	PowerOnHrs int    `json:"power_on_hours"`
}

// SmartctlJSON runs `smartctl -j -a <device>` and parses key fields.
func SmartctlJSON(ctx context.Context, device string) (DriveSmart, error) {
	out, err := RunCombined(ctx, 30*time.Second, "smartctl", "-j", "-a", device)
	// smartctl exits non-zero for some statuses; still parse JSON.
	var raw map[string]any
	if jerr := json.Unmarshal([]byte(out), &raw); jerr != nil {
		if err != nil {
			return DriveSmart{}, err
		}
		return DriveSmart{}, jerr
	}
	ds := DriveSmart{Device: device}
	if v, ok := raw["model_name"].(string); ok {
		ds.Model = v
	}
	if v, ok := raw["serial_number"].(string); ok {
		ds.Serial = v
	}
	if sm, ok := raw["smart_status"].(map[string]any); ok {
		if v, ok := sm["passed"].(bool); ok {
			if v {
				ds.Health = "PASSED"
			} else {
				ds.Health = "FAILED"
			}
		}
	}
	if t, ok := raw["temperature"].(map[string]any); ok {
		if c, ok := t["current"].(float64); ok {
			ds.TempC = int(c)
		}
	}
	if at, ok := raw["ata_smart_attributes"].(map[string]any); ok {
		if tbl, ok := at["table"].([]any); ok {
			for _, e := range tbl {
				m, _ := e.(map[string]any)
				if name, _ := m["name"].(string); name == "Power_On_Hours" {
					if rawv, ok := m["raw"].(map[string]any); ok {
						if v, ok := rawv["value"].(float64); ok {
							ds.PowerOnHrs = int(v)
						}
					}
				}
			}
		}
	}
	_ = time.Second
	return ds, nil
}
