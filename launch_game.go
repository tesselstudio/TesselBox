package main

import (
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/tesselstudio/TesselBox/pkg/opengl"
	"github.com/tesselstudio/TesselBox/pkg/player"
	"github.com/tesselstudio/TesselBox/pkg/types"
	"github.com/tesselstudio/TesselBox/pkg/world"
)

func main() {
	println("🚀 Launching TesselBox Hexagonal Prism World")

	// Create world
	gameWorld := world.NewWorld("TesselBox World", 12345)

	// Initialize chunk loading around spawn point
	spawn := gameWorld.GetSpawnPoint()
	tempPlayerPos := types.NewVec3(float32(spawn.X), 70.0, float32(spawn.Z))
	gameWorld.GetChunkManager().InitializeChunkLoading(tempPlayerPos)

	// Wait for chunks to generate
	println("⏳ Generating vast hexagonal prism world...")
	time.Sleep(3 * time.Second)

	// Check chunk generation
	stats := gameWorld.GetChunkManager().GetStats()
	println("📊 Generated", stats.LoadedChunks, "chunks with hexagonal prisms")

	if stats.LoadedChunks == 0 {
		println("❌ No chunks generated!")
		return
	}

	// Wait for mesh generation to complete
	println("⏳ Waiting for mesh generation...")
	maxWaitTime := 10 * time.Second
	waitInterval := 500 * time.Millisecond
	startTime := time.Now()

	for time.Since(startTime) < maxWaitTime {
		// Force mesh rebuild
		gameWorld.GetChunkManager().RebuildDirtyMeshes()

		// Check if chunks have meshes
		allChunksHaveMeshes := true
		loadedChunks := gameWorld.GetChunkManager().GetLoadedChunks()

		for coord, chunk := range loadedChunks {
			mesh := chunk.GetMesh()
			if mesh == nil || len(mesh.InterleavedVertices) == 0 {
				allChunksHaveMeshes = false
				println("🔍 DEBUG: Chunk", coord.X, coord.Z, "still needs mesh")
				break
			}
		}

		if allChunksHaveMeshes {
			println("✅ All chunks have meshes!")
			break
		}

		time.Sleep(waitInterval)
	}

	// Final check
	finalStats := gameWorld.GetChunkManager().GetStats()
	println("📊 Final stats - Loaded:", finalStats.LoadedChunks, "Generated:", finalStats.GeneratedChunks, "Meshed:", finalStats.MeshedChunks)

	// Find safe spawn height
	safeY := gameWorld.GetSafeSpawnHeight(int(spawn.X), int(spawn.Z))

	playerPos := types.NewVec3(float32(spawn.X), float32(safeY), float32(spawn.Z))

	println("🔍 DEBUG: Safe spawn height:", safeY)
	println("🔍 DEBUG: Spawn position - X:", spawn.X, "Y:", safeY, "Z:", spawn.Z)
	println("🔍 DEBUG: Player position - X:", playerPos.X, "Y:", playerPos.Y, "Z:", playerPos.Z)

	// Create player
	player := player.NewPlayer(gameWorld)
	player.SetPosition(playerPos)

	// Create OpenGL engine
	println("🔧 Creating OpenGL window for hexagonal prism world...")
	engine, err := opengl.NewEngine(1024, 768, "TesselBox - Hexagonal Prism World")
	if err != nil {
		println("❌ Failed to create OpenGL engine:", err.Error())
		return
	}
	defer engine.Cleanup()

	println("✅ OpenGL window created!")
	println("🌍 Vast hexagonal prism world with different terrains is ready!")
	println("👤 Player spawned at X=", spawn.X, " Y=", safeY, " Z=", spawn.Z)

	// Generate meshes for all chunks
	println("🎨 Generating hexagonal prism meshes...")
	gameWorld.GetChunkManager().UpdateEngineWithMeshes(engine)

	// Debug: Verify mesh transfer
	meshCount := engine.GetLoadedMeshCount()
	println("🔍 DEBUG: Engine has", meshCount, "loaded meshes after UpdateEngineWithMeshes")

	if meshCount == 0 {
		println("❌ CRITICAL: No meshes were transferred to engine! This explains grey screen.")
		println("🔍 DEBUG: Checking chunk manager stats...")
		stats := gameWorld.GetChunkManager().GetStats()
		println("📊 Chunk Manager Stats - Loaded:", stats.LoadedChunks, "Generated:", stats.GeneratedChunks, "Meshed:", stats.MeshedChunks)
	} else {
		println("✅ Good: Engine received", meshCount, "meshes for rendering")
	}

	// Game loop - render the vast hexagonal world
	println("🎮 Starting rendering loop...")
	frameCount := 0
	for !engine.ShouldClose() {
		// Update game logic
		player.Update(0.016) // ~60 FPS
		playerPos := player.GetPosition()
		gameWorld.Update(0.016, world.NewVec3(playerPos.X, playerPos.Y, playerPos.Z))

		// Update camera from player
		playerRot := player.GetRotation()
		engine.UpdateCameraFromPlayer(
			mgl32.Vec3{playerPos.X, playerPos.Y, playerPos.Z},
			mgl32.Vec3{playerRot.X, playerRot.Y, playerRot.Z},
		)

		// Render frame
		engine.BeginFrame()
		engine.Render(nil)
		engine.EndFrame()
		engine.PollEvents()

		// Print status every 60 frames (1 second)
		frameCount++
		if frameCount%60 == 0 {
			println("🎨 Rendering hexagonal prism world... Frame:", frameCount/60)
		}
	}

	println("🎮 Game ended")
}
