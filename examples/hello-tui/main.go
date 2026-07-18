package main

import (
	"github.com/zztkm/solod-vs-go/soopentui"
	"solod.dev/so/time"
)

func main() {
	r := soopentui.CreateRenderer(80, 24)
	if r == soopentui.InvalidHandle {
		panic("createRenderer failed")
	}
	defer soopentui.Destroy(r)

	soopentui.SetupTerminal(r, true)
	defer soopentui.RestoreTerminal(r)

	buf := soopentui.NextBuffer(r)
	if buf == soopentui.InvalidHandle {
		panic("getNextBuffer failed")
	}

	var bg, fg, muted soopentui.RGBA
	soopentui.SetRGB(&bg, 20, 24, 32)
	soopentui.SetRGB(&fg, 120, 220, 160)
	soopentui.SetRGB(&muted, 180, 180, 190)
	soopentui.Clear(buf, bg)
	soopentui.DrawText(buf, "Hello from Solod + OpenTUI", 2, 2, fg)
	soopentui.DrawText(buf, "Closing in 2 seconds...", 2, 4, muted)

	if soopentui.Render(r, true) != 0 {
		panic("render failed")
	}

	time.Sleep(2 * time.Second)
}
