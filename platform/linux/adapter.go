// Package linux implements the Alap platform adapter for Linux (Wayland, X11, AppImage, Flatpak).
package linux

import (
	"fmt"
	"os"
	"path/filepath"
)

// Adapter handles generating Linux native bundles and desktop entries.
type Adapter struct {
	Display string // "wayland" | "x11" | "drm"
	AppName string
	AppID   string
}

func New() *Adapter {
	return &Adapter{
		Display: "wayland",
		AppName: "AlapApp",
		AppID:   "org.alap.app",
	}
}

// GenerateProject builds a complete Linux desktop application bundle.
func (a *Adapter) GenerateProject(outputDir string, bytecode []byte) error {
	linuxDir := filepath.Join(outputDir, "linux")
	appDir := filepath.Join(linuxDir, "AppDir")
	usrBin := filepath.Join(appDir, "usr", "bin")
	usrShareApps := filepath.Join(appDir, "usr", "share", "applications")
	usrShareIcons := filepath.Join(appDir, "usr", "share", "icons", "hicolor", "256x256", "apps")

	dirs := []string{linuxDir, appDir, usrBin, usrShareApps, usrShareIcons}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed creating directory %s: %w", d, err)
		}
	}

	// 1. Desktop Entry (XDG .desktop)
	desktopFile := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=%s
Exec=alap_app
Icon=%s
Categories=Utility;Development;
Terminal=false
StartupNotify=true
StartupWMClass=%s
`, a.AppName, a.AppID, a.AppName)

	// 2. AppRun launcher for AppImage
	appRun := `#!/bin/bash
HERE="$(dirname "$(readlink -f "${0}")")"
export PATH="${HERE}/usr/bin:${PATH}"
export LD_LIBRARY_PATH="${HERE}/usr/lib:${LD_LIBRARY_PATH}"
exec "${HERE}/usr/bin/alap_app" "$@"
`

	// 3. Flatpak Manifest (org.alap.app.json)
	flatpakManifest := fmt.Sprintf(`{
    "app-id": "%s",
    "runtime": "org.freedesktop.Platform",
    "runtime-version": "23.08",
    "sdk": "org.freedesktop.Sdk",
    "command": "alap_app",
    "finish-args": [
        "--socket=wayland",
        "--socket=fallback-x11",
        "--share=ipc",
        "--share=network",
        "--filesystem=home:ro"
    ],
    "modules": [
        {
            "name": "alap_app",
            "buildsystem": "simple",
            "build-commands": [
                "install -D alap_app /app/bin/alap_app",
                "install -D main.nabc /app/bin/main.nabc"
            ]
        }
    ]
}
`, a.AppID)

	// 4. Standalone launch script
	runSh := `#!/bin/bash
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo "Starting Alap Application on Linux ($XDG_SESSION_TYPE)..."
exec nilc -in "$DIR/main.nabc" -run "$@"
`

	fileMap := map[string]string{
		filepath.Join(linuxDir, "run.sh"):                               runSh,
		filepath.Join(linuxDir, fmt.Sprintf("%s.json", a.AppID)):        flatpakManifest,
		filepath.Join(appDir, "AppRun"):                                 appRun,
		filepath.Join(appDir, fmt.Sprintf("%s.desktop", a.AppID)):       desktopFile,
		filepath.Join(usrShareApps, fmt.Sprintf("%s.desktop", a.AppID)): desktopFile,
	}

	for path, content := range fileMap {
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			return fmt.Errorf("failed writing %s: %w", path, err)
		}
	}

	if len(bytecode) > 0 {
		_ = os.WriteFile(filepath.Join(linuxDir, "main.nabc"), bytecode, 0644)
		_ = os.WriteFile(filepath.Join(usrBin, "main.nabc"), bytecode, 0644)
	}

	return nil
}
