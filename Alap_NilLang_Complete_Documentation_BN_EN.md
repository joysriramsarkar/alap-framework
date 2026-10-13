ALAP FRAMEWORK + NILLANG

সম্পূর্ণ প্রযুক্তিগত ডকুমেন্টেশন / Complete Technical Documentation

Version 0.1 Architecture Specification • 1 September 2026

Target: Onuron, Android, iOS, Linux • Primary language: NilLang

> **Naming note (13 Oct 2026):** the Alap CLI binaries are now `alap`/`alapc`/`alap-lsp`
> (`nil` belongs to the canonical NilLang toolchain). Examples below read `alap …`.


Executive note / নির্বাহী নোট

Alap is designed as a native-first cross-platform application platform for Onuron. It deliberately borrows the declarative UI model and developer ergonomics of ArkTS/ArkUI, the static type discipline and familiar syntax of TypeScript, and the concurrency/tooling philosophy associated with Go. It is not a source-level clone of ArkTS, nor a TypeScript transpiler with platform-specific hacks.

Alap-কে এমনভাবে নকশা করা হয়েছে যাতে Onuron (অনুরণ ওএস)-এর নিজস্ব UI/runtime/security stack মূল প্ল্যাটফর্ম হয়, আর Android, iOS এবং Linux-এ thin adaptation layer-এর মাধ্যমে একই application source ব্যবহার করা যায়।


## Document scope / ডকুমেন্টের পরিধি

This document defines the proposed language, runtime, UI framework, compiler, ABI, package format, platform adapters, security model, developer tooling, repository structure, compatibility strategy, testing strategy, roadmap, and implementation rules. It is an architecture/specification document, not a claim that all pieces already exist.

এই ডকুমেন্টটি implementation-ready blueprint হিসেবে লেখা; তবে এটি বর্তমান repository-তে সবকিছু ইতিমধ্যে implemented আছে—এমন দাবি করে না।


## Table of contents / সূচিপত্র

1. Vision and principles / লক্ষ্য ও নীতি

2. Onuron baseline / বর্তমান Onuron ভিত্তি

3. System architecture / সামগ্রিক স্থাপত্য

4. NilLang language design / ভাষা নকশা

5. Grammar / ব্যাকরণ

6. Type system / টাইপ সিস্টেম

7. Concurrency / concurrency model

8. Declarative UI / UI model

9. State, lifecycle and navigation / state

10. Standard library / standard API

11. Compiler / compiler architecture

12. IR and bytecode / IR ও bytecode

13. Runtime / runtime design

14. Graphics and rendering / rendering

15. Platform adaptation / platform bridge

16. Android / Android port

17. iOS / iOS port

18. Linux / Linux port

19. Onuron / Onuron port

20. Native extensions and FFI / FFI

21. Security and capabilities / নিরাপত্তা

22. Packaging and distribution / packaging

23. Build system and CLI / build tools

24. Tooling and IDE / tooling

25. Testing and CI / testing

26. Performance / performance

27. Compatibility and migration / migration

28. Repository blueprint / repository

29. Governance and versioning / governance

30. Implementation roadmap / roadmap

31. Reference application / reference app

32. Acceptance criteria / release gate

33. References / তথ্যসূত্র


## 1. লক্ষ্য ও নীতি / Vision and Principles

বাংলা

Alap-এর মূল লক্ষ্য হলো একবার application logic ও UI লিখে Onuron, Android, iOS এবং Linux-এ ship করা। Platform-specific code থাকবে, কিন্তু সেটি boundary-এর মধ্যে থাকবে। Framework-এর কেন্দ্রে থাকবে stable ABI এবং capability-based API।

English

Alap aims to let developers build once and ship the same application model to Onuron, Android, iOS, and Linux. Platform-specific implementation remains necessary, but it is isolated behind explicit boundaries. The most important long-term asset is the stable application ABI and capability model, not a particular renderer or VM implementation.


## 2. বর্তমান Onuron ভিত্তি / Current Onuron Baseline

বাংলা

বর্তমান public Onuron repository-তে kernel configuration, HAL, `nilrt` sandbox, `nilui`, Vulkan-based `nilui-gpu`, Wayland compositor (`nilshell`), SoftBus, signed package manager এবং Android compatibility layer ইতিমধ্যে architectural building blocks হিসেবে দেখা যাচ্ছে। GitHub README অনুযায়ী repository-টি x86_64/ARM64 image build এবং QEMU run script-ও দেয়।

