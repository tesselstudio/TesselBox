package ui

import (
	"kaijuengine.com/matrix"
)

// ModernStyle represents modern styling properties
type ModernStyle struct {
	*Style
	Gradient     *GradientStyle
	Shadow       *ShadowStyle
	Blur         float32
	BorderRadius float32
	Opacity      float32
}

// GradientStyle represents gradient properties
type GradientStyle struct {
	Type      string // "linear", "radial"
	Colors    []matrix.Color
	Direction matrix.Vec2 // for linear gradients
}

// Card represents a modern card component with shadows and rounded corners
type Card struct {
	*UIComponent
	Header      *Label
	Content     *UIComponent
	Footer      *UIComponent
	ModernStyle *ModernStyle
}

// NewCard creates a new modern card
func NewCard(id string, position, size matrix.Vec2) *Card {
	component := NewUIComponent(id, ComponentTypePanel, position, size)

	modernStyle := &ModernStyle{
		Style: &Style{
			BackgroundColor: matrix.NewColor(20/255.0, 20/255.0, 20/255.0, 0.9),
			ForegroundColor: matrix.ColorWhite(),
			BorderColor:     matrix.NewColor(255/255.0, 255/255.0, 255/255.0, 0.1),
			BorderWidth:     1,
			CornerRadius:    12,
			FontSize:        16,
			Padding:         matrix.NewVec4(20, 20, 20, 20),
			Margin:          matrix.NewVec4(10, 10, 10, 10),
		},
		Shadow: &ShadowStyle{
			Color:  matrix.NewColor(0, 0, 0, 0.3),
			Offset: matrix.NewVec2(0, 4),
			Blur:   12,
			Spread: 0,
		},
		BorderRadius: 12,
		Opacity:      0.95,
	}

	component.Style = modernStyle.Style

	return &Card{
		UIComponent: component,
		ModernStyle: modernStyle,
	}
}

// SetHeader sets the card header
func (c *Card) SetHeader(header *Label) {
	c.Header = header
	c.AddChild(header.UIComponent)
}

// SetContent sets the card content
func (c *Card) SetContent(content *UIComponent) {
	c.Content = content
	c.AddChild(content)
}

// SetFooter sets the card footer
func (c *Card) SetFooter(footer *UIComponent) {
	c.Footer = footer
	c.AddChild(footer)
}

// IconButton represents a button with icon support
type IconButton struct {
	*Button
	Icon        *Image
	IconText    string
	IconSize    matrix.Vec2
	TextSpacing float32
}

// NewIconButton creates a new icon button
func NewIconButton(id string, iconID string, text string, position, size matrix.Vec2) *IconButton {
	button := NewButton(id, text, position, size)

	// Create icon if iconID provided
	var icon *Image
	if iconID != "" {
		iconSize := matrix.NewVec2(24, 24) // Default icon size
		icon = NewImage(id+"_icon", iconID, matrix.NewVec2(10, (size.Y()-iconSize.Y())/2), iconSize)
	}

	// Modern styling
	button.Style.BackgroundColor = matrix.NewColor(59/255.0, 130/255.0, 246/255.0, 1.0) // GitHub blue
	button.Style.ForegroundColor = matrix.ColorWhite()
	button.Style.CornerRadius = 8
	button.Style.Padding = matrix.NewVec4(16, 20, 16, 20)

	return &IconButton{
		Button:      button,
		Icon:        icon,
		IconText:    text,
		IconSize:    matrix.NewVec2(24, 24),
		TextSpacing: 12,
	}
}

// GetIcon returns the icon component
func (ib *IconButton) GetIcon() *Image {
	return ib.Icon
}

// SetIconSize sets the icon size
func (ib *IconButton) SetIconSize(size matrix.Vec2) {
	ib.IconSize = size
	if ib.Icon != nil {
		ib.Icon.SetSize(size)
	}
}

