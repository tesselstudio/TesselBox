package ui

import (
	"strconv"
	"sync"
	"time"

	"github.com/tesselstudio/TesselBox/pkg/world"
	"kaijuengine.com/engine"
	"kaijuengine.com/matrix"
)

// GitHubUser represents a GitHub user profile (simplified for UI integration)
type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// MenuType represents different menu types
type MenuType int

const (
	MenuTypeMain MenuType = iota
	MenuTypeLogin
	MenuTypeGameSelect
	MenuTypePause
	MenuTypeSettings
	MenuTypeInventory
	MenuTypeCrafting
	MenuTypeWorldSelect
)

// MenuManager manages all game menus
type MenuManager struct {
	mu sync.RWMutex

	currentMenu MenuType
	menus       map[MenuType]*Menu
	visible     bool

	// Fyne integration
	fyneBridge *FyneKaijuBridge
	useFyne    bool

	// Callbacks
	onStartGame     func(worldName string, seed int64)
	onResumeGame    func()
	onSaveGame      func()
	onLoadGame      func(worldName string)
	onQuitToMenu    func()
	onQuitToDesktop func()
}

// Menu represents a game menu
type Menu struct {
	ID       string
	Type     MenuType
	Title    string
	Panel    *Panel
	Visible  bool
	Elements []UIElement
}

// UIElement represents a UI element in a menu
type UIElement struct {
	Type     string
	ID       string
	Label    string
	Position matrix.Vec2
	Size     matrix.Vec2
	Callback func()
}

// NewMenuManager creates a new menu manager
func NewMenuManager(screenWidth, screenHeight float32) *MenuManager {
	mm := &MenuManager{
		menus:   make(map[MenuType]*Menu),
		visible: false,
	}

	mm.createMenus(screenWidth, screenHeight)

	return mm
}

// createMenus creates all game menus
func (mm *MenuManager) createMenus(screenWidth, screenHeight float32) {
	// Login Screen
	loginScreen := NewLoginScreen(screenWidth, screenHeight)
	mm.menus[MenuTypeLogin] = loginScreen.Menu

	// Game Select Screen
	gameSelectScreen := NewGameSelectScreen(screenWidth, screenHeight)
	mm.menus[MenuTypeGameSelect] = gameSelectScreen.Menu

	// Main Menu
	mainMenu := &Menu{
		ID:    "main_menu",
		Type:  MenuTypeMain,
		Title: "TesselBox 3D",
		Panel: NewPanel("main_menu_panel", matrix.NewVec2(screenWidth/2-200, screenHeight/2-250), matrix.NewVec2(400, 500)),
	}
	mainMenu.Panel.Style.BackgroundColor = matrix.NewColor(0, 0, 0, 200/255.0)
	mainMenu.Panel.Style.CornerRadius = 10

	// Add title
	titleLabel := NewLabel("main_title", "TesselBox 3D", matrix.NewVec2(100, 30), matrix.NewVec2(200, 40))
	titleLabel.Style.FontSize = 32
	titleLabel.Style.ForegroundColor = matrix.ColorWhite()
	mainMenu.Panel.AddChild(titleLabel.UIComponent)

	// Add buttons
	buttons := []struct {
		label    string
		callback func()
	}{
		{"New World", mm.showWorldCreation},
		{"Load World", mm.showWorldSelect},
		{"Settings", mm.showSettings},
		{"Quit", mm.quitToDesktop},
	}

	for i, btn := range buttons {
		button := NewButton("main_btn_"+strconv.Itoa(i), btn.label, matrix.NewVec2(100, float32(100+i*60)), matrix.NewVec2(200, 50))
		button.Style.BackgroundColor = matrix.NewColor(100/255.0, 100/255.0, 100/255.0, 255/255.0)
		button.Style.ForegroundColor = matrix.ColorWhite()
		button.Style.CornerRadius = 5
		mainMenu.Panel.AddChild(button.UIComponent)
	}

	mm.menus[MenuTypeMain] = mainMenu

	// Pause Menu
	pauseMenu := &Menu{
		ID:    "pause_menu",
		Type:  MenuTypePause,
		Title: "Paused",
		Panel: NewPanel("pause_menu_panel", matrix.NewVec2(screenWidth/2-150, screenHeight/2-200), matrix.NewVec2(300, 400)),
	}
	pauseMenu.Panel.Style.BackgroundColor = matrix.NewColor(0, 0, 0, 200/255.0)
	pauseMenu.Panel.Style.CornerRadius = 10

	// Pause title
	pauseTitle := NewLabel("pause_title", "Paused", matrix.NewVec2(100, 30), matrix.NewVec2(100, 40))
	pauseTitle.Style.FontSize = 28
	pauseTitle.Style.ForegroundColor = matrix.ColorWhite()
	pauseMenu.Panel.AddChild(pauseTitle.UIComponent)

	// Pause buttons
	pauseButtons := []struct {
		label    string
		callback func()
	}{
		{"Resume", mm.resumeGame},
		{"Save Game", mm.saveGame},
		{"Settings", mm.showSettings},
		{"Quit to Menu", mm.quitToMenu},
		{"Quit to Desktop", mm.quitToDesktop},
	}

	for i, btn := range pauseButtons {
		button := NewButton("pause_btn_"+strconv.Itoa(i), btn.label, matrix.NewVec2(50, float32(100+i*60)), matrix.NewVec2(200, 50))
		button.Style.BackgroundColor = matrix.NewColor(100/255.0, 100/255.0, 100/255.0, 255/255.0)
		button.Style.ForegroundColor = matrix.ColorWhite()
		button.Style.CornerRadius = 5
		pauseMenu.Panel.AddChild(button.UIComponent)
	}

	mm.menus[MenuTypePause] = pauseMenu
}

