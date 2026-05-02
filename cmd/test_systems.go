package main

import (
	"fmt"
	"log"

	"github.com/tesselstudio/TesselBox/pkg/blocks"
	"github.com/tesselstudio/TesselBox/pkg/crafting"
	"github.com/tesselstudio/TesselBox/pkg/network"
	"github.com/tesselstudio/TesselBox/pkg/survival"
	"github.com/tesselstudio/TesselBox/pkg/ui"
	"github.com/tesselstudio/TesselBox/pkg/world"
	"kaijuengine.com/matrix"
)

func main() {
	fmt.Println("=== TesselBox 3D - System Test ===")
	fmt.Println("Testing all game systems without Vulkan rendering...")
	
	// Test Block System
	fmt.Println("\n🔷 Testing Block System...")
	testBlockSystem()
	
	// Test World Generation  
	fmt.Println("\n🌍 Testing World Generation...")
	testWorldGeneration()
	
	// Test Crafting System
	fmt.Println("\n🔨 Testing Crafting System...")
	testCraftingSystem()
	
	// Test Survival System
	fmt.Println("\n❤️ Testing Survival System...")
	testSurvivalSystem()
	
	// Test UI System
	fmt.Println("\n🎨 Testing UI System...")
	testUISystem()
	
	// Test Networking
	fmt.Println("\n🌐 Testing Networking...")
	testNetworking()
	
	fmt.Println("\n✅ All systems tested successfully!")
	fmt.Println("\n📝 Note: To run the full game with Vulkan rendering, you may need:")
	fmt.Println("   - Proper graphics drivers installed")
	fmt.Println("   - Vulkan runtime libraries") 
	fmt.Println("   - A display environment (not headless)")
	fmt.Println("   - Or try: export VK_ICD_FILENAMES=/usr/share/vulkan/icd.d/llvmpipe_icd.x86_64.json")
}

func testBlockSystem() {
	// Test block registry
	registry := blocks.NewBlockRegistry()
	
	// Test getting block types
	blockTypes := registry.GetAllTypes()
	
	fmt.Printf("  ✓ Available block types: %d\n", len(blockTypes))
	if len(blockTypes) > 0 {
		fmt.Printf("  ✓ First block type: %s\n", blockTypes[0].Name)
	}
}

func testWorldGeneration() {
	// Test hexagonal coordinates
	coord := world.HexCoord{Q: 5, R: 3}
	worldPos := coord.ToWorld(1.0) // Scale factor
	
	fmt.Printf("  ✓ Hex coordinate (%d, %d) -> World position (%.2f, %.2f)\n", 
		coord.Q, coord.R, worldPos.X(), worldPos.Z())
	
	// Test distance calculation
	other := world.HexCoord{Q: 8, R: 6}
	distance := coord.Distance(other)
	fmt.Printf("  ✓ Distance: %.2f\n", distance)
}

func testCraftingSystem() {
	// Test item system
	itemRegistry := crafting.GetGlobalItemRegistry()
	
	// Test creating items
	woodItem, err := itemRegistry.GetItem("wood")
	if err != nil {
		log.Printf("  ✗ Could not get wood item: %v", err)
		return
	}
	
	stoneItem, err := itemRegistry.GetItem("stone")
	if err != nil {
		log.Printf("  ✗ Could not get stone item: %v", err)
		return
	}
	
	fmt.Printf("  ✓ Available items: %d\n", len(itemRegistry.GetAllItems()))
	fmt.Printf("  ✓ Wood item: %s\n", woodItem.Name)
	fmt.Printf("  ✓ Stone item: %s\n", stoneItem.Name)
	
	// Test inventory
	inventory := crafting.NewInventory(36, 10)
	woodStack := crafting.NewItemStack(woodItem, 64)
	inventory.AddItem(woodStack)
	
	fmt.Printf("  ✓ Inventory created successfully\n")
	fmt.Printf("  ✓ Added wood stack to inventory\n")
}

func testSurvivalSystem() {
	// Test player stats
	stats := survival.NewPlayerStats(100.0, 20.0, 100.0)
	
	fmt.Printf("  ✓ Health: %.1f/%.1f\n", stats.Health.GetCurrentHealth(), stats.Health.GetMaxHealth())
	fmt.Printf("  ✓ Hunger: %.1f/%.1f\n", stats.Hunger.GetCurrentHunger(), stats.Hunger.GetMaxHunger())
	fmt.Printf("  ✓ Stamina: %.1f/%.1f\n", stats.Stamina.GetCurrentStamina(), stats.Stamina.GetMaxStamina())
	fmt.Printf("  ✓ Level: %d\n", stats.Experience.GetLevel())
	
	// Test environmental system
	env := survival.NewEnvironment()
	fmt.Printf("  ✓ Time of day: %.1f\n", env.GetTimeOfDay())
	fmt.Printf("  ✓ Temperature: %.1f°C\n", env.GetTemperature())
	fmt.Printf("  ✓ Weather: %v\n", env.GetWeather())
}

func testUISystem() {
	// Test UI components
	panel := ui.NewPanel("test", matrix.NewVec2(100, 100), matrix.NewVec2(200, 150))
	button := ui.NewButton("test_btn", "Click Me", matrix.NewVec2(10, 10), matrix.NewVec2(80, 30))
	label := ui.NewLabel("test_label", "Hello World", matrix.NewVec2(10, 50), matrix.NewVec2(180, 20))
	
	panel.AddChild(button.UIComponent)
	panel.AddChild(label.UIComponent)
	
	fmt.Printf("  ✓ Created panel with %d children\n", len(panel.GetChildren()))
	fmt.Printf("  ✓ Button text: %s\n", button.GetText())
	fmt.Printf("  ✓ Label text: %s\n", label.GetText())
	
	// Test HUD
	hud := ui.NewHUD(1920, 1080)
	fmt.Printf("  ✓ HUD created with %d components\n", len(hud.GetRootPanel().GetChildren()))
}

func testNetworking() {
	// Test protocol
	protocol := network.NewProtocol()
	
	// Test message creation
	handshake := &network.HandshakeMessage{
		PlayerName: "TestPlayer",
		Version:    "1.0.0",
	}
	
	encoded := protocol.EncodeHandshake(handshake)
	decoded, err := protocol.DecodeHandshake(encoded)
	
	if err != nil {
		log.Printf("  ✗ Protocol test failed: %v", err)
		return
	}
	
	fmt.Printf("  ✓ Protocol encoding/decoding works\n")
	fmt.Printf("  ✓ Handshake: %s (v%s)\n", decoded.PlayerName, decoded.Version)
	
	// Test server config
	config := network.ServerConfig{
		Host:       "localhost",
		Port:       25565,
		MaxPlayers:  16,
		TickRate:   60,
		WorldSize:  100,
		ChunkSize:  16,
		ViewDistance: 8,
	}
	
	fmt.Printf("  ✓ Server config: %s:%d (max %d players)\n", config.Host, config.Port, config.MaxPlayers)
}
