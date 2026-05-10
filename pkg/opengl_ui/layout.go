package opengl_ui

// Layout represents a layout manager for arranging UI components
type Layout interface {
	Apply(components []UIComponent, bounds Rect)
}

// VerticalLayout arranges components vertically
type VerticalLayout struct {
	spacing float32
	padding float32
}

// NewVerticalLayout creates a new vertical layout
func NewVerticalLayout(spacing, padding float32) *VerticalLayout {
	return &VerticalLayout{
		spacing: spacing,
		padding: padding,
	}
}

// Apply applies vertical layout to components
func (v *VerticalLayout) Apply(components []UIComponent, bounds Rect) {
	currentY := bounds.Y + v.padding
	availableWidth := bounds.Width - 2*v.padding

	for _, component := range components {
		component.SetPosition(bounds.X + v.padding, currentY)
		component.SetSize(availableWidth, component.GetBounds().Height)
		currentY += component.GetBounds().Height + v.spacing
	}
}

// HorizontalLayout arranges components horizontally
type HorizontalLayout struct {
	spacing float32
	padding float32
}

// NewHorizontalLayout creates a new horizontal layout
func NewHorizontalLayout(spacing, padding float32) *HorizontalLayout {
	return &HorizontalLayout{
		spacing: spacing,
		padding: padding,
	}
}

// Apply applies horizontal layout to components
func (h *HorizontalLayout) Apply(components []UIComponent, bounds Rect) {
	currentX := bounds.X + h.padding
	availableHeight := bounds.Height - 2*h.padding

	for _, component := range components {
		component.SetPosition(currentX, bounds.Y + h.padding)
		component.SetSize(component.GetBounds().Width, availableHeight)
		currentX += component.GetBounds().Width + h.spacing
	}
}

// AbsoluteLayout allows components to be positioned absolutely
type AbsoluteLayout struct{}

// NewAbsoluteLayout creates a new absolute layout
func NewAbsoluteLayout() *AbsoluteLayout {
	return &AbsoluteLayout{}
}

// Apply does nothing for absolute layout (components keep their positions)
func (a *AbsoluteLayout) Apply(components []UIComponent, bounds Rect) {
	// Components maintain their absolute positions
}