// ShowMenu shows a specific menu
func (mm *MenuManager) ShowMenu(menuType MenuType) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Hide all menus
	for _, menu := range mm.menus {
		menu.Panel.SetVisible(false)
	}

	// Show requested menu
	if menu, exists := mm.menus[menuType]; exists {
		menu.Panel.SetVisible(true)
		mm.currentMenu = menuType
		mm.visible = true
	}
}

// HideMenu hides the current menu
func (mm *MenuManager) HideMenu() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	for _, menu := range mm.menus {
		menu.Panel.SetVisible(false)
	}

	mm.visible = false
}

// IsVisible returns true if any menu is visible
func (mm *MenuManager) IsVisible() bool {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	return mm.visible
}

// GetCurrentMenu returns the current menu type
func (mm *MenuManager) GetCurrentMenu() MenuType {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	return mm.currentMenu
}

// GetMenuPanel returns the panel for a menu type
func (mm *MenuManager) GetMenuPanel(menuType MenuType) *Panel {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if menu, exists := mm.menus[menuType]; exists {
		return menu.Panel
	}
	return nil
}

// SetOnStartGame sets the start game callback
func (mm *MenuManager) SetOnStartGame(callback func(worldName string, seed int64)) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.onStartGame = callback
}

// SetOnResumeGame sets the resume game callback
func (mm *MenuManager) SetOnResumeGame(callback func()) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.onResumeGame = callback
}

// SetOnSaveGame sets the save game callback
func (mm *MenuManager) SetOnSaveGame(callback func()) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.onSaveGame = callback
}

// SetOnLoadGame sets the load game callback
func (mm *MenuManager) SetOnLoadGame(callback func(worldName string)) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.onLoadGame = callback
}

// SetOnQuitToMenu sets the quit to menu callback
func (mm *MenuManager) SetOnQuitToMenu(callback func()) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.onQuitToMenu = callback
}

// SetOnQuitToDesktop sets the quit to desktop callback
func (mm *MenuManager) SetOnQuitToDesktop(callback func()) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.onQuitToDesktop = callback
}

// Menu action handlers
func (mm *MenuManager) showWorldCreation() {
	// Create world creation menu if it doesn't exist
	if _, exists := mm.menus[MenuTypeWorldSelect]; !exists {
		mm.createWorldSelectMenu()
	}
	// Show creation dialog as overlay or separate menu
	mm.createWorldCreationDialog()
}

