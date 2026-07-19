// Package soopentui is a third-party Solod (So) binding for the OpenTUI native C ABI.
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

// Cursor style constants (OpenTUI CursorStyle).
const (
	CursorBlock     uint8 = 0
	CursorUnderline uint8 = 1
	CursorBar       uint8 = 2
)

// RGBA is OpenTUI's packed color (4x uint16, low byte = channel).
type RGBA [4]uint16

// CursorStyleOptions matches OtuiCursorStyleOptions in opentui.h.
//
//so:extern OtuiCursorStyleOptions
type CursorStyleOptions struct {
	style    uint8
	blinking uint8
	color    *uint16
	cursor   uint8
}

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
func setClearOnShutdown(renderer Handle, clear bool)

//so:extern
func setBackgroundColor(renderer Handle, color *uint16)

//so:extern
func setupTerminal(renderer Handle, useAlternateScreen bool)

//so:extern
func restoreTerminalModes(renderer Handle)

//so:extern
func clearTerminal(renderer Handle)

//so:extern
func setTerminalTitle(renderer Handle, title string, titleLen uint32)

//so:extern
func resizeRenderer(renderer Handle, width, height uint32)

//so:extern
func setCursorPosition(renderer Handle, x, y int32, visible bool)

//so:extern
func setCursorColor(renderer Handle, color *uint16)

//so:extern
func setCursorStyleOptions(renderer Handle, options *CursorStyleOptions)

//so:extern
func enableMouse(renderer Handle, enableMovement bool)

//so:extern
func disableMouse(renderer Handle)

//so:extern
func processCapabilityResponse(renderer Handle, response string, responseLen uint32)

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
func bufferFillRect(buffer Handle, x, y, width, height uint32, bg *uint16)

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

// SetClearOnShutdown controls whether the terminal is cleared on destroy.
func SetClearOnShutdown(renderer Handle, clear bool) {
	setClearOnShutdown(renderer, clear)
}

// SetBackgroundColor sets the renderer default background color.
func SetBackgroundColor(renderer Handle, color RGBA) {
	setBackgroundColor(renderer, &color[0])
}

// SetupTerminal enters the OpenTUI terminal modes (optionally alternate screen).
func SetupTerminal(renderer Handle, alternateScreen bool) {
	setupTerminal(renderer, alternateScreen)
}

// RestoreTerminal restores the terminal after SetupTerminal.
func RestoreTerminal(renderer Handle) {
	restoreTerminalModes(renderer)
}

// ClearTerminal clears the terminal via the renderer.
func ClearTerminal(renderer Handle) {
	clearTerminal(renderer)
}

// SetTerminalTitle sets the terminal window title.
func SetTerminalTitle(renderer Handle, title string) {
	setTerminalTitle(renderer, title, uint32(len(title)))
}

// Resize updates the renderer dimensions (e.g. after a terminal resize).
func Resize(renderer Handle, width, height uint32) {
	resizeRenderer(renderer, width, height)
}

// SetCursorPosition moves the cursor; coordinates are 1-based in OpenTUI.
func SetCursorPosition(renderer Handle, x, y int32, visible bool) {
	setCursorPosition(renderer, x, y, visible)
}

// SetCursorColor sets the cursor color.
func SetCursorColor(renderer Handle, color RGBA) {
	setCursorColor(renderer, &color[0])
}

// SetCursorStyle sets cursor shape and blink (color unchanged).
func SetCursorStyle(renderer Handle, style uint8, blinking bool) {
	var blink uint8
	if blinking {
		blink = 1
	}
	opts := CursorStyleOptions{
		style:    style,
		blinking: blink,
		color:    nil,
		cursor:   0,
	}
	setCursorStyleOptions(renderer, &opts)
}

// EnableMouse enables terminal mouse reporting (decode input in the app).
func EnableMouse(renderer Handle, enableMovement bool) {
	enableMouse(renderer, enableMovement)
}

// DisableMouse disables terminal mouse reporting.
func DisableMouse(renderer Handle) {
	disableMouse(renderer)
}

// ProcessCapabilityResponse feeds a stdin chunk to OpenTUI's capability parser.
// Call this for replies to SetupTerminal queries (and optionally other chunks).
func ProcessCapabilityResponse(renderer Handle, data string) {
	processCapabilityResponse(renderer, data, uint32(len(data)))
}

// NextBuffer returns the buffer to draw into for the next frame.
func NextBuffer(renderer Handle) Handle {
	return getNextBuffer(renderer)
}

// CurrentBuffer returns the currently displayed buffer handle.
func CurrentBuffer(renderer Handle) Handle {
	return getCurrentBuffer(renderer)
}

// BufferWidth returns the buffer width in cells.
func BufferWidth(buffer Handle) uint32 {
	return getBufferWidth(buffer)
}

// BufferHeight returns the buffer height in cells.
func BufferHeight(buffer Handle) uint32 {
	return getBufferHeight(buffer)
}

// Clear fills the buffer with bg.
func Clear(buffer Handle, bg RGBA) {
	bufferClear(buffer, &bg[0])
}

// FillRect fills a rectangle with bg.
func FillRect(buffer Handle, x, y, width, height uint32, bg RGBA) {
	bufferFillRect(buffer, x, y, width, height, &bg[0])
}

// DrawText draws text at (x, y) with the given foreground color.
func DrawText(buffer Handle, text string, x, y uint32, fg RGBA) {
	bufferDrawText(buffer, text, uint32(len(text)), x, y, &fg[0], nil, 0)
}

// DrawTextAttr draws text with optional background and attributes.
// Pass bg == nil to leave the cell background unchanged.
func DrawTextAttr(buffer Handle, text string, x, y uint32, fg RGBA, bg *RGBA, attributes uint32) {
	var bgPtr *uint16
	if bg != nil {
		bgPtr = &bg[0]
	}
	bufferDrawText(buffer, text, uint32(len(text)), x, y, &fg[0], bgPtr, attributes)
}

// Render presents the next buffer to the terminal.
func Render(renderer Handle, force bool) uint8 {
	return render(renderer, force)
}
