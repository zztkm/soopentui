// Package soopentui is a thin Solod binding for the OpenTUI native C ABI.
package soopentui

//so:include "opentui.h"

// Handle is an OpenTUI native object handle.
type Handle = uint32

const InvalidHandle Handle = 0

// DestinationStdout writes rendered frames to the terminal.
const DestinationStdout uint8 = 0

// DestinationMemory keeps frames in memory (no terminal I/O).
const DestinationMemory uint8 = 1

// RemoteAuto lets OpenTUI detect local vs remote terminal behavior.
const RemoteAuto uint8 = 0

// RemoteLocal forces local terminal mode.
const RemoteLocal uint8 = 1

// RemoteRemote forces remote terminal mode.
const RemoteRemote uint8 = 2

// RGBA is OpenTUI's packed color (4x uint16, low byte = channel).
type RGBA [4]uint16

// SetRGB writes an opaque RGB color into out.
func SetRGB(out *RGBA, r, g, b uint8) {
	out[0] = uint16(r)
	out[1] = uint16(g)
	out[2] = uint16(b)
	out[3] = 255
}

//so:extern
func createRenderer(width, height uint32, bufferedDestinationKind, remoteMode uint8, feedPtr *byte) Handle

//so:extern
func destroyRenderer(renderer Handle)

//so:extern
func setupTerminal(renderer Handle, useAlternateScreen bool)

//so:extern
func restoreTerminalModes(renderer Handle)

//so:extern
func clearTerminal(renderer Handle)

//so:extern
func getNextBuffer(renderer Handle) Handle

//so:extern
func getCurrentBuffer(renderer Handle) Handle

//so:extern
func getBufferWidth(buffer Handle) uint32

//so:extern
func getBufferHeight(buffer Handle) uint32

//so:extern
func bufferClear(buffer Handle, bg *uint16)

//so:extern
func bufferDrawText(buffer Handle, text string, textLen uint32, x, y uint32, fg, bg *uint16, attributes uint32)

//so:extern
func render(renderer Handle, force bool) uint8

// CreateRenderer creates a renderer for the given terminal size.
func CreateRenderer(width, height uint32) Handle {
	return createRenderer(width, height, DestinationStdout, RemoteLocal, nil)
}

// Destroy releases a renderer.
func Destroy(renderer Handle) {
	destroyRenderer(renderer)
}

// SetupTerminal enters the OpenTUI terminal modes (optionally alternate screen).
func SetupTerminal(renderer Handle, alternateScreen bool) {
	setupTerminal(renderer, alternateScreen)
}

// RestoreTerminal restores the terminal after SetupTerminal.
func RestoreTerminal(renderer Handle) {
	restoreTerminalModes(renderer)
}

// NextBuffer returns the buffer to draw into for the next frame.
func NextBuffer(renderer Handle) Handle {
	return getNextBuffer(renderer)
}

// Clear fills the buffer with bg.
func Clear(buffer Handle, bg RGBA) {
	bufferClear(buffer, &bg[0])
}

// DrawText draws text at (x, y) with the given foreground color.
func DrawText(buffer Handle, text string, x, y uint32, fg RGBA) {
	bufferDrawText(buffer, text, uint32(len(text)), x, y, &fg[0], nil, 0)
}

// Render presents the next buffer to the terminal.
func Render(renderer Handle, force bool) uint8 {
	return render(renderer, force)
}
