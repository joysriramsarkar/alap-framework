# Alap Framework + NilLang — Complete Blueprint
Version: 0.1 (architecture draft)
Target platforms: Onuron, Android, iOS, Linux
Interpretation: "ISO" in the request is treated as iOS. A bootable ISO image target can be added separately.

> **Naming note (13 Oct 2026):** the Alap CLI binaries are now `alap`/`alapc`/`alap-lsp`.
> The name `nil` belongs to the canonical NilLang toolchain (`nilLang` repo).
> Command examples below read `alap …`.

---

## 0. Executive definition

**Alap** is a cross-platform application framework whose primary application language is **NilLang**.

NilLang combines:
- ArkTS-style declarative UI and component model
- TypeScript-style static typing and familiar syntax
- Go-inspired concurrency, channels, packages, tooling simplicity and explicit error handling
- A platform-neutral standard API (`nil.*`)
- Native extension ABI for Onuron, Android, iOS and Linux

The target is:

```text
                    ┌──────────────────────────────┐
                    │          NilLang App          │
                    │ .nil files + resources        │
                    └──────────────┬───────────────┘
                                   │
                         nilc / Alap compiler
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
          Typed IR              UI IR              Native IR
              │                    │                    │
              └────────────────────┼────────────────────┘
                                   │
                         NIL Bytecode / AOT
                                   │
                         ┌─────────▼─────────┐
                         │    Nil Runtime    │
                         │ GC / scheduler    │
                         │ VM / AOT / FFI    │
                         └─────────┬─────────┘
                                   │
              ┌────────────────────┼─────────────────────┐
              │                    │                     │
           Onuron                Android                  iOS
              │                    │                     │
        Wayland/Vulkan      Surface/NDK/JNI        Metal/Swift/ObjC
              │                    │                     │
              └────────────────────┴─────────────────────┘
                                   │
                                Linux
                            Wayland/X11/DRM
```

The same source project should produce:

```text
Onuron  -> .nilax
Android -> .apk / .aab
iOS     -> .app / .ipa / XCFramework
Linux   -> AppImage / Flatpak / native bundle
```

---

# 1. Project naming

Recommended names:

- Framework: `Alap`
- Language: `NilLang`
- compiler: `nilc`
- package/build tool: `nil`
- language server: `nills`
- formatter: `nilfmt`
- runtime: `nilrt`
- UI engine: `nilui`
- graphics backend: `nilgfx`
- platform ABI: `nilabi`
- native bridge: `nilbridge`
- package registry: `nilpm`
- bytecode: `.nabc`
- UI intermediate format: `.nui`
- application manifest: `alap.yaml`

Example:

```text
nil create counter
cd counter
alap run
alap build android
alap build ios
alap build onuron
alap build linux
```

---

# 2. Design principles

1. One application source tree.
2. UI is declarative and platform-neutral.
3. Business logic is platform-neutral.
4. Platform APIs are accessed through `nil.*`.
5. Native-only code is isolated under `platform/<target>/`.
6. No platform code leaks into normal application code.
7. The UI engine owns layout/rendering; operating systems provide windows, surfaces, input and system services.
8. A stable ABI matters more than a stable internal implementation.
9. A small official standard library; optional capabilities live in packages.
10. Deterministic builds and signed packages.
11. A11y, localization, deep links and lifecycle are first-class concepts.
12. The framework must work without an app store.

---

# 3. Onuron integration with the existing repository

The current Onuron repository already has the correct conceptual places for this framework:

- `runtime/nilui` — declarative UI foundation
- `runtime/nilui-gpu` — Vulkan renderer
- `runtime/nilrt` — sandbox/runtime area
- `runtime/nilhal` — hardware abstraction
- `runtime/nilbus-client` — distributed IPC
- `shell/` — Wayland compositor
- `android/` — Android compatibility layer
- Rust workspace under `Cargo.toml`

Therefore Alap should become a layer above the existing OS runtime, not replace Onuron kernel/HAL.

Recommended addition:

```text
onuron/
├── framework/
│   └── alap/
│       ├── compiler/
│       ├── runtime/
│       ├── ui/
│       ├── abi/
│       ├── sdk/
│       ├── cli/
│       ├── lsp/
│       ├── tools/
│       └── templates/
├── platform/
│   ├── onuron/
│   ├── android/
│   ├── ios/
│   └── linux/
└── apps/
```

Keep the existing Rust system services. Use Go heavily for developer tooling, service orchestration and portable framework tooling; do not rewrite working low-level Rust OS components merely to make the framework "all Go."

---

# 4. Alap architecture layers

## Layer A — Application source

```text
src/
  main.nil
  App.nil
  pages/
  components/
  state/
  services/
  models/
  native/
    android/
    ios/
    onuron/
    linux/
assets/
resources/
alap.yaml
```

## Layer B — Language frontend

Responsibilities:
- lexer
- parser
- type checker
- import/module resolver
- const evaluator
- decorator processing
- UI DSL lowering
- diagnostics
- source maps

## Layer C — Nil IR

Use one canonical intermediate representation:

```text
AST
 ↓
Typed AST
 ↓
HIR
 ↓
MIR
 ↓
NIR
 ↓
NABC / AOT
```

Where:

- HIR = user-language constructs
- MIR = explicit control flow
- NIR = runtime/platform-neutral instructions
- NUI = declarative UI tree

## Layer D — Runtime

- garbage collector
- event loop
- actor/task scheduler
- futures/promises
- channel implementation
- reflection metadata
- FFI
- filesystem abstraction
- networking abstraction
- timers
- lifecycle
- capability enforcement

## Layer E — UI engine

- retained scene tree
- diff/reconciliation
- flex/grid/stack layout
- text shaping
- accessibility tree
- input dispatch
- animation
- gestures
- clipping
- scrolling
- image decode
- theme system

