package ui

import (
	"sync"

	"kaijuengine.com/matrix"

	"github.com/tesselstudio/TesselBox/pkg/crafting"
	"github.com/tesselstudio/TesselBox/pkg/survival"
)

// HUD represents the heads-up display
type HUD struct {
	mu            sync.RWMutex
	visible       bool
	rootPanel     *Panel
	healthBar     *ProgressBar
	hungerBar     *ProgressBar
	staminaBar    *ProgressBar
	experienceBar *ProgressBar
	hotbar        *Hotbar
	crosshair     *Crosshair
	debugInfo     *DebugInfo
	playerStats   *survival.PlayerStats
	inventory     *crafting.Inventory
}

// NewHUD creates a new HUD
func NewHUD(screenWidth, screenHeight float32) *HUD {
	hud := &HUD{
		visible: true,
	}

	hud.createComponents(screenWidth, screenHeight)

	return hud
}

// createComponents creates all HUD components
func (h *HUD) createComponents(screenWidth, screenHeight float32) {
	// Root panel
	h.rootPanel = NewPanel("hud_root", matrix.NewVec2(0, 0), matrix.NewVec2(screenWidth, screenHeight))
	h.rootPanel.Style.BackgroundColor = matrix.ColorTransparent()

	// Health bar (top left)
	h.healthBar = NewProgressBar("health_bar", matrix.NewVec2(20, 20), matrix.NewVec2(200, 20))
	h.healthBar.Style.ForegroundColor = matrix.ColorRed()
	h.healthBar.Style.BackgroundColor = matrix.NewColor(50/255.0, 0, 0, 200/255.0)
	h.healthBar.SetShowText(false)
	h.rootPanel.AddChild(h.healthBar.UIComponent)

	// Hunger bar (below health)
	h.hungerBar = NewProgressBar("hunger_bar", matrix.NewVec2(20, 45), matrix.NewVec2(200, 20))
	h.hungerBar.Style.ForegroundColor = matrix.ColorOrange()
	h.hungerBar.Style.BackgroundColor = matrix.NewColor(50/255.0, 25/255.0, 0, 200/255.0)
	h.hungerBar.SetShowText(false)
	h.rootPanel.AddChild(h.hungerBar.UIComponent)

	// Stamina bar (below hunger)
	h.staminaBar = NewProgressBar("stamina_bar", matrix.NewVec2(20, 70), matrix.NewVec2(200, 20))
	h.staminaBar.Style.ForegroundColor = matrix.ColorYellow()
	h.staminaBar.Style.BackgroundColor = matrix.NewColor(50/255.0, 50/255.0, 0, 200/255.0)
	h.staminaBar.SetShowText(false)
	h.rootPanel.AddChild(h.staminaBar.UIComponent)

	// Experience bar (bottom center)
	h.experienceBar = NewProgressBar("experience_bar", matrix.NewVec2(screenWidth/2-100, screenHeight-30), matrix.NewVec2(200, 20))
	h.experienceBar.Style.ForegroundColor = matrix.ColorGreen()
	h.experienceBar.Style.BackgroundColor = matrix.NewColor(0, 50/255.0, 0, 200/255.0)
	h.experienceBar.SetShowText(true)
	h.rootPanel.AddChild(h.experienceBar.UIComponent)

	// Hotbar (bottom center, above experience)
	h.hotbar = NewHotbar("hotbar", matrix.NewVec2(screenWidth/2-200, screenHeight-80), 10)
	h.rootPanel.AddChild(h.hotbar.Panel.UIComponent)

	// Crosshair (center)
	h.crosshair = NewCrosshair("crosshair", matrix.NewVec2(screenWidth/2-10, screenHeight/2-10))
	h.rootPanel.AddChild(h.crosshair.Panel.UIComponent)

	// Debug info (top right)
	h.debugInfo = NewDebugInfo("debug_info", matrix.NewVec2(screenWidth-300, 20))
	h.debugInfo.SetVisible(false) // Hidden by default
	h.rootPanel.AddChild(h.debugInfo.Panel.UIComponent)
}

// SetPlayerStats sets the player stats reference
func (h *HUD) SetPlayerStats(stats *survival.PlayerStats) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.playerStats = stats
}

// SetInventory sets the inventory reference
func (h *HUD) SetInventory(inventory *crafting.Inventory) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.inventory = inventory
	h.hotbar.SetInventory(inventory)
}

// Update updates the HUD with current stats
func (h *HUD) Update(deltaTime float32) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.visible || h.playerStats == nil {
		return
	}

	// Update health bar
	h.healthBar.SetValue(h.playerStats.Health.GetHealthPercentage())

	// Update hunger bar
	h.hungerBar.SetValue(h.playerStats.Hunger.GetHungerPercentage())

	// Update stamina bar
	h.staminaBar.SetValue(h.playerStats.Stamina.GetStaminaPercentage())

	// Update experience bar
	if h.playerStats.Experience.GetExperienceToNext() > 0 {
		expPercentage := float32(h.playerStats.Experience.GetCurrentExperience()) / float32(h.playerStats.Experience.GetExperienceToNext())
		h.experienceBar.SetValue(expPercentage)
	}

	// Update hotbar
	if h.hotbar != nil {
		h.hotbar.Update()
	}

	// Update debug info
	if h.debugInfo != nil && h.debugInfo.IsVisible() {
		h.debugInfo.Update()
	}
}

