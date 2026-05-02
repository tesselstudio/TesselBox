package main

import (
	"reflect"

	"github.com/tesselstudio/TesselBox/pkg/blocks"
	"kaijuengine.com/bootstrap"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

// TesselBoxGame represents the main game state
type TesselBoxGame struct {
	host       *engine.Host
	testEntity *engine.Entity
	updateId   engine.UpdateId
}

// PluginRegistry returns the plugin types for this game
func (g *TesselBoxGame) PluginRegistry() []reflect.Type {
	return []reflect.Type{}
}

// ContentDatabase returns the game content database
func (g *TesselBoxGame) ContentDatabase() (assets.Database, error) {
	// Use file database for game content
	return assets.NewFileDatabase("game_content")
}

// Launch initializes the game
func (g *TesselBoxGame) Launch(host *engine.Host) {
	g.host = host

	// Create a hexagonal prism mesh
	hexPrism := blocks.NewHexPrism(matrix.NewVec3(0, 0, 0), 1.0, 2.0)
	vertices := hexPrism.GenerateVertices()
	indices := hexPrism.GenerateIndices()
	normals := hexPrism.GenerateNormals()
	uvs := hexPrism.GenerateUVCoordinates()

	// Convert to Kaiju Engine Vertex format
	meshVertices := make([]rendering.Vertex, len(vertices))
	for i := 0; i < len(vertices); i++ {
		meshVertices[i] = rendering.Vertex{
			Position: vertices[i],
			Normal:   normals[i],
			UV0:      uvs[i],
			Color:    matrix.ColorWhite(),
		}
	}

	// Create mesh from hexagonal prism data
	mesh := host.MeshCache().Mesh("hex_prism", meshVertices, indices)

	// Create shader data for material
	sd := shader_data_registry.Create("basic")
	sd.(*shader_data_registry.ShaderDataStandard).Color = matrix.ColorBlue()

	// Create an entity with transform
	g.testEntity = engine.NewEntity(host.WorkGroup())

	// Get material and texture from caches
	mat, err := host.MaterialCache().Material(assets.MaterialDefinitionBasic)
	if err != nil {
		panic("Material not found - check asset database path")
	}
	tex, err := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	if err != nil {
		panic("Texture not found - check asset database path")
	}

	// Create drawing
	draw := rendering.Drawing{
		Material:   mat.CreateInstance([]*rendering.Texture{tex}),
		Mesh:       mesh,
		ShaderData: sd,
		Transform:  &g.testEntity.Transform,
		ViewCuller: &host.Cameras.Primary,
	}

	// Add drawing to rendering system
	host.Drawings.AddDrawing(draw)

	// Register update function for game loop
	g.updateId = host.Updater.AddUpdate(g.update)

	// Cleanup when entity is destroyed
	g.testEntity.OnDestroy.Add(func() {
		sd.Destroy()
		host.Updater.RemoveUpdate(&g.updateId)
	})
}

// Update handles game logic updates
func (g *TesselBoxGame) update(deltaTime float64) {
	// Animate test entity
	x := matrix.Sin(matrix.Float(g.host.Runtime()))
	g.testEntity.Transform.SetPosition(matrix.NewVec3(x, 0, -3))
}

// getGame returns the game instance for bootstrap
func getGame() bootstrap.GameInterface {
	return &TesselBoxGame{}
}

func main() {
	bootstrap.Main(getGame(), nil)
}
