package player

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tesselstudio/TesselBox/pkg/crafting"
	"kaijuengine.com/matrix"
)

// PlayerData represents all serialized player state
type PlayerData struct {
	// Version for migration support
	Version int `json:"version"`

	// Position
	PosX float32 `json:"pos_x"`
	PosY float32 `json:"pos_y"`
	PosZ float32 `json:"pos_z"`

	// Rotation
	RotX float32 `json:"rot_x"`
	RotY float32 `json:"rot_y"`
	RotZ float32 `json:"rot_z"`

	// Stats
	Health      float32 `json:"health"`
	MaxHealth   float32 `json:"max_health"`
	Hunger      float32 `json:"hunger"`
	MaxHunger   float32 `json:"max_hunger"`
	Stamina     float32 `json:"stamina"`
	MaxStamina  float32 `json:"max_stamina"`
	Saturation  float32 `json:"saturation"`
	Level       int32   `json:"level"`
	Experience  int32   `json:"experience"`
	SkillPoints int32   `json:"skill_points"`

	// Inventory
	InventorySlots []InventorySlotData `json:"inventory_slots"`
	HotbarSlot     int                 `json:"hotbar_slot"`

	// Timestamps
	LastSaved time.Time `json:"last_saved"`
}

// InventorySlotData represents a single inventory slot
type InventorySlotData struct {
	Slot     int    `json:"slot"`
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

// CurrentDataVersion is the current save format version
const CurrentDataVersion = 1

// SavePlayer saves player data to a file
func SavePlayer(player *Player, worldName string) error {
	data := SerializePlayer(player)

	// Create saves directory if needed
	saveDir := filepath.Join("saves", worldName)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return fmt.Errorf("failed to create save directory: %w", err)
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal player data: %w", err)
	}

	// Write to temp file first, then rename for atomicity
	playerPath := filepath.Join(saveDir, "player.json")
	tempPath := playerPath + ".tmp"

	if err := os.WriteFile(tempPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write player data: %w", err)
	}

	if err := os.Rename(tempPath, playerPath); err != nil {
		return fmt.Errorf("failed to finalize player save: %w", err)
	}

	return nil
}

// LoadPlayer loads player data from a file and applies it to the player
func LoadPlayer(player *Player, worldName string) error {
	playerPath := filepath.Join("saves", worldName, "player.json")

	// Check if file exists
	if _, err := os.Stat(playerPath); os.IsNotExist(err) {
		// No saved player data, use defaults
		return nil
	}

	// Read file
	jsonData, err := os.ReadFile(playerPath)
	if err != nil {
		return fmt.Errorf("failed to read player data: %w", err)
	}

	// Parse JSON
	var data PlayerData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal player data: %w", err)
	}

	// Apply to player
	DeserializePlayer(player, &data)

	return nil
}

// SerializePlayer converts a Player to PlayerData
func SerializePlayer(player *Player) *PlayerData {
	player.mu.RLock()
	defer player.mu.RUnlock()

	data := &PlayerData{
		Version:    CurrentDataVersion,
		PosX:       player.position.X(),
		PosY:       player.position.Y(),
		PosZ:       player.position.Z(),
		RotX:       player.rotation.X(),
		RotY:       player.rotation.Y(),
		RotZ:       player.rotation.Z(),
		HotbarSlot: player.hotbarSlot,
		LastSaved:  time.Now(),
	}

	// Serialize stats
	if player.stats != nil {
		data.Health = player.stats.Health.GetCurrentHealth()
		data.MaxHealth = player.stats.Health.GetMaxHealth()
		data.Hunger = player.stats.Hunger.GetCurrentHunger()
		data.MaxHunger = player.stats.Hunger.GetMaxHunger()
		data.Saturation = player.stats.Hunger.GetSaturation()
		data.Stamina = player.stats.Stamina.GetCurrentStamina()
		data.MaxStamina = player.stats.Stamina.GetMaxStamina()
		data.Level = player.stats.Experience.GetLevel()
		data.Experience = player.stats.Experience.GetCurrentExperience()
		data.SkillPoints = player.stats.Experience.GetSkillPoints()
	}

	// Serialize inventory
	if player.inventory != nil {
		data.InventorySlots = serializeInventory(player.inventory)
	}

	return data
}

// DeserializePlayer applies PlayerData to a Player
func DeserializePlayer(player *Player, data *PlayerData) {
	player.mu.Lock()
	defer player.mu.Unlock()

	// Restore position
	player.position = matrix.NewVec3(data.PosX, data.PosY, data.PosZ)
	if player.entity != nil {
		player.entity.Transform.SetPosition(player.position)
	}

	// Restore rotation
	player.rotation = matrix.NewVec3(data.RotX, data.RotY, data.RotZ)

	// Restore hotbar slot
	if data.HotbarSlot >= 0 && data.HotbarSlot < 9 {
		player.hotbarSlot = data.HotbarSlot
	}

	// Restore stats
	if player.stats != nil && data.MaxHealth > 0 {
		player.stats.Health.SetMaxHealth(data.MaxHealth)
		// Heal to restore current health
		healthDiff := data.Health - player.stats.Health.GetCurrentHealth()
		if healthDiff > 0 {
			player.stats.Health.Heal(healthDiff)
		}
		// Note: Hunger and Stamina don't have setters, they use Eat() and Reset() patterns
		// For now, we just restore what we can
	}

	// Restore inventory
	if player.inventory != nil && len(data.InventorySlots) > 0 {
		deserializeInventory(player.inventory, data.InventorySlots)
	}
}

// serializeInventory converts inventory to serializable format
func serializeInventory(inv *crafting.Inventory) []InventorySlotData {
	var slots []InventorySlotData

	// Get all items from inventory
	items := inv.GetItems()
	for _, stack := range items {
		if stack != nil && !stack.IsEmpty() {
			// Find the slot index for this item
			// We need to iterate to find which slot this is in
			for i := 0; i < 36; i++ {
				slotStack := inv.GetSlot(i)
				if slotStack == stack {
					slots = append(slots, InventorySlotData{
						Slot:     i,
						ItemID:   stack.Item.ID,
						Quantity: stack.Quantity,
					})
					break
				}
			}
		}
	}

	return slots
}

// deserializeInventory restores inventory from saved data
func deserializeInventory(inv *crafting.Inventory, slots []InventorySlotData) {
	// Clear existing inventory
	inv.Clear()

	// Get item registry for lookups
	registry := crafting.GetGlobalItemRegistry()

	// Restore slots
	for _, slotData := range slots {
		if slotData.Slot < 0 || slotData.Slot >= 36 {
			continue
		}

		item, exists := registry.GetItem(slotData.ItemID)
		if !exists {
			continue // Item no longer exists in registry
		}

		stack := crafting.NewItemStack(item, slotData.Quantity)
		inv.SetSlot(slotData.Slot, stack)
	}
}

// PlayerSaveExists checks if player save data exists for a world
func PlayerSaveExists(worldName string) bool {
	playerPath := filepath.Join("saves", worldName, "player.json")
	_, err := os.Stat(playerPath)
	return !os.IsNotExist(err)
}

// DeletePlayerSave removes player save data for a world
func DeletePlayerSave(worldName string) error {
	playerPath := filepath.Join("saves", worldName, "player.json")
	if err := os.Remove(playerPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