// createWorldCreationDialog creates the world creation UI
func (mm *MenuManager) createWorldCreationDialog() {
	screenWidth := float32(1920)
	screenHeight := float32(1080)

	// Use world select menu as base, add creation dialog overlay
	creationPanel := NewPanel("world_creation_panel", matrix.NewVec2(screenWidth/2-250, screenHeight/2-250), matrix.NewVec2(500, 500))
	creationPanel.Style.BackgroundColor = matrix.NewColor(0, 0, 0, 240/255.0)
	creationPanel.Style.CornerRadius = 10

	// Title
	titleLabel := NewLabel("creation_title", "Create New World", matrix.NewVec2(150, 20), matrix.NewVec2(200, 40))
	titleLabel.Style.FontSize = 24
	titleLabel.Style.ForegroundColor = matrix.ColorWhite()
	creationPanel.AddChild(titleLabel.UIComponent)

	// World name input
	nameLabel := NewLabel("creation_name_label", "World Name:", matrix.NewVec2(30, 80), matrix.NewVec2(120, 25))
	nameLabel.Style.ForegroundColor = matrix.ColorWhite()
	nameLabel.Style.FontSize = 16
	creationPanel.AddChild(nameLabel.UIComponent)

	// Name input field (using a button as placeholder for text input)
	nameInput := NewButton("creation_name_input", "New World", matrix.NewVec2(160, 80), matrix.NewVec2(300, 30))
	nameInput.Style.BackgroundColor = matrix.NewColor(60/255.0, 60/255.0, 60/255.0, 1.0)
	nameInput.Style.ForegroundColor = matrix.ColorWhite()
	nameInput.Style.CornerRadius = 3
	creationPanel.AddChild(nameInput.UIComponent)

	// Seed input
	seedLabel := NewLabel("creation_seed_label", "Seed:", matrix.NewVec2(30, 130), matrix.NewVec2(120, 25))
	seedLabel.Style.ForegroundColor = matrix.ColorWhite()
	seedLabel.Style.FontSize = 16
	creationPanel.AddChild(seedLabel.UIComponent)

	seedInput := NewButton("creation_seed_input", "Random", matrix.NewVec2(160, 130), matrix.NewVec2(200, 30))
	seedInput.Style.BackgroundColor = matrix.NewColor(60/255.0, 60/255.0, 60/255.0, 1.0)
	seedInput.Style.ForegroundColor = matrix.ColorWhite()
	seedInput.Style.CornerRadius = 3
	creationPanel.AddChild(seedInput.UIComponent)

	// Randomize seed button
	randomSeedBtn := NewButton("creation_random_seed", "Randomize", matrix.NewVec2(370, 130), matrix.NewVec2(90, 30))
	randomSeedBtn.Style.BackgroundColor = matrix.NewColor(80/255.0, 80/255.0, 80/255.0, 1.0)
	randomSeedBtn.Style.ForegroundColor = matrix.ColorWhite()
	randomSeedBtn.Style.CornerRadius = 3
	randomSeedBtn.OnClick = func() {
		seedInput.SetText("Random")
	}
	creationPanel.AddChild(randomSeedBtn.UIComponent)

	// Game mode selection
	modeLabel := NewLabel("creation_mode_label", "Game Mode:", matrix.NewVec2(30, 180), matrix.NewVec2(120, 25))
	modeLabel.Style.ForegroundColor = matrix.ColorWhite()
	modeLabel.Style.FontSize = 16
	creationPanel.AddChild(modeLabel.UIComponent)

	modes := []string{"Survival", "Creative", "Adventure"}
	var modeButtons []*Button
	selectedMode := 0 // Survival default

	for i, mode := range modes {
		modeBtn := NewButton("creation_mode_"+mode, mode, matrix.NewVec2(float32(160+i*110), 180), matrix.NewVec2(100, 30))
		if i == 0 {
			modeBtn.Style.BackgroundColor = matrix.NewColor(0, 150/255.0, 0, 1.0) // Selected
		} else {
			modeBtn.Style.BackgroundColor = matrix.NewColor(60/255.0, 60/255.0, 60/255.0, 1.0)
		}
		modeBtn.Style.ForegroundColor = matrix.ColorWhite()
		modeBtn.Style.CornerRadius = 3

		modeIndex := i
		modeBtn.OnClick = func() {
			selectedMode = modeIndex
			for j, btn := range modeButtons {
				if j == modeIndex {
					btn.Style.BackgroundColor = matrix.NewColor(0, 150/255.0, 0, 1.0)
				} else {
					btn.Style.BackgroundColor = matrix.NewColor(60/255.0, 60/255.0, 60/255.0, 1.0)
				}
			}
		}
		modeButtons = append(modeButtons, modeBtn)
		creationPanel.AddChild(modeBtn.UIComponent)
	}

	// Store selected mode for use when creating world
	_ = selectedMode // Used in closure above

	// Difficulty selection
	diffLabel := NewLabel("creation_diff_label", "Difficulty:", matrix.NewVec2(30, 230), matrix.NewVec2(120, 25))
	diffLabel.Style.ForegroundColor = matrix.ColorWhite()
	diffLabel.Style.FontSize = 16
	creationPanel.AddChild(diffLabel.UIComponent)

	difficulties := []string{"Peaceful", "Easy", "Normal", "Hard"}
	var diffButtons []*Button
	selectedDiff := 2 // Normal default

	for i, diff := range difficulties {
		diffBtn := NewButton("creation_diff_"+diff, diff, matrix.NewVec2(float32(160+i*85), 230), matrix.NewVec2(80, 30))
		if i == 2 {
			diffBtn.Style.BackgroundColor = matrix.NewColor(0, 150/255.0, 0, 1.0) // Selected
		} else {
			diffBtn.Style.BackgroundColor = matrix.NewColor(60/255.0, 60/255.0, 60/255.0, 1.0)
		}
		diffBtn.Style.ForegroundColor = matrix.ColorWhite()
		diffBtn.Style.CornerRadius = 3

		diffIndex := i
		diffBtn.OnClick = func() {
			selectedDiff = diffIndex
			for j, btn := range diffButtons {
				if j == diffIndex {
					btn.Style.BackgroundColor = matrix.NewColor(0, 150/255.0, 0, 1.0)
				} else {
					btn.Style.BackgroundColor = matrix.NewColor(60/255.0, 60/255.0, 60/255.0, 1.0)
				}
			}
		}
		diffButtons = append(diffButtons, diffBtn)
		creationPanel.AddChild(diffBtn.UIComponent)
	}

	// Store selected difficulty for use when creating world
	_ = selectedDiff // Used in closure above

	// More options
	cheatsLabel := NewLabel("creation_cheats_label", "Allow Cheats:", matrix.NewVec2(30, 280), matrix.NewVec2(120, 25))
	cheatsLabel.Style.ForegroundColor = matrix.ColorWhite()
	cheatsLabel.Style.FontSize = 16
	creationPanel.AddChild(cheatsLabel.UIComponent)

	cheatsBtn := NewButton("creation_cheats", "OFF", matrix.NewVec2(160, 280), matrix.NewVec2(80, 30))
	cheatsBtn.Style.BackgroundColor = matrix.NewColor(60/255.0, 60/255.0, 60/255.0, 1.0)
	cheatsBtn.Style.ForegroundColor = matrix.ColorWhite()
	cheatsBtn.Style.CornerRadius = 3
	cheatsEnabled := false
	cheatsBtn.OnClick = func() {
		cheatsEnabled = !cheatsEnabled
		if cheatsEnabled {
			cheatsBtn.SetText("ON")
			cheatsBtn.Style.BackgroundColor = matrix.NewColor(0, 150/255.0, 0, 1.0)
		} else {
			cheatsBtn.SetText("OFF")
			cheatsBtn.Style.BackgroundColor = matrix.NewColor(60/255.0, 60/255.0, 60/255.0, 1.0)
		}
	}
	creationPanel.AddChild(cheatsBtn.UIComponent)

	// Create World button
	createBtn := NewButton("creation_create", "Create World", matrix.NewVec2(150, 380), matrix.NewVec2(200, 50))
	createBtn.Style.BackgroundColor = matrix.NewColor(0, 150/255.0, 0, 1.0)
	createBtn.Style.ForegroundColor = matrix.ColorWhite()
	createBtn.Style.CornerRadius = 5
	createBtn.OnClick = func() {
		// Get world name
		worldName := nameInput.GetText()
		if worldName == "" {
			worldName = "New World"
		}

		// Check if world already exists
		if world.WorldExists(worldName) {
			// Show error or append number
			worldName = worldName + " (1)"
		}

		// Generate seed
		var seed int64
		if seedInput.GetText() == "Random" {
			seed = time.Now().UnixNano()
		} else {
			// Parse seed from text (simplified)
			seed = time.Now().UnixNano()
		}

		// Hide creation panel and start game
		creationPanel.SetVisible(false)

		if mm.onStartGame != nil {
			mm.onStartGame(worldName, seed)
		}
	}
	creationPanel.AddChild(createBtn.UIComponent)

	// Cancel button
	cancelBtn := NewButton("creation_cancel", "Cancel", matrix.NewVec2(20, 430), matrix.NewVec2(100, 40))
	cancelBtn.Style.BackgroundColor = matrix.NewColor(100/255.0, 100/255.0, 100/255.0, 1.0)
	cancelBtn.Style.ForegroundColor = matrix.ColorWhite()
	cancelBtn.Style.CornerRadius = 5
	cancelBtn.OnClick = func() {
		creationPanel.SetVisible(false)
		mm.ShowMenu(MenuTypeWorldSelect)
	}
	creationPanel.AddChild(cancelBtn.UIComponent)

	// Add to world select menu
	if worldSelectMenu, exists := mm.menus[MenuTypeWorldSelect]; exists {
		worldSelectMenu.Panel.AddChild(creationPanel.UIComponent)
	}
}

