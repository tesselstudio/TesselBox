package opengl_ui

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// UIComponent represents a base UI element
type UIComponent interface {
	Render(renderer *UIRenderer)
	SetPosition(x, y float32)
	SetSize(width, height float32)
	GetBounds() Rect
	HandleMouse(x, y float32, button glfw.MouseButton, action glfw.Action) bool
	HandleKey(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) bool
}

// Rect represents a rectangular area
type Rect struct {
	X, Y, Width, Height float32
}

// Contains checks if a point is inside the rectangle
func (r Rect) Contains(x, y float32) bool {
	return x >= r.X && x <= r.X+r.Width && y >= r.Y && y <= r.Y+r.Height
}

// BaseComponent provides common functionality for UI components
type BaseComponent struct {
	bounds     Rect
	visible    bool
	enabled    bool
	background mgl32.Vec4
}

// NewBaseComponent creates a new base component
func NewBaseComponent(x, y, width, height float32) *BaseComponent {
	return &BaseComponent{
		bounds:     Rect{X: x, Y: y, Width: width, Height: height},
		visible:    true,
		enabled:    true,
		background: mgl32.Vec4{0.2, 0.2, 0.2, 1.0}, // Default gray
	}
}

// SetPosition sets the component position
func (b *BaseComponent) SetPosition(x, y float32) {
	b.bounds.X = x
	b.bounds.Y = y
}

// SetSize sets the component size
func (b *BaseComponent) SetSize(width, height float32) {
	b.bounds.Width = width
	b.bounds.Height = height
}

// GetBounds returns the component bounds
func (b *BaseComponent) GetBounds() Rect {
	return b.bounds
}

// SetVisible sets component visibility
func (b *BaseComponent) SetVisible(visible bool) {
	b.visible = visible
}

// SetEnabled sets component enabled state
func (b *BaseComponent) SetEnabled(enabled bool) {
	b.enabled = enabled
}

// SetBackground sets the background color
func (b *BaseComponent) SetBackground(color mgl32.Vec4) {
	b.background = color
}

// HandleMouse default implementation
func (b *BaseComponent) HandleMouse(x, y float32, button glfw.MouseButton, action glfw.Action) bool {
	return false
}

// HandleKey default implementation
func (b *BaseComponent) HandleKey(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) bool {
	return false
}

// Button represents a clickable button
type Button struct {
	*BaseComponent
	text       string
	textColor  mgl32.Vec4
	hoverColor mgl32.Vec4
	clickColor mgl32.Vec4
	isHovered  bool
	onClick    func()
}

// NewButton creates a new button
func NewButton(x, y, width, height float32, text string) *Button {
	return &Button{
		BaseComponent: NewBaseComponent(x, y, width, height),
		text:          text,
		textColor:     mgl32.Vec4{1, 1, 1, 1}, // White text
		hoverColor:    mgl32.Vec4{0.3, 0.3, 0.3, 1.0},
		clickColor:    mgl32.Vec4{0.4, 0.4, 0.4, 1.0},
	}
}

// Render renders the button
func (b *Button) Render(renderer *UIRenderer) {
	if !b.visible {
		return
	}

	// Determine background color based on state
	color := b.background
	if b.isHovered {
		color = b.hoverColor
	}

	// Render button background
	renderer.RenderQuad(b.bounds.X, b.bounds.Y, b.bounds.Width, b.bounds.Height, color)

	// Render text centered in button
	textScale := float32(1.0)
	textWidth := float32(len(b.text)) * 8.0 * textScale * 0.8 // Approximate text width
	textHeight := 16.0 * textScale
	textX := b.bounds.X + (b.bounds.Width-textWidth)/2
	textY := b.bounds.Y + (b.bounds.Height-textHeight)/2 + textHeight

	renderer.RenderText(b.text, textX, textY, textScale, b.textColor)
}

// HandleMouse handles mouse events for the button
func (b *Button) HandleMouse(x, y float32, button glfw.MouseButton, action glfw.Action) bool {
	if !b.enabled || !b.visible {
		return false
	}

	wasHovered := b.isHovered
	b.isHovered = b.bounds.Contains(x, y)

	// Handle click
	if b.isHovered && button == glfw.MouseButtonLeft && action == glfw.Press {
		if b.onClick != nil {
			b.onClick()
		}
		return true
	}

	// Return true if hover state changed
	return b.isHovered != wasHovered
}

// SetOnClick sets the click handler
func (b *Button) SetOnClick(onClick func()) {
	b.onClick = onClick
}

// SetText sets the button text
func (b *Button) SetText(text string) {
	b.text = text
}

// GetText returns the button text
func (b *Button) GetText() string {
	return b.text
}

// Label represents a text label
type Label struct {
	*BaseComponent
	text      string
	textColor mgl32.Vec4
	fontSize  float32
}

