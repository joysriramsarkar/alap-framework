/* nilabi.h — Alap Stable C ABI
 * Version: 0.1.0
 *
 * This header defines the stable C ABI used by all platform adapters.
 * All Alap plugin calls cross this boundary.
 *
 * Copyright (c) 2026 Joy Sarkar / Onuron Project
 * License: AGPL-3.0 (see LICENSE)
 */
#pragma once
#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

/* ─── Version ──────────────────────────────────────────────────────────────── */
#define NILABI_VERSION_MAJOR 0
#define NILABI_VERSION_MINOR 1
#define NILABI_VERSION_PATCH 0

/* ─── Value tag ─────────────────────────────────────────────────────────────── */
typedef enum NilValueKind {
    NIL_VAL_NULL      = 0,
    NIL_VAL_BOOL      = 1,
    NIL_VAL_INT       = 2,
    NIL_VAL_FLOAT     = 3,
    NIL_VAL_STRING    = 4,
    NIL_VAL_BYTES     = 5,
    NIL_VAL_ARRAY     = 6,
    NIL_VAL_OBJECT    = 7,
    NIL_VAL_FUNC      = 8,
    NIL_VAL_CHANNEL   = 9,
    NIL_VAL_ERROR     = 10,
} NilValueKind;

/* ─── Opaque handles ────────────────────────────────────────────────────────── */
typedef void* NilHandle;    /* opaque object handle */
typedef void* NilContext;   /* runtime context */
typedef void* NilChannel;   /* channel<T> */
typedef void* NilFuture;    /* future<T> */

/* ─── Value struct ──────────────────────────────────────────────────────────── */
typedef struct NilValue {
    NilValueKind kind;
    union {
        bool        bool_val;
        int64_t     int_val;
        double      float_val;
        struct {
            const char* ptr;
            size_t      len;
        } str_val;
        struct {
            const uint8_t* ptr;
            size_t          len;
        } bytes_val;
        NilHandle   handle;
    };
} NilValue;

/* ─── Error ─────────────────────────────────────────────────────────────────── */
typedef struct NilError {
    int32_t     code;
    const char* message;
} NilError;

/* ─── Result ────────────────────────────────────────────────────────────────── */
typedef struct NilResult {
    bool        ok;
    NilValue    value;
    NilError    err;
} NilResult;

/* ─── Function pointer types ────────────────────────────────────────────────── */
typedef NilValue (*NilNativeFunc)(NilContext ctx, NilValue* args, int32_t argc);
typedef void     (*NilCallback)(NilValue event);
typedef void     (*NilUIEventHandler)(NilValue event, void* user_data);

/* ─── Runtime lifecycle ─────────────────────────────────────────────────────── */
NilContext nilx_runtime_create(void);
void       nilx_runtime_destroy(NilContext ctx);
NilResult  nilx_runtime_run(NilContext ctx, uint8_t* nabc, size_t nabc_len);
bool       nilx_runtime_dispatch_touch(NilContext ctx, float x, float y);
char*      nilx_runtime_get_ui_json(NilContext ctx);

/* ─── Value constructors ────────────────────────────────────────────────────── */
NilValue nilx_val_null(void);
NilValue nilx_val_bool(bool b);
NilValue nilx_val_int(int64_t n);
NilValue nilx_val_float(double f);
NilValue nilx_val_string(const char* s, size_t len);
NilValue nilx_val_bytes(const uint8_t* data, size_t len);

/* ─── Value accessors ───────────────────────────────────────────────────────── */
bool       nilx_val_is_null(NilValue v);
bool       nilx_val_to_bool(NilValue v);
int64_t    nilx_val_to_int(NilValue v);
double     nilx_val_to_float(NilValue v);
const char* nilx_val_to_cstring(NilValue v);

/* ─── Object / Array ────────────────────────────────────────────────────────── */
NilHandle  nilx_object_create(NilContext ctx);
void       nilx_object_set(NilHandle obj, const char* key, NilValue val);
NilValue   nilx_object_get(NilHandle obj, const char* key);

NilHandle  nilx_array_create(NilContext ctx, size_t capacity);
void       nilx_array_push(NilHandle arr, NilValue val);
NilValue   nilx_array_get(NilHandle arr, size_t index);
size_t     nilx_array_len(NilHandle arr);

/* ─── Channel ───────────────────────────────────────────────────────────────── */
NilChannel nilx_channel_create(NilContext ctx, size_t capacity);
void       nilx_channel_send(NilChannel ch, NilValue val);
NilValue   nilx_channel_recv(NilChannel ch);
bool       nilx_channel_try_recv(NilChannel ch, NilValue* out);

/* ─── Task / concurrency ────────────────────────────────────────────────────── */
void       nilx_task_spawn(NilContext ctx, NilNativeFunc fn, NilValue* args, int32_t argc);
NilFuture  nilx_future_create(NilContext ctx);
void       nilx_future_resolve(NilFuture f, NilValue val);
NilValue   nilx_future_await(NilFuture f);

/* ─── UI ────────────────────────────────────────────────────────────────────── */
NilHandle nilx_ui_create_widget(NilContext ctx, const char* widget_type);
void      nilx_ui_set_prop(NilHandle widget, const char* key, NilValue val);
void      nilx_ui_add_child(NilHandle parent, NilHandle child);
void      nilx_ui_set_event(NilHandle widget, const char* event, NilUIEventHandler handler, void* user_data);
void      nilx_ui_commit(NilContext ctx);  /* flush UI tree changes to renderer */

/* ─── Platform ──────────────────────────────────────────────────────────────── */
typedef struct NilPlatformInfo {
    const char* platform;    /* "onuron" | "android" | "ios" | "linux" */
    const char* arch;        /* "arm64" | "x86_64" | "riscv64" */
    const char* os_version;
    int32_t     screen_width;
    int32_t     screen_height;
    float       pixel_ratio;
    bool        has_camera;
    bool        has_gps;
    bool        has_biometric;
} NilPlatformInfo;

NilPlatformInfo nilx_platform_info(NilContext ctx);

/* ─── Permissions / capabilities ────────────────────────────────────────────── */
typedef enum NilCapability {
    NIL_CAP_STORAGE_READ    = 0x01,
    NIL_CAP_STORAGE_WRITE   = 0x02,
    NIL_CAP_CAMERA          = 0x04,
    NIL_CAP_MICROPHONE      = 0x08,
    NIL_CAP_LOCATION        = 0x10,
    NIL_CAP_NOTIFICATIONS   = 0x20,
    NIL_CAP_NETWORK         = 0x40,
    NIL_CAP_CONTACTS        = 0x80,
    NIL_CAP_BIOMETRIC       = 0x100,
    NIL_CAP_BLUETOOTH       = 0x200,
} NilCapability;

bool nilx_capability_check(NilContext ctx, NilCapability cap);
void nilx_capability_request(NilContext ctx, NilCapability cap, NilCallback result_cb);

/* ─── Onuron-specific (only available on Onuron target) ──────────────────────── */
#if defined(ALAP_TARGET_ONURON) || defined(NILX_TARGET_NILOS)
const char* onuron_kernel_version(void);
#define nilos_kernel_version onuron_kernel_version
int32_t     onuron_trigger_sensor(int32_t sensor_id);
#define nilos_trigger_sensor onuron_trigger_sensor
int32_t     onuron_bus_call(const char* service, const char* method,
                           const uint8_t* args, size_t args_len,
                           uint8_t* out, size_t* out_len);
#define nilos_bus_call onuron_bus_call
#endif

#ifdef __cplusplus
}  /* extern "C" */
#endif