func (mm *MenuManager) showWorldSelect() {
	// Create world select menu if it doesn't exist (recreate to refresh list)
	mm.createWorldSelectMenu()
	mm.ShowMenu(MenuTypeWorldSelect)
}

// createWorldSelectMenu creates the world selection menu
func (mm *MenuManager) createWorldSelectMenu() {
	screenWidth := float32(1920)
	screenHeight := float32(1080)

	worldSelectMenu := &Menu{
		ID:    "world_select_menu",
		Type:  MenuTypeWorldSelect,
		Title: "Select World",
		Panel: NewPanel("world_select_panel", matrix.NewVec2(screenWidth/2-350, screenHeight/2-300), matrix.NewVec2(700, 600)),
	}
	worldSelectMenu.Panel.Style.BackgroundColor = matrix.NewColor(0, 0, 0, 220/255.0)
	worldSelectMenu.Panel.Style.CornerRadius = 10

	// Title
	titleLabel := NewLabel("world_select_title", "Select World", matrix.NewVec2(280, 20), matrix.NewVec2(140, 40))
	titleLabel.Style.FontSize = 28
	titleLabel.Style.ForegroundColor = matrix.ColorWhite()
	worldSelectMenu.Panel.AddChild(titleLabel.UIComponent)

	// World list panel
	listPanel := NewPanel("world_list_panel", matrix.NewVec2(20, 80), matrix.NewVec2(660, 380))
	listPanel.Style.BackgroundColor = matrix.NewColor(30/255.0, 30/255.0, 30/255.0, 1.0)
	listPanel.Style.CornerRadius = 5
	worldSelectMenu.Panel.AddChild(listPanel.UIComponent)

	// Get world list
	worlds, err := mm.getWorldList()
	if err != nil {
		worlds = []worldInfo{}
	}

	if len(worlds) == 0 {
		// Empty state
		emptyLabel := NewLabel("world_empty", "No worlds found", matrix.NewVec2(230, 170), matrix.NewVec2(200, 30))
		emptyLabel.Style.ForegroundColor = matrix.NewColor(150/255.0, 150/255.0, 150/255.0, 1.0)
		emptyLabel.Style.FontSize = 18
		listPanel.AddChild(emptyLabel.UIComponent)

		hintLabel := NewLabel("world_hint", "Create a new world to get started", matrix.NewVec2(180, 200), matrix.NewVec2(300, 20))
		hintLabel.Style.ForegroundColor = matrix.NewColor(100/255.0, 100/255.0, 100/255.0, 1.0)
		hintLabel.Style.FontSize = 14
		listPanel.AddChild(hintLabel.UIComponent)
	} else {
		// World list items
		for i, world := range worlds {
			if i >= 6 { // Show max 6 worlds
				break
			}
			mm.createWorldListItem(listPanel, world, i)
		}
	}

	// Create New World button
	createBtn := NewButton("world_create_new", "Create New World", matrix.NewVec2(20, 480), matrix.NewVec2(200, 50))
	createBtn.Style.BackgroundColor = matrix.NewColor(0, 150/255.0, 0, 1.0)
	createBtn.Style.ForegroundColor = matrix.ColorWhite()
	createBtn.Style.CornerRadius = 5
	createBtn.OnClick = func() {
		mm.showWorldCreation()
	}
	worldSelectMenu.Panel.AddChild(createBtn.UIComponent)

	// Back button
	backBtn := NewButton("world_select_back", "Back", matrix.NewVec2(300, 480), matrix.NewVec2(100, 50))
	backBtn.Style.BackgroundColor = matrix.NewColor(100/255.0, 100/255.0, 100/255.0, 1.0)
	backBtn.Style.ForegroundColor = matrix.ColorWhite()
	backBtn.Style.CornerRadius = 5
	backBtn.OnClick = func() {
		mm.ShowMenu(MenuTypeMain)
	}
	worldSelectMenu.Panel.AddChild(backBtn.UIComponent)

	mm.menus[MenuTypeWorldSelect] = worldSelectMenu
}

