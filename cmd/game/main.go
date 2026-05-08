package main

import (
	"log"
	"math"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tesselstudio/TesselBox/pkg/opengl"
)

const (
	width  = 1024
	height = 768
)

func main() {
	// Create OpenGL engine
	engine, err := opengl.NewEngine(width, height, "Pure OpenGL 3D Game")
	if err != nil {
		log.Fatalf("Failed to create OpenGL engine: %v", err)
	}
	defer engine.Cleanup()

	// Set up input callbacks
	setupInput(engine.GetWindow())

	// Game loop
	lastTime := time.Now()
	angle := 0.0

	for !engine.ShouldClose() {
		// Calculate delta time
		currentTime := time.Now()
		deltaTime := currentTime.Sub(lastTime).Seconds()
		lastTime = currentTime

		// Update
		angle += deltaTime * 45.0 // Rotate 45 degrees per second

		// Update camera position for rotation effect
		radius := float32(5.0)
		x := float32(angle) * 3.14159 / 180.0
		cameraPos := mgl32.Vec3{
			float32(math.Cos(float64(x))) * radius,
			0.0,
			float32(math.Sin(float64(x))) * radius,
		}
		engine.SetCameraPosition(cameraPos)

		// Render
		engine.BeginFrame()
		engine.Render()
		engine.EndFrame()

		// Handle events
		engine.PollEvents()
	}
}

func setupInput(window *glfw.Window) {
	// Key callback
	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if key == glfw.KeyEscape && action == glfw.Press {
			w.SetShouldClose(true)
		}
	})

	// Window resize callback
	window.SetFramebufferSizeCallback(func(w *glfw.Window, fbWidth int, fbHeight int) {
		gl.Viewport(0, 0, int32(fbWidth), int32(fbHeight))
	})
}