// NewLabel creates a new label
func NewLabel(x, y, width, height float32, text string) *Label {
	return &Label{
		BaseComponent: NewBaseComponent(x, y, width, height),
		text:          text,
		textColor:     mgl32.Vec4{1, 1, 1, 1}, // White text
		fontSize:      16.0,
	}
}

// Render renders the label
func (l *Label) Render(renderer *UIRenderer) {
	if !l.visible {
		return
	}

	// Render background if not transparent
	if l.background[3] > 0 {
		renderer.RenderQuad(l.bounds.X, l.bounds.Y, l.bounds.Width, l.bounds.Height, l.background)
	}

	// Render text
	textScale := l.fontSize / 16.0 // Scale based on font size
	textX := l.bounds.X
	textY := l.bounds.Y + l.fontSize

	renderer.RenderText(l.text, textX, textY, textScale, l.textColor)
}

// SetText sets the label text
func (l *Label) SetText(text string) {
	l.text = text
}

// GetText returns the label text
func (l *Label) GetText() string {
	return l.text
}

// SetTextColor sets the text color
func (l *Label) SetTextColor(color mgl32.Vec4) {
	l.textColor = color
}

// SetFontSize sets the font size
func (l *Label) SetFontSize(size float32) {
	l.fontSize = size
}

// Container represents a container that holds other UI components
type Container struct {
	*BaseComponent
	components []UIComponent
	layout     Layout
}

// NewContainer creates a new container
func NewContainer(x, y, width, height float32) *Container {
	container := &Container{
		BaseComponent: NewBaseComponent(x, y, width, height),
		components:    make([]UIComponent, 0),
	}
	container.SetBackground(mgl32.Vec4{0, 0, 0, 0}) // Transparent by default
	return container
}

// AddComponent adds a component to the container
func (c *Container) AddComponent(component UIComponent) {
	c.components = append(c.components, component)
}

// RemoveComponent removes a component from the container
func (c *Container) RemoveComponent(component UIComponent) {
	for i, comp := range c.components {
		if comp == component {
			c.components = append(c.components[:i], c.components[i+1:]...)
			break
		}
	}
}

// SetLayout sets the layout manager for the container
func (c *Container) SetLayout(layout Layout) {
	c.layout = layout
}

// Render renders the container and all its components
func (c *Container) Render(renderer *UIRenderer) {
	if !c.visible {
		return
	}

	// Render container background
	if c.background[3] > 0 {
		renderer.RenderQuad(c.bounds.X, c.bounds.Y, c.bounds.Width, c.bounds.Height, c.background)
	}

	// Apply layout if set
	if c.layout != nil {
		c.layout.Apply(c.components, c.bounds)
	}

	// Render all components
	for _, component := range c.components {
		component.Render(renderer)
	}
}

// HandleMouse handles mouse events for the container
func (c *Container) HandleMouse(x, y float32, button glfw.MouseButton, action glfw.Action) bool {
	if !c.enabled || !c.visible {
		return false
	}

	// Check components in reverse order (top to bottom)
	for i := len(c.components) - 1; i >= 0; i-- {
		if c.components[i].HandleMouse(x, y, button, action) {
			return true
		}
	}

	return false
}

// HandleKey handles keyboard events for the container
func (c *Container) HandleKey(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) bool {
	if !c.enabled || !c.visible {
		return false
	}

	// Check components in reverse order (top to bottom)
	for i := len(c.components) - 1; i >= 0; i-- {
		if c.components[i].HandleKey(key, action, mods) {
			return true
		}
	}

	return false
}

// ImageButton represents a button with an image background
type ImageButton struct {
	*BaseComponent
	textureID    uint32
	hoverTexture uint32
	imageBounds  image.Rectangle
	isHovered    bool
	onClick      func()
	hasHover     bool
}

// NewImageButton creates a new image button
func NewImageButton(x, y, width, height float32, imagePath string) *ImageButton {
	button := &ImageButton{
		BaseComponent: NewBaseComponent(x, y, width, height),
		hasHover:      false,
	}

	// Load the image texture
	fmt.Printf("Loading image texture from: %s\n", imagePath)
	textureID, bounds, err := loadImageTexture(imagePath)
	if err != nil {
		fmt.Printf("Failed to load image texture: %v\n", err)
		// Fallback to a colored button
		button.textureID = 0
		button.imageBounds = image.Rectangle{}
	} else {
		button.textureID = textureID
		button.imageBounds = bounds
		fmt.Printf("Successfully loaded image texture with ID: %d, bounds: %v\n", textureID, bounds)
	}

	return button
}

