#pragma once

#include <stdbool.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/* Opaque handle used by the OpenTUI C ABI (see Zig NativeHandle). */
typedef uint32_t OtuiHandle;

#define OTUI_INVALID_HANDLE ((OtuiHandle)0)

/* Cursor style for setCursorStyleOptions.style (OpenTUI CursorStyle). */
#define OTUI_CURSOR_BLOCK 0
#define OTUI_CURSOR_UNDERLINE 1
#define OTUI_CURSOR_BAR 2

/* Matches OpenTUI CursorStyleOptions (Zig extern struct). */
typedef struct OtuiCursorStyleOptions {
    uint8_t style;
    uint8_t blinking; /* 0 or 1; other values keep current */
    const uint16_t *color; /* optional packed RGBA; NULL keeps current */
    uint8_t cursor; /* reserved / OpenTUI cursor field */
} OtuiCursorStyleOptions;

/* createRenderer: buffered_destination_kind — 0 = stdout, 1 = memory */
/* createRenderer: remote_mode — 0 = auto, 1 = local, 2 = remote */
OtuiHandle createRenderer(
    uint32_t width,
    uint32_t height,
    uint8_t buffered_destination_kind,
    uint8_t remote_mode,
    void *feed_ptr);
void destroyRenderer(OtuiHandle renderer);

void setClearOnShutdown(OtuiHandle renderer, bool clear);
void setBackgroundColor(OtuiHandle renderer, const uint16_t *color);

void setupTerminal(OtuiHandle renderer, bool use_alternate_screen);
void restoreTerminalModes(OtuiHandle renderer);
void clearTerminal(OtuiHandle renderer);
void setTerminalTitle(OtuiHandle renderer, const char *title, uint32_t title_len);

void resizeRenderer(OtuiHandle renderer, uint32_t width, uint32_t height);

void setCursorPosition(OtuiHandle renderer, int32_t x, int32_t y, bool visible);
void setCursorColor(OtuiHandle renderer, const uint16_t *color);
void setCursorStyleOptions(OtuiHandle renderer, const OtuiCursorStyleOptions *options);

void enableMouse(OtuiHandle renderer, bool enable_movement);
void disableMouse(OtuiHandle renderer);

OtuiHandle getNextBuffer(OtuiHandle renderer);
OtuiHandle getCurrentBuffer(OtuiHandle renderer);
uint32_t getBufferWidth(OtuiHandle buffer);
uint32_t getBufferHeight(OtuiHandle buffer);

/* Colors are OpenTUI packed RGBA: 4x uint16_t (low byte = channel). */
void bufferClear(OtuiHandle buffer, const uint16_t *bg);
void bufferFillRect(
    OtuiHandle buffer,
    uint32_t x,
    uint32_t y,
    uint32_t width,
    uint32_t height,
    const uint16_t *bg);
void bufferDrawText(
    OtuiHandle buffer,
    const char *text,
    uint32_t text_len,
    uint32_t x,
    uint32_t y,
    const uint16_t *fg,
    const uint16_t *bg,
    uint32_t attributes);

/* Returns a RenderStatus discriminant (0 = success in current OpenTUI). */
uint8_t render(OtuiHandle renderer, bool force);

#ifdef __cplusplus
}
#endif
