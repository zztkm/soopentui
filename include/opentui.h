#pragma once

#include <stdbool.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/* Opaque handle used by the OpenTUI C ABI (see Zig NativeHandle). */
typedef uint32_t OtuiHandle;

#define OTUI_INVALID_HANDLE ((OtuiHandle)0)

/* createRenderer: buffered_destination_kind — 0 = stdout, 1 = memory */
/* createRenderer: remote_mode — 0 = auto, 1 = local, 2 = remote */
OtuiHandle createRenderer(
    uint32_t width,
    uint32_t height,
    uint8_t buffered_destination_kind,
    uint8_t remote_mode,
    void *feed_ptr);
void destroyRenderer(OtuiHandle renderer);

void setupTerminal(OtuiHandle renderer, bool use_alternate_screen);
void restoreTerminalModes(OtuiHandle renderer);
void clearTerminal(OtuiHandle renderer);

OtuiHandle getNextBuffer(OtuiHandle renderer);
OtuiHandle getCurrentBuffer(OtuiHandle renderer);
uint32_t getBufferWidth(OtuiHandle buffer);
uint32_t getBufferHeight(OtuiHandle buffer);

/* Colors are OpenTUI packed RGBA: 4x uint16_t (low byte = channel). */
void bufferClear(OtuiHandle buffer, const uint16_t *bg);
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
