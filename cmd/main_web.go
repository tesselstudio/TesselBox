//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"syscall/js"
	"time"

	"github.com/tesselstudio/TesselBox/pkg/game"
	"github.com/tesselstudio/TesselBox/pkg/webgl"
)

var (
	controller *game.Controller
	renderer   *webgl.Renderer
)

func main() {
	fmt.Println("Starting TesselBox Web Version...")

	// Create game controller
	controller = game.NewController()

	// Create WebGL renderer
	renderer = webgl.NewRenderer()

	// Set up callbacks for JS
	c := make(chan struct{}, 0)

	// Register JS functions
	js.Global().Set("startWorld", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		worldName := args[0].String()
		controller.StartWorld(worldName, time.Now().UnixNano())
		return nil
	}))

	js.Global().Set("update", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		controller.Update()
		return nil
	}))

	js.Global().Set("render", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if renderer != nil {
			// Update meshes from chunk manager
			if world := controller.GetWorld(); world != nil {
				world.GetChunkManager().UpdateEngineWithMeshes(renderer)
			}

			// Update camera from player
			if player := controller.GetPlayer(); player != nil {
				pos := player.GetPosition()
				rot := player.GetRotation()
				renderer.UpdateCamera(pos, rot)
			}
			renderer.BeginFrame()
			renderer.Render()
			renderer.EndFrame()
		}
		return nil
	}))

	js.Global().Set("cleanup", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		controller.Stop()
		if renderer != nil {
			renderer.Cleanup()
		}
		return nil
	}))

	js.Global().Set("handleKeyInput", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		keyCode := args[0].Int()
		keyState := args[1].Int()
		controller.HandleKeyInput(keyCode, keyState)
		return nil
	}))

	js.Global().Set("handleMouseMove", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		dx := float32(args[0].Float())
		dy := float32(args[1].Float())
		controller.HandleMouseMove(dx, dy)
		return nil
	}))

	js.Global().Set("handleMouseInput", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		button := args[0].Int()
		state := args[1].Int()
		controller.HandleMouseInput(button, state)
		return nil
	}))

	js.Global().Set("initWebGL", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		canvas := args[0]
		if err := renderer.Initialize(canvas); err != nil {
			fmt.Printf("Failed to initialize WebGL: %v\n", err)
			return false
		}
		return true
	}))

	js.Global().Set("resizeCanvas", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		width := args[0].Int()
		height := args[1].Int()
		if renderer != nil {
			renderer.Resize(width, height)
		}
		return nil
	}))

	// Register the update loop callback
	js.Global().Call("registerUpdateLoop")
	fmt.Println("WebGL initialized")

	// Keep the program running
	<-c
}