English

The current public Onuron repository exposes architectural building blocks that are directly useful to Alap: kernel configuration, HAL, `nilrt` sandboxing, `nilui`, Vulkan-based `nilui-gpu`, a wlroots-based Wayland compositor (`nilshell`), SoftBus, a signed package manager, and an Android compatibility layer. Its README also documents x86_64/ARM64 image building and QEMU execution.

onuron/
├── kernel/
├── hal/
├── runtime/
│   ├── nilrt/
│   ├── nilui/
│   ├── nilui-gpu/
│   └── nilbus-client/
├── shell/
├── softbus/
├── pkg/nilpkg/
├── services/
├── android/
├── apps/
└── security/selinux/

Source: https://github.com/joysriramsarkar/onuron

উৎস / Source: Onuron GitHub repository, accessed 1 September 2026.


## 3. সামগ্রিক স্থাপত্য / System Architecture

বাংলা

Alap পাঁচটি প্রধান স্তরকে পৃথক করবে: language/compiler, runtime, UI, graphics, এবং platform adapter. Application source থেকে NIR এবং NUI তৈরি হবে; runtime bytecode বা AOT code চালাবে; UI engine scene tree-কে layout/paint pipeline-এ পাঠাবে; platform adapter OS services এবং surface lifecycle সরবরাহ করবে।

English

Alap separates language/compiler, runtime, UI, graphics, and platform adaptation. Application source lowers into a neutral intermediate representation and a UI representation; the runtime executes bytecode or AOT code; the UI engine turns state into a retained scene tree; platform adapters provide lifecycle, surfaces, input, security and system services.

NilLang source
   │
   ├── Typed AST ──> HIR ──> MIR ──> NIR ──> NABC / Native
   │
   └── UI DSL ──> NUI ──> State Diff ──> Layout ──> Paint ──> GPU
                                                  │
        ┌─────────────────────────────────────────┼───────────────┐
        │                                         │               │
      Onuron                                    Android            iOS
      Wayland                                 Surface/JNI      UIKit/Metal
        │                                         │               │
        └──────────────────────── Linux ──────────┘


## 4. NilLang ভাষার নকশা / NilLang Language Design

বাংলা

NilLang হবে TypeScript/ArkTS পরিবারের কাছাকাছি syntax, কিন্তু আলাদা semantics-সহ একটি statically typed language। Go থেকে syntax কপি না করে concurrency primitives, package simplicity, tooling philosophy এবং error-as-data ধারণা নেওয়া হবে।

English

NilLang should feel familiar to TypeScript and ArkTS developers, but its semantics are independently specified. It should borrow concurrency primitives, package simplicity, tooling philosophy, and explicit error handling ideas associated with Go rather than copying Go syntax wholesale.

let count: i32 = 0
const title: string = "Alap"

function add(a: i32, b: i32): i32 {
  return a + b
}

async function load(): Future<User> {
  return await api.get<User>("/me")
}


## 5. ব্যাকরণ / Grammar

বাংলা

প্রথম সংস্করণে grammar ছোট এবং predictable রাখা হবে। Dynamic metaprogramming কমিয়ে compiler diagnostics শক্তিশালী করা হবে।

English

The first grammar intentionally stays small and predictable. Metaprogramming is constrained so that the compiler can provide strong diagnostics and reliable optimization.

program       = { declaration } ;
declaration   = importDecl | typeDecl | enumDecl | interfaceDecl
              | classDecl | functionDecl | componentDecl | appDecl ;
functionDecl  = ["public"] "function" identifier "(" [parameters] ")"
                [":" type] block ;
parameters    = parameter { "," parameter } ;
parameter     = identifier ":" type ;
letDecl       = ["const"] "let" identifier [":" type] "=" expression ";" ;
type          = primitive | identifier | arrayType | unionType
              | genericType | optionalType ;
arrayType     = type "[]" ;
optionalType  = type "?" ;
unionType     = type "|" type { "|" type } ;
block         = "{" { statement } "}" ;


## 6. টাইপ সিস্টেম / Type System

বাংলা

