package opengl

import (
	"sync"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/tesselstudio/TesselBox/pkg/world"
)

// MeshRenderer manages rendering of chunk meshes
type MeshRenderer struct {
	mu sync.RWMutex

	// Mesh cache: ChunkCoord -> (VAO, VertexCount)
	meshes map[world.ChunkCoord]*RenderedMesh
}

// RenderedMesh holds OpenGL resources for a chunk mesh
type RenderedMesh struct {
	VAO         uint32
	VBO         uint32
	EBO         uint32
	VertexCount int32
	IndexCount  int32
	Generated   bool
}

// NewMeshRenderer creates a new mesh renderer
func NewMeshRenderer() *MeshRenderer {
	return &MeshRenderer{
		meshes: make(map[world.ChunkCoord]*RenderedMesh),
	}
}

// AddMesh adds a chunk mesh to the renderer
func (mr *MeshRenderer) AddMesh(coord world.ChunkCoord, meshData *ChunkMeshData) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	println("🎨 Adding mesh for chunk", coord.X, coord.Z, "with", meshData.VertexCount, "vertices")

	// Create VAO
	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	// Bind VAO
	gl.BindVertexArray(vao)

	// Bind and fill VBO
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(meshData.Vertices)*4, gl.Ptr(meshData.Vertices), gl.STATIC_DRAW)

	// Bind and fill EBO
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(meshData.Indices)*4, gl.Ptr(meshData.Indices), gl.STATIC_DRAW)

	// Set vertex attributes
	// Position (3 floats)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	// Color (3 floats)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	// Unbind
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	mr.meshes[coord] = &RenderedMesh{
		VAO:         vao,
		VBO:         vbo,
		EBO:         ebo,
		VertexCount: meshData.VertexCount,
		IndexCount:  meshData.IndexCount,
		Generated:   true,
	}

	println("✅ Mesh added successfully for chunk", coord.X, coord.Z)
}

// RemoveMesh removes a chunk mesh from the renderer
func (mr *MeshRenderer) RemoveMesh(coord world.ChunkCoord) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if mesh, exists := mr.meshes[coord]; exists && mesh.Generated {
		gl.DeleteBuffers(1, &mesh.EBO)
		gl.DeleteBuffers(1, &mesh.VBO)
		gl.DeleteVertexArrays(1, &mesh.VAO)
	}

	delete(mr.meshes, coord)
}

// GetMesh retrieves a chunk mesh
func (mr *MeshRenderer) GetMesh(coord world.ChunkCoord) *RenderedMesh {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return mr.meshes[coord]
}

// RenderMesh renders a specific chunk mesh
func (mr *MeshRenderer) RenderMesh(coord world.ChunkCoord) {
	mr.mu.RLock()
	mesh := mr.meshes[coord]
	mr.mu.RUnlock()

	if mesh == nil || !mesh.Generated {
		return
	}

	gl.BindVertexArray(mesh.VAO)
	gl.DrawElements(gl.TRIANGLES, mesh.IndexCount, gl.UNSIGNED_INT, nil)
	gl.BindVertexArray(0)
}

// RenderAllMeshes renders all chunk meshes (simple batch rendering)
func (mr *MeshRenderer) RenderAllMeshes() {
	mr.mu.RLock()
	meshes := mr.meshes
	totalMeshes := len(meshes)
	mr.mu.RUnlock()

	meshCount := 0
	for coord, mesh := range meshes {
		if mesh != nil && mesh.Generated {
			gl.BindVertexArray(mesh.VAO)
			gl.DrawElements(gl.TRIANGLES, mesh.IndexCount, gl.UNSIGNED_INT, nil)
			gl.BindVertexArray(0)
			meshCount++

			// Debug: Log first few meshes being rendered
			if meshCount <= 3 {
				println("🔍 DEBUG: Rendering chunk", coord.X, coord.Z, "with", mesh.IndexCount, "indices")
			}
		}
	}

	if meshCount > 0 {
		println("🎨 Rendered", meshCount, "of", totalMeshes, "available chunk meshes")
	} else {
		println("❌ DEBUG: No meshes rendered! Available meshes:", totalMeshes)
		if totalMeshes > 0 {
			println("🔍 DEBUG: Meshes exist but Generated flag may be false")
			mr.mu.RLock()
			for coord, mesh := range meshes {
				if mesh != nil {
					println("  - Chunk", coord.X, coord.Z, "Generated:", mesh.Generated, "IndexCount:", mesh.IndexCount)
				}
			}
			mr.mu.RUnlock()
		}
	}
}

// GetLoadedMeshCount returns the number of loaded meshes
func (mr *MeshRenderer) GetLoadedMeshCount() int {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return len(mr.meshes)
}

// Clear removes all meshes
func (mr *MeshRenderer) Clear() {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	for coord, mesh := range mr.meshes {
		if mesh != nil && mesh.Generated {
			gl.DeleteBuffers(1, &mesh.EBO)
			gl.DeleteBuffers(1, &mesh.VBO)
			gl.DeleteVertexArrays(1, &mesh.VAO)
		}
		delete(mr.meshes, coord)
	}
}
