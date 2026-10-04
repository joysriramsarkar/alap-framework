// Package onuron implements the native platform adapter for Onuron (Onuron OS).
// Connects Alap apps directly to Onuron services:
//   - nilui / nilui-gpu (Vulkan renderer & Wayland surface manager)
//   - nilbus-client (distributed IPC)
//   - nilhal (hardware & sensor abstraction)
//   - nilpkg / nilrt (signed sandbox & app lifecycle supervisor)
package onuron

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/joysriramsarkar/alap-framework/compiler/codegen"
	"github.com/joysriramsarkar/alap-framework/platform/onuron/nilbus"
	"github.com/joysriramsarkar/alap-framework/platform/onuron/nilhal"
	"github.com/joysriramsarkar/alap-framework/platform/onuron/nilui"
	"github.com/joysriramsarkar/alap-framework/runtime/vm"
)

// Adapter is the full Onuron native platform adapter.
type Adapter struct {
	mu            sync.RWMutex
	KernelVersion string
	DisplayServer string // "wayland"
	GPUBackend    string // "vulkan"
	BusClient     *nilbus.Client
	Renderer      *nilui.Renderer
	HAL           *nilhal.HAL
	initialized   bool
	activeVM      *vm.VM
	capabilities  map[string]bool
	AppName       string
	Version       string
	PackageID     string
}

// Event represents a platform input event.
type Event struct {
	Type    EventType `json:"type"`
	X       float64   `json:"x"`
	Y       float64   `json:"y"`
	Key     string    `json:"key,omitempty"`
	Char    rune      `json:"char,omitempty"`
	ScrollX float64   `json:"scrollX,omitempty"`
	ScrollY float64   `json:"scrollY,omitempty"`
}

type EventType int

const (
	EventTouchDown EventType = iota
	EventTouchMove
	EventTouchUp
	EventKeyDown
	EventKeyUp
	EventCharInput
	EventScroll
	EventResize
	EventQuit
	EventFocus
	EventBlur
	EventBackButton
)

// New creates a new Onuron platform adapter with full system service clients.
func New() *Adapter {
	return &Adapter{
		KernelVersion: "Onuron-0.1-vulkan",
		DisplayServer: "wayland",
		GPUBackend:    "vulkan",
		AppName:       "AlapApp",
		Version:       "0.1.0",
		PackageID:     "org.onuron.app",
		BusClient:     nilbus.NewClient(""),
		Renderer:      nilui.NewRenderer("AlapApp", 1080, 1920),
		HAL:           nilhal.NewHAL(),
		capabilities:  make(map[string]bool),
	}
}

// Init connects to the Onuron system services (NilBus, NilUI, NilHAL).
func (a *Adapter) Init() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.initialized {
		return nil
	}

	// 1. Connect to NilBus IPC
	if err := a.BusClient.Connect(); err != nil {
		return fmt.Errorf("failed connecting to nilbus: %w", err)
	}

	// 2. Grant default runtime capabilities
	a.capabilities["storage.read"] = true
	a.capabilities["notifications"] = true
	a.capabilities["sensors"] = true
	a.capabilities["network"] = true

	a.initialized = true
	return nil
}

// Shutdown disconnects from all Onuron system services.
func (a *Adapter) Shutdown() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.initialized {
		if a.BusClient != nil {
			a.BusClient.Close()
		}
		a.initialized = false
	}
}

// GetKernelVersion returns the active Onuron kernel version string.
func (a *Adapter) GetKernelVersion() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.KernelVersion
}

// CreateWindow creates a Wayland xdg_surface with Vulkan swapchain.
func (a *Adapter) CreateWindow(title string, width, height int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Renderer = nilui.NewRenderer(title, width, height)
	return nil
}

// CheckCapability verifies if the app has been granted a specific permission.
func (a *Adapter) CheckCapability(capName string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.capabilities[capName]
}

// SendNotification dispatches a system notification via NilBus.
func (a *Adapter) SendNotification(title, body string) error {
	if !a.CheckCapability("notifications") {
		return fmt.Errorf("permission denied: notifications")
	}

	payload := []byte(fmt.Sprintf(`{"title":%q,"body":%q}`, title, body))
	_, err := a.BusClient.Call("org.onuron.NotificationService", "Notify", payload)
	if err != nil {
		// Fallback to legacy service identifier if needed
		_, err = a.BusClient.Call("org.nilos.NotificationService", "Notify", payload)
	}
	return err
}

// ReadSensorData reads hardware telemetry via NilHAL.
func (a *Adapter) ReadSensorData(sensorID int) ([]float64, error) {
	if !a.CheckCapability("sensors") {
		return nil, fmt.Errorf("permission denied: sensors")
	}
	return a.HAL.ReadSensor(nilhal.SensorKind(sensorID))
}

