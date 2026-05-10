package game

import (
	"sync"

	"github.com/go-gl/glfw/v3.3/glfw"
)

// InputHandler manages keyboard and mouse input
type InputHandler struct {
	mu sync.RWMutex

	// Current input state
	keys         map[glfw.Key]bool
	mouseButtons map[glfw.MouseButton]bool
	mouseDX      float64
	mouseDY      float64
	lastMouseX   float64
	lastMouseY   float64

	// Window reference
	window *glfw.Window

	// Input callbacks
	onKeyCallback       func(key glfw.Key, action glfw.Action)
	onMouseCallback     func(button glfw.MouseButton, action glfw.Action)
	onMouseMoveCallback func(x, y float64)

	// Additional input state
	hotbarSlot       int
	mouseSensitivity float32
	mouseLocked      bool
	firstMouse       bool
}

// NewInputHandler creates a new input handler
func NewInputHandler(window *glfw.Window) *InputHandler {
	ih := &InputHandler{
		window:           window,
		keys:             make(map[glfw.Key]bool),
		mouseButtons:     make(map[glfw.MouseButton]bool),
		hotbarSlot:       0,
		mouseSensitivity: 0.1,
		mouseLocked:      false,
		firstMouse:       true,
	}

	// Set up GLFW callbacks
	window.SetKeyCallback(ih.handleKeyCallback)
	window.SetMouseButtonCallback(ih.handleMouseCallback)
	window.SetCursorPosCallback(ih.handleMouseMoveCallback)
	window.SetScrollCallback(ih.handleScrollCallback)

	return ih
}

// handleKeyCallback is called by GLFW on key events
func (ih *InputHandler) handleKeyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	ih.mu.Lock()
	defer ih.mu.Unlock()

	switch action {
	case glfw.Press:
		ih.keys[key] = true
	case glfw.Release:
		ih.keys[key] = false
	}

	if ih.onKeyCallback != nil {
		ih.onKeyCallback(key, action)
	}
}

// handleMouseCallback is called by GLFW on mouse button events
func (ih *InputHandler) handleMouseCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	ih.mu.Lock()
	defer ih.mu.Unlock()

	switch action {
	case glfw.Press:
		ih.mouseButtons[button] = true
	case glfw.Release:
		ih.mouseButtons[button] = false
	}

	if ih.onMouseCallback != nil {
		ih.onMouseCallback(button, action)
	}
}

// handleMouseMoveCallback is called by GLFW on mouse movement
func (ih *InputHandler) handleMouseMoveCallback(window *glfw.Window, xpos, ypos float64) {
	ih.mu.Lock()
	defer ih.mu.Unlock()

	if ih.firstMouse {
		ih.lastMouseX = xpos
		ih.lastMouseY = ypos
		ih.firstMouse = false
		return
	}

	ih.mouseDX = xpos - ih.lastMouseX
	ih.mouseDY = ypos - ih.lastMouseY
	ih.lastMouseX = xpos
	ih.lastMouseY = ypos

	if ih.onMouseMoveCallback != nil {
		ih.onMouseMoveCallback(xpos, ypos)
	}
}

// handleScrollCallback is called by GLFW on mouse scroll
func (ih *InputHandler) handleScrollCallback(window *glfw.Window, xoffset, yoffset float64) {
	ih.mu.Lock()
	defer ih.mu.Unlock()

	// Handle hotbar scrolling
	if yoffset > 0 {
		// Scroll up - previous slot
		ih.hotbarSlot = (ih.hotbarSlot - 1 + 9) % 9
	} else if yoffset < 0 {
		// Scroll down - next slot
		ih.hotbarSlot = (ih.hotbarSlot + 1) % 9
	}
}

// IsKeyPressed returns whether a key is currently pressed
func (ih *InputHandler) IsKeyPressed(key glfw.Key) bool {
	ih.mu.RLock()
	defer ih.mu.RUnlock()
	return ih.keys[key]
}

// IsMouseButtonPressed returns whether a mouse button is currently pressed
func (ih *InputHandler) IsMouseButtonPressed(button glfw.MouseButton) bool {
	ih.mu.RLock()
	defer ih.mu.RUnlock()
	return ih.mouseButtons[button]
}