## Layer F — graphics

Primary:
- Vulkan on Onuron/Linux/Android where available
- Metal on iOS
- optional OpenGL ES fallback
- software rasterizer for CI/headless

## Layer G — platform adapter

```text
NilPlatform
  ├── Window
  ├── Display
  ├── Input
  ├── Clipboard
  ├── Files
  ├── Camera
  ├── Audio
  ├── Sensors
  ├── Network
  ├── Notifications
  ├── SecureStorage
  ├── Biometrics
  ├── Contacts
  ├── Telephony
  └── Share
```

---

# 5. The most important language decision

Do NOT define NilLang as:

```text
50% Go syntax + 25% TypeScript syntax + 25% ArkTS syntax
```

That creates an awkward language.

Instead:

```text
Syntax          = TypeScript/ArkTS family
Type system     = strict static typing
UI model        = ArkUI-inspired declarative components
Concurrency     = Go-inspired tasks + channels
Modules         = Go-like simple package model
Errors          = explicit Result
Native bridge   = explicit capabilities
Build tooling   = Go-like single-command workflow
```

This gives the language a coherent grammar.

---

# 6. NilLang file model

Normal code:

```text
*.nil
```

UI code:

```text
*.nil
```

Native extensions:

```text
*.go
*.rs
*.kt
*.swift
*.cpp
```

Resources:

```text
*.nilasset
*.json
*.toml
*.png
*.svg
*.woff2
```

---

# 7. NilLang primitive types

```text
void
bool

i8
i16
i32
i64

u8
u16
u32
u64

f32
f64

bigint

char
string

bytes

null
undefined
```

Default policy:

```text
number
```

is forbidden.

The programmer must choose:

```text
i32
i64
f32
f64
```

This avoids silent numeric conversions in performance-sensitive code.

---

# 8. Composite types

```text
type User = {
  id: string
  name: string
  age: i32
}
```

Arrays:

```text
let names: string[] = ["Joy", "Nil"]
```

Map:

```text
let scores: Map<string, i32> = new Map()
```

Set:

```text
let ids: Set<string> = new Set()
```

Tuples:

```text
let point: [f32, f32] = [1.0, 2.0]
```

Enums:

```text
enum Theme {
  Light,
  Dark,
  System,
}
```

Union:

```text
type Status = "idle" | "loading" | "ready" | "error"
```

Option:

```text
let user: User?
```

Equivalent concept:

```text
Option<User>
```

---

# 9. Interfaces and classes

Interfaces:

```text
interface Repository<T> {
  get(id: string): Future<T?>
  save(value: T): Future<void>
}
```

Classes:

```text
class UserService {
  private repo: UserRepository

  constructor(repo: UserRepository) {
    this.repo = repo
  }

  async getUser(id: string): Future<User?> {
    return await this.repo.get(id)
  }
}
```

NilLang should restrict runtime object-layout mutation.

A class's fields cannot appear/disappear dynamically.

Forbidden:

```text
user["newField"] = 10
```

unless explicitly stored in:

```text
Map<string, Any>
```

This follows the practical static-layout philosophy used by ArkTS. citeturn759431search3

---

# 10. Type inference

Allowed:

```text
let count = 10
let title = "Onuron"
let enabled = true
```

But public APIs should be explicit:

```text
public function calculate(x: i64, y: i64): i64
```

---

# 11. Functions

```text
function add(a: i32, b: i32): i32 {
  return a + b
}
```

Arrow function:

```text
let add = (a: i32, b: i32): i32 => a + b
```

Function type:

```text
type Handler = (Event) => void
```

---

# 12. Error model

Do not use exceptions as the normal control flow.

Primary type:

```text
Result<T, E>
```

Example:

```text
function readUser(id: string): Result<User, UserError> {
  ...
}
```

Usage:

```text
let user = try readUser("42")
```

Or:

```text
match readUser("42") {
  Ok(value) => print(value.name)
  Err(error) => log(error)
}
```

Exceptions remain available for programmer/runtime faults:

```text
throw RuntimeError("invalid UI state")
```

---

# 13. Go-inspired concurrency

Core primitive:

```text
task
```

Example:

```text
task {
  let data = await api.fetch()
  state.update(data)
}
```

Channels:

```text
let ch: Channel<i32> = Channel(16)

task {
  await ch.send(42)
}

let value = await ch.receive()
```

Select:

```text
select {
  value <- ch => handle(value)
  timeout <- timer(1000) => handleTimeout()
}
```

Actor:

```text
actor Counter {
  private value: i64 = 0

  on Increment(amount: i64) {
    value += amount
  }

  on Get(reply: Reply<i64>) {
    reply.send(value)
  }
}
```

No shared mutable state across actors by default.

This is compatible with the direction of ArkCompiler's task/actor concurrency model. citeturn759431search0turn759431search5

---

# 14. Sendability

Any value crossing an actor/task boundary must be:

```text
Sendable
```

Primitive values are Sendable.

Immutable records can be Sendable.

Mutable UI objects cannot be transferred.

```text
sendable type DeviceInfo = {
  model: string
  os: string
}
```

---

# 15. Async model

```text
async function load(): Future<Data> {
  let response = await http.get("/data")
  return response.json<Data>()
}
```

Cancellation:

```text
let task = spawn(load(), cancelToken)

await task.cancel()
```

Structured concurrency:

```text
withTaskGroup(async group => {
  group.add(loadUsers())
  group.add(loadMessages())
  group.add(loadSettings())
})
```

---

# 16. Declarative UI

The UI syntax is intentionally ArkUI-like.

```text
component HelloCard {
  prop name: string

  build() {
    Column {
      Text("Hello " + name)
      Button("Open") {
        onClick => openHome()
      }
    }
    .padding(16)
    .spacing(10)
  }
}
```

