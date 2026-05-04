# TesselBox API Reference

Complete API documentation for all public packages.

## Table of Contents

- [pkg/audio](#pkgaudio) - Audio and sound system
- [pkg/blocks](#pkgblocks) - Block system
- [pkg/crafting](#pkgcrafting) - Crafting and items
- [pkg/game](#pkggame) - Game controller
- [pkg/network](#pkgnetwork) - Multiplayer networking
- [pkg/player](#pkgplayer) - Player entity
- [pkg/survival](#pkgsurvival) - Survival mechanics
- [pkg/ui](#pkgui) - User interface
- [pkg/world](#pkgworld) - World generation and management

---

## pkg/audio

The audio package provides sound effects, background music, and audio management.

### Types

#### SoundID

```go
type SoundID int

const (
    SFXBlockBreak SoundID = iota
    SFXBlockPlace
    SFXFootstep
    SFXUIClick
    SFXUIHover
    SFXItemPickup
    SFXPlayerDamage
    SFXJump
)
```

Sound effect identifiers.

#### GameEvent

```go
type GameEvent int

const (
    EventBlockBreak GameEvent = iota
    EventBlockPlace
    EventFootstep
    EventUIClick
    EventUIHover
    EventItemPickup
    EventPlayerDamage
    EventJump
)
```

Game events that can trigger sounds.

### Manager

```go
import "github.com/tesselstudio/TesselBox/pkg/audio"

// Get the singleton audio manager
manager := audio.GetManager()

// Initialize audio system
err := manager.Initialize()

// Play a sound effect
manager.TriggerGameEvent(audio.EventBlockBreak)

// Play background music
manager.PlayMusic("exploration")

// Control volumes
manager.SetMasterVolume(1.0)
manager.SetSFXVolume(0.8)
manager.SetMusicVolume(0.7)
manager.SetMuted(false)
```

### Positional Audio

```go
// Play sound with 3D positioning
listenerPos := audio.Vec3{X: 0, Y: 0, Z: 0}
soundPos := audio.Vec3{X: 5, Y: 0, Z: 3}

manager.PlaySFXAt(audio.SFXBlockBreak, listenerPos, soundPos)
```

### Configuration

```go
// Load audio config
config, err := audio.LoadConfig("config/audio.json")

// Apply config to manager
config.Apply(manager)

// Save config
config.Save("config/audio.json")
```

---

## pkg/blocks

The blocks package provides hexagonal prism block types, geometry, and placement mechanics.

### Types

#### BlockType

```go
type BlockType struct {
    ID          BlockID
    Name        string
    Solid       bool
    Transparent bool
    Material    MaterialType
    Hardness    float32
    BlastResistance float32
    LightEmission uint8
    HarvestLevel  int
}
```

Represents a type of block in the game.

#### BlockID

```go
type BlockID uint16
```

Unique identifier for block types.

### Block Registry

#### NewBlockRegistry

```go
func NewBlockRegistry() *BlockRegistry
```

Creates a new block registry with all default block types.

**Example:**
```go
registry := blocks.NewBlockRegistry()
stone, exists := registry.GetBlock(blocks.BlockStone)
```

#### BlockRegistry.GetBlock

```go
func (r *BlockRegistry) GetBlock(id BlockID) (BlockType, bool)
```

Retrieves a block type by ID.

#### BlockRegistry.GetAllBlocks

```go
func (r *BlockRegistry) GetAllBlocks() []BlockType
```

Returns all registered block types.

### Geometry

#### CreateHexPrismVertices

```go
func CreateHexPrismVertices(size float32) []matrix.Vec3
```

Generates vertex positions for a hexagonal prism.

**Parameters:**
- `size` - The radius of the hexagon

**Returns:**
- Array of 24 vertex positions (4 vertices per face × 6 faces)

#### CreateHexPrismNormals

```go
func CreateHexPrismNormals() []matrix.Vec3
```

Returns normal vectors for each face of a hexagonal prism.

### Placement

#### ValidatePlacement

```go
func ValidatePlacement(world *world.World, pos world.BlockPos, block BlockType) bool
```

Validates if a block can be placed at the specified position.

**Parameters:**
- `world` - The world instance
- `pos` - Block position
- `block` - Block type to place

**Returns:**
- `true` if placement is valid

---

## pkg/crafting

Crafting system with items, recipes, and inventory management.

### Types

#### Item

```go
type Item struct {
    ID          ItemID
    Name        string
    Description string
    MaxStack    int
    ToolType    ToolType
    ToolLevel   int
    AttackDamage float32
    Durability  int
    IsFood      bool
    FoodValue   float32
    IsBlock     bool
    BlockID     blocks.BlockID
}
```

Represents an item in the game.

#### ItemStack

```go
type ItemStack struct {
    Item  *Item
    Count int
}
```

A stack of items in inventory.

### Item Registry

#### GetGlobalItemRegistry

```go
func GetGlobalItemRegistry() *ItemRegistry
```

Returns the singleton item registry.

#### ItemRegistry.Register

```go
func (r *ItemRegistry) Register(item Item)
```

Registers a new item type.

### Inventory

#### NewInventory

```go
func NewInventory(size int, hotbarSize int) *Inventory
```

Creates a new inventory.

**Parameters:**
- `size` - Total inventory size (typically 36)
- `hotbarSize` - Hotbar slots (typically 9)

#### Inventory.AddItem

```go
func (inv *Inventory) AddItem(stack ItemStack) (remaining ItemStack, success bool)
```

Adds an item stack to the inventory.

**Returns:**
- `remaining` - Items that couldn't be added
- `success` - Whether any items were added

#### Inventory.GetItem

```go
func (inv *Inventory) GetItem(slot int) ItemStack
```

Gets the item in a specific slot.

#### Inventory.SetHotbarSlot

```go
func (inv *Inventory) SetHotbarSlot(slot int)
```

Sets the active hotbar slot.

### Recipes

#### Recipe

```go
type Recipe struct {
    ID       RecipeID
    Name     string
    Inputs   []RecipeInput
    Output   ItemStack
    Shape    [][]rune // nil for shapeless
}
```

A crafting recipe definition.

#### RecipeRegistry.FindMatch

```go
func (r *RecipeRegistry) FindMatch(grid [3][3]ItemStack) (*Recipe, bool)
```

Finds a matching recipe for a 3x3 crafting grid.

---

## pkg/game

Game controller that manages all systems and state transitions.

### GameState

```go
type GameState int

const (
    GameStateLogin GameState = iota
    GameStateMainMenu
    GameStateWorldSelect
    GameStateMultiplayer
    GameStateLoading
    GameStatePlaying
    GameStatePaused
    GameStateInventory
    GameStateCrafting
    GameStateSettings
)
```

### Controller

#### NewController

```go
func NewController(host *engine.Host) *Controller
```

Creates a new game controller.

#### Controller.StartWorld

```go
func (c *Controller) StartWorld(worldName string, seed int64)
```

Starts a new game world.

#### Controller.TransitionTo

```go
func (c *Controller) TransitionTo(newState GameState) bool
```

Transitions to a new game state if valid.

#### Controller.Update

```go
func (c *Controller) Update()
```

Updates the game state (call every frame).

#### Controller.SaveGame

```go
func (c *Controller) saveGame()
```

Saves the current game state.

---

## pkg/network

Multiplayer networking system.

### Types

#### ServerConfig

```go
type ServerConfig struct {
    Host         string
    Port         int
    MaxPlayers   int
    TickRate     int
    WorldSize    int
    ChunkSize    int
    ViewDistance int
}
```

### Server

#### NewServer

```go
func NewServer(config ServerConfig) *Server
```

Creates a new game server.

#### Server.Start

```go
func (s *Server) Start() error
```

Starts the server.

#### Server.Stop

```go
func (s *Server) Stop()
```

Stops the server.

### Client

#### NewClient

```go
func NewClient(host string, port int) *Client
```

Creates a new network client.

#### Client.Connect

```go
func (c *Client) Connect() error
```

Connects to a server.

---

## pkg/player

Player entity and controller.

### Player

#### NewPlayer

```go
func NewPlayer(host *engine.Host, world *world.World) *Player
```

Creates a new player.

#### Player.GetPosition

```go
func (p *Player) GetPosition() matrix.Vec3
```

Gets player position.

#### Player.SetPosition

```go
func (p *Player) SetPosition(pos matrix.Vec3)
```

Sets player position.

#### Player.Move

```go
func (p *Player) Move(forward, right float32, sprint bool)
```

Moves the player based on input.

#### Player.Jump

```go
func (p *Player) Jump()
```

Makes the player jump.

#### Player.BreakBlock

```go
func (p *Player) BreakBlock()
```

Breaks the block the player is looking at.

#### Player.PlaceBlock

```go
func (p *Player) PlaceBlock()
```

Places a block at the target location.

#### Player.GetStats

```go
func (p *Player) GetStats() *survival.PlayerStats
```

Gets player survival stats.

---

## pkg/survival

Survival mechanics including health, hunger, and environment.

### PlayerStats

```go
type PlayerStats struct {
    Health   *HealthSystem
    Hunger   *HungerSystem
    Stamina  *StaminaSystem
    Experience *ExperienceSystem
}
```

#### NewPlayerStats

```go
func NewPlayerStats(maxHealth, maxHunger, maxStamina float32) *PlayerStats
```

Creates player stats with specified maximums.

### HealthSystem

#### HealthSystem.TakeDamage

```go
func (h *HealthSystem) TakeDamage(amount float32, damageType DamageType)
```

Applies damage to the player.

#### HealthSystem.Heal

```go
func (h *HealthSystem) Heal(amount float32)
```

Heals the player.

### HungerSystem

#### HungerSystem.Consume

```go
func (h *HungerSystem) Consume(foodValue float32)
```

Consumes food, restoring hunger.

#### HungerSystem.Update

```go
func (h *HungerSystem) Update(deltaTime float64)
```

Updates hunger (decreases over time).

---

## pkg/ui

User interface components.

### HUD

#### NewHUD

```go
func NewHUD(width, height int) *HUD
```

Creates a new HUD.

#### HUD.SetPlayerStats

```go
func (h *HUD) SetPlayerStats(stats *survival.PlayerStats)
```

Sets the player stats to display.

#### HUD.SetInventory

```go
func (h *HUD) SetInventory(inv *crafting.Inventory)
```

Sets the inventory to display.

#### HUD.ToggleDebug

```go
func (h *HUD) ToggleDebug()
```

Toggles debug information display.

#### HUD.Update

```go
func (h *HUD) Update(deltaTime float32)
```

Updates the HUD.

### Components

#### NewPanel

```go
func NewPanel(id string, position, size matrix.Vec2) *Panel
```

Creates a new UI panel.

#### NewButton

```go
func NewButton(id, text string, position, size matrix.Vec2) *Button
```

Creates a new button.

#### NewLabel

```go
func NewLabel(id, text string, position, size matrix.Vec2) *Label
```

Creates a new text label.

---

## pkg/world

World generation and chunk management.

### World

#### NewWorld

```go
func NewWorld(name string, seed int64) *World
```

Creates a new world instance.

#### World.GetBlock

```go
func (w *World) GetBlock(pos BlockPos) blocks.BlockID
```

Gets the block at a position.

#### World.SetBlock

```go
func (w *World) SetBlock(pos BlockPos, block blocks.BlockID)
```

Sets a block at a position.

#### World.GetSafeSpawnHeight

```go
func (w *World) GetSafeSpawnHeight(x, z int) int
```

Finds a safe Y coordinate for spawning.

### ChunkManager

#### ChunkManager.GetChunk

```go
func (cm *ChunkManager) GetChunk(coord ChunkCoord) *Chunk
```

Gets or loads a chunk.

#### ChunkManager.Update

```go
func (cm *ChunkManager) Update(deltaTime float64, center matrix.Vec3)
```

Updates chunks based on player position.

### Coordinates

#### HexCoord

```go
type HexCoord struct {
    Q, R int
}
```

Axial coordinates for hexagonal grid.

#### HexCoord.ToWorld

```go
func (h HexCoord) ToWorld(scale float32) matrix.Vec3
```

Converts hex coordinates to world position.

---

## Examples

### Creating a Block

```go
import "github.com/tesselstudio/TesselBox/pkg/blocks"

func createCustomBlock() {
    block := blocks.BlockType{
        ID:       blocks.BlockID(100),
        Name:     "Custom Block",
        Solid:    true,
        Material: blocks.MaterialStone,
        Hardness: 2.0,
    }
    // Register with world...
}
```

### Adding an Item

```go
import "github.com/tesselstudio/TesselBox/pkg/crafting"

func createCustomItem() {
    item := crafting.Item{
        ID:       crafting.ItemID(1000),
        Name:     "Magic Sword",
        MaxStack: 1,
        ToolType: crafting.ToolSword,
        AttackDamage: 10.0,
        Durability:   500,
    }
    
    registry := crafting.GetGlobalItemRegistry()
    registry.Register(item)
}
```

### Creating a Recipe

```go
import "github.com/tesselstudio/TesselBox/pkg/crafting"

func createRecipe() {
    stick, _ := crafting.GetGlobalItemRegistry().GetItem("stick")
    stone, _ := crafting.GetGlobalItemRegistry().GetItem("stone")
    
    recipe := crafting.Recipe{
        ID:   1,
        Name: "Stone Pickaxe",
        Inputs: []crafting.RecipeInput{
            {Item: stone, Count: 3, X: 0, Y: 0},
            {Item: stick, Count: 2, X: 1, Y: 1},
        },
        Output: crafting.NewItemStack(stonePickaxe, 1),
    }
}
```

### World Generation

```go
import "github.com/tesselstudio/TesselBox/pkg/world"

func generateWorld() {
    w := world.NewWorld("MyWorld", 12345)
    
    // Get spawn point
    spawn := w.GetSpawnPoint()
    safeY := w.GetSafeSpawnHeight(int(spawn.X()), int(spawn.Z()))
    
    // Set a block
    pos := world.BlockPos{X: 0, Y: safeY, Z: 0}
    w.SetBlock(pos, blocks.BlockStone)
}
```