Primitive types হবে bool, signed/unsigned integers, f32/f64, bigint, char, string, bytes এবং void। `number`-এর মতো implicit catch-all numeric type থাকবে না। Optional value `T?`, Result `Result<T,E>`, এবং immutable records প্রথম-class হবে।

English

Primitive types include bool, signed/unsigned integers, f32/f64, bigint, char, string, bytes and void. There is no implicit catch-all numeric `number` type. Optional values use `T?`; fallible operations use `Result<T,E>`; immutable records are first-class.

type User = {
  id: string
  name: string
  age: i32
}

type Status = "idle" | "loading" | "ready" | "error"

function parseId(text: string): Result<i64, ParseError> {
  ...
}


## 7. Concurrency model / Concurrency Model

বাংলা

Concurrency হবে structured। `task`, `Future`, `Channel`, `select` এবং actor model থাকবে। UI thread-এ blocking I/O নিষিদ্ধ/সতর্ক করা হবে। Shared mutable state সরাসরি worker thread-এ ছড়ানো হবে না।

English

Concurrency is structured. The language provides tasks, futures, channels, select, and actors. Blocking I/O on the UI thread is forbidden or strongly warned. Shared mutable state is not casually shared across worker tasks.

let events: Channel<Event> = Channel(32)

task {
  while true {
    let event = await events.receive()
    handle(event)
  }
}

select {
  e <- events => handle(e)
  _ <- timer(1000) => timeout()
}


## 8. Declarative UI model / Declarative UI Model

বাংলা

UI হবে declarative এবং retained. `build()` UI description দেয়; state পরিবর্তনে compiler/runtime minimal diff করে scene tree update করবে। Render engine পুরো screen প্রতিবার নতুন করে আঁকবে না, যদি না backend তাই প্রয়োজন করে।

English

UI is declarative and retained. `build()` describes UI structure; state changes produce a minimal scene-tree diff. The renderer is not required to rebuild the entire screen for every state mutation.

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


## 9. State, lifecycle এবং navigation / State, Lifecycle and Navigation

বাংলা

State-এর scopes: local component state, app store state এবং external resource state। Lifecycle hooks side effects-এর জন্য, `build()` side-effect-এর জন্য নয়। Navigation URL-like route-এর উপর কাজ করবে।

English

State exists at component, application-store, and external-resource scopes. Lifecycle hooks are for side effects; `build()` should remain close to a pure description. Navigation uses URL-like routes so deep links can map naturally across platforms.

app Notes {
  state query: string = ""

  onAppear() {
    loadNotes()
  }

  build() {
    Column {
      Input(query) { onChange => query = it }
      NoteList(filter: query)
    }
  }
}

router {
  route("/", HomePage)
  route("/settings", SettingsPage)
  route("/note/:id", NotePage)
}


## 10. Standard library এবং API / Standard Library and API

বাংলা

Standard library ছোট কিন্তু নির্ভরযোগ্য হবে। OS capability APIs `nil.*` namespace-এ থাকবে। Network, storage, camera, media, notifications, sensors, location, accessibility এবং distributed services common interfaces দিয়ে প্রকাশ করা হবে।

English

The standard library should remain small and reliable. OS-facing APIs live under `nil.*`. Networking, storage, camera, media, notifications, sensors, location, accessibility, and distributed services are exposed through common capability-oriented interfaces.

nil.core
nil.collections
nil.async
nil.time
nil.text
nil.fs
nil.path
nil.json
nil.log
nil.net
nil.http
nil.crypto
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
nil.accessibility
nil.distributed


## 11. Compiler architecture / Compiler Architecture

বাংলা

প্রথম compiler Go-তে লেখা হবে। কারণ CLI, parser, package graph, cross-compilation এবং developer tooling দ্রুত এগোনোর সুবিধা আছে। Runtime আলাদা implementation language-এ হতে পারে; Onuron-এর বিদ্যমান Rust components অকারণে rewrite করা যাবে না।

English

The first compiler should be implemented in Go. Go is a good fit for the CLI, parser, package graph, cross-compilation, and developer tooling. The runtime may use another implementation language; existing Rust components in Onuron should not be rewritten merely for language consistency.

Source
  ↓ lexer/parser
Typed AST
  ↓ resolver/type checker
HIR
  ↓ control-flow lowering
MIR
  ↓ platform-neutral lowering
NIR
  ├── NABC bytecode
  └── AOT/native code


