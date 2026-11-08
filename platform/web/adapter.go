// Package web implements the Alap platform adapter for Web, Browsers, and PWA.
package web

import (
	"fmt"
	"os"
	"path/filepath"
)

// Adapter handles generating Web and PWA application bundles.
type Adapter struct {
	AppName    string
	AppID      string
	Title      string
	ThemeColor string
	Hydration  bool
}

// New creates a default Web platform adapter
func New() *Adapter {
	return &Adapter{
		AppName:    "AlapWebApp",
		AppID:      "org.alap.web",
		Title:      "Alap Enterprise Web Application",
		ThemeColor: "#00d4ff",
		Hydration:  true,
	}
}

// GenerateProject builds a complete Web client bundle.
func (a *Adapter) GenerateProject(outputDir string, bytecode []byte) error {
	webDir := filepath.Join(outputDir, "web")
	staticDir := filepath.Join(webDir, "static")
	jsDir := filepath.Join(webDir, "js")

	dirs := []string{webDir, staticDir, jsDir}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed creating directory %s: %w", d, err)
		}
	}

	// 1. index.html with hydration container and meta tags
	indexHTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s</title>
  <link rel="manifest" href="manifest.json">
  <meta name="theme-color" content="%s">
  <link rel="stylesheet" href="style.css">
</head>
<body>
  <div id="alap-root" class="alap-container">
    <div class="alap-loader">
      <div class="spinner"></div>
      <p>Loading %s...</p>
    </div>
  </div>

  <script id="__NILANG_STATE__" type="application/json">{}</script>
  <script src="js/runtime.js"></script>
  <script>
    window.addEventListener('DOMContentLoaded', () => {
      AlapRuntime.boot({
        bytecodeUrl: 'main.nabc',
        rootId: 'alap-root',
        hydration: %t
      });
    });
  </script>
</body>
</html>
`, a.Title, a.ThemeColor, a.AppName, a.Hydration)

	if err := os.WriteFile(filepath.Join(webDir, "index.html"), []byte(indexHTML), 0644); err != nil {
		return err
	}

	// 2. runtime.js (Browser runtime bridge)
	runtimeJS := `/**
 * Alap Web Browser Runtime & Hydration Engine
 * Built for Nilang & Alap Framework
 */
const AlapRuntime = {
  state: {},
  components: new Map(),

  boot: function(config) {
    console.log('[Alap Web] Initializing runtime with config:', config);
    this.hydrate();
    this.bindEvents();
    this.initRouter();
  },

  hydrate: function() {
    const stateEl = document.getElementById('__NILANG_STATE__');
    if (stateEl && stateEl.textContent.trim()) {
      try {
        this.state = JSON.parse(stateEl.textContent);
        window.__NILANG_INITIAL_STATE__ = this.state;
        console.log('[Alap Web] Hydrated state:', Object.keys(this.state).length, 'keys');
      } catch (e) {
        console.warn('[Alap Web] State parse error:', e);
      }
    }
  },

  bindEvents: function() {
    document.addEventListener('click', (e) => {
      const target = e.target.closest('[data-alap-event="click"]');
      if (target) {
        const id = target.getAttribute('data-alap-id');
        console.log('[Alap Web] Click event on:', id);
      }
    });
  },

  initRouter: function() {
    window.addEventListener('popstate', (e) => {
      console.log('[Alap Web] Navigation popped:', location.pathname);
    });
  }
};
`
	if err := os.WriteFile(filepath.Join(jsDir, "runtime.js"), []byte(runtimeJS), 0644); err != nil {
		return err
	}

	// 3. style.css (Default design system)
	styleCSS := `/* Alap Web Design System */
:root {
  --primary: #00d4ff;
  --bg: #0a0a0f;
  --surface: #161622;
  --text: #f8fafc;
  --muted: #94a3b8;
  --radius: 12px;
}

body {
  margin: 0;
  padding: 0;
  background-color: var(--bg);
  color: var(--text);
  font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
}

.alap-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px;
}

.alap-loader {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 60vh;
}

.spinner {
  width: 48px;
  height: 48px;
  border: 4px solid rgba(0, 212, 255, 0.2);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
`
	if err := os.WriteFile(filepath.Join(webDir, "style.css"), []byte(styleCSS), 0644); err != nil {
		return err
	}

	// 4. PWA manifest.json
	manifestJSON := fmt.Sprintf(`{
  "name": "%s",
  "short_name": "%s",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#0a0a0f",
  "theme_color": "%s"
}
`, a.Title, a.AppName, a.ThemeColor)

	if err := os.WriteFile(filepath.Join(webDir, "manifest.json"), []byte(manifestJSON), 0644); err != nil {
		return err
	}

	// 5. Write the NABC bytecode. This is NOT WebAssembly: no Wasm
	// compilation exists yet, so the file is honestly named .nabc and a
	// browser runtime would need a NABC interpreter to execute it.
	return os.WriteFile(filepath.Join(webDir, "main.nabc"), bytecode, 0644)
}