// worldInfo contains world metadata for display
type worldInfo struct {
	name       string
	lastPlayed string
	gameMode   string
	seed       int64
}

// getWorldList retrieves world information from saves directory
func (mm *MenuManager) getWorldList() ([]worldInfo, error) {
	worldNames, err := world.ListWorlds()
	if err != nil {
		return nil, err
	}

	worlds := make([]worldInfo, 0, len(worldNames))
	for _, name := range worldNames {
		saveManager, err := world.NewSaveManager(name)
		if err != nil {
			continue
		}

		info, err := saveManager.LoadWorldInfo()
		saveManager.Close()

		if err != nil {
			// World exists but no info file, use defaults
			worlds = append(worlds, worldInfo{
				name:       name,
				lastPlayed: "Unknown",
				gameMode:   "Survival",
				seed:       0,
			})
		} else {
			worlds = append(worlds, worldInfo{
				name:       info.Name,
				lastPlayed: mm.formatLastPlayed(info.LastPlayed),
				gameMode:   "Survival", // Would come from info in full implementation
				seed:       info.Seed,
			})
		}
	}

	return worlds, nil
}

// formatLastPlayed formats timestamp for display
func (mm *MenuManager) formatLastPlayed(timestamp int64) string {
	if timestamp == 0 {
		return "Never"
	}
	t := time.Unix(timestamp, 0)
	return t.Format("Jan 02, 2006")
}

