package ui

import (
	"sync"
	"kaijuengine.com/matrix"
)

// UIComponent represents a base UI component
type UIComponent struct {
	ID          string
	Type        ComponentType
	Position    matrix.Vec2
	Size        matrix.Vec2
	Visible     bool
	Enabled     bool
	Parent      *UIComponent
	Children    []*UIComponent
	Style       *Style
	OnClick     func()
	OnHover     func()
	OnFocus     func()
	OnBlur      func()
	Data        interface{}
	mu          sync.RWMutex
}

// ComponentType represents different UI component types
type ComponentType int

const (
	ComponentTypePanel ComponentType = iota
	ComponentTypeButton
	ComponentTypeLabel
	ComponentTypeTextField
	ComponentTypeImage
	ComponentTypeProgressBar
	ComponentTypeSlider
	ComponentTypeCheckBox
	ComponentTypeRadioButton
	ComponentTypeDropdown
	ComponentTypeListBox
	ComponentTypeScrollView
	ComponentTypeTabView
)

// Style represents styling properties for UI components
type Style struct {
	BackgroundColor matrix.Color
	ForegroundColor matrix.Color
	BorderColor   matrix.Color
	BorderWidth   float32
	CornerRadius  float32
	Font          string
	FontSize      float32
	Padding       matrix.Vec4 // top, right, bottom, left
	Margin        matrix.Vec4 // top, right, bottom, left
	Opacity       float32
	Shadow        ShadowStyle
}

// ShadowStyle represents shadow properties
type ShadowStyle struct {
	Color  matrix.Color
	Offset matrix.Vec2
	Blur   float32
	Spread float32
}

// NewUIComponent creates a new UI component
func NewUIComponent(id string, componentType ComponentType, position, size matrix.Vec2) *UIComponent {
	return &UIComponent{
		ID:       id,
		Type:     componentType,
		Position: position,
		Size:     size,
		Visible:  true,
		Enabled:  true,
		Children: make([]*UIComponent, 0),
		Style:    &Style{
			BackgroundColor: matrix.ColorWhite(),
			ForegroundColor: matrix.ColorBlack(),
			BorderColor:   matrix.ColorTransparent(),
			BorderWidth:   0,
			CornerRadius:  0,
			FontSize:      16,
			Padding:       matrix.NewVec4(5, 5, 5, 5),
			Margin:        matrix.NewVec4(0, 0, 0, 0),
			Opacity:       1.0,
		},
	}
}

// AddChild adds a child component
func (c *UIComponent) AddChild(child *UIComponent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	child.Parent = c
	c.Children = append(c.Children, child)
}

// RemoveChild removes a child component
func (c *UIComponent) RemoveChild(childID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for i, child := range c.Children {
		if child.ID == childID {
			c.Children = append(c.Children[:i], c.Children[i+1:]...)
			child.Parent = nil
			break
		}
	}
}

// GetChild returns a child component by ID
func (c *UIComponent) GetChild(childID string) *UIComponent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	for _, child := range c.Children {
		if child.ID == childID {
			return child
		}
	}
	return nil
}

// GetChildren returns all child components
func (c *UIComponent) GetChildren() []*UIComponent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	children := make([]*UIComponent, len(c.Children))
	copy(children, c.Children)
	return children
}

// SetPosition sets the component position
func (c *UIComponent) SetPosition(position matrix.Vec2) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Position = position
}

// GetPosition returns the component position
func (c *UIComponent) GetPosition() matrix.Vec2 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Position
}

// SetSize sets the component size
func (c *UIComponent) SetSize(size matrix.Vec2) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Size = size
}

// GetSize returns the component size
func (c *UIComponent) GetSize() matrix.Vec2 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Size
}

// SetVisible sets component visibility
func (c *UIComponent) SetVisible(visible bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Visible = visible
}

// IsVisible returns component visibility
func (c *UIComponent) IsVisible() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Visible
}

// SetEnabled sets component enabled state
func (c *UIComponent) SetEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Enabled = enabled
}

// IsEnabled returns component enabled state
func (c *UIComponent) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Enabled
}

// GetBounds returns the component bounds
func (c *UIComponent) GetBounds() UIRect {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return UIRect{
		X:      c.Position.X(),
		Y:      c.Position.Y(),
		Width:  c.Size.X(),
		Height: c.Size.Y(),
	}
}

