package main

import (
	"github.com/zztkm/soopentui"
)

func main() {
	r := soopentui.CreateRenderer(80, 24)
	if r == soopentui.InvalidHandle {
		panic("createRenderer failed")
	}
	defer soopentui.Destroy(r)

	soopentui.SetClearOnShutdown(r, true)
	soopentui.SetupTerminal(r, true)
	defer soopentui.RestoreTerminal(r)

	if !soopentui.EnableRawStdin() {
		panic("enable raw stdin failed")
	}
	defer soopentui.RestoreStdin()

	soopentui.DrainCapabilities(r, 200)

	soopentui.SetTerminalTitle(r, "soopentui quit-tui")

	var bg, fg, muted soopentui.RGBA
	soopentui.SetRGB(&bg, 20, 24, 32)
	soopentui.SetRGB(&fg, 120, 220, 160)
	soopentui.SetRGB(&muted, 180, 180, 190)
	soopentui.SetBackgroundColor(r, bg)

	buf := make([]byte, 256)
	var pending [256]byte
	pendingLen := 0
	done := false
	for !done {
		frame := soopentui.NextBuffer(r)
		if frame == soopentui.InvalidHandle {
			panic("getNextBuffer failed")
		}
		soopentui.Clear(frame, bg)
		soopentui.DrawText(frame, "Press q to quit (Ctrl+C also works)", 2, 2, fg)
		soopentui.DrawText(frame, "stdin is raw; capability replies are drained", 2, 4, muted)
		if soopentui.Render(r, true) != 0 {
			panic("render failed")
		}

		n := soopentui.ReadStdin(buf)
		if n <= 0 {
			done = true
			continue
		}
		chunk := string(buf[:n])
		soopentui.ProcessCapabilityResponse(r, chunk)

		pendingLen = appendPending(&pending, pendingLen, chunk)
		if shouldQuit(string(pending[:pendingLen])) {
			done = true
		}
	}
}

func appendPending(pending *[256]byte, pendingLen int, chunk string) int {
	for i := 0; i < len(chunk); i++ {
		if pendingLen >= len(pending) {
			copy(pending[:64], pending[pendingLen-64:pendingLen])
			pendingLen = 64
		}
		pending[pendingLen] = chunk[i]
		pendingLen++
	}
	return pendingLen
}

func shouldQuit(data string) bool {
	for i := 0; i < len(data); i++ {
		c := data[i]
		if c == 'q' || c == 'Q' || c == 0x03 {
			return true
		}
	}
	// OpenTUI SetupTerminal enables modifyOtherKeys; Ctrl+C is then an
	// escape sequence rather than byte 0x03 / SIGINT.
	if contains(data, "\x1b[27;5;99~") {
		return true
	}
	if contains(data, "\x1b[99;5u") {
		return true
	}
	if contains(data, "\x1b[99;5:") {
		return true
	}
	return false
}

func contains(s, sub string) bool {
	n := len(sub)
	if n == 0 {
		return true
	}
	if n > len(s) {
		return false
	}
	for i := 0; i <= len(s)-n; i++ {
		ok := true
		for j := 0; j < n; j++ {
			if s[i+j] != sub[j] {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}