// UserProfile represents a user profile display component
type UserProfile struct {
	*UIComponent
	Avatar      *Image
	Username    *Label
	Status      *Label
	IsConnected bool
}

// NewUserProfile creates a new user profile component
func NewUserProfile(id string, position, size matrix.Vec2) *UserProfile {
	component := NewUIComponent(id, ComponentTypePanel, position, size)

	// Create avatar placeholder
	avatarSize := matrix.NewVec2(40, 40)
	avatar := NewImage(id+"_avatar", "default_avatar", matrix.NewVec2(0, 0), avatarSize)
	avatar.Style.CornerRadius = 20 // Circular

	// Create username label
	username := NewLabel(id+"_username", "@dev_player", matrix.NewVec2(50, 5), matrix.NewVec2(150, 20))
	username.Style.ForegroundColor = matrix.ColorWhite()
	username.Style.FontSize = 16

	// Create status label
	status := NewLabel(id+"_status", "Connected", matrix.NewVec2(50, 25), matrix.NewVec2(150, 15))
	status.Style.ForegroundColor = matrix.NewColor(76/255.0, 175/255.0, 80/255.0, 1.0) // Green
	status.Style.FontSize = 12

	// Component styling
	component.Style.BackgroundColor = matrix.NewColor(30/255.0, 30/255.0, 30/255.0, 0.8)
	component.Style.CornerRadius = 8
	component.Style.Padding = matrix.NewVec4(10, 15, 10, 15)

	userProfile := &UserProfile{
		UIComponent: component,
		Avatar:      avatar,
		Username:    username,
		Status:      status,
		IsConnected: true,
	}

	// Add children
	component.AddChild(avatar.UIComponent)
	component.AddChild(username.UIComponent)
	component.AddChild(status.UIComponent)

	return userProfile
}

// SetUsername sets the username display
func (up *UserProfile) SetUsername(username string) {
	up.Username.SetText("@" + username)
}

// SetStatus sets the connection status
func (up *UserProfile) SetStatus(status string, connected bool) {
	up.Status.SetText(status)
	up.IsConnected = connected

	if connected {
		up.Status.Style.ForegroundColor = matrix.NewColor(76/255.0, 175/255.0, 80/255.0, 1.0)
	} else {
		up.Status.Style.ForegroundColor = matrix.NewColor(239/255.0, 68/255.0, 68/255.0, 1.0)
	}
}

// BackgroundPanel represents a full-screen background panel
type BackgroundPanel struct {
	*Panel
	TextureID    string
	OverlayColor matrix.Color
	Blur         float32
}

// NewBackgroundPanel creates a new background panel
func NewBackgroundPanel(id string, textureID string, screenWidth, screenHeight float32) *BackgroundPanel {
	panel := NewPanel(id, matrix.NewVec2(0, 0), matrix.NewVec2(screenWidth, screenHeight))

	// Dark overlay for better text readability
	overlayColor := matrix.NewColor(0, 0, 0, 0.3)

	backgroundPanel := &BackgroundPanel{
		Panel:        panel,
		TextureID:    textureID,
		OverlayColor: overlayColor,
		Blur:         0,
	}

	panel.Style.BackgroundColor = overlayColor

	return backgroundPanel
}

// SetTexture sets the background texture
func (bp *BackgroundPanel) SetTexture(textureID string) {
	bp.TextureID = textureID
	// In a full implementation, this would update the rendering
}

// SetOverlay sets the overlay color and opacity
func (bp *BackgroundPanel) SetOverlay(color matrix.Color) {
	bp.OverlayColor = color
	bp.Style.BackgroundColor = color
}

// GameModeCard represents a game mode selection card
type GameModeCard struct {
	*Card
	Icon        *Image
	Title       *Label
	Description *Label
	Mode        string
}

