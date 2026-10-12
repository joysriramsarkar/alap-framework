# NilLang (Alap dialect) — Language Reference

This is the syntax reference for the NilLang **dialect implemented by this
prototype** (the Alap toolchain). The canonical language contract is owned by
[`nilLang`](../nilLang/); see [`../RELATIONSHIP.md`](../RELATIONSHIP.md).

> Items marked **prototype** are parsed and/or compiled but are not fully
> enforced at runtime yet.

---

## 1. Program structure

A program is a sequence of top-level declarations and statements. Execution
starts at the first top-level statement (or an explicit `main()` call).

```nil
// comments use //
/* block comments are supported */

function add(a: i32, b: i32): i32 {
    return a + b
}

let total: i32 = add(2, 3)
print(total)
```

---

## 2. Variables

| Keyword | Meaning |
|---------|---------|
| `let` | Immutable binding |
| `const` | Compile-time constant binding |
| `var` | Mutable binding |

```nil
let name: string = "Onuron"
const PI: f64 = 3.14159
var count: i32 = 0
count = count + 1
```

A mutable state field inside a UI component uses `@State` (see §11).

---

## 3. Types

Built-in scalar types: `i8`, `i16`, `i32`, `i64`, `f32`, `f64`, `bool`,
`string`, `void`.

Composite forms:

```nil
let nums: i32[]   = [1, 2, 3]     // array
let grid: i32[][] = [[1, 2]]      // nested array
let maybe: ?string                // nullable (prototype)
```

Type annotations are optional where the initialiser makes the type obvious.

---

## 4. Operators

| Category | Operators |
|----------|-----------|
| Arithmetic | `+ - * / %` |
| Comparison | `== != < > <= >=` |
| Logical | `&& \|\| !` |
| Assignment | `= += -= *= /=` |
| String | `+` (concatenation) |
| Ternary | `cond ? a : b` |
| Null-coalescing | `??` (prototype) |

```nil
let a: i32 = 10
let b: i32 = 3
print(a / b)          // 3 (integer division)
print(a % b)          // 1
print("sum=" + (a + b).toString())
```

---

## 5. Control flow

### `if` / `else`

```nil
if score > 90 {
    print("excellent")
} else if score > 75 {
    print("good")
} else {
    print("keep going")
}
```

### `while`

```nil
let i: i32 = 0
while i < 5 {
    print(i)
    i = i + 1
}
```

### `for` / `for … in`

```nil
for i in 0..10 {
    print(i)
}

let nums: i32[] = [1, 2, 3]
for n in nums {
    print(n)
}
```

### `loop` / `break` / `continue`

```nil
let n: i32 = 0
loop {
    n = n + 1
    if n == 3 { break }
}
```

### `match`

```nil
function grade(score: i32): string {
    match score {
        100 => { return "perfect" }
        90  => { return "excellent" }
    }
    return "ok"
}
```

---

## 6. Functions

```nil
function greet(name: string): void {
    print("Hello, " + name + "!")
}

function add(a: i32, b: i32): i32 {
    return a + b
}
```

**Prototype:** generic parameters (`function id<T>(x: T): T`) are parsed but
not fully enforced.

---

## 7. Arrays and maps

```nil
let nums: i32[] = [10, 20, 30]
print(nums[0])          // 10
print(nums.length)      // 3

let config: map = {
    "host": "localhost",
    "port": 4324
}
```

---

## 8. Structs, classes, enums, interfaces

```nil
struct User {
    id: i32,
    username: string,
    isAdmin: bool,
}

let dev: User = User {
    id: 1,
    username: "joysri",
    isAdmin: true,
}
print(dev.username)
```

```nil
enum Result<T, E> {
    Ok(T),
    Err(E),
}

class Counter {
    count: i32 = 0

    function increment(): void {
        this.count = this.count + 1
    }
}
```

`interface` and `type` (alias) declarations are parsed (**prototype**).

---

## 9. Actors (message-passing concurrency)

```nil
actor Counter {
    count: i32 = 0

    on Increment(n: i32) {
        count = count + n
    }
}
```

At the runtime level, `runtime/actor` provides isolated state plus `Send`
(fire-and-forget) and `Ask` (request/reply).

---

## 10. Tasks, channels, error handling (prototype)

```nil
task compute() {
    return 42
}

spawn worker()

let ch: Channel<i32> = Channel.new(8)
ch.send(1)
let v = ch.recv()

try {
    throw "something went wrong"
} catch (e) {
    print("caught: " + e)
}
```

`throw`/`try`/`catch` instructions exist in NABC; full semantics are still
being completed (see [`ROADMAP.md`](./ROADMAP.md)).

---

## 11. Declarative UI

UI is expressed as a component with a `build()` block containing a widget tree.
Modifiers attach with `.` and events use arrow handlers.

```nil
@Entry
@Component
component App {
    @State count: i32 = 0
    @State message: string = "Click the button!"

    build() {
        Column {
            Text("Count: " + count.toString())
                .fontSize(24)
                .color("#333333")

            Button("+") {
                onClick => {
                    count += 1
                    message = "Incremented to " + count.toString()
                }
            }
            .padding({ horizontal: 24, vertical: 12 })
            .backgroundColor("#176BFF")
            .color("#FFFFFF")
            .borderRadius(8)
        }
    }
}
```

Built-in widgets: `Window`, `Column`, `Row`, `Text`, `Button`, `TextInput`,
`Image`, `Card`, `Divider`, `Spacer`, `Switch` (see `ui/widgets`).

State changes drive re-rendering through `ui/state` signals and the
`ui/state` reconciler.

---

## 12. Modules

```nil
import { Table } from "alap/data"
export function publicApi(): void { /* ... */ }
```

`import`/`export` are parsed (**prototype**); module resolution is still being
fleshed out.

---

## 13. Standard library builtins

Functions available to running programs are registered by the VM and bridge to
`stdlib/*`:

| Group | Examples |
|-------|----------|
| Strings | `.length`, `.toString()`, concatenation |
| Math | `math_sqrt`, `math_abs`, `math_min`, `math_max` (and `stdlib/math`) |
| Crypto | `crypto_sha256`, `crypto_uuid` (and `stdlib/crypto`) |
| Time | `time_now` (and `stdlib/time`) |
| JSON | `stdlib/json` (`Stringify`, `Parse`, `Valid`) |
| IO | `stdlib/fs` (`ReadTextFile`, `WriteTextFile`, …) |
| Net | `stdlib/net` (`Get`, `Post`, `Fetch`) |

See [`DATA.md`](./DATA.md) and [`WEB.md`](./WEB.md) for the data and web
libraries.