// ContainsPoint checks if a point is within the component bounds
func (c *UIComponent) ContainsPoint(point matrix.Vec2) bool {
	bounds := c.GetBounds()
	return point.X() >= bounds.X && point.X() <= bounds.X+bounds.Width &&
		point.Y() >= bounds.Y && point.Y() <= bounds.Y+bounds.Height
}

// GetAbsolutePosition returns the absolute position (relative to root)
func (c *UIComponent) GetAbsolutePosition() matrix.Vec2 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.Parent == nil {
		return c.Position
	}
	
	parentPos := c.Parent.GetAbsolutePosition()
	return matrix.NewVec2(c.Position.X()+parentPos.X(), c.Position.Y()+parentPos.Y())
}

// Panel represents a panel UI component
type Panel struct {
	*UIComponent
}

// NewPanel creates a new panel
func NewPanel(id string, position, size matrix.Vec2) *Panel {
	component := NewUIComponent(id, ComponentTypePanel, position, size)
	return &Panel{UIComponent: component}
}

// Button represents a button UI component
type Button struct {
	*UIComponent
	Text     string
	Pressed  bool
	Hovered  bool
}

// NewButton creates a new button
func NewButton(id string, text string, position, size matrix.Vec2) *Button {
	component := NewUIComponent(id, ComponentTypeButton, position, size)
	return &Button{
		UIComponent: component,
		Text:        text,
	}
}

// SetText sets the button text
func (b *Button) SetText(text string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Text = text
}

// GetText returns the button text
func (b *Button) GetText() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Text
}

// Label represents a label UI component
type Label struct {
	*UIComponent
	Text string
}

// NewLabel creates a new label
func NewLabel(id string, text string, position, size matrix.Vec2) *Label {
	component := NewUIComponent(id, ComponentTypeLabel, position, size)
	return &Label{
		UIComponent: component,
		Text:        text,
	}
}

// SetText sets the label text
func (l *Label) SetText(text string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Text = text
}

// GetText returns the label text
func (l *Label) GetText() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.Text
}

// ProgressBar represents a progress bar UI component
type ProgressBar struct {
	*UIComponent
	Value    float32 // 0-1
	ShowText bool
}

// NewProgressBar creates a new progress bar
func NewProgressBar(id string, position, size matrix.Vec2) *ProgressBar {
	component := NewUIComponent(id, ComponentTypeProgressBar, position, size)
	return &ProgressBar{
		UIComponent: component,
		Value:       0,
		ShowText:    true,
	}
}

// SetValue sets the progress value (0-1)
func (pb *ProgressBar) SetValue(value float32) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	
	if value < 0 {
		value = 0
	} else if value > 1 {
		value = 1
	}
	pb.Value = value
}

// GetValue returns the progress value
func (pb *ProgressBar) GetValue() float32 {
	pb.mu.RLock()
	defer pb.mu.RUnlock()
	return pb.Value
}

// SetShowText sets whether to show percentage text
func (pb *ProgressBar) SetShowText(show bool) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.ShowText = show
}

// GetShowText returns whether percentage text is shown
func (pb *ProgressBar) GetShowText() bool {
	pb.mu.RLock()
	defer pb.mu.RUnlock()
	return pb.ShowText
}

// Image represents an image UI component
type Image struct {
	*UIComponent
	TextureID string
}

// NewImage creates a new image
func NewImage(id string, textureID string, position, size matrix.Vec2) *Image {
	component := NewUIComponent(id, ComponentTypeImage, position, size)
	return &Image{
		UIComponent: component,
		TextureID:   textureID,
	}
}

// SetTextureID sets the texture ID
func (img *Image) SetTextureID(textureID string) {
	img.mu.Lock()
	defer img.mu.Unlock()
	img.TextureID = textureID
}

// GetTextureID returns the texture ID
func (img *Image) GetTextureID() string {
	img.mu.RLock()
	defer img.mu.RUnlock()
	return img.TextureID
}

// UIRect represents a UI rectangle
type UIRect struct {
	X      float32
	Y      float32
	Width  float32
	Height float32
}

