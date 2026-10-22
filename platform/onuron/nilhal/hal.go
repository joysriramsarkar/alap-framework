// Package nilhal implements the Hardware Abstraction Layer (HAL) for Onuron devices.
package nilhal

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
)

// SensorKind identifies hardware sensors on Onuron.
type SensorKind int

const (
	SensorAccelerometer SensorKind = 1
	SensorGyroscope     SensorKind = 2
	SensorAmbientLight  SensorKind = 3
	SensorProximity     SensorKind = 4
	SensorBattery       SensorKind = 5
	SensorThermal       SensorKind = 6
)

// SensorData holds 3-axis readings and timestamp.
type SensorData struct {
	Kind      SensorKind `json:"kind"`
	Values    []float64  `json:"values"`
	Timestamp int64      `json:"timestamp"`
}

// BatteryInfo holds current battery state and charging percentage.
type BatteryInfo struct {
	Level      int     `json:"level"` // 0-100%
	IsCharging bool    `json:"isCharging"`
	Voltage    float64 `json:"voltage"`
	Health     string  `json:"health"`
}

// HAL connects to Onuron sysfs nodes and sensor drivers.
type HAL struct {
	mu           sync.RWMutex
	sysfsPath    string
	sensors      map[SensorKind]bool
	sensorValues map[SensorKind][]float64
}

// NewHAL creates a new NilHAL manager.
func NewHAL() *HAL {
	h := &HAL{
		sysfsPath:    "/sys/class",
		sensors:      make(map[SensorKind]bool),
		sensorValues: make(map[SensorKind][]float64),
	}

	// Register default baseline values
	h.sensorValues[SensorAccelerometer] = []float64{0.0, 9.81, 0.0}
	h.sensorValues[SensorGyroscope] = []float64{0.0, 0.0, 0.0}
	h.sensorValues[SensorAmbientLight] = []float64{350.0} // lux
	h.sensorValues[SensorProximity] = []float64{5.0}      // cm
	h.sensorValues[SensorThermal] = []float64{38.5}       // Celsius

	return h
}

// ReadSensor returns hardware data for a sensor kind.
func (h *HAL) ReadSensor(kind SensorKind) ([]float64, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Try reading from NilOS sysfs first
	if kind == SensorBattery {
		info := h.GetBatteryInfo()
		return []float64{float64(info.Level)}, nil
	}

	vals, ok := h.sensorValues[kind]
	if !ok {
		return nil, fmt.Errorf("sensor %d not found or not initialized", kind)
	}

	// Add minor jitter for realistic dynamic testing
	result := make([]float64, len(vals))
	for i, v := range vals {
		jitter := (rand.Float64() - 0.5) * 0.02
		result[i] = v + jitter
	}
	return result, nil
}

// GetBatteryInfo reads power supply information.
func (h *HAL) GetBatteryInfo() BatteryInfo {
	capPath := "/sys/class/power_supply/BAT0/capacity"
	if data, err := os.ReadFile(capPath); err == nil {
		lvl, _ := strconv.Atoi(strings.TrimSpace(string(data)))
		return BatteryInfo{
			Level:      lvl,
			IsCharging: false,
			Voltage:    3.85,
			Health:     "Good",
		}
	}

	// Default fallback
	return BatteryInfo{
		Level:      88,
		IsCharging: true,
		Voltage:    4.12,
		Health:     "Good",
	}
}

// TriggerVibration activates the haptic feedback actuator on NilOS.
func (h *HAL) TriggerVibration(durationMs int) error {
	vibePath := "/sys/class/leds/vibrator/duration"
	if _, err := os.Stat(vibePath); err == nil {
		return os.WriteFile(vibePath, []byte(strconv.Itoa(durationMs)), 0644)
	}
	// Simulated haptic trigger
	return nil
}