// DispatchTouchEvent routes Wayland touch events to the active UI engine.
func (a *Adapter) DispatchTouchEvent(x, y float64) bool {
	a.mu.RLock()
	activeVM := a.activeVM
	a.mu.RUnlock()

	if activeVM == nil {
		return false
	}

	tree := activeVM.GetUITree()
	if tree == nil {
		return false
	}

	return tree.DispatchTouch(x, y)
}

// RenderCurrentFrame computes UI layout and exports Vulkan render commands.
func (a *Adapter) RenderCurrentFrame() (string, error) {
	a.mu.RLock()
	activeVM := a.activeVM
	renderer := a.Renderer
	a.mu.RUnlock()

	if activeVM == nil || renderer == nil {
		return "{}", nil
	}

	tree := activeVM.GetUITree()
	if tree == nil || tree.Root == nil {
		return "{}", nil
	}

	activeVM.ComputeUILayout(float64(renderer.Width), float64(renderer.Height))
	packet := renderer.RenderTree(tree.Root)
	return renderer.ExportFrameJSON(packet)
}

// RunApp launches and supervises a native .nilax package on Onuron.
func (a *Adapter) RunApp(bundlePath string) error {
	if err := a.Init(); err != nil {
		return err
	}

	nabcPath := filepath.Join(bundlePath, "bin", "main.nabc")
	if _, err := os.Stat(nabcPath); os.IsNotExist(err) {
		nabcPath = filepath.Join(bundlePath, "main.nabc")
	}

	data, err := os.ReadFile(nabcPath)
	if err != nil {
		return fmt.Errorf("failed reading .nilax bytecode: %w", err)
	}

	mod, err := codegen.Deserialize(data)
	if err != nil {
		return fmt.Errorf("failed deserializing bytecode: %w", err)
	}

	runner := vm.New(mod)
	a.mu.Lock()
	a.activeVM = runner
	a.mu.Unlock()

	return runner.Run()
}

// GenerateProject builds a complete native Onuron .nilax bundle structure.
func (a *Adapter) GenerateProject(outputDir string, bytecode []byte) error {
	onuronDir := filepath.Join(outputDir, "onuron")
	binDir := filepath.Join(onuronDir, "bin")
	resDir := filepath.Join(onuronDir, "res")

	dirs := []string{onuronDir, binDir, resDir}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed creating directory %s: %w", d, err)
		}
	}

	name := a.AppName
	if name == "" {
		name = "AlapApp"
	}
	ver := a.Version
	if ver == "" {
		ver = "0.1.0"
	}
	manifest := fmt.Sprintf(`[NilAx]
Name = %s
Version = %s
Platform = onuron
Entry = bin/main.nabc
Permissions = storage.read, notifications, sensors, network
DisplayServer = %s
GPUBackend = %s
KernelMinVersion = 0.1.0
`, name, ver, a.DisplayServer, a.GPUBackend)
	launcherSh := `#!/bin/sh
# onuron-launcher.sh — Native Onuron Wayland/Vulkan app launcher
DIR="$(cd "$(dirname "$0")" && pwd)"
export ONURON_DISPLAY="wayland-0"
export ONURON_GPU="vulkan"
exec nilc -in "$DIR/bin/main.nabc" -run "$@"
`
	serviceUnit := `[Unit]
Description=Alap Native Application on Onuron
After=nilbus.service nilui.service

[Service]
Type=simple
ExecStart=/usr/bin/nilc -in /app/bin/main.nabc -run
Restart=on-failure
Environment="NILBUS_SOCKET=/run/nilbus/system.sock"

[Install]
WantedBy=graphical-session.target
`

	if err := os.WriteFile(filepath.Join(onuronDir, "app.alapmanifest"), []byte(manifest), 0644); err != nil {
		return fmt.Errorf("failed writing manifest: %w", err)
	}
	if err := os.WriteFile(filepath.Join(onuronDir, "onuron-launcher.sh"), []byte(launcherSh), 0755); err != nil {
		return fmt.Errorf("failed writing launcher: %w", err)
	}
	if err := os.WriteFile(filepath.Join(onuronDir, "alap-app.service"), []byte(serviceUnit), 0644); err != nil {
		return fmt.Errorf("failed writing service unit: %w", err)
	}

	if len(bytecode) > 0 {
		if err := os.WriteFile(filepath.Join(binDir, "main.nabc"), bytecode, 0644); err != nil {
			return fmt.Errorf("failed writing bytecode: %w", err)
		}
	}

	return nil
}