// ContainsPoint checks if a point is within the rectangle
func (r UIRect) ContainsPoint(point matrix.Vec2) bool {
	return point.X() >= r.X && point.X() <= r.X+r.Width &&
		point.Y() >= r.Y && point.Y() <= r.Y+r.Height
}

// GetCenter returns the center point of the rectangle
func (r UIRect) GetCenter() matrix.Vec2 {
	return matrix.NewVec2(r.X+r.Width/2, r.Y+r.Height/2)
}

// UIEvent represents a UI event
type UIEvent struct {
	Type      UIEventType
	Component *UIComponent
	Position  matrix.Vec2
	Data      interface{}
}

// UIEventType represents different UI event types
type UIEventType int

const (
	UIEventTypeClick UIEventType = iota
	UIEventTypeHover
	UIEventTypeUnhover
	UIEventTypeFocus
	UIEventTypeBlur
	UIEventTypeKeyPress
	UIEventTypeKeyRelease
	UIEventTypeMouseMove
	UIEventTypeMouseScroll
)

// UIEventManager manages UI events
type UIEventManager struct {
	mu      sync.RWMutex
	handlers map[UIEventType][]func(UIEvent)
}

// NewUIEventManager creates a new UI event manager
func NewUIEventManager() *UIEventManager {
	return &UIEventManager{
		handlers: make(map[UIEventType][]func(UIEvent)),
	}
}

// AddHandler adds an event handler
func (em *UIEventManager) AddHandler(eventType UIEventType, handler func(UIEvent)) {
	em.mu.Lock()
	defer em.mu.Unlock()
	
	em.handlers[eventType] = append(em.handlers[eventType], handler)
}

// RemoveHandler removes an event handler
func (em *UIEventManager) RemoveHandler(eventType UIEventType, handler func(UIEvent)) {
	em.mu.Lock()
	defer em.mu.Unlock()
	
	handlers := em.handlers[eventType]
	for i, h := range handlers {
		// Simple reference comparison - in practice would use more sophisticated matching
		if &h == &handler {
			em.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// DispatchEvent dispatches an event to all handlers
func (em *UIEventManager) DispatchEvent(event UIEvent) {
	em.mu.RLock()
	handlers := em.handlers[event.Type]
	em.mu.RUnlock()
	
	for _, handler := range handlers {
		handler(event)
	}
}

// UILayout represents a layout system for UI components
type UILayout struct {
	Components []*UIComponent
	Spacing    float32
	Padding    matrix.Vec4
}

// NewUILayout creates a new UI layout
func NewUILayout(spacing float32, padding matrix.Vec4) *UILayout {
	return &UILayout{
		Components: make([]*UIComponent, 0),
		Spacing:    spacing,
		Padding:    padding,
	}
}

// AddComponent adds a component to the layout
func (l *UILayout) AddComponent(component *UIComponent) {
	l.Components = append(l.Components, component)
}

// ArrangeHorizontal arranges components horizontally
func (l *UILayout) ArrangeHorizontal(startPos matrix.Vec2) {
	currentX := startPos.X() + l.Padding.X()
	
	for _, component := range l.Components {
		component.SetPosition(matrix.NewVec2(currentX, startPos.Y()+l.Padding.Y()))
		currentX += component.GetSize().X() + l.Spacing
	}
}

// ArrangeVertical arranges components vertically
func (l *UILayout) ArrangeVertical(startPos matrix.Vec2) {
	currentY := startPos.Y() + l.Padding.Y()
	
	for _, component := range l.Components {
		component.SetPosition(matrix.NewVec2(startPos.X()+l.Padding.X(), currentY))
		currentY += component.GetSize().Y() + l.Spacing
	}
}

// ArrangeGrid arranges components in a grid
func (l *UILayout) ArrangeGrid(startPos matrix.Vec2, columns int) {
	currentX := startPos.X() + l.Padding.X()
	currentY := startPos.Y() + l.Padding.Y()
	col := 0
	
	for _, component := range l.Components {
		component.SetPosition(matrix.NewVec2(currentX, currentY))
		
		col++
		if col >= columns {
			col = 0
			currentX = startPos.X() + l.Padding.X()
			currentY += component.GetSize().Y() + l.Spacing
		} else {
			currentX += component.GetSize().X() + l.Spacing
		}
	}
}
