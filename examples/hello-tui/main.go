package main

import (
	"github.com/zztkm/soopentui"
	"solod.dev/so/time"
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

	soopentui.SetTerminalTitle(r, "soopentui hello-tui")

	var bg, panel, fg, muted, cursor soopentui.RGBA
	soopentui.SetRGB(&bg, 20, 24, 32)
	soopentui.SetRGB(&panel, 32, 40, 56)
	soopentui.SetRGB(&fg, 120, 220, 160)
	soopentui.SetRGB(&muted, 180, 180, 190)
	soopentui.SetRGB(&cursor, 120, 220, 160)

	soopentui.SetBackgroundColor(r, bg)
	soopentui.SetCursorStyle(r, soopentui.CursorBar, true)
	soopentui.SetCursorColor(r, cursor)
	soopentui.SetCursorPosition(r, 3, 6, true)

	buf := soopentui.NextBuffer(r)
	if buf == soopentui.InvalidHandle {
		panic("getNextBuffer failed")
	}

	soopentui.Clear(buf, bg)
	soopentui.FillRect(buf, 1, 1, soopentui.BufferWidth(buf)-2, 6, panel)
	soopentui.DrawText(buf, "Hello from Solod + OpenTUI", 3, 2, fg)
	soopentui.DrawTextAttr(buf, "FillRect + cursor + title", 3, 3, muted, nil, 0)
	soopentui.DrawText(buf, "Closing in 2 seconds...", 3, 5, muted)

	if soopentui.Render(r, true) != 0 {
		panic("render failed")
	}

	time.Sleep(2 * time.Second)
}