// SetVisible sets HUD visibility
func (h *HUD) SetVisible(visible bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.visible = visible
	h.rootPanel.SetVisible(visible)
}

// IsVisible returns HUD visibility
func (h *HUD) IsVisible() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.visible
}

// GetRootPanel returns the root panel for rendering
func (h *HUD) GetRootPanel() *Panel {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.rootPanel
}

// GetHotbar returns the hotbar for slot selection
func (h *HUD) GetHotbar() *Hotbar {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.hotbar
}

// ToggleDebug toggles debug information display
func (h *HUD) ToggleDebug() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.debugInfo != nil {
		h.debugInfo.SetVisible(!h.debugInfo.IsVisible())
	}
}

// Hotbar represents the hotbar UI
type Hotbar struct {
	*Panel
	inventory    *crafting.Inventory
	slots        []*HotbarSlot
	selectedSlot int
	slotCount    int
}

// NewHotbar creates a new hotbar
func NewHotbar(id string, position matrix.Vec2, slotCount int) *Hotbar {
	panel := NewPanel(id, position, matrix.NewVec2(float32(slotCount*40+10), 50))
	panel.Style.BackgroundColor = matrix.NewColor(0, 0, 0, 150/255.0)
	panel.Style.CornerRadius = 5

	hotbar := &Hotbar{
		Panel:        panel,
		slots:        make([]*HotbarSlot, slotCount),
		selectedSlot: 0,
		slotCount:    slotCount,
	}

	// Create hotbar slots
	for i := 0; i < slotCount; i++ {
		slot := NewHotbarSlot("hotbar_slot_"+string(rune(i)), matrix.NewVec2(float32(i*40+5), 5))
		hotbar.slots[i] = slot
		hotbar.Panel.AddChild(slot.Panel.UIComponent)
	}

	// Highlight selected slot
	hotbar.updateSelectedSlot()

	return hotbar
}

// SetInventory sets the inventory for the hotbar
func (h *Hotbar) SetInventory(inventory *crafting.Inventory) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.inventory = inventory
}

// Update updates the hotbar display
func (h *Hotbar) Update() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.inventory == nil {
		return
	}

	// Update each slot with inventory items
	for i := 0; i < h.slotCount && i < 10; i++ {
		slot := h.slots[i]
		item := h.inventory.GetSlot(i)

		if item != nil && !item.IsEmpty() {
			slot.SetItem(item)
		} else {
			slot.SetItem(nil)
		}
	}
}

// SelectSlot selects a hotbar slot
func (h *Hotbar) SelectSlot(slot int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if slot >= 0 && slot < h.slotCount {
		h.selectedSlot = slot
		h.updateSelectedSlot()
	}
}

// GetSelectedSlot returns the selected slot index
func (h *Hotbar) GetSelectedSlot() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.selectedSlot
}

// updateSelectedSlot updates the visual selection indicator
func (h *Hotbar) updateSelectedSlot() {
	for i, slot := range h.slots {
		if i == h.selectedSlot {
			slot.Panel.Style.BorderColor = matrix.ColorWhite()
			slot.Panel.Style.BorderWidth = 2
		} else {
			slot.Panel.Style.BorderColor = matrix.ColorTransparent()
			slot.Panel.Style.BorderWidth = 0
		}
	}
}

// HotbarSlot represents a single hotbar slot
type HotbarSlot struct {
	*Panel
	item       *crafting.ItemStack
	icon       *Image
	countLabel *Label
}

// NewHotbarSlot creates a new hotbar slot
func NewHotbarSlot(id string, position matrix.Vec2) *HotbarSlot {
	panel := NewPanel(id, position, matrix.NewVec2(40, 40))
	panel.Style.BackgroundColor = matrix.NewColor(50/255.0, 50/255.0, 50/255.0, 100/255.0)
	panel.Style.CornerRadius = 3

	slot := &HotbarSlot{
		Panel: panel,
	}

	// Item icon
	slot.icon = NewImage(id+"_icon", "", matrix.NewVec2(5, 5), matrix.NewVec2(30, 30))
	panel.AddChild(slot.icon.UIComponent)

	// Count label
	slot.countLabel = NewLabel(id+"_count", "", matrix.NewVec2(2, 22), matrix.NewVec2(36, 16))
	slot.countLabel.Style.ForegroundColor = matrix.ColorWhite()
	slot.countLabel.Style.FontSize = 12
	panel.AddChild(slot.countLabel.UIComponent)

	return slot
}

