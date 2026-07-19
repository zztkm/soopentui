package soopentui

//so:include "soopentui_stdin.h"
//so:include.c <unistd.h>

//so:extern
func soopentui_enable_raw_stdin() int32

//so:extern
func soopentui_restore_stdin()

//so:extern
func soopentui_poll_stdin(timeoutMs int32) int32

//so:extern
func soopentui_read_stdin_nonblock(buf *byte, len uint32) int32

//so:extern read
func cRead(fd int32, buf *byte, count uint) int

// EnableRawStdin puts stdin into non-canonical raw mode.
// Pair with RestoreStdin (typically via defer).
func EnableRawStdin() bool {
	return soopentui_enable_raw_stdin() == 0
}

// RestoreStdin restores termios saved by EnableRawStdin.
func RestoreStdin() {
	soopentui_restore_stdin()
}

// PollStdin reports whether stdin is readable within timeoutMs.
func PollStdin(timeoutMs int) bool {
	return soopentui_poll_stdin(int32(timeoutMs)) != 0
}

// ReadStdin reads up to len(buf) bytes from stdin via read(2).
// Prefer this over os.Stdin.Read after EnableRawStdin so FILE* buffering
// cannot mix with DrainCapabilities' direct reads.
func ReadStdin(buf []byte) int {
	if len(buf) == 0 {
		return 0
	}
	n := cRead(0, &buf[0], uint(len(buf)))
	if n < 0 {
		return 0
	}
	return n
}

// ReadStdinNonblock reads stdin without blocking. Returns 0 if no data.
func ReadStdinNonblock(buf []byte) int {
	if len(buf) == 0 {
		return 0
	}
	n := soopentui_read_stdin_nonblock(&buf[0], uint32(len(buf)))
	if n < 0 {
		return 0
	}
	return int(n)
}

// DrainCapabilities reads stdin for up to timeoutMs for the first chunk,
// then continues while more data arrives within 50ms, feeding each chunk
// to ProcessCapabilityResponse. Use right after SetupTerminal.
// Uses non-blocking reads so POLLIN races / hangups cannot stall startup.
func DrainCapabilities(renderer Handle, timeoutMs int) {
	if timeoutMs < 0 {
		timeoutMs = 0
	}
	if !PollStdin(timeoutMs) {
		return
	}
	var buf [4096]byte
	for i := 0; i < 256; i++ {
		n := ReadStdinNonblock(buf[:])
		if n > 0 {
			ProcessCapabilityResponse(renderer, string(buf[:n]))
			continue
		}
		if !PollStdin(50) {
			return
		}
		n = ReadStdinNonblock(buf[:])
		if n <= 0 {
			return
		}
		ProcessCapabilityResponse(renderer, string(buf[:n]))
	}
}