// NewGameModeCard creates a new game mode card
func NewGameModeCard(id string, mode string, iconID string, title string, description string, position, size matrix.Vec2) *GameModeCard {
	card := NewCard(id, position, size)

	// Modern card styling
	card.ModernStyle.Style.BackgroundColor = matrix.NewColor(30/255.0, 30/255.0, 30/255.0, 0.9)
	card.ModernStyle.Style.CornerRadius = 16
	card.ModernStyle.Style.Padding = matrix.NewVec4(24, 24, 24, 24)
	card.ModernStyle.Shadow = &ShadowStyle{
		Color:  matrix.NewColor(0, 0, 0, 0.4),
		Offset: matrix.NewVec2(0, 8),
		Blur:   16,
		Spread: 0,
	}

	// Create icon
	iconSize := matrix.NewVec2(48, 48)
	icon := NewImage(id+"_icon", iconID, matrix.NewVec2(0, 0), iconSize)

	// Create title
	titleLabel := NewLabel(id+"_title", title, matrix.NewVec2(60, 8), matrix.NewVec2(200, 30))
	titleLabel.Style.ForegroundColor = matrix.ColorWhite()
	titleLabel.Style.FontSize = 20

	// Create description
	descLabel := NewLabel(id+"_desc", description, matrix.NewVec2(60, 40), matrix.NewVec2(250, 40))
	descLabel.Style.ForegroundColor = matrix.NewColor(200/255.0, 200/255.0, 200/255.0, 1.0)
	descLabel.Style.FontSize = 14

	gameModeCard := &GameModeCard{
		Card:        card,
		Icon:        icon,
		Title:       titleLabel,
		Description: descLabel,
		Mode:        mode,
	}

	// Add components to card
	card.AddChild(icon.UIComponent)
	card.AddChild(titleLabel.UIComponent)
	card.AddChild(descLabel.UIComponent)

	return gameModeCard
}

// SetHoverState sets the hover appearance
func (gmc *GameModeCard) SetHoverState(isHovered bool) {
	if isHovered {
		gmc.ModernStyle.Style.BackgroundColor = matrix.NewColor(50/255.0, 50/255.0, 50/255.0, 0.95)
		gmc.ModernStyle.Shadow.Offset = matrix.NewVec2(0, 12)
		gmc.ModernStyle.Shadow.Blur = 20
	} else {
		gmc.ModernStyle.Style.BackgroundColor = matrix.NewColor(30/255.0, 30/255.0, 30/255.0, 0.9)
		gmc.ModernStyle.Shadow.Offset = matrix.NewVec2(0, 8)
		gmc.ModernStyle.Shadow.Blur = 16
	}
}

// ModernButton represents a modern styled button
type ModernButton struct {
	*Button
	HoverColor  matrix.Color
	NormalColor matrix.Color
	ClickColor  matrix.Color
}

// NewModernButton creates a new modern button
func NewModernButton(id string, text string, position, size matrix.Vec2) *ModernButton {
	button := NewButton(id, text, position, size)

	normalColor := matrix.NewColor(59/255.0, 130/255.0, 246/255.0, 1.0)
	hoverColor := matrix.NewColor(79/255.0, 150/255.0, 255/255.0, 1.0)

	button.Style.BackgroundColor = normalColor
	button.Style.ForegroundColor = matrix.ColorWhite()
	button.Style.CornerRadius = 8
	button.Style.Padding = matrix.NewVec4(12, 24, 12, 24)
	button.Style.BorderWidth = 0

	return &ModernButton{
		Button:      button,
		NormalColor: normalColor,
		HoverColor:  hoverColor,
		ClickColor:  matrix.NewColor(49/255.0, 110/255.0, 200/255.0, 1.0),
	}
}

// SetHoverState sets button hover appearance
func (mb *ModernButton) SetHoverState(isHovered bool) {
	if isHovered {
		mb.Style.BackgroundColor = mb.HoverColor
	} else {
		mb.Style.BackgroundColor = mb.NormalColor
	}
}