## 12. IR এবং bytecode / IR and Bytecode

বাংলা

NIR হবে runtime-independent, typed এবং explicit-control-flow representation। Bytecode format versioned হবে; debug symbols, source maps এবং capability metadata binary-তে থাকবে।

English

NIR is runtime-independent, typed, and explicit about control flow. The bytecode format is versioned and carries debug symbols, source maps, capability metadata, and signature information.

NABC header
  magic
  version
  target-triple
  flags
  type table
  string table
  constant pool
  function table
  bytecode
  debug data
  capability manifest
  signature

LOAD_CONST
LOAD_LOCAL
STORE_LOCAL
LOAD_FIELD
STORE_FIELD
ADD_I64
CALL
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
UI_PROP
UI_EVENT
UI_END
THROW


## 13. Runtime design / Runtime Design

বাংলা

Nil Runtime (`nilrt`) bytecode VM দিয়ে শুরু হবে; পরে AOT support যোগ হবে। Runtime-এ scheduler, GC, actor/channel, timers, FFI, reflection metadata এবং security capability checks থাকবে।

English

Nil Runtime (`nilrt`) should begin with a bytecode VM and later gain AOT. It contains scheduling, garbage collection, actors/channels, timers, FFI, reflection metadata, lifecycle management, and capability checks.

nilrt/
├── vm/
├── gc/
├── scheduler/
├── async/
├── actor/
├── channel/
├── timers/
├── io/
├── ffi/
├── security/
└── platform/

Recommendation: use a generational moving collector for the application heap, but keep native resource lifetimes explicit with `defer`/`using`-style cleanup and opaque handles.


## 14. Graphics এবং rendering / Graphics and Rendering

বাংলা

Onuron/Linux/Android-এ Vulkan হবে primary renderer যেখানে supported; iOS-এ Metal; software renderer CI/headless environment-এর জন্য। Existing Onuron `nilui-gpu`-কে first backend candidate হিসেবে reuse করা হবে।

English

Vulkan is the primary renderer on Onuron/Linux/Android where supported; Metal is primary on iOS; a software renderer provides CI/headless coverage. The existing Onuron `nilui-gpu` should be the first backend candidate rather than creating a second Vulkan UI stack.

NilLang/UI
   ↓
Scene Tree
   ↓
Layout
   ↓
Paint/Clip
   ↓
Render Graph
   ↓
Backend
   ├── Vulkan
   ├── Metal
   ├── GLES fallback
   └── Software


## 15. Platform adaptation এবং bridge / Platform Adaptation and Bridge

বাংলা

Cross-platform architecture-এর মূল হলো shared core + thin platform adapter। ArkUI-X-এর public documentation-এও shared ArkTS code-এর সঙ্গে platform-specific adaptation ও bridge layer-এর কথা বলা হয়েছে। Alap একই general architectural principle নেবে, কিন্তু নিজের runtime/ABI ব্যবহার করবে।

English

The key to cross-platform architecture is shared core plus a thin platform adapter. ArkUI-X documentation explicitly describes shared application code with platform adaptation and a bridge layer. Alap follows that architectural principle but defines its own runtime and ABI.

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
  audio()
  notifications()
  sensors()
  location()
  share()
  deepLinks()
  lifecycle()
  accessibility()
  gpu()
}


## 16. Android port / Android Port

বাংলা

Android-এ Kotlin/Java bootstrap + JNI/C ABI + Nil Runtime architecture ব্যবহার করা হবে। Activity lifecycle, Surface, intents/deep links, permissions, notifications, camera, sensors, share sheet এবং storage APIs adapter-এর দায়িত্ব।

English

Android uses a Kotlin/Java bootstrap layer, a JNI/C ABI boundary, and the Nil Runtime. The adapter owns Activity/Process lifecycle, Surface acquisition, intents/deep links, permissions, notifications, camera, sensors, share sheet, and storage integration.

Android App
├── AndroidManifest.xml
├── Kotlin bootstrap
├── libalap_runtime.so
├── libalap_app.so
└── assets/

Activity
  ↓ JNI/C ABI
Nil Runtime
  ↓
NilLang application
  ↓
Vulkan/GLES


## 17. iOS port / iOS Port

বাংলা