// GetMouseDelta returns the mouse movement since last frame
func (ih *InputHandler) GetMouseDelta() (float64, float64) {
	ih.mu.RLock()
	defer ih.mu.RUnlock()
	return ih.mouseDX, ih.mouseDY
}

// ResetMouseDelta resets the mouse delta to zero
func (ih *InputHandler) ResetMouseDelta() {
	ih.mu.Lock()
	defer ih.mu.Unlock()
	ih.mouseDX = 0
	ih.mouseDY = 0
}

// SetKeyCallback sets a callback for key events
func (ih *InputHandler) SetKeyCallback(callback func(key glfw.Key, action glfw.Action)) {
	ih.mu.Lock()
	defer ih.mu.Unlock()
	ih.onKeyCallback = callback
}

// SetMouseCallback sets a callback for mouse button events
func (ih *InputHandler) SetMouseCallback(callback func(button glfw.MouseButton, action glfw.Action)) {
	ih.mu.Lock()
	defer ih.mu.Unlock()
	ih.onMouseCallback = callback
}

// SetMouseMoveCallback sets a callback for mouse movement
func (ih *InputHandler) SetMouseMoveCallback(callback func(x, y float64)) {
	ih.mu.Lock()
	defer ih.mu.Unlock()
	ih.onMouseMoveCallback = callback
}

// LockMouse locks the cursor to the window
func (ih *InputHandler) LockMouse() {
	ih.mu.Lock()
	defer ih.mu.Unlock()

	if ih.window != nil {
		ih.window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
		ih.mouseLocked = true
		ih.firstMouse = true // Reset first mouse to avoid jumps
	}
}

// UnlockMouse unlocks the cursor
func (ih *InputHandler) UnlockMouse() {
	ih.mu.Lock()
	defer ih.mu.Unlock()

	if ih.window != nil {
		ih.window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
		ih.mouseLocked = false
	}
}

// HandleKeyCallback is the public key callback handler
func (ih *InputHandler) HandleKeyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	ih.handleKeyCallback(window, key, scancode, action, mods)
}

// HandleMouseCallback is the public mouse callback handler
func (ih *InputHandler) HandleMouseCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	ih.handleMouseCallback(window, button, action, mods)
}

// HandleMouseMoveCallback is the public mouse move callback handler
func (ih *InputHandler) HandleMouseMoveCallback(window *glfw.Window, xpos, ypos float64) {
	ih.handleMouseMoveCallback(window, xpos, ypos)
}

// ProcessInput applies current input to the controller
func (ih *InputHandler) ProcessInput(controller *Controller) {
	if controller == nil {
		return
	}

	ih.mu.Lock()
	defer ih.mu.Unlock()

	// Update input state in controller
	input := controller.input
	if input == nil {
		return
	}

	// Movement
	input.Forward = ih.keys[glfw.KeyW]
	input.Backward = ih.keys[glfw.KeyS]
	input.Left = ih.keys[glfw.KeyA]
	input.Right = ih.keys[glfw.KeyD]

	// Actions
	input.Jump = ih.keys[glfw.KeySpace]
	input.Sprint = ih.keys[glfw.KeyLeftShift]
	input.Sneak = ih.keys[glfw.KeyLeftControl]
	input.Inventory = ih.keys[glfw.KeyI]
	input.Pause = ih.keys[glfw.KeyEscape]
	input.Debug = ih.keys[glfw.KeyF3]
	input.CameraSwitch = ih.keys[glfw.KeyRightControl] || ih.keys[glfw.KeyC]
	input.MenuReturn = ih.keys[glfw.KeyQ]

	// Combat
	input.Attack = ih.mouseButtons[glfw.MouseButtonLeft]
	input.Use = ih.mouseButtons[glfw.MouseButtonRight]

	// Hotbar
	input.HotbarSlot = ih.hotbarSlot

	// Mouse
	input.MouseDX = float32(ih.mouseDX * float64(ih.mouseSensitivity))
	input.MouseDY = float32(ih.mouseDY * float64(ih.mouseSensitivity))
}
