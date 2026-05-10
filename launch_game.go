package main

import (
	"fmt"
	"time"

	"github.com/tesselstudio/TesselBox/pkg/game"
	"github.com/tesselstudio/TesselBox/pkg/opengl"
)

func main() {
	println("🚀 Launching TesselBox with Proper Input Handling")

	// Create game controller
	controller := game.NewController()

	// Start world
	controller.StartWorld("TesselBox World", 12345)

	// Create OpenGL engine with proper resolution
	println("🔧 Creating OpenGL window...")
	engine, err := opengl.NewEngine(1024, 768, "TesselBox - Hexagonal Prism World")
	if err != nil {
		println("❌ Failed to create OpenGL engine:", err.Error())
		return
	}
	defer engine.Cleanup()

	// Create input handler and connect to engine
	window := engine.GetWindow()
	inputHandler := game.NewInputHandler(window)

	// Set up mouse movement callback
	inputHandler.SetMouseMoveCallback(func(x, y float64) {
		controller.HandleMouseMove(float32(x), float32(y))
	})

	// Lock mouse after a short delay to ensure window is properly established
	go func() {
		time.Sleep(100 * time.Millisecond)
		inputHandler.LockMouse()
		println("🎯 Mouse locked for first-person control")
	}()

	println("✅ OpenGL window and input system created!")
	println("🌍 World ready with WASD movement and mouse camera control!")
	println("📝 Controls: WASD=Move, Mouse=Camera, ESC=Pause/Quit, C=Switch Camera")

	// Main game loop with proper input processing
	println("🎮 Starting game loop...")
	frameCount := 0

	for !engine.ShouldClose() && controller.GetState() != game.GameStateMainMenu {
		frameStart := time.Now()

		// Process input
		inputHandler.ProcessInput(controller)
		inputHandler.ResetMouseDelta()

		// Update game logic
		controller.Update()

		// Update camera from player
		player := controller.GetPlayer()
		if player != nil {
			// Update engine camera using camera manager
			cameraManager := controller.GetCameraManager()
			if cameraManager != nil {
				// Get current camera position and update engine
				camPos := cameraManager.GetPosition()
				camTarget := camPos.Add(cameraManager.GetFront())
				engine.SetCameraPosition(camPos)
				engine.SetCameraTarget(camTarget)
			}
		}

		// Render frame
		engine.BeginFrame()
		engine.Render(controller)
		engine.EndFrame()
		engine.PollEvents()

		// Frame rate limiting
		elapsed := time.Since(frameStart)
		if elapsed < time.Second/60 {
			time.Sleep(time.Second/60 - elapsed)
		}

		// Print status every 60 frames (1 second)
		frameCount++
		if frameCount%60 == 0 {
			gameState := controller.GetState()
			cameraManager := controller.GetCameraManager()
			cameraMode := "Unknown"
			if cameraManager != nil {
				cameraMode = fmt.Sprintf("%v", cameraManager.GetMode())
			}
			println("🎨 Rendering - Frame:", frameCount/60, "State:", fmt.Sprintf("%v", gameState), "Camera:", cameraMode)
		}
	}

	println("🎮 Game ended")
}