// NewImageButtonWithHover creates a new image button with hover state
func NewImageButtonWithHover(x, y, width, height float32, imagePath, hoverImagePath string) *ImageButton {
	button := &ImageButton{
		BaseComponent: NewBaseComponent(x, y, width, height),
		hasHover:      true,
	}

	// Load the normal image texture
	textureID, bounds, err := loadImageTexture(imagePath)
	if err != nil {
		fmt.Printf("Failed to load image texture: %v\n", err)
		button.textureID = 0
		button.imageBounds = image.Rectangle{}
	} else {
		button.textureID = textureID
		button.imageBounds = bounds
	}

	// Load the hover image texture
	hoverTextureID, _, err := loadImageTexture(hoverImagePath)
	if err != nil {
		fmt.Printf("Failed to load hover image texture: %v\n", err)
		button.hoverTexture = button.textureID // Fallback to normal texture
	} else {
		button.hoverTexture = hoverTextureID
	}

	return button
}

// loadImageTexture loads an image file and creates an OpenGL texture
func loadImageTexture(imagePath string) (uint32, image.Rectangle, error) {
	// Open the image file
	file, err := os.Open(imagePath)
	if err != nil {
		return 0, image.Rectangle{}, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	// Decode the PNG image
	img, err := png.Decode(file)
	if err != nil {
		return 0, image.Rectangle{}, fmt.Errorf("failed to decode PNG: %w", err)
	}

	// Convert image to RGBA format
	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0, image.Rectangle{}, fmt.Errorf("unsupported stride")
	}

	// Draw the image to the RGBA buffer
	draw(rgba, img)

	// Create OpenGL texture
	var textureID uint32
	gl.GenTextures(1, &textureID)
	gl.BindTexture(gl.TEXTURE_2D, textureID)

	// Set texture parameters
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	// Upload texture data
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(rgba.Rect.Size().X), int32(rgba.Rect.Size().Y), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))

	// Unbind texture
	gl.BindTexture(gl.TEXTURE_2D, 0)

	return textureID, img.Bounds(), nil
}

// draw converts any image to RGBA
func draw(dst *image.RGBA, src image.Image) {
	b := dst.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Set(x, y, src.At(x, y))
		}
	}
}

// Render renders the image button
func (b *ImageButton) Render(renderer *UIRenderer) {
	if !b.visible || b.textureID == 0 {
		return
	}

	// Determine which texture to use
	textureID := b.textureID
	if b.isHovered && b.hasHover {
		textureID = b.hoverTexture
	}

	// Calculate aspect ratio preserving dimensions
	imageWidth := float32(b.imageBounds.Dx())
	imageHeight := float32(b.imageBounds.Dy())

	// Calculate scale to fit within button bounds while preserving aspect ratio
	buttonWidth := b.bounds.Width
	buttonHeight := b.bounds.Height

	scaleX := buttonWidth / imageWidth
	scaleY := buttonHeight / imageHeight
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// Calculate final dimensions (centered)
	finalWidth := imageWidth * scale
	finalHeight := imageHeight * scale
	finalX := b.bounds.X + (buttonWidth-finalWidth)/2
	finalY := b.bounds.Y + (buttonHeight-finalHeight)/2

	// Render the button with texture at correct aspect ratio
	renderer.RenderTexturedQuad(finalX, finalY, finalWidth, finalHeight, textureID)
}

// HandleMouse handles mouse events for the image button
func (b *ImageButton) HandleMouse(x, y float32, button glfw.MouseButton, action glfw.Action) bool {
	if !b.enabled || !b.visible {
		return false
	}

	wasHovered := b.isHovered
	b.isHovered = b.bounds.Contains(x, y)

	// Debug output
	if button == glfw.MouseButtonLeft && action == glfw.Press {
		fmt.Printf("ImageButton click at (%.1f, %.1f), bounds: (%.1f, %.1f)-(%.1f, %.1f), contains: %v\n",
			x, y, b.bounds.X, b.bounds.Y, b.bounds.X+b.bounds.Width, b.bounds.Y+b.bounds.Height, b.isHovered)
		fmt.Printf("Button size: %.1f x %.1f\n", b.bounds.Width, b.bounds.Height)
	}

	// Handle click
	if b.isHovered && button == glfw.MouseButtonLeft && action == glfw.Press {
		if b.onClick != nil {
			fmt.Printf("ImageButton clicked! Triggering onClick callback\n")
			b.onClick()
		}
		return true
	}

	// Return true if hover state changed
	return b.isHovered != wasHovered
}

// SetOnClick sets the click handler
func (b *ImageButton) SetOnClick(onClick func()) {
	b.onClick = onClick
}

// Cleanup releases the texture resources
func (b *ImageButton) Cleanup() {
	if b.textureID != 0 {
		gl.DeleteTextures(1, &b.textureID)
	}
	if b.hoverTexture != 0 && b.hoverTexture != b.textureID {
		gl.DeleteTextures(1, &b.hoverTexture)
	}
}
