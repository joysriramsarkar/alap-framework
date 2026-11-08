# Alap Framework — Platform Adapters

Alap compiles a project to portable **NABC bytecode**, then hands that bytecode
to a platform adapter. Each single-target adapter implements the same shape:

```go
adapter := pkg.New()
err := adapter.GenerateProject(outputDir, bytecode)
```

The CLI selects the adapter via `alap build <platform>`.

| Target | Package | Output |
|--------|---------|--------|
| `onuron` / `nilos` | `platform/onuron` | Onuron `.nilax` bundle + native runtime hooks |
| `android` | `platform/android` | Android Studio / Gradle project with a JNI bridge |
| `ios` | `platform/ios` | iOS Swift / Metal project |
| `linux` | `platform/linux` | AppImage / Flatpak-style desktop bundle |
| `web` | `platform/web` | Web host bundle |

---

## Onuron (`platform/onuron`)

The richest adapter — it drives a live Onuron host and generates a bundle.

```go
a := onuron.New()
_ = a.Init()
_ = a.CreateWindow("My App", 1080, 1920)

if a.CheckCapability("storage.read") {
    _ = a.SendNotification("Alap", "running")
}

sensors, _ := a.ReadSensorData(1)
_ = a.DispatchTouchEvent(120, 340)
frame, _ := a.RenderCurrentFrame()
_ = a.RunApp("build/my-app.nilax")

_ = a.GenerateProject("out/onuron", bytecode)
a.Shutdown()
```

### Onuron subpackages

| Package | Responsibility |
|---------|----------------|
| `platform/onuron/nilhal` | Hardware abstraction: sensors, display, touch (evdev), audio (ALSA), power. |
| `platform/onuron/nilui` | Renders an `engine.UITree` to the Onuron surface. |
| `platform/onuron/nilbus` | IPC client to the Onuron message bus. |

---

## Android (`platform/android`)

`GenerateProject` emits a Gradle project containing the generated Java/Kotlin
host, JNI/NDK glue, `assets/` (bytecode + resources), and an Android manifest.

```bash
alap build android
```

The generated app loads the bytecode through the JNI glue and `nilabi.h`
header generated into the project's own `cpp/` directory (there is no
top-level abi package in this repo), and can honour `permissions`
declared in `alap.yaml`.

---

## iOS (`platform/ios`)

`GenerateProject` emits a Swift/Metal view hierarchy plus the embedded bytecode
and a bridging header.

```bash
alap build ios
```

---

## Linux (`platform/linux`)

`GenerateProject` emits a desktop bundle (AppDir + launcher) suitable for
AppImage/Flatpak-style packaging, targeting Wayland or X11.

```bash
alap build linux
```

---

## Web (`platform/web`)

Emits a web host bundle (`index.html`, `js/runtime.js`, `style.css`,
`manifest.json`) via:

```bash
alap build web   # → build/web/
```

The bytecode is written as `main.nabc`. It is **NABC bytecode, not
WebAssembly** — no Wasm compilation exists yet, so no `.wasm` file is
produced or claimed; a browser runtime would need a NABC interpreter to
execute it. The server-side runtime logic lives in `web/*` (see
[`WEB.md`](./WEB.md)).

---

## Native embedding

There is no top-level abi package or stable C ABI in this repository.
Generated native projects carry their own glue: e.g. the Android adapter
emits JNI code and a `nilabi.h` header inside the generated project's
`cpp/` directory. Treat that generated scaffold as project-local, not as a
versioned ABI.

---

## Status

| Adapter | Maturity |
|---------|----------|
| Onuron | Functional prototype (host simulation) |
| Android | Project generator + JNI bridge (buildable, not store-ready) |
| iOS | Project generator (not device-tested) |
| Linux | Project generator (not packaged for release) |
| Web | Functional (routing/middleware/SSR) |

Hardware backends are represented but **not production-ready**. See the
ecosystem maturity map in `onuron_project/docs/ECOSYSTEM_MATURITY.md`.
