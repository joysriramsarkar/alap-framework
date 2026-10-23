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

The generated app loads the bytecode through the C ABI (`abi/nilabi.h`) and can
honour `permissions` declared in `alap.yaml`.

---

## iOS (`platform/ios`)

`GenerateProject` emits a Swift/Metal view hierarchy plus the embedded bytecode
and a bridging header for the C ABI.

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

Emits a web host bundle. The runtime logic lives in `web/*` (see
[`WEB.md`](./WEB.md)); this adapter wires it to a build output.

---

## The C ABI (all native targets)

Native hosts embed the runtime through `abi/nilabi.h`:

```c
NilContext ctx = nilx_runtime_create();
NilResult  r   = nilx_runtime_run(ctx, bytecode, size);
if (!r.ok) { /* r.err.message */ }

bool handled = nilx_runtime_dispatch_touch(ctx, x, y);
char *json   = nilx_runtime_get_ui_json(ctx);
/* ... */
nilx_runtime_destroy(ctx);
```

The Go side boxes integer handles in C memory so it never converts a `uintptr`
back to an `unsafe.Pointer` (see `abi/nilabi.go`).

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