iOS-এ Xcode project generation করা হবে। Alap runtime/engine XCFramework হিসেবে embed হবে। UIViewController/scene lifecycle থেকে Alap app lifecycle-এ bridge করা হবে এবং Metal layer-এর মাধ্যমে rendering হবে।

English

iOS uses an Xcode-generated project. The Alap runtime/engine is distributed as an XCFramework. UIKit scene lifecycle is translated into the Alap lifecycle, and a Metal-backed surface provides rendering.

Xcode App
  ↓
App/Scene delegate
  ↓
Alap iOS Adapter
  ↓
Nil Runtime
  ↓
NilUI
  ↓
CAMetalLayer / Metal


## 18. Linux port / Linux Port

বাংলা

Linux-এ Wayland primary, X11 compatibility এবং headless test backend থাকবে। Desktop window integration, clipboard, file picker, notifications এবং desktop portals adapter করবে।

English

Linux targets Wayland first, X11 for compatibility, and a headless backend for CI. Window integration, clipboard, file pickers, notifications, and desktop portal integration live in the Linux adapter.

alap run --target linux
alap build linux --arch x86_64
alap build linux --arch arm64



## 19. Onuron port / Onuron Port

বাংলা

Onuron হবে Alap-এর reference platform। `nilui`/`nilui-gpu`, `nilrt`, `nilbus-client`, SELinux policy, permission broker এবং `nilpkg`-এর সঙ্গে direct integration হবে। App package `.nilax` হিসেবে ship করবে।

English

Onuron is the reference platform. Alap integrates directly with `nilui`/`nilui-gpu`, `nilrt`, `nilbus-client`, SELinux policy, the permission broker, and `nilpkg`. Applications are distributed as signed `.nilax` packages.

myapp.nilax/
├── manifest.json
├── app.nabc
├── ui.nui
├── assets/
├── permissions.sig
└── signature.ed25519


## 20. Native extensions ও FFI / Native Extensions and FFI

বাংলা

Native ABI কখনও Go, Rust বা C++ object layout-এর উপর নির্ভর করবে না। Opaque handles, primitive ABI types এবং versioned functions ব্যবহার হবে। Go, Rust, C/C++, Kotlin/Java এবং Swift/ObjC adapter এই boundary ব্যবহার করবে।

English

The stable ABI must never depend on Go, Rust, or C++ object layouts. It uses opaque handles, primitive ABI types, explicit ownership, and versioned functions. Go, Rust, C/C++, Kotlin/Java, and Swift/ObjC integrations all terminate at this ABI.

typedef uint64_t NilHandle;
typedef int32_t  NilStatus;

typedef struct {
  const uint8_t* ptr;
  size_t len;
} NilBytes;

NilStatus nil_plugin_open(...);
NilStatus nil_plugin_call(NilHandle h, ...);
void nil_release(NilHandle h);


## 21. Security and capabilities / Security and Capabilities

বাংলা

App manifest-এ permissions declare করবে; runtime capability broker-এর মাধ্যমে অনুমতি দেবে। App direct device nodes বা privileged system IPC ব্যবহার করতে পারবে না। Capability token short-lived ও scoped হতে পারবে।

English

Applications declare permissions in the manifest and obtain runtime-scoped capabilities through a broker. Applications should not directly access privileged device nodes or unrestricted system IPC. Capability tokens can be short-lived and scoped.

permissions:
  - network
  - storage.read
  - storage.write
  - camera
  - notifications
  - location


## 22. Packaging এবং distribution / Packaging and Distribution

বাংলা

Package format হবে platform-neutral metadata সহ target-specific payload। Version, dependency lock, capabilities, signing key identity, ABI requirement এবং minimum runtime version manifest-এ থাকবে।

English

Packages carry platform-neutral metadata plus target-specific payloads. The manifest records version, dependencies, capabilities, signing identity, ABI requirements, and minimum runtime version.

name: dev.nil.demo
version: 0.1.0
runtime: ">=0.1 <0.2"
entry: src/main.nil
permissions:
  - network
targets:
  - onuron-arm64
  - android-arm64
  - ios-arm64
  - linux-x86_64


## 23. Build system এবং CLI / Build System and CLI

বাংলা

Developer experience-এর কেন্দ্র হবে একটি `nil` command। একই CLI project creation, dependency resolution, check, test, build, run, package, install এবং doctor command দেবে।

English

