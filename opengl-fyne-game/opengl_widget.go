package main

import (
	"log"

	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// OpenGLWidget is a custom Fyne widget that simulates 3D rendering
type OpenGLWidget struct {
	widget.BaseWidget
	renderer *OpenGLRenderer
}

// NewOpenGLWidget creates a new OpenGL widget
func NewOpenGLWidget() *OpenGLWidget {
	widget := &OpenGLWidget{}
	widget.ExtendBaseWidget(widget)
	return widget
}

// CreateRenderer creates the renderer for this widget
func (o *OpenGLWidget) CreateRenderer() fyne.WidgetRenderer {
	if o.renderer == nil {
		o.renderer = &OpenGLRenderer{widget: o}
		o.renderer.setup()
	}
	return o.renderer
}

// OpenGLRenderer handles the rendering using Fyne's canvas system
type OpenGLRenderer struct {
	widget      *OpenGLWidget
	objects     []fyne.CanvasObject
	rect        *canvas.Rectangle
	lines       []*canvas.Line
	initialized bool
}

// setup initializes the renderer objects
func (r *OpenGLRenderer) setup() {
	if r.initialized {
		return
	}

	// Create a dark background for the 3D viewport
	r.rect = canvas.NewRectangle(color.RGBA{R: 26, G: 26, B: 51, A: 255})

	// Create lines to simulate a 3D cube wireframe
	r.lines = make([]*canvas.Line, 12)

	// Front face lines (square)
	r.lines[0] = canvas.NewLine(color.RGBA{R: 255, G: 128, B: 51, A: 255})
	r.lines[1] = canvas.NewLine(color.RGBA{R: 255, G: 128, B: 51, A: 255})
	r.lines[2] = canvas.NewLine(color.RGBA{R: 255, G: 128, B: 51, A: 255})
	r.lines[3] = canvas.NewLine(color.RGBA{R: 255, G: 128, B: 51, A: 255})

	// Back face lines (square)
	r.lines[4] = canvas.NewLine(color.RGBA{R: 255, G: 128, B: 51, A: 255})
	r.lines[5] = canvas.NewLine(color.RGBA{R: 255, G: 128, B: 51, A: 255})
	r.lines[6] = canvas.NewLine(color.RGBA{R: 255, G: 128, B: 51, A: 255})
	r.lines[7] = canvas.NewLine(color.RGBA{R: 255, G: 128, B: 51, A: 255})

	// Connecting lines between front and back faces
	r.lines[8] = canvas.NewLine(color.RGBA{R: 200, G: 100, B: 40, A: 255})
	r.lines[9] = canvas.NewLine(color.RGBA{R: 200, G: 100, B: 40, A: 255})
	r.lines[10] = canvas.NewLine(color.RGBA{R: 200, G: 100, B: 40, A: 255})
	r.lines[11] = canvas.NewLine(color.RGBA{R: 200, G: 100, B: 40, A: 255})

	// Combine all objects
	r.objects = []fyne.CanvasObject{r.rect}
	for _, line := range r.lines {
		r.objects = append(r.objects, line)
	}

	r.initialized = true
	log.Println("OpenGL widget initialized successfully")
}

// Layout handles layout changes
func (r *OpenGLRenderer) Layout(size fyne.Size) {
	if !r.initialized {
		return
	}

	// Layout the background rectangle
	r.rect.Resize(size)
	r.rect.Move(fyne.NewPos(0, 0))

	// Calculate cube position and size
	centerX := size.Width / 2
	centerY := size.Height / 2
	cubeSize := float32(100)

	// Front face (square)
	frontLeft := fyne.NewPos(centerX-cubeSize/2, centerY-cubeSize/2)
	frontRight := fyne.NewPos(centerX+cubeSize/2, centerY-cubeSize/2)
	frontBottomRight := fyne.NewPos(centerX+cubeSize/2, centerY+cubeSize/2)
	frontBottomLeft := fyne.NewPos(centerX-cubeSize/2, centerY+cubeSize/2)

	// Back face (smaller square, offset)
	backOffset := float32(30)
	backSize := cubeSize * 0.8
	backLeft := fyne.NewPos(centerX-backSize/2+backOffset, centerY-backSize/2-backOffset)
	backRight := fyne.NewPos(centerX+backSize/2+backOffset, centerY-backSize/2-backOffset)
	backBottomRight := fyne.NewPos(centerX+backSize/2+backOffset, centerY+backSize/2-backOffset)
	backBottomLeft := fyne.NewPos(centerX-backSize/2+backOffset, centerY+backSize/2-backOffset)

	// Front face lines
	r.lines[0].Position1 = frontLeft
	r.lines[0].Position2 = frontRight
	r.lines[1].Position1 = frontRight
	r.lines[1].Position2 = frontBottomRight
	r.lines[2].Position1 = frontBottomRight
	r.lines[2].Position2 = frontBottomLeft
	r.lines[3].Position1 = frontBottomLeft
	r.lines[3].Position2 = frontLeft

	// Back face lines
	r.lines[4].Position1 = backLeft
	r.lines[4].Position2 = backRight
	r.lines[5].Position1 = backRight
	r.lines[5].Position2 = backBottomRight
	r.lines[6].Position1 = backBottomRight
	r.lines[6].Position2 = backBottomLeft
	r.lines[7].Position1 = backBottomLeft
	r.lines[7].Position2 = backLeft

	// Connecting lines
	r.lines[8].Position1 = frontLeft
	r.lines[8].Position2 = backLeft
	r.lines[9].Position1 = frontRight
	r.lines[9].Position2 = backRight
	r.lines[10].Position1 = frontBottomRight
	r.lines[10].Position2 = backBottomRight
	r.lines[11].Position1 = frontBottomLeft
	r.lines[11].Position2 = backBottomLeft

	// Set line width for 3D effect
	for _, line := range r.lines {
		line.StrokeWidth = 2
	}
}

// MinSize returns the minimum size for this widget
func (r *OpenGLRenderer) MinSize() fyne.Size {
	return fyne.NewSize(400, 300)
}

// Refresh triggers a redraw
func (r *OpenGLRenderer) Refresh() {
	// Refresh all objects
	for _, obj := range r.objects {
		if refresher, ok := obj.(fyne.Widget); ok {
			refresher.Refresh()
		}
	}
}

// Destroy cleans up resources
func (r *OpenGLRenderer) Destroy() {
	// Clean up resources if needed
}

// ApplyTheme applies theme changes
func (r *OpenGLRenderer) ApplyTheme() {
	// Handle theme changes if needed
}

// Objects returns the objects that make up this widget
func (r *OpenGLRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}