App:

```text
app MyApp {
  window {
    title = "Onuron Demo"
    size = .fit
  }

  build() {
    HomePage()
  }
}
```

---

# 17. State

```text
state count: i32 = 0
```

Binding:

```text
Text(count.toString())
```

Mutation:

```text
Button("+") {
  onClick => count += 1
}
```

Computed state:

```text
computed doubleCount: i32 {
  return count * 2
}
```

External store:

```text
store AppStore {
  user: User?
  theme: Theme = .system

  async loadUser() {
    user = await api.currentUser()
  }
}
```

---

# 18. UI lifecycle

Supported:

```text
onCreate()
onMount()
onAppear()
onDisappear()
onUpdate()
onDispose()
```

Application lifecycle:

```text
onLaunch()
onBackground()
onForeground()
onSuspend()
onResume()
onTerminate()
```

---

# 19. UI primitives

Required first-party widgets:

```text
Text
RichText
Icon
Image

Button
IconButton
Toggle
Checkbox
Radio
Slider
Progress
Input
TextArea

Row
Column
Stack
Grid
List
LazyList
Scroll
Wrap

Card
Dialog
Sheet
Menu
Popup
Tooltip

Navigation
TabBar
Toolbar
BottomBar

Canvas
WebView
Video
Map
```

---

# 20. Layout system

Use constraint-based layout internally.

Public layout primitives:

```text
.width()
.height()
.minWidth()
.maxWidth()

.padding()
.margin()

.fill()
.fit()
.aspect()

.align()
.center()

.row()
.column()
.stack()
.grid()
```

Grid:

```text
Grid(columns: 2) {
  ProductCard(product)
  ProductCard(product2)
}
```

---

# 21. Styling

Prefer modifiers instead of CSS.

```text
Text("Onuron")
  .font(size: 24, weight: .bold)
  .foreground(.primary)
  .padding(horizontal: 16)
```

Theme:

```text
theme AppTheme {
  colors {
    primary = "#176BFF"
    background = "#FFFFFF"
  }

  typography {
    body = Font.system(16)
  }

  shape {
    cardRadius = 16
  }
}
```

---

# 22. Animation

```text
.animate(
  property: .opacity,
  duration: 180,
  curve: .easeOut
)
```

Spring:

```text
.spring(
  response: 0.35,
  damping: 0.82
)
```

Gesture:

```text
onDrag { event ->
  translateX = event.x
}
```

---

# 23. Navigation

```text
router {
  route("/", HomePage)
  route("/settings", SettingsPage)
  route("/profile/:id", ProfilePage)
}
```

Navigate:

```text
await nav.push("/settings")
```

Deep link:

```text
deepLink("nil://profile/:id")
```

---

# 24. Accessibility

Every component must expose:

```text
a11y.label
a11y.role
a11y.hint
a11y.value
a11y.enabled
a11y.focusable
```

Example:

```text
Button("Delete") {
  a11y.label = "Delete account permanently"
}
```

The UI tree and accessibility tree should be separate but synchronized.

---

# 25. Localization

```text
Text(t("settings.title"))
```

Resource:

```text
{
  "settings.title": {
    "en": "Settings",
    "bn": "সেটিংস",
    "hi": "सेटिंग्स"
  }
}
```

Plural rules and right-to-left layout are mandatory features.

---

# 26. Standard library namespaces

```text
nil.core
nil.collections
nil.async
nil.math
nil.text
nil.time
nil.fs
nil.path
nil.crypto
nil.net
nil.http
nil.json
nil.log
nil.process
nil.ui
nil.graphics
nil.audio
nil.camera
nil.sensors
nil.location
nil.notifications
nil.storage
nil.security
nil.device
nil.share
nil.media
nil.bluetooth
nil.wifi
nil.telephony
nil.accessibility
nil.distributed
```

---

# 27. Capability security

Do not expose unrestricted OS APIs.

Application manifest:

```yaml
name: com.example.notes
permissions:
  - storage.read
  - storage.write
  - notifications
  - camera
```

Code:

```text
let camera = await capabilities.camera.request()
```

The application never directly opens privileged device files.

Onuron should connect these capabilities to its existing sandbox/permission broker.

---

# 28. Native bridge

Syntax:

```text
native android {
  import "com.example.CameraBridge"
}

native ios {
  import "CameraBridge"
}
```

Portable code:

```text
let camera = platform.camera()
```

Platform-specific implementation stays outside normal source.

ABI:

```text
NilValue
NilHandle
NilString
NilBytes
NilFuture
NilError
NilObject
```

Calls cross the ABI only through stable C-compatible entry points.

---

# 29. Go integration

Go should be a first-class extension language, but not the syntax of the whole language.

Example:

```text
native go "example.com/nil/camera"
```

Go:

```go
package camera

type DeviceInfo struct {
    Model string
}

func Open() (*DeviceInfo, error) {
    return &DeviceInfo{Model: "Generic"}, nil
}
```

The `nilc` tool generates a bridge manifest:

```text
Go type -> Nil type
string   -> string
int64    -> i64
[]byte   -> bytes
error    -> Result<T, Error>
context.Context -> CancellationContext
```

Go is especially suitable for:
- CLI
- package tooling
- build orchestration
- network services
- portable plugins
- selected native/system adapters

Go's mobile ecosystem already has Android/iOS binding/build tooling, although the official `golang/mobile` repository labels it experimental; Alap should therefore define its own stable ABI instead of making gomobile's generated API the framework ABI. citeturn649263search0turn649263search1

---

# 30. TypeScript compatibility

NilLang should support a deliberate TS compatibility mode:

```text
nilc --from-ts src/*.ts
```

Supported:
- interfaces
- type aliases
- enums
- generics
- async/await
- classes
- standard collections
- common ECMAScript syntax

