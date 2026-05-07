package ui

import (
	"time"

	"kaijuengine.com/matrix"
)

// LoginScreen represents modern login screen
type LoginScreen struct {
	*Menu
	BackgroundPanel  *BackgroundPanel
	MenuPanel        *Panel
	LoginCard        *Card
	LogoLabel        *Label
	TitleLabel       *Label
	DescriptionLabel *Label
	GitHubButton     *IconButton
	TermsLabel       *Label
	PrivacyLabel     *Label
	HelpLink         *Label
	PrivacyLink      *Label
	TermsLink        *Label
	VersionLabel     *Label
}

// NewLoginScreen creates a new login screen
func NewLoginScreen(screenWidth, screenHeight float32) *LoginScreen {
	// Create background panel
	background := NewBackgroundPanel("login_background", "winter_landscape", screenWidth, screenHeight)

	// Create main panel for the menu
	panelWidth := float32(500)
	panelHeight := float32(600)
	panelX := float32(50) // Position on left side
	panelY := (screenHeight - panelHeight) / 2

	menuPanel := NewPanel("login_menu_panel", matrix.NewVec2(panelX, panelY), matrix.NewVec2(panelWidth, panelHeight))
	menuPanel.Style.BackgroundColor = matrix.NewColor(0, 0, 0, 0) // Transparent

	// Create login card
	cardWidth := float32(400)
	cardHeight := float32(500)
	cardX := float32(50) // Position relative to panel
	cardY := float32(50)

	loginCard := NewCard("login_card", matrix.NewVec2(cardX, cardY), matrix.NewVec2(cardWidth, cardHeight))
	loginCard.ModernStyle.Style.BackgroundColor = matrix.NewColor(20/255.0, 20/255.0, 20/255.0, 0.85)
	loginCard.ModernStyle.Style.CornerRadius = 16
	loginCard.ModernStyle.Style.Padding = matrix.NewVec4(40, 40, 40, 40)
	loginCard.ModernStyle.Shadow = &ShadowStyle{
		Color:  matrix.NewColor(0, 0, 0, 0.5),
		Offset: matrix.NewVec2(0, 16),
		Blur:   32,
		Spread: 0,
	}

	// OAuth Login label
	oauthLabel := NewLabel("oauth_login", "OAuth Login", matrix.NewVec2(0, 20), matrix.NewVec2(200, 30))
	oauthLabel.Style.ForegroundColor = matrix.NewColor(150/255.0, 150/255.0, 150/255.0, 1.0)
	oauthLabel.Style.FontSize = 14

	// TesselBox logo/brand
	logoLabel := NewLabel("tesselbox_logo", "TesselBox", matrix.NewVec2(0, 70), matrix.NewVec2(200, 50))
	logoLabel.Style.ForegroundColor = matrix.ColorWhite()
	logoLabel.Style.FontSize = 36

	// Welcome message
	titleLabel := NewLabel("welcome_title", "Welcome back", matrix.NewVec2(0, 140), matrix.NewVec2(200, 40))
	titleLabel.Style.ForegroundColor = matrix.ColorWhite()
	titleLabel.Style.FontSize = 28

	// Description
	descriptionLabel := NewLabel("welcome_desc", "Sign in with your GitHub account to access your worlds and sync your progress.", matrix.NewVec2(0, 190), matrix.NewVec2(300, 60))
	descriptionLabel.Style.ForegroundColor = matrix.NewColor(200/255.0, 200/255.0, 200/255.0, 1.0)
	descriptionLabel.Style.FontSize = 14

	// GitHub login button
	gitHubButton := NewIconButton("github_login", "github_icon", "Continue with GitHub", matrix.NewVec2(0, 270), matrix.NewVec2(320, 50))

	// Terms and privacy text
	termsText := "By continuing, you agree to our Terms of Service and Privacy Policy."
	termsLabel := NewLabel("terms_text", termsText, matrix.NewVec2(0, 350), matrix.NewVec2(300, 40))
	termsLabel.Style.ForegroundColor = matrix.NewColor(150/255.0, 150/255.0, 150/255.0, 1.0)
	termsLabel.Style.FontSize = 12

	// Footer links
	helpLink := NewLabel("help_link", "Help", matrix.NewVec2(0, 400), matrix.NewVec2(40, 20))
	helpLink.Style.ForegroundColor = matrix.NewColor(150/255.0, 150/255.0, 150/255.0, 1.0)
	helpLink.Style.FontSize = 12

	privacyLink := NewLabel("privacy_link", "Privacy", matrix.NewVec2(60, 400), matrix.NewVec2(50, 20))
	privacyLink.Style.ForegroundColor = matrix.NewColor(150/255.0, 150/255.0, 150/255.0, 1.0)
	privacyLink.Style.FontSize = 12

	termsLink := NewLabel("terms_link", "Terms", matrix.NewVec2(130, 400), matrix.NewVec2(40, 20))
	termsLink.Style.ForegroundColor = matrix.NewColor(150/255.0, 150/255.0, 150/255.0, 1.0)
	termsLink.Style.FontSize = 12

	// Version label
	versionLabel := NewLabel("version_label", "v2.4.1 Early Access", matrix.NewVec2(screenWidth-200, screenHeight-30), matrix.NewVec2(180, 20))
	versionLabel.Style.ForegroundColor = matrix.NewColor(150/255.0, 150/255.0, 150/255.0, 1.0)
	versionLabel.Style.FontSize = 12

	// Add components to login card
	loginCard.AddChild(oauthLabel.UIComponent)
	loginCard.AddChild(logoLabel.UIComponent)
	loginCard.AddChild(titleLabel.UIComponent)
	loginCard.AddChild(descriptionLabel.UIComponent)
	loginCard.AddChild(gitHubButton.UIComponent)
	loginCard.AddChild(termsLabel.UIComponent)
	loginCard.AddChild(helpLink.UIComponent)
	loginCard.AddChild(privacyLink.UIComponent)
	loginCard.AddChild(termsLink.UIComponent)

	// Add card to panel
	menuPanel.AddChild(loginCard.UIComponent)

	// Create menu structure
	loginScreen := &LoginScreen{
		Menu: &Menu{
			ID:       "login_screen",
			Type:     MenuTypeLogin,
			Title:    "Login",
			Panel:    menuPanel,
			Visible:  false,
			Elements: []UIElement{},
		},
		BackgroundPanel:  background,
		MenuPanel:        menuPanel,
		LoginCard:        loginCard,
		LogoLabel:        logoLabel,
		TitleLabel:       titleLabel,
		DescriptionLabel: descriptionLabel,
		GitHubButton:     gitHubButton,
		TermsLabel:       termsLabel,
		PrivacyLabel:     nil,
		HelpLink:         helpLink,
		PrivacyLink:      privacyLink,
		TermsLink:        termsLink,
		VersionLabel:     versionLabel,
	}

	// Set up GitHub button click handler
	gitHubButton.OnClick = func() {
		loginScreen.handleGitHubLogin()
	}

	// Set up link handlers (these would open browser or show in-game browser)
	helpLink.OnClick = func() {
		loginScreen.openHelpLink()
	}

	privacyLink.OnClick = func() {
		loginScreen.openPrivacyLink()
	}

	termsLink.OnClick = func() {
		loginScreen.openTermsLink()
	}

	return loginScreen
}