// createWorldListItem creates a single world list item UI
func (mm *MenuManager) createWorldListItem(parent *Panel, world worldInfo, index int) {
	yPos := float32(10 + index*62)

	// Item panel
	itemPanel := NewPanel("world_item_"+world.name, matrix.NewVec2(10, yPos), matrix.NewVec2(640, 58))
	itemPanel.Style.BackgroundColor = matrix.NewColor(50/255.0, 50/255.0, 50/255.0, 1.0)
	itemPanel.Style.CornerRadius = 3
	parent.AddChild(itemPanel.UIComponent)

	// World name
	nameLabel := NewLabel("world_name_"+world.name, world.name, matrix.NewVec2(15, 8), matrix.NewVec2(300, 24))
	nameLabel.Style.ForegroundColor = matrix.ColorWhite()
	nameLabel.Style.FontSize = 18
	itemPanel.AddChild(nameLabel.UIComponent)

	// World info
	infoText := "Last played: " + world.lastPlayed + " | Seed: " + strconv.Itoa(int(world.seed))
	infoLabel := NewLabel("world_info_"+world.name, infoText, matrix.NewVec2(15, 32), matrix.NewVec2(350, 18))
	infoLabel.Style.ForegroundColor = matrix.NewColor(150/255.0, 150/255.0, 150/255.0, 1.0)
	infoLabel.Style.FontSize = 12
	itemPanel.AddChild(infoLabel.UIComponent)

	// Play button
	playBtn := NewButton("world_play_"+world.name, "Play", matrix.NewVec2(480, 14), matrix.NewVec2(70, 30))
	playBtn.Style.BackgroundColor = matrix.NewColor(0, 150/255.0, 0, 1.0)
	playBtn.Style.ForegroundColor = matrix.ColorWhite()
	playBtn.Style.CornerRadius = 3
	worldName := world.name
	playBtn.OnClick = func() {
		if mm.onLoadGame != nil {
			mm.onLoadGame(worldName)
		}
	}
	itemPanel.AddChild(playBtn.UIComponent)

	// Delete button
	deleteBtn := NewButton("world_delete_"+world.name, "Delete", matrix.NewVec2(560, 14), matrix.NewVec2(70, 30))
	deleteBtn.Style.BackgroundColor = matrix.NewColor(150/255.0, 0, 0, 1.0)
	deleteBtn.Style.ForegroundColor = matrix.ColorWhite()
	deleteBtn.Style.CornerRadius = 3
	deleteBtn.OnClick = func() {
		mm.deleteWorld(worldName)
	}
	itemPanel.AddChild(deleteBtn.UIComponent)
}

// deleteWorld deletes a world and refreshes the list
func (mm *MenuManager) deleteWorld(name string) {
	if err := world.DeleteWorld(name); err == nil {
		// Refresh the world list
		mm.createWorldSelectMenu()
		mm.ShowMenu(MenuTypeWorldSelect)
	}
}

func (mm *MenuManager) showSettings() {
	// Create settings menu if it doesn't exist
	if _, exists := mm.menus[MenuTypeSettings]; !exists {
		mm.createSettingsMenu()
	}
	mm.ShowMenu(MenuTypeSettings)
}

// createSettingsMenu creates the settings menu
func (mm *MenuManager) createSettingsMenu() {
	screenWidth := float32(1920)
	screenHeight := float32(1080)

	settingsMenu := &Menu{
		ID:    "settings_menu",
		Type:  MenuTypeSettings,
		Title: "Settings",
		Panel: NewPanel("settings_panel", matrix.NewVec2(screenWidth/2-300, screenHeight/2-250), matrix.NewVec2(600, 500)),
	}
	settingsMenu.Panel.Style.BackgroundColor = matrix.NewColor(0, 0, 0, 220/255.0)
	settingsMenu.Panel.Style.CornerRadius = 10

	// Settings title
	titleLabel := NewLabel("settings_title", "Settings", matrix.NewVec2(250, 20), matrix.NewVec2(100, 40))
	titleLabel.Style.FontSize = 28
	titleLabel.Style.ForegroundColor = matrix.ColorWhite()
	settingsMenu.Panel.AddChild(titleLabel.UIComponent)

	// Category tabs
	categories := []string{"Video", "Audio", "Controls", "Gameplay"}
	var categoryButtons []*Button
	var categoryPanels []*Panel

	for i, category := range categories {
		btn := NewButton("settings_tab_"+category, category, matrix.NewVec2(float32(20+i*140), 80), matrix.NewVec2(130, 40))
		btn.Style.BackgroundColor = matrix.NewColor(80/255.0, 80/255.0, 80/255.0, 1.0)
		btn.Style.ForegroundColor = matrix.ColorWhite()
		btn.Style.CornerRadius = 5
		settingsMenu.Panel.AddChild(btn.UIComponent)
		categoryButtons = append(categoryButtons, btn)

		// Create category panel
		panel := NewPanel("settings_panel_"+category, matrix.NewVec2(20, 130), matrix.NewVec2(560, 300))
		panel.Style.BackgroundColor = matrix.NewColor(30/255.0, 30/255.0, 30/255.0, 1.0)
		panel.Style.CornerRadius = 5
		panel.SetVisible(i == 0) // Show first category by default
		settingsMenu.Panel.AddChild(panel.UIComponent)
		categoryPanels = append(categoryPanels, panel)

		// Add click handler for tab switching
		catIndex := i
		btn.OnClick = func() {
			// Highlight active tab
			for j, b := range categoryButtons {
				if j == catIndex {
					b.Style.BackgroundColor = matrix.NewColor(100/255.0, 100/255.0, 100/255.0, 1.0)
				} else {
					b.Style.BackgroundColor = matrix.NewColor(80/255.0, 80/255.0, 80/255.0, 1.0)
				}
			}
			// Show selected panel, hide others
			for j, p := range categoryPanels {
				p.SetVisible(j == catIndex)
			}
		}
	}

	// Populate Video settings
	mm.createVideoSettings(categoryPanels[0])
	// Populate Audio settings
	mm.createAudioSettings(categoryPanels[1])
	// Populate Controls settings
	mm.createControlsSettings(categoryPanels[2])
	// Populate Gameplay settings
	mm.createGameplaySettings(categoryPanels[3])

	// Back button
	backBtn := NewButton("settings_back", "Back", matrix.NewVec2(250, 445), matrix.NewVec2(100, 40))
	backBtn.Style.BackgroundColor = matrix.NewColor(100/255.0, 100/255.0, 100/255.0, 1.0)
	backBtn.Style.ForegroundColor = matrix.ColorWhite()
	backBtn.Style.CornerRadius = 5
	backBtn.OnClick = func() {
		mm.HideMenu()
		if mm.onResumeGame != nil {
			mm.onResumeGame()
		}
	}
	settingsMenu.Panel.AddChild(backBtn.UIComponent)

	mm.menus[MenuTypeSettings] = settingsMenu
}