Unsupported or discouraged:
- eval
- dynamic object shape mutation
- monkey patching
- implicit any
- prototype mutation
- unbounded reflection

This aligns with ArkTS's emphasis on static analysis and fixed object layout. citeturn759431search3

---

# 31. ArkTS compatibility

Do not claim binary compatibility with OpenHarmony.

Instead create:

```text
ArkTS Source
     ↓
arkts-import
     ↓
NilLang AST
     ↓
NIR
```

Provide compatibility packages:

```text
@nil/arkui
@nil/ohos
@nil/storage
@nil/network
@nil/media
```

ArkUI-X demonstrates the practical model: one main ArkTS codebase plus platform-specific adaptation. It currently supports OpenHarmony/HarmonyOS, Android and iOS. citeturn288101search0turn288101search5

---

# 32. Compiler architecture

Recommended implementation:

```text
nilc
├── lexer
├── parser
├── syntax
├── typecheck
├── resolver
├── diagnostics
├── hir
├── mir
├── nir
├── ui-lowering
├── borrow/ownership checks for selected values
├── optimizer
├── codegen
│   ├── nabc
│   ├── native
│   └── wasm (optional)
└── package
```

The first compiler can be written in Go.

Reason:
- fast developer iteration
- excellent CLI/library ecosystem
- easy cross-compilation
- easy JSON/YAML/TOML tooling
- strong concurrency
- simple distribution as one executable

The runtime does not need to be written in the same language as the compiler.

---

# 33. Runtime architecture

Recommended first implementation:

```text
nilrt
├── scheduler
├── eventloop
├── heap
├── gc
├── object
├── string
├── array
├── map
├── async
├── actor
├── channel
├── timers
├── io
├── ffi
├── reflection
├── security
└── platform
```

There are two realistic runtime strategies.

## Strategy A — bytecode first

```text
NilLang -> NABC -> nilrt VM
```

Pros:
- easiest bootstrap
- deterministic
- debuggable
- same runtime on all platforms

## Strategy B — native/AOT

```text
NilLang -> NIR -> LLVM/native
```

Pros:
- highest performance

Recommended order:
1. bytecode
2. profiler
3. AOT
4. optional JIT

ArkCompiler itself uses bytecode plus interpreter/compiler/runtime subsystems and AOT capabilities, so this phased model is proven. citeturn759431search0

---

# 34. NABC bytecode

Example conceptual format:

```text
NABC
  header
  version
  target
  metadata
  type table
  string table
  constant pool
  function table
  bytecode
  debug info
  signatures
```

Instructions:

```text
LOAD_CONST
LOAD_LOCAL
STORE_LOCAL
LOAD_FIELD
STORE_FIELD

ADD_I64
SUB_I64
MUL_I64
DIV_I64

CALL
CALL_NATIVE
RETURN

JUMP
JUMP_IF_FALSE

NEW_OBJECT
NEW_ARRAY

AWAIT
SPAWN
SEND
RECEIVE

UI_CREATE
UI_CHILD
UI_PROP
UI_EVENT
UI_END

THROW
TRY
CATCH
```

---

# 35. UI compiler lowering

Input:

```text
Column {
  Text("Hello")
  Button("Tap") {
    onClick => count += 1
  }
}
```

Lower to:

```text
UI_CREATE Column
UI_CREATE Text
UI_PROP text "Hello"
UI_END

UI_CREATE Button
UI_PROP text "Tap"
UI_EVENT click fn#42
UI_END

UI_END
```

Runtime produces a retained UI tree:

```text
App
└── Column
    ├── Text
    └── Button
```

Updates should be property-level, not whole-tree redraws.

---

# 36. Rendering pipeline

```text
NilLang
 ↓
UI tree
 ↓
state diff
 ↓
layout
 ↓
paint commands
 ↓
scene graph
 ↓
GPU command encoder
 ↓
Vulkan / Metal / GLES
```

Use the existing Onuron Vulkan UI renderer as the first backend foundation rather than creating a second graphics stack. The repository already identifies `nilui-gpu` as a Vulkan 2D renderer. citeturn759431view0

Linux:

```text
Wayland -> Vulkan
X11     -> Vulkan
DRM/KMS -> Vulkan/software fallback
```

Android:

```text
Activity
  ↓
Surface
  ↓
Alap Android adapter
  ↓
graphics backend
```

iOS:

```text
UIViewController
  ↓
CAMetalLayer
  ↓
Alap iOS adapter
  ↓
Metal renderer
```

Onuron:

```text
nilshell / Wayland
  ↓
surface
  ↓
nilui-gpu / Vulkan
```

---

# 37. Platform adapter contracts

Each platform implements:

```text
interface Platform {
  init()
  shutdown()

  window()
  display()
  input()
  clipboard()

  filesystem()
  secureStorage()

  network()
  camera()
  microphone()
  speaker()

  notifications()
  sensors()
  location()

  share()
  deepLinks()

  appLifecycle()
  accessibility()

  gpu()
}
```

The application only sees this interface.

---

# 38. Android architecture

```text
APK
├── AndroidManifest.xml
├── Kotlin/Java bootstrap
├── libalap_runtime.so
├── libalap_app.so
└── assets/
```

Flow:

```text
Kotlin Activity
      ↓ JNI/C ABI
Alap Runtime
      ↓
NilLang app
      ↓
Vulkan/OpenGL ES
```

Needed Android integration:
- Activity lifecycle
- saved state
- permissions
- intent/deep links
- clipboard
- notifications
- camera
- audio
- sensors
- file picker
- biometric APIs
- Android share sheet

Do not attempt to run inside an Android application by pretending Linux/Wayland exists.

---

# 39. iOS architecture