The developer experience centers on one `nil` command. It handles project creation, dependency resolution, checking, testing, building, running, packaging, installation, and environment diagnostics.

nil create notes
alap check
alap test
alap run --target linux
alap run --target onuron
alap build android
alap build ios
nil package
nil install app.nilax
nil devices
nil logs
nil doctor


## 24. Tooling এবং IDE / Tooling and IDE

বাংলা

Language Server Protocol implementation (`nills`) editor independence দেবে। Formatter (`nilfmt`), static checker, debugger, UI inspector, profiler এবং package manager একই toolchain-এর অংশ হবে।

English

An LSP server (`nills`) keeps the language editor-independent. Formatter, static checker, debugger, UI inspector, profiler, and package manager form one coherent toolchain.


## 25. Testing এবং CI / Testing and CI

বাংলা

Unit, integration, UI golden, runtime stress, ABI compatibility, packaging and platform matrix tests বাধ্যতামূলক। Headless software renderer CI-কে platform availability থেকে আলাদা রাখবে।

English

Unit, integration, UI golden, runtime stress, ABI compatibility, package signing, and platform matrix tests are mandatory. A headless software renderer keeps most UI tests independent of physical device availability.

test matrix
├── compiler
├── runtime
├── ui-golden
├── abi
├── security
├── linux-wayland
├── android-emulator
├── ios-simulator
└── onuron-qemu


## 26. Performance targets / Performance Targets

বাংলা

প্রাথমিক engineering targets: 60 FPS baseline, 120 FPS capability where supported, small app startup 300 ms-এর আশেপাশে, UI frame budget 16.6 ms at 60 Hz এবং 8.3 ms at 120 Hz। এগুলো target, guarantee নয়।

English

Initial engineering targets are a 60 FPS baseline, 120 FPS where supported, approximately 300 ms startup for a small app, and frame budgets of 16.6 ms at 60 Hz and 8.3 ms at 120 Hz. These are engineering targets, not guarantees.


## 27. Compatibility এবং migration / Compatibility and Migration

বাংলা

NilLang সরাসরি ArkTS source compatibility claim করবে না। বরং `arkts-import` source transformer, compatibility libraries এবং automated diagnostics থাকবে। TypeScript থেকে migration-এ interfaces, type aliases, generics, async/await, classes ইত্যাদি সহজে আনা যাবে।

English

NilLang should not claim binary or full source compatibility with ArkTS. Instead, provide `arkts-import`, compatibility packages, and migration diagnostics. TypeScript migration should cover interfaces, type aliases, generics, async/await, and classes where semantics are compatible.

TypeScript/ArkTS
      ↓
parser + migration rules
      ↓
NilLang AST
      ↓
compatibility diagnostics
      ↓
NilLang source


## 28. Repository blueprint / Repository Blueprint

বাংলা

Prototype পর্যায়ে monorepo রাখা সবচেয়ে ভালো। Compiler এবং tooling Go-তে; runtime ও platform engine-এ Rust/C++/Swift/Kotlin প্রয়োজন অনুসারে থাকবে; Onuron-specific existing crates untouched থাকবে যতটা সম্ভব।

English

A monorepo is preferable during the prototype phase. Compiler and tooling are implemented in Go; runtime and platform pieces can use Rust/C++/Swift/Kotlin where appropriate; existing Onuron crates remain intact wherever practical.

alap/
├── cmd/
│   ├── nil/
│   └── nilc/
├── compiler/
├── runtime/
├── ui/
├── gfx/
├── abi/
├── platform/
│   ├── onuron/
│   ├── android/
│   ├── ios/
│   └── linux/
├── lsp/
├── packages/
├── examples/
├── tests/
├── spec/
└── docs/


## 29. Governance ও versioning / Governance and Versioning

বাংলা

Language version এবং runtime ABI version আলাদা হবে। Semantic language changes major version bump করবে; ABI break হলে runtime ABI major bump হবে। RFC process ছাড়া syntax/semantic breaking change merge করা উচিত নয়।

English

Language versions and runtime ABI versions are independent. A semantic language break requires a major language release; an ABI break requires a major ABI release. Breaking changes should require an RFC rather than ad-hoc merging.

Language: NilLang 0.1 / 0.2 / 1.0
Runtime ABI: NILABI 1.x
Bytecode: NABC v1
Package format: NILPKG v1


## 30. Implementation roadmap / Implementation Roadmap