// createVideoSettings creates video settings UI
func (mm *MenuManager) createVideoSettings(panel *Panel) {
	settings := []struct {
		label string
		value string
	}{
		{"Resolution", "1920x1080"},
		{"Fullscreen", "Off"},
		{"VSync", "On"},
		{"Render Distance", "8 chunks"},
		{"FOV", "70"},
		{"Brightness", "100%"},
	}

	for i, setting := range settings {
		yPos := float32(20 + i*40)

		label := NewLabel("video_label_"+setting.label, setting.label, matrix.NewVec2(20, yPos), matrix.NewVec2(200, 30))
		label.Style.ForegroundColor = matrix.ColorWhite()
		label.Style.FontSize = 16
		panel.AddChild(label.UIComponent)

		valueBtn := NewButton("video_value_"+setting.label, setting.value, matrix.NewVec2(300, yPos), matrix.NewVec2(200, 30))
		valueBtn.Style.BackgroundColor = matrix.NewColor(60/255.0, 60/255.0, 60/255.0, 1.0)
		valueBtn.Style.ForegroundColor = matrix.ColorWhite()
		valueBtn.Style.CornerRadius = 3
		panel.AddChild(valueBtn.UIComponent)
	}
}

// createAudioSettings creates audio settings UI
func (mm *MenuManager) createAudioSettings(panel *Panel) {
	settings := []struct {
		label string
		value string
	}{
		{"Master Volume", "80%"},
		{"Music", "60%"},
		{"Sound Effects", "90%"},
		{"Ambient", "70%"},
		{"UI Sounds", "100%"},
	}

	for i, setting := range settings {
		yPos := float32(20 + i*45)

		label := NewLabel("audio_label_"+setting.label, setting.label, matrix.NewVec2(20, yPos), matrix.NewVec2(200, 30))
		label.Style.ForegroundColor = matrix.ColorWhite()
		label.Style.FontSize = 16
		panel.AddChild(label.UIComponent)

		valueBtn := NewButton("audio_value_"+setting.label, setting.value, matrix.NewVec2(300, yPos), matrix.NewVec2(200, 30))
		valueBtn.Style.BackgroundColor = matrix.NewColor(60/255.0, 60/255.0, 60/255.0, 1.0)
		valueBtn.Style.ForegroundColor = matrix.ColorWhite()
		valueBtn.Style.CornerRadius = 3
		panel.AddChild(valueBtn.UIComponent)
	}
}

// createControlsSettings creates controls settings UI
func (mm *MenuManager) createControlsSettings(panel *Panel) {
	settings := []struct {
		action string
		key    string
	}{
		{"Move Forward", "W"},
		{"Move Backward", "S"},
		{"Move Left", "A"},
		{"Move Right", "D"},
		{"Jump", "Space"},
		{"Sneak", "Shift"},
		{"Sprint", "Ctrl"},
		{"Inventory", "E"},
		{"Drop Item", "Q"},
		{"Attack", "LMB"},
		{"Use", "RMB"},
	}

	for i, setting := range settings {
		yPos := float32(15 + i*28)

		label := NewLabel("controls_label_"+setting.action, setting.action, matrix.NewVec2(20, yPos), matrix.NewVec2(200, 24))
		label.Style.ForegroundColor = matrix.ColorWhite()
		label.Style.FontSize = 14
		panel.AddChild(label.UIComponent)

		keyBtn := NewButton("controls_key_"+setting.action, setting.key, matrix.NewVec2(300, yPos), matrix.NewVec2(100, 24))
		keyBtn.Style.BackgroundColor = matrix.NewColor(60/255.0, 60/255.0, 60/255.0, 1.0)
		keyBtn.Style.ForegroundColor = matrix.ColorWhite()
		keyBtn.Style.CornerRadius = 3
		panel.AddChild(keyBtn.UIComponent)
	}
}

