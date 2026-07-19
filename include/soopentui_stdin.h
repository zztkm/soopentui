#pragma once

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/* Enable non-canonical raw mode on STDIN_FILENO. Returns 0 on success. */
int soopentui_enable_raw_stdin(void);

/* Restore stdin termios saved by soopentui_enable_raw_stdin. */
void soopentui_restore_stdin(void);

/* poll(STDIN_FILENO) with timeout_ms. Returns 1 if POLLIN, 0 otherwise. */
int soopentui_poll_stdin(int timeout_ms);

/* Non-blocking read from STDIN_FILENO. Returns bytes read, 0 if EAGAIN, -1 on error. */
int soopentui_read_stdin_nonblock(void *buf, uint32_t len);

#ifdef __cplusplus
}
#endif