// handleGitHubLogin initiates GitHub OAuth flow
func (ls *LoginScreen) handleGitHubLogin() {
	// Set loading state
	ls.SetLoadingState(true)

	// This would integrate with existing OAuth system from pkg/oauth
	// In a real implementation, this would:
	// 1. Generate OAuth URL using existing GitHubOAuth handler
	// 2. Open system browser to GitHub OAuth URL
	// 3. Listen for OAuth callback via local server
	// 4. Process authentication result
	// 5. Extract user information and token
	// 6. Transition to game selection screen

	// For now, simulate successful authentication after a delay
	// In production, this would be handled by OAuth callback handlers
	go func() {
		// Simulate network delay
		time.Sleep(2 * time.Second)

		// Simulate successful authentication
		// In real implementation, this would be called by OAuth callback
		ls.onAuthenticationSuccess("dev_player", "default_avatar")
	}()
}

// onAuthenticationSuccess handles successful authentication
func (ls *LoginScreen) onAuthenticationSuccess(username string, avatarID string) {
	ls.SetLoadingState(false)

	// This would call MenuManager to transition to game select screen
	// For now, we'll just update the UI to show success
	ls.DescriptionLabel.SetText("Authentication successful! Redirecting...")
	ls.DescriptionLabel.Style.ForegroundColor = matrix.NewColor(76/255.0, 175/255.0, 80/255.0, 1.0)
}

// openHelpLink opens help documentation
func (ls *LoginScreen) openHelpLink() {
	// Open help documentation in browser or show in-game help
	// Implementation depends on game's browser integration
}

// openPrivacyLink opens privacy policy
func (ls *LoginScreen) openPrivacyLink() {
	// Open privacy policy in browser
}

// openTermsLink opens terms of service
func (ls *LoginScreen) openTermsLink() {
	// Open terms of service in browser
}

// Show displays the login screen
func (ls *LoginScreen) Show() {
	ls.BackgroundPanel.Panel.SetVisible(true)
	ls.Menu.Panel.SetVisible(true)
	ls.Visible = true
}

// Hide hides the login screen
func (ls *LoginScreen) Hide() {
	ls.BackgroundPanel.Panel.SetVisible(false)
	ls.Menu.Panel.SetVisible(false)
	ls.Visible = false
}

// SetLoadingState sets the loading state for the login button
func (ls *LoginScreen) SetLoadingState(loading bool) {
	if loading {
		ls.GitHubButton.SetText("Signing in...")
		ls.GitHubButton.SetEnabled(false)
	} else {
		ls.GitHubButton.SetText("Continue with GitHub")
		ls.GitHubButton.SetEnabled(true)
	}
}

// SetError sets an error message on the login screen
func (ls *LoginScreen) SetError(message string) {
	// Could add an error label to display authentication errors
	ls.DescriptionLabel.SetText("Error: " + message)
	ls.DescriptionLabel.Style.ForegroundColor = matrix.NewColor(239/255.0, 68/255.0, 68/255.0, 1.0)
	ls.SetLoadingState(false)
}