```text
Xcode project
├── Swift/ObjC bootstrap
├── Alap.xcframework
├── app bundle
└── resources
```

Flow:

```text
UIViewController
      ↓
Alap iOS Adapter
      ↓
NilRT
      ↓
NilLang
      ↓
Metal
```

Needed iOS integration:
- scene lifecycle
- URL schemes/universal links
- keychain
- notifications
- camera/microphone
- photo library
- location
- share sheet
- haptics
- background execution where permitted

The framework can generate an Xcode project/workspace rather than trying to replace Apple's build/signing machinery.

ArkUI-X's public project layout similarly keeps Android/iOS-specific native projects beside the shared application code. citeturn288101search5

---

# 40. Linux architecture

Targets:

```text
linux-x86_64
linux-aarch64
```

Window systems:

```text
Wayland (primary)
X11 (compatibility)
headless (CI)
```

Outputs:

```text
AppImage
Flatpak
.tar.zst
```

Linux is also the easiest desktop debugging target.

---

# 41. Onuron architecture

Onuron output:

```text
myapp.nilax
```

Example package:

```text
myapp.nilax/
├── manifest.json
├── app.nabc
├── ui.nui
├── assets/
├── permissions.sig
└── signature.ed25519
```

Installation:

```bash
nilpkg install myapp.nilax
```

Sandbox:

```text
namespace
seccomp
SELinux/AppArmor policy
capability broker
read-only system
private app data
```

Connect this to Onuron's existing `nilrt` security/runtime layer rather than bypassing it. The existing repository already describes namespace sandboxing, seccomp and a permission broker. citeturn759431view0

---

# 42. Build system

Primary command:

```bash
alap build
```

Targets:

```bash
alap build onuron
alap build android
alap build ios
alap build linux
```

Architectures:

```text
arm64
armeabi-v7a
x86_64
riscv64 (future)
```

Configuration:

```yaml
name: com.example.notes
version: 0.1.0

language:
  version: "0.1"

targets:
  - onuron-arm64
  - android-arm64
  - ios-arm64
  - linux-x86_64

permissions:
  - storage.read
  - storage.write

entry: src/main.nil
```

---

# 43. CLI

Required commands:

```text
nil create
alap init
alap build
alap run
alap test
alap fmt
alap check
nil lint
nil doctor
nil install
nil uninstall
nil package
nil publish

nil devices
nil logs
nil shell

nil generate
nil migrate
alap clean
nil upgrade
```

Examples:

```bash
nil create hello
alap run --target linux
alap run --target android
alap run --target onuron
alap check
alap test
nil package
```

---

# 44. Hot reload

Development mode:

```text
source change
  ↓
incremental compiler
  ↓
UI diff
  ↓
runtime patch
```

State should survive component replacement whenever possible.

---

# 45. Debugging

`nil debug` needs:

```text
breakpoints
step over
step into
stack traces
variable inspection
async task inspection
actor inspection
UI tree inspector
layout inspector
network inspector
memory profiler
frame profiler
```

---

# 46. Developer tools

Recommended:

```text
VS Code extension
IntelliJ plugin later
Neovim LSP
nilfmt
niltest
nilbench
niltrace
nilprof
```

Language server capabilities:

```text
completion
go-to-definition
rename
references
diagnostics
hover
code actions
semantic tokens
```

---

# 47. Testing

Unit:

```text
@test
function addTest() {
  assert(add(2, 3) == 5)
}
```

UI:

```text
@testUI
function buttonTest() {
  let app = mount(App)
  app.find("Button").tap()
  assert(app.text("count") == "1")
}
```

Golden tests:

```text
render -> screenshot -> compare
```

Platform matrix:

```text
Onuron x86_64
Onuron ARM64
Android emulator
Android ARM64 device
iOS simulator
iOS device
Linux Wayland
Linux X11
```

---

# 48. Performance targets

Initial targets:

```text
startup:
  < 300 ms on modern ARM64 target for a small app

UI:
  60 FPS minimum
  120 FPS target on supported displays

input-to-render:
  < 16 ms at 60 Hz
  < 8 ms at 120 Hz

release binary:
  aggressively strip unused modules

memory:
  avoid per-widget heavy allocations
```

These are engineering targets, not guarantees.

---

# 49. Package manager

`nilpm`

Registry:

```text
registry.onuron.dev
```

Package:

```yaml
name: nilhttp
version: 1.2.0
license: MIT

dependencies:
  nil/json: ^1.0
```

Security:
- Ed25519 signing
- immutable version
- lockfile
- checksum
- provenance
- reproducible package metadata

---

# 50. Interop with existing ecosystems

Web:

```text
NilLang -> JS/TS adapter
```

Android native:

```text
NilLang -> C ABI/JNI -> Kotlin/Java
```

iOS native:

```text
NilLang -> C ABI -> Swift/ObjC
```

Go:

```text
NilLang -> C ABI -> Go package
```

Rust:

```text
NilLang -> C ABI -> Rust
```

C/C++:

```text
NilLang -> stable C ABI
```

This is more important than direct source compatibility.

---

# 51. Stable ABI

ABI must not depend on Go memory layout, Rust layout or C++ ABI.

Every boundary uses opaque handles:

```c
typedef uint64_t NilHandle;
typedef int32_t  NilStatus;

typedef struct {
    const uint8_t *ptr;
    size_t len;
} NilBytes;
```

Never expose:
- Go pointers across long-lived boundaries
- Rust trait objects
- C++ STL containers
- language-specific exceptions

---

# 52. Memory model

Use GC for ordinary application objects.

Use explicit resource wrappers for:
- file handles
- sockets
- GPU resources
- native camera objects
- OS handles

Example:

```text
using camera = await Camera.open()

defer camera.close()
```

---

# 53. Resource lifetime

Every native resource follows:

```text
acquire
use
cancel/close
dispose
```

The compiler should warn about:
- un-awaited Future
- unused Result
- leaked native handles
- actor message misuse

---

# 54. Security model

Three permission classes:

```text
user capability
system capability
developer capability
```

Examples:

```text
camera
microphone
location
contacts
bluetooth
notifications
filesystem
network
sensors
telephony
```

Onuron can enforce these with its existing security architecture.

---

# 55. Distributed APIs

For Onuron's SoftBus layer:

```text
device.discover()
device.connect()
device.call()
```

Example:

```text
let devices = await nil.distributed.discover("camera")

let stream = await devices.first.openCamera()
```

The application sees a capability, not a socket.

This matches the existing Onuron direction around SoftBus/P2P IPC. citeturn759431view0

---

# 56. Web/PWA future target

Phase 2/3:

```text
NilLang
  ↓
Web IR
  ↓
WASM + JS
  ↓
Canvas/WebGPU/DOM adapters
```

This is optional for v1.

---

# 57. WASM optional architecture

Some pure application modules can compile to WASM:

```text
NilLang library
  ↓
NIR
  ↓
WASM
  ↓
embedded WASM runtime
```

A Go implementation may use a WASM runtime for sandboxed plugins. TinyGo can generate WASI WebAssembly and runtimes such as wazero can embed WASM in Go applications. citeturn288101search4

Do not make WASM mandatory for the UI path in v1; native UI rendering should remain the primary path.

---

# 58. Framework source repository

Recommended:

```text
alap/
├── compiler/
│   ├── lexer/
│   ├── parser/
│   ├── types/
│   ├── resolver/
│   ├── hir/
│   ├── mir/
│   ├── nir/
│   ├── codegen/
│   └── diagnostics/
├── runtime/
│   ├── nilrt/
│   ├── gc/
│   ├── scheduler/
│   ├── ffi/
│   └── std/
├── ui/
│   ├── nilui/
│   ├── layout/
│   ├── widgets/
│   ├── animation/
│   ├── a11y/
│   └── theme/
├── gfx/
│   ├── vulkan/
│   ├── metal/
│   ├── gles/
│   └── software/
├── abi/
│   ├── include/
│   ├── generated/
│   └── versioning/
├── platform/
│   ├── onuron/
│   ├── android/
│   ├── ios/
│   └── linux/
├── tools/
│   ├── nil/
│   ├── nilfmt/
│   └── nills/
├── packages/
├── examples/
├── tests/
└── docs/
```

---

# 59. Minimal compiler bootstrap

First milestone should support:

```text
let x: i32 = 10
let y: i32 = 20

function add(a: i32, b: i32): i32 {
  return a + b
}

print(add(x, y))
```

Compiler:

```text
NilLang -> NABC
```

Runtime:

```text
alap run hello.nil
```

Do not start with UI.

---

# 60. Minimal UI milestone

Then:

```text
app Hello {
  build() {
    Column {
      Text("Hello Nil")
      Button("Tap") {
        onClick => print("tap")
      }
    }
  }
}
```

Run:

```bash
alap run --target linux
```

Then port the same app to:

```bash
alap run --target android
alap run --target onuron
alap run --target ios
```

---

# 61. First example application

```nil
import { App, Text, Column, Button } from "nil/ui"

app CounterApp {
  state count: i32 = 0

  build() {
    Column {
      Text("Count: " + count.toString())
        .font(size: 32)

      Button("+1") {
        onClick => count += 1
      }

      Button("Async") {
        onClick => load()
      }
    }
    .padding(24)
    .spacing(12)
  }

  async function load() {
    let result = await api.get("/counter")
    log.info(result)
  }
}
```

This is the canonical demonstration of language + UI + async.

---

# 62. Generated application flow

```text
src/*.nil
   ↓
nilc
   ↓
NIR
   ├── app.nabc
   ├── ui.nui
   └── metadata.json
   ↓
platform packager
   ├── Onuron package
   ├── Android APK/AAB
   ├── iOS Xcode app
   └── Linux bundle
```

---

# 63. Android/iOS/Linux/Onuron support matrix

| Capability | Onuron | Android | iOS | Linux |
|---|---|---|---|---|
| Declarative UI | yes | yes | yes | yes |
| Vulkan | yes | yes* | no | yes |
| Metal | no | no | yes | no |
| Wayland | yes | no | no | yes |
| X11 | no | no | no | yes |
| Camera | yes | yes | yes | yes* |
| Telephony | yes | yes | limited | limited |
| Notifications | yes | yes | yes | yes |
| SoftBus | yes | optional | optional | optional |
| Sandboxing | yes | OS sandbox | OS sandbox | framework + OS |
| Native extensions | yes | yes | yes | yes |

*depends on device/driver.

---

# 64. What should be borrowed from ArkUI-X

Use as architectural inspiration:

```text
shared application code
+
platform-specific native adaptation
+
cross-platform SDK
+
CLI build tool
```

ArkUI-X publicly follows that pattern and supplies separate Android/iOS adaptation repositories, an app framework and a CLI. citeturn288101search1turn288101search2

Do NOT make Alap depend on ArkUI-X's internal application model.

---

# 65. What should NOT be copied

Do not make:
- proprietary service dependencies
- Huawei-only package assumptions
- OpenHarmony-only system APIs mandatory
- Android/iOS-specific UI behavior leak into `nil.ui`
- GPL-incompatible third-party code part of the framework without license review

Onuron currently declares GPLv3 for its repository. citeturn759431view0

For Alap itself, choose licensing carefully:
- compiler/runtime: Apache-2.0 or MIT
- optional copyleft adapters: separate modules
- Onuron-specific integration: can follow Onuron licensing constraints

A legal/license review is required before importing third-party code.

---

# 66. Roadmap

