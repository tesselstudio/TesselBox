# Modern UI System for TesselBox

This document describes the modern UI implementation that matches the provided screenshots for login and game mode selection screens.

## Overview

The modern UI system extends the existing TesselBox UI framework with contemporary design patterns while maintaining compatibility with the existing MenuManager system.

## Components

### New Modern Components

#### Card
Modern card component with shadows, rounded corners, and glassmorphism effects.
```go
card := NewCard("card_id", position, size)
card.ModernStyle.Style.BackgroundColor = matrix.NewColor(30/255.0, 30/255.0, 30/255.0, 0.9)
```

#### IconButton
Button with integrated icon support for GitHub login and other actions.
```go
githubButton := NewIconButton("github_login", "github_icon", "Continue with GitHub", position, size)
```

#### UserProfile
User profile display with avatar, username, and connection status.
```go
profile := NewUserProfile("user_profile", position, size)
profile.SetUsername("dev_player")
profile.SetStatus("Connected", true)
```

#### BackgroundPanel
Full-screen background with overlay support for winter landscape images.
```go
background := NewBackgroundPanel("bg", "winter_landscape", screenWidth, screenHeight)
```

#### GameModeCard
Specialized card for game mode selection with hover effects.
```go
modeCard := NewGameModeCard("singleplayer", "singleplayer", "icon", "Title", "Description", position, size)
```

#### ModernButton
Enhanced button with hover states and modern styling.
```go
button := NewModernButton("btn_id", "Text", position, size)
```

## Screens

### Login Screen (`login_screen.go`)

**Features:**
- Split-screen layout (left: login form, right: background visual)
- GitHub OAuth integration
- Modern glassmorphism design
- Legal links (Terms, Privacy, Help)
- Version display
- Loading states and error handling

**Usage:**
```go
loginScreen := NewLoginScreen(screenWidth, screenHeight)
loginScreen.Show()
```

### Game Selection Screen (`game_select_screen.go`)

**Features:**
- Top bar with TesselBox logo and user profile
- Three game mode cards (Singleplayer, Multiplayer, Settings)
- User authentication status display
- Sign out functionality
- Hover effects and transitions

**Usage:**
```go
gameSelectScreen := NewGameSelectScreen(screenWidth, screenHeight)
gameSelectScreen.SetUser("username", "avatar_id", true)
gameSelectScreen.Show()
```

## Integration with MenuManager

### New Menu Types
```go
const (
    MenuTypeLogin      // Modern login screen
    MenuTypeGameSelect // Game mode selection
    // ... existing menu types
)
```

### Helper Functions
```go
menuManager.ShowLoginScreen()
menuManager.ShowGameSelectScreen()
menuManager.HandleAuthenticationSuccess(username, avatarID)
menuManager.HandleSignOut()
```

## OAuth Integration

The login screen integrates with the existing `pkg/oauth` system:

1. **GitHub OAuth Flow:**
   - User clicks "Continue with GitHub"
   - Browser opens to GitHub OAuth URL
   - System listens for OAuth callback
   - Authentication result processed
   - User profile extracted and stored

2. **Authentication States:**
   - Loading state during OAuth flow
   - Success state with user profile
   - Error state with retry option

## Styling

### Modern Design Principles
- **Dark Theme:** Primary background with semi-transparent overlays
- **Glassmorphism:** Blur effects and transparency
- **Shadows:** Soft shadows for depth and hierarchy
- **Rounded Corners:** Consistent 8-16px radius
- **Typography:** Clean sans-serif fonts
- **Colors:**
  - Primary: `rgb(59, 130, 246)` (GitHub blue)
  - Success: `rgb(76, 175, 80)` (Green)
  - Error: `rgb(239, 68, 68)` (Red)
  - Background: `rgb(20, 20, 20)` (Dark)
  - Text: `rgb(255, 255, 255)` (White)

### Responsive Design
Components automatically scale based on screen dimensions:
```go
cardWidth, cardHeight, spacing := ResponsiveLayout(screenWidth, screenHeight)
```

## Usage Example

```go
// Create modern menu manager
menuManager := CreateModernMenuManager(1920, 1080)

// Show login screen
menuManager.ShowLoginScreen()

// Handle authentication
menuManager.HandleAuthenticationSuccess("dev_player", "github_avatar")

// Show game selection
menuManager.ShowGameSelectScreen()

// Handle game mode selection
// (Implemented via card click handlers)
```

## File Structure

```
pkg/ui/
├── components.go           # Original UI components
├── modern_components.go   # New modern components
├── menus.go             # Enhanced MenuManager
├── login_screen.go       # Login screen implementation
├── game_select_screen.go  # Game selection screen
├── example_usage.go      # Usage examples
└── README_MODERN_UI.md  # This documentation
```

## Assets Required

To complete the visual implementation, add these assets to `game_content/`:

### Textures
- `winter_landscape` - Background image for login screen
- `game_select_bg` - Background for game selection
- `github_icon` - GitHub logo (24x24)
- `singleplayer_icon` - Singleplayer icon (48x48)
- `multiplayer_icon` - Multiplayer icon (48x48)
- `settings_icon` - Settings icon (48x48)
- `default_avatar` - Default user avatar (40x40)

### Fonts
- Modern sans-serif font for UI text
- Bold variant for headers

## Next Steps

1. **Asset Integration:** Add required textures and fonts
2. **OAuth Server:** Integrate with existing `pkg/oauth` handlers
3. **Browser Integration:** Implement system browser opening
4. **Animation:** Add smooth transitions between screens
5. **Accessibility:** Add keyboard navigation and screen reader support
6. **Testing:** Comprehensive UI testing across different resolutions

## Migration Path

The modern UI system is designed to work alongside existing menus:

1. **Phase 1:** Deploy modern UI as optional alternative
2. **Phase 2:** Set modern UI as default for new installations
3. **Phase 3:** Gradually migrate existing users
4. **Phase 4:** Remove legacy menu system

This approach ensures backward compatibility while providing a modern user experience.