// SetItem sets the item in the slot
func (hs *HotbarSlot) SetItem(item *crafting.ItemStack) {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	hs.item = item

	if item != nil && !item.IsEmpty() {
		hs.icon.SetTextureID(item.Item.ID)
		hs.icon.SetVisible(true)

		if item.Quantity > 1 {
			hs.countLabel.SetText(string(rune(item.Quantity)))
			hs.countLabel.SetVisible(true)
		} else {
			hs.countLabel.SetVisible(false)
		}
	} else {
		hs.icon.SetVisible(false)
		hs.countLabel.SetVisible(false)
	}
}

// GetItem returns the item in the slot
func (hs *HotbarSlot) GetItem() *crafting.ItemStack {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	return hs.item
}

// Crosshair represents the aiming crosshair
type Crosshair struct {
	*Panel
	horizontal *Panel
	vertical   *Panel
}

// NewCrosshair creates a new crosshair
func NewCrosshair(id string, position matrix.Vec2) *Crosshair {
	panel := NewPanel(id, position, matrix.NewVec2(20, 20))
	panel.Style.BackgroundColor = matrix.ColorTransparent()

	crosshair := &Crosshair{
		Panel: panel,
	}

	// Horizontal line
	crosshair.horizontal = NewPanel(id+"_horizontal", matrix.NewVec2(0, 9), matrix.NewVec2(20, 2))
	crosshair.horizontal.Style.BackgroundColor = matrix.ColorWhite()
	panel.AddChild(crosshair.horizontal.UIComponent)

	// Vertical line
	crosshair.vertical = NewPanel(id+"_vertical", matrix.NewVec2(9, 0), matrix.NewVec2(2, 20))
	crosshair.vertical.Style.BackgroundColor = matrix.ColorWhite()
	panel.AddChild(crosshair.vertical.UIComponent)

	return crosshair
}

// SetColor sets the crosshair color
func (c *Crosshair) SetColor(color matrix.Color) {
	c.horizontal.Style.BackgroundColor = color
	c.vertical.Style.BackgroundColor = color
}

// DebugInfo represents debug information display
type DebugInfo struct {
	*Panel
	labels      []*Label
	fps         float32
	position    matrix.Vec3
	timeOfDay   float32
	temperature float32
}

// NewDebugInfo creates a new debug info panel
func NewDebugInfo(id string, position matrix.Vec2) *DebugInfo {
	panel := NewPanel(id, position, matrix.NewVec2(280, 120))
	panel.Style.BackgroundColor = matrix.NewColor(0, 0, 0, 180/255.0)
	panel.Style.CornerRadius = 3

	debug := &DebugInfo{
		Panel:  panel,
		labels: make([]*Label, 0),
	}

	// Create debug labels
	debugLabels := []string{
		"FPS: 0",
		"Position: (0, 0, 0)",
		"Time: 06:00",
		"Temperature: 20°C",
		"Biome: Plains",
		"Entities: 0",
	}

	for i, text := range debugLabels {
		label := NewLabel(id+"_label_"+string(rune(i)), text, matrix.NewVec2(10, float32(i*20+10)), matrix.NewVec2(260, 16))
		label.Style.ForegroundColor = matrix.ColorWhite()
		label.Style.FontSize = 14
		debug.labels = append(debug.labels, label)
		debug.Panel.AddChild(label.UIComponent)
	}

	return debug
}

// SetFPS sets the FPS display
func (di *DebugInfo) SetFPS(fps float32) {
	di.mu.Lock()
	defer di.mu.Unlock()

	di.fps = fps
	if len(di.labels) > 0 {
		di.labels[0].SetText("FPS: " + string(rune(fps)))
	}
}

// SetPosition sets the position display
func (di *DebugInfo) SetPosition(pos matrix.Vec3) {
	di.mu.Lock()
	defer di.mu.Unlock()

	di.position = pos
	if len(di.labels) > 1 {
		di.labels[1].SetText("Position: (" + string(rune(pos.X())) + ", " + string(rune(pos.Y())) + ", " + string(rune(pos.Z())) + ")")
	}
}

// SetTimeOfDay sets the time display
func (di *DebugInfo) SetTimeOfDay(timeOfDay float32) {
	di.mu.Lock()
	defer di.mu.Unlock()

	di.timeOfDay = timeOfDay
	if len(di.labels) > 2 {
		hours := int(timeOfDay)
		minutes := int((timeOfDay - float32(hours)) * 60)
		di.labels[2].SetText("Time: " + string(rune(hours)) + ":" + string(rune(minutes)))
	}
}

// SetTemperature sets the temperature display
func (di *DebugInfo) SetTemperature(temp float32) {
	di.mu.Lock()
	defer di.mu.Unlock()

	di.temperature = temp
	if len(di.labels) > 3 {
		di.labels[3].SetText("Temperature: " + string(rune(temp)) + "°C")
	}
}

// Update updates all debug information
func (di *DebugInfo) Update() {
	// This would be called with actual game data
	// For now, it's a placeholder
}
