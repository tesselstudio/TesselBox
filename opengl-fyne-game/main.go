package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// Initialize Fyne application
	myApp := app.New()
	myApp.SetIcon(nil)
	myWindow := myApp.NewWindow("OpenGL 3D Game")
	myWindow.Resize(fyne.NewSize(1024, 768))

	// Create OpenGL widget
	openglWidget := NewOpenGLWidget()

	// Create main layout with OpenGL viewport and controls
	content := container.NewBorder(
		// Top
		container.NewHBox(
			widget.NewLabel("OpenGL 3D Game"),
			widget.NewButton("Reset", func() {
				log.Println("Reset camera")
			}),
		),
		// Bottom
		container.NewHBox(
			widget.NewButton("Start Game", func() {
				log.Println("Game started!")
			}),
			widget.NewButton("Exit", func() {
				myApp.Quit()
			}),
		),
		// Left
		container.NewVBox(
			widget.NewLabel("Controls"),
			widget.NewLabel("W/S: Forward/Back"),
			widget.NewLabel("A/D: Left/Right"),
			widget.NewLabel("Q/E: Up/Down"),
		),
		// Right
		nil,
		// Center
		openglWidget,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
