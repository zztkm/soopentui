#include "soopentui_stdin.h"

#include <errno.h>
#include <fcntl.h>
#include <poll.h>
#include <stdio.h>
#include <termios.h>
#include <unistd.h>

static struct termios soopentui_saved_termios;
static int soopentui_saved_valid = 0;

int soopentui_enable_raw_stdin(void) {
    struct termios raw;

    if (!isatty(STDIN_FILENO)) {
        return -1;
    }
    if (tcgetattr(STDIN_FILENO, &soopentui_saved_termios) != 0) {
        return -1;
    }
    soopentui_saved_valid = 1;

    raw = soopentui_saved_termios;
    raw.c_lflag &= (tcflag_t) ~(ICANON | ECHO | ECHOE | ECHOK | ECHONL | ISIG | IEXTEN);
    raw.c_iflag &= (tcflag_t) ~(IXON | IXOFF | ICRNL | INLCR | IGNCR);
    raw.c_oflag &= (tcflag_t) ~(OPOST);
    raw.c_cc[VMIN] = 1;
    raw.c_cc[VTIME] = 0;

    if (tcsetattr(STDIN_FILENO, TCSANOW, &raw) != 0) {
        soopentui_saved_valid = 0;
        return -1;
    }
    /* Avoid FILE* stdin line-buffering waiting for newline after raw mode. */
    (void)setvbuf(stdin, NULL, _IONBF, 0);
    return 0;
}

void soopentui_restore_stdin(void) {
    if (!soopentui_saved_valid) {
        return;
    }
    (void)tcsetattr(STDIN_FILENO, TCSANOW, &soopentui_saved_termios);
    soopentui_saved_valid = 0;
}

int soopentui_poll_stdin(int timeout_ms) {
    struct pollfd pfd;
    int rc;

    pfd.fd = STDIN_FILENO;
    pfd.events = POLLIN;
    pfd.revents = 0;

    do {
        rc = poll(&pfd, 1, timeout_ms);
    } while (rc < 0 && errno == EINTR);

    if (rc <= 0) {
        return 0;
    }
    /* POLLHUP alone must not count as readable: a following blocking read can hang. */
    return (pfd.revents & POLLIN) != 0 ? 1 : 0;
}

int soopentui_read_stdin_nonblock(void *buf, uint32_t len) {
    int flags;
    ssize_t n;

    if (buf == NULL || len == 0) {
        return 0;
    }
    flags = fcntl(STDIN_FILENO, F_GETFL, 0);
    if (flags < 0) {
        return -1;
    }
    if (fcntl(STDIN_FILENO, F_SETFL, flags | O_NONBLOCK) != 0) {
        return -1;
    }
    do {
        n = read(STDIN_FILENO, buf, (size_t)len);
    } while (n < 0 && errno == EINTR);
    (void)fcntl(STDIN_FILENO, F_SETFL, flags);
    if (n < 0) {
        if (errno == EAGAIN || errno == EWOULDBLOCK) {
            return 0;
        }
        return -1;
    }
    return (int)n;
}