// createGameplaySettings creates gameplay settings UI
func (mm *MenuManager) createGameplaySettings(panel *Panel) {
	settings := []struct {
		label string
		value string
	}{
		{"Difficulty", "Normal"},
		{"Game Mode", "Survival"},
		{"Auto-Save", "On"},
		{"Show Coordinates", "Off"},
		{"Fireworks on Level Up", "On"},
	}

	for i, setting := range settings {
		yPos := float32(20 + i*50)

		label := NewLabel("gameplay_label_"+setting.label, setting.label, matrix.NewVec2(20, yPos), matrix.NewVec2(200, 30))
		label.Style.ForegroundColor = matrix.ColorWhite()
		label.Style.FontSize = 16
		panel.AddChild(label.UIComponent)

		valueBtn := NewButton("gameplay_value_"+setting.label, setting.value, matrix.NewVec2(300, yPos), matrix.NewVec2(200, 30))
		valueBtn.Style.BackgroundColor = matrix.NewColor(60/255.0, 60/255.0, 60/255.0, 1.0)
		valueBtn.Style.ForegroundColor = matrix.ColorWhite()
		valueBtn.Style.CornerRadius = 3
		panel.AddChild(valueBtn.UIComponent)
	}
}

func (mm *MenuManager) resumeGame() {
	mm.HideMenu()
	if mm.onResumeGame != nil {
		mm.onResumeGame()
	}
}

func (mm *MenuManager) saveGame() {
	if mm.onSaveGame != nil {
		mm.onSaveGame()
	}
}

func (mm *MenuManager) loadWorld(worldName string) {
	if mm.onLoadGame != nil {
		mm.onLoadGame(worldName)
	}
}

func (mm *MenuManager) quitToMenu() {
	if mm.onQuitToMenu != nil {
		mm.onQuitToMenu()
	}
	mm.ShowMenu(MenuTypeMain)
}

func (mm *MenuManager) quitToDesktop() {
	if mm.onQuitToDesktop != nil {
		mm.onQuitToDesktop()
	}
}

// GetLoginScreen returns the login screen
func (mm *MenuManager) GetLoginScreen() *LoginScreen {
	if _, exists := mm.menus[MenuTypeLogin]; exists {
		// Type assert to LoginScreen - this is a simplified approach
		// In a full implementation, you'd store the actual screen objects
		_ = exists // Use blank identifier to avoid unused variable error
		return nil // Placeholder - would return actual LoginScreen
	}
	return nil
}

// GetGameSelectScreen returns the game selection screen
func (mm *MenuManager) GetGameSelectScreen() *GameSelectScreen {
	if _, exists := mm.menus[MenuTypeGameSelect]; exists {
		// Type assert to GameSelectScreen
		_ = exists // Use blank identifier to avoid unused variable error
		return nil // Placeholder - would return actual GameSelectScreen
	}
	return nil
}

// ShowLoginScreen shows the login screen
func (mm *MenuManager) ShowLoginScreen() {
	mm.ShowMenu(MenuTypeLogin)
}

// ShowGameSelectScreen shows the game selection screen
func (mm *MenuManager) ShowGameSelectScreen() {
	mm.ShowMenu(MenuTypeGameSelect)
}

// SetAuthenticatedUser sets the authenticated user information
func (mm *MenuManager) SetAuthenticatedUser(user *GitHubUser) {
	if loginScreen := mm.GetLoginScreen(); loginScreen != nil {
		// TODO: Implement SetUser method in LoginScreen
		// loginScreen.SetUser(user)
	}
}

// HandleAuthenticationSuccess handles successful authentication
func (mm *MenuManager) HandleAuthenticationSuccess(user *GitHubUser) {
	mm.SetAuthenticatedUser(user)
	mm.ShowGameSelectScreen()
}

// HandleSignOut handles user sign out
func (mm *MenuManager) HandleSignOut() {
	mm.SetAuthenticatedUser(nil)
	mm.ShowLoginScreen()
}

// EnableFyneUI enables Fyne UI integration
func (mm *MenuManager) EnableFyneUI(host *engine.Host) {
	if mm.fyneBridge == nil {
		mm.fyneBridge = NewFyneKaijuBridge(host, mm)
	}
	mm.useFyne = true
}

// DisableFyneUI disables Fyne UI integration
func (mm *MenuManager) DisableFyneUI() {
	mm.useFyne = false
	if mm.fyneBridge != nil {
		mm.fyneBridge.HideFyne()
	}
}

// ShowFyneLogin displays Fyne login screen
func (mm *MenuManager) ShowFyneLogin() {
	if mm.fyneBridge != nil {
		mm.fyneBridge.ShowFyneLogin()
	}
}

// IsFyneActive returns true if Fyne UI is currently active
func (mm *MenuManager) IsFyneActive() bool {
	if mm.fyneBridge != nil {
		return mm.fyneBridge.IsFyneActive()
	}
	return false
}