বাংলা

Roadmap-এর প্রথম লক্ষ্য হলো ভাষা + runtime; তারপর Linux UI; তারপর Onuron; তারপর Android; শেষে iOS। কারণ iOS build/signing ecosystem এবং device constraints prototype phase-এ early blocker না হওয়াই ভালো।

English

The roadmap should stabilize language and runtime first, then Linux UI, then Onuron, then Android, and finally iOS. This keeps iOS signing/toolchain constraints from blocking the core platform early.


## 31. Reference application / Reference Application

বাংলা

Reference app হবে Counter/Notes ধরনের ছোট কিন্তু stateful application যা UI, async network, local storage, navigation এবং native capability একসঙ্গে ব্যবহার করে। প্রথম public compatibility test এই app-কে চার target-এ চালানো।

English

The reference application should be a small but stateful app that combines UI, async networking, local storage, navigation, and one native capability. It becomes the first cross-platform compatibility test.

import { Column, Text, Button } from "nil/ui"

app Counter {
  state count: i32 = 0

  build() {
    Column {
      Text("Count: " + count.toString())
      Button("Increment") {
        onClick => count += 1
      }
    }
    .padding(24)
  }
}


## 32. Release acceptance criteria / Release Acceptance Criteria

বাংলা

Alap 1.0 বলা যাবে তখনই যখন একই source tree থেকে Onuron ARM64, Android ARM64, iOS ARM64 এবং Linux x86_64 build সফল হবে এবং reference app-এ UI/state semantics সামঞ্জস্যপূর্ণ থাকবে। Platform-specific differences documented এবং intentional হতে হবে।

English

Alap 1.0 is acceptable only when one source tree can build a reference app for Onuron ARM64, Android ARM64, iOS ARM64, and Linux x86_64 with consistent UI/state semantics. Platform differences must be documented and intentional.


## 33. তথ্যসূত্র / References

1. Onuron repository: https://github.com/joysriramsarkar/onuron

2. ArkUI-X organization: https://github.com/arkui-x

3. ArkUI-X overview: https://github.com/arkui-x/docs/blob/master/en/ArkUI-X-Overview.md

4. ArkUI cross-platform design: https://github.com/arkui-x/docs/blob/master/en/framework-dev/design/design-overview.md

5. ArkUI-X SDK: https://github.com/arkui-x/docs/blob/master/en/application-dev/tools/how-to-use-arkui-x-sdk.md

6. ArkUI-X platform bridge: https://github.com/arkui-x/docs/blob/master/zh-cn/application-dev/quick-start/platform-bridge-introduction.md

7. ArkUI-X iOS framework integration: https://github.com/arkui-x/docs/blob/master/zh-cn/application-dev/tutorial/how-to-use-library-on-ios.md


## Appendix A - Language conventions / ভাষাগত convention

Use 2 spaces for indentation. UTF-8 source. Files use LF line endings. Public API names use camelCase; types/components use PascalCase; constants use UPPER_SNAKE_CASE only for compile-time constants. One statement per line is preferred but semicolons remain legal for migration familiarity.

ইচ্ছাকৃতভাবে Go, TypeScript এবং ArkTS-এর জনপ্রিয় syntax-এর অংশগুলো কাছাকাছি রাখা হয়েছে, কিন্তু semantics NilLang specification দ্বারা স্বতন্ত্রভাবে নির্ধারিত হবে।


## Appendix B - Non-goals / ইচ্ছাকৃতভাবে যা করা হবে না

• Onuron kernel-এর বিকল্প runtime বানানো নয়।

• Android/iOS-এ পুরো Onuron emulation বানানো নয়।

• Full JavaScript dynamic semantics reproduce করা নয়।

• OpenHarmony binary compatibility promise করা নয়।

• Every native API-কে common abstraction-এ ঠেসে ঢোকানো নয়; rare APIs explicit platform extensions হবে।

• Version 0.x-এ AOT compiler দিয়ে শুরু করা নয়।


## Appendix C - Recommended first repository change / প্রথম repository change

The first implementation PR should add `framework/alap/` or an equivalent top-level subtree, wire a Go workspace for `nil` and `nilc`, define `spec/grammar.md`, add a minimal NABC VM, and ship a `hello` example that runs under a headless backend. Only after that should the existing NilUI GPU path be connected.