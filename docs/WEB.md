# Alap Framework — Web (`web/*`)

Alap ships an HTTP stack: a radix-tree router, a middleware chain, server-side
rendering, Server-Sent Events, and a WebSocket manager.

```text
web/
├── server/     # net/http server wrapper
├── router/     # radix-tree router, Context, groups
├── middleware/ # Recovery, RequestID, SecurityHeaders, CORS, RateLimit
├── ssr/        # server-side rendering + state hydration
├── sse/        # Server-Sent Events stream
└── websocket/  # WebSocket client manager
```

---

## Server

```go
srv := server.New("UserService")
_ = srv.Listen(":4324")            // blocks
_ = srv.Shutdown(ctx)
```

`server.Server` implements `http.Handler` (`ServeHTTP`), so it can also be
mounted inside an existing Go server.

---

## Router

`router.New()` returns a radix-tree (trie) router that matches in `O(k)` time
over path segments.

```go
r := router.New()
r.Use(middleware.Recovery(), middleware.SecurityHeaders())

r.GET("/users/:id", func(ctx *router.Context) (interface{}, error) {
    id := ctx.Param("id")
    return map[string]string{"id": id}, nil
})

// Constrained parameters: {name:type}
r.GET("/items/{id:int}", getItem)
r.GET("/orders/{id:uuid}", getOrder)
```

| Kind | Example | Notes |
|------|---------|-------|
| Static | `/health` | Literal segment |
| Named param | `/users/:id` | `ctx.Param("id")` / `ctx.ParamInt("id")` |
| Constrained param | `/items/{id:int}` | `int` / `uuid` constraints |
| Methods | `GET POST PUT DELETE` | Also `AddRoute(method, pattern, …)` |

Route groups apply a prefix and shared middleware:

```go
admin := r.Group("/admin", requireRole("admin"))
admin.GET("/dashboard", dashboardHandler)
```

`r.Routes()` returns `[]RouteInfo` (useful for introspection and tests).

---

## Middleware

```go
type Middleware func(next router.HandlerFunc) router.HandlerFunc
```

| Middleware | Purpose |
|------------|---------|
| `Recovery()` | Convert panics into 500 responses. |
| `RequestID()` | Attach a unique request ID. |
| `SecurityHeaders()` | Set hardening headers. |
| `CORS(CORSConfig{...})` | Configurable CORS. |
| `RateLimit(requestsPerMin)` | Per-key rate limiting (`RateLimiter.Allow`). |

---

## Server-Side Rendering (`web/ssr`)

```go
renderer := ssr.New("My App")
html := renderer.RenderPage(content, map[string]interface{}{
    "user": "Sarkar",
})
```

`RenderPage` escapes text content and serialises hydration state into the page
(the `__NILANG_STATE__` payload) so the client can adopt it.

---

## Server-Sent Events (`web/sse`)

```go
stream, err := sse.NewStream(w)
if err == nil {
    _ = stream.Emit("tick", map[string]int{"n": 1})
}
```

---

## WebSocket (`web/websocket`)

```go
m := websocket.NewManager()
m.Register(client)
m.Broadcast(map[string]string{"type": "update"})
m.Unregister(client)
```

---

## Status

The web stack is functional and unit-tested at the routing/middleware level.
TLS termination, production load testing, and full WebSocket upgrade handling
are not yet hardened — see [`ROADMAP.md`](./ROADMAP.md).