## Phase 0 — 0.x language core

Deliver:
- lexer
- parser
- static types
- functions
- structs/classes
- arrays/maps
- Result
- async syntax
- basic NABC
- command line compiler

## Phase 1 — runtime

Deliver:
- VM
- GC
- scheduler
- Future
- channels
- filesystem
- networking
- logging

## Phase 2 — UI

Deliver:
- component syntax
- state
- layout
- text
- input
- buttons
- scrolling
- animation
- accessibility tree

## Phase 3 — Linux

Deliver:
- Wayland
- Vulkan
- X11 compatibility
- desktop packaging
- hot reload

## Phase 4 — Onuron

Deliver:
- Wayland integration
- nilpkg
- sandbox
- capabilities
- SoftBus
- system services

## Phase 5 — Android

Deliver:
- APK/AAB
- Android lifecycle
- permissions
- camera/audio/sensors
- notification/share/deeplink

## Phase 6 — iOS

Deliver:
- Xcode project generation
- XCFramework
- Metal
- lifecycle
- permissions
- camera/share/deeplink

## Phase 7 — AOT

Deliver:
- NIR optimizer
- native code generation
- profile-guided optimization
- link-time dead stripping

---

# 67. First 12 repositories/modules

1. `alap-compiler`
2. `alap-runtime`
3. `alap-ui`
4. `alap-gfx`
5. `alap-abi`
6. `alap-platform-onuron`
7. `alap-platform-android`
8. `alap-platform-ios`
9. `alap-platform-linux`
10. `alap-cli`
11. `alap-lsp`
12. `alap-packages`

At the start, keep them in one monorepo to avoid synchronization overhead.

---

# 68. Exact v0.1 repository layout

```text
alap/
├── go.work
├── go.mod
├── LICENSE
├── README.md
├── spec/
│   ├── language.md
│   ├── grammar.md
│   ├── types.md
│   ├── concurrency.md
│   ├── ui.md
│   ├── abi.md
│   └── bytecode.md
├── compiler/
│   ├── cmd/nilc/
│   ├── lexer/
│   ├── parser/
│   ├── typecheck/
│   └── ir/
├── runtime/
│   └── nilrt/
├── ui/
│   └── nilui/
├── abi/
│   └── include/nil_abi.h
├── platform/
│   ├── linux/
│   ├── onuron/
│   ├── android/
│   └── ios/
├── cli/
│   └── nil/
├── lsp/
├── examples/
│   ├── hello/
│   └── counter/
└── tests/
```

---

# 69. Bootstrap Go module

`go.mod`:

```go
module github.com/joysriramsarkar/alap

go 1.26
```

Core packages:

```text
internal/lexer
internal/parser
internal/types
internal/ir
internal/diagnostic
cmd/nilc
cmd/nil
```

Keep the language compiler independent from the OS at this layer.

---

# 70. Minimal syntax grammar

EBNF-like draft:

```text
program        = { declaration } ;

declaration    = importDecl
               | typeDecl
               | enumDecl
               | interfaceDecl
               | classDecl
               | functionDecl
               | componentDecl
               | appDecl
               | storeDecl ;

importDecl     = "import" importSpec "from" string ";" ;

functionDecl   = ["public"] "function"
                 identifier "(" [parameters] ")"
                 [":" type]
                 block ;

parameters     = parameter { "," parameter } ;
parameter      = identifier ":" type ;

block          = "{" { statement } "}" ;

statement      = letDecl
               | assignment
               | expressionStmt
               | returnStmt
               | ifStmt
               | matchStmt
               | whileStmt
               | forStmt
               | taskStmt
               | tryStmt ;

letDecl        = ["const"] "let" identifier
                 [":" type] "=" expression ";" ;

type           = primitive
               | identifier
               | arrayType
               | unionType
               | genericType
               | optionalType ;

arrayType      = type "[]" ;
optionalType   = type "?" ;
unionType      = type "|" type { "|" type } ;

componentDecl  = "component" identifier
                 block ;

appDecl        = "app" identifier
                 block ;
```

This grammar is intentionally close enough to TypeScript/ArkTS for migration tooling, while leaving room for Nil-specific UI constructs.

---

# 71. Compiler warnings

Mandatory warnings:

```text
W001 unused import
W002 unused variable
W003 ignored Result
W004 un-awaited Future
W005 mutable capture across task
W006 non-Sendable value crossing actor boundary
W007 native handle not disposed
W008 permission requested but not declared
W009 blocking call in UI thread
W010 excessive UI allocation in build()
```

Errors:

```text
E001 type mismatch
E002 missing return
E003 invalid await
E004 illegal mutation
E005 unsupported dynamic property
E006 invalid platform API
```

---

# 72. UI performance rules enforced by compiler

Compiler should flag:

```text
blocking file I/O in build()
network call in build()
infinite rebuild loop
state mutation during layout
heavy allocation inside animation frame
```

Preferred model:

```text
build()
  = pure-ish UI description

side effects
  = event handlers / lifecycle / tasks
```

---

# 73. Threading model

UI thread:

```text
one thread
render safe
input safe
state commit safe
```

Workers:

```text
task pool
network
database
crypto
image decode
compression
```

GPU thread:

```text
render command submission
```

Native platform threads remain managed by platform adapters.

---

# 74. Database

Do not put SQL in the language.

Provide:

```text
nil.db
```

Adapters:

```text
sqlite
postgres
mysql
remote API
```

Mobile default:

```text
SQLite
```

Onuron default:

```text
SQLite / RocksDB optional
```

---

# 75. Networking

High-level:

```text
http.get()
http.post()
websocket()
tcp()
udp()
```

Security:

```text
TLS 1.3
certificate validation
secure defaults
```

---

# 76. Media

Common API:

```text
Camera.open()
Microphone.open()
AudioPlayer.play()
VideoPlayer.open()
Image.decode()
```

Implementation varies by platform.

---

# 77. Device APIs

```text
device.model
device.os
device.arch
device.locale
device.orientation
device.battery
device.screen
```

No application should need to know how the value was obtained.

---

# 78. System integration

Onuron-only extensions:

```text
nil.onuron.softbus
nil.onuron.package
nil.onuron.security
nil.onuron.services
```

These should be optional imports.

Example:

```text
try import nil.onuron.softbus
```

so Android/iOS/Linux builds can still compile with a fallback.

---

# 79. Build manifest example

```yaml
id: dev.onuron.demo
name: Nil Demo
version: 0.1.0

entry: src/main.nil

ui:
  renderer: auto
  minRefreshRate: 60
  targetRefreshRate: 120

targets:
  onuron:
    architectures: [arm64, x86_64]

  android:
    architectures: [arm64, x86_64]

  ios:
    architectures: [arm64]

  linux:
    architectures: [x86_64, arm64]

permissions:
  - network
  - notifications

features:
  softbus: false
```

---

# 80. Build pipeline by platform

## Onuron

```text
nilc
 ↓
nabc
 ↓
nilrt + nilui
 ↓
nilpkg package
 ↓
signed .nilax
```

## Android

```text
nilc
 ↓
nabc
 ↓
native Android bootstrap
 ↓
APK/AAB
```

## iOS

```text
nilc
 ↓
nabc
 ↓
XCFramework
 ↓
Xcode project
 ↓
.app / .ipa
```

## Linux

```text
nilc
 ↓
nabc
 ↓
native launcher
 ↓
AppImage / Flatpak
```

---

# 81. Why this can scale

The platform-specific pieces are intentionally small:

```text
                Shared
       ┌─────────────────────┐
       │ Language             │
       │ Runtime              │
       │ UI                   │
       │ State                │
       │ Networking            │
       │ Storage API           │
       │ App lifecycle model  │
       └──────────┬──────────┘
                  │
          thin platform layer
      ┌───────────┼───────────┐
      │           │           │
    Onuron      Android       iOS
      │           │           │
   Linux-ish   Android OS   Darwin
```

That is the key architectural property.

---

# 82. Critical technical constraints

1. Do not target "all phones" initially.
2. Start with one Android device family + x86_64 Linux + Onuron x86_64.
3. Add ARM64 Onuron next.
4. Add iOS only after the language/runtime ABI is stable.
5. Do not rewrite Onuron's Rust services merely to satisfy a Go language strategy.
6. Keep native API capability-oriented.
7. Keep the UI tree platform-neutral.
8. Make the compiler and package manager reproducible.
9. Keep ABI versioning explicit.
10. Treat binary compatibility as a release engineering problem, not a syntax problem.

---

# 83. First concrete implementation order

```text
Week 1:
  lexer
  parser
  AST
  diagnostics

Week 2:
  types
  functions
  variables
  modules

Week 3:
  NIR
  NABC writer
  minimal VM

Week 4:
  print
  strings
  arrays
  Result

Week 5:
  Future
  task
  channel

Week 6:
  UI AST
  Column
  Row
  Text
  Button

Week 7:
  state
  events
  reconciliation

Week 8:
  Linux Wayland
  Vulkan renderer

Week 9:
  Onuron adapter
  `.nilax` package

Week 10:
  Android bootstrap
  APK

Week 11:
  iOS bootstrap
  Xcode/XCFramework

Week 12:
  LSP
  formatter
  examples
```

This is a realistic bootstrap sequence for a prototype; production hardening will take substantially longer.

---

# 84. Final architecture decision

The final stack should be:

```text
                    NILX
┌───────────────────────────────────────────────────────────────┐
│                    Developer Experience                       │
│ nil CLI • LSP • Formatter • Debugger • Package Manager       │
├───────────────────────────────────────────────────────────────┤
│                         NilLang                               │
│ TypeScript/ArkTS syntax + strict types + Go-like concurrency  │
├───────────────────────────────────────────────────────────────┤
│                         Nil Compiler                          │
│ AST → HIR → MIR → NIR → NABC/AOT                             │
├───────────────────────────────────────────────────────────────┤
│                         Nil Runtime                           │
│ VM • GC • Scheduler • Actors • Channels • FFI • Security      │
├───────────────────────────────────────────────────────────────┤
│                           NilUI                               │
│ Components • State • Layout • Animation • A11y • Theme        │
├───────────────────────────────────────────────────────────────┤
│                           NilGFX                              │
│ Vulkan • Metal • GLES • Software                              │
├───────────────────────────────────────────────────────────────┤
│                        Platform ABI                           │
│ Onuron • Android • iOS • Linux                                 │
└───────────────────────────────────────────────────────────────┘
```

## The key strategic choice

**Do not build another ArkUI-X clone.**

Build a framework where:

```text
ArkTS/TypeScript developers
        ↓
can naturally write NilLang
        ↓
and get a stable app ABI
        ↓
that is native on Onuron
        ↓
while using thin adapters on Android/iOS/Linux.
```

That gives Onuron a real native application ecosystem instead of making the OS permanently dependent on a compatibility layer.

---

# 85. Definition of "done" for Alap 1.0

Alap 1.0 is complete only when this exact app can be built from one source tree:

```nil
import { App, Column, Text, Button } from "nil/ui"

app Hello {
  state count: i32 = 0

  build() {
    Column {
      Text("Hello from Alap")
      Text(count.toString())

      Button("Increment") {
        onClick => count += 1
      }
    }
  }
}
```

And:

```bash
alap build onuron
alap build android
alap build ios
alap build linux
```

produce functioning applications with the same UI and state logic, while platform-specific services are provided through adapters.

That is the architectural north star for the project.
