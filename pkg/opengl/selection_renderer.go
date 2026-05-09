package opengl

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tesselstudio/TesselBox/pkg/types"
	"github.com/tesselstudio/TesselBox/pkg/world"
)

// SelectionRenderer renders the block selection highlight
type SelectionRenderer struct {
	vao              uint32
	vbo              uint32
	shaderProgram    uint32
	modelLoc         int32
	viewLoc          int32
	projLoc          int32
	colorLoc         int32
	initialized      bool
	selectedBlock    *world.RaycastHit
	selectionColor   mgl32.Vec4
	outlineThickness float32
}

// NewSelectionRenderer creates a new selection renderer
func NewSelectionRenderer() *SelectionRenderer {
	sr := &SelectionRenderer{
		selectionColor:   mgl32.Vec4{1, 1, 0, 0.5}, // Yellow with transparency
		outlineThickness: 1.02,                     // Slightly larger than block
	}
	sr.init()
	return sr
}

// init initializes OpenGL resources for selection rendering
func (sr *SelectionRenderer) init() {
	// Vertex shader source
	vertexShaderSource := `
		#version 410 core
		layout (location = 0) in vec3 aPos;
		
		uniform mat4 model;
		uniform mat4 view;
		uniform mat4 projection;
		
		void main() {
			gl_Position = projection * view * model * vec4(aPos, 1.0);
		}
	`

	// Fragment shader source
	fragmentShaderSource := `
		#version 410 core
		uniform vec4 color;
		out vec4 FragColor;
		
		void main() {
			FragColor = color;
		}
	`

	// Compile shaders
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	csource, free := gl.Strs(vertexShaderSource)
	gl.ShaderSource(vertexShader, 1, csource, nil)
	free()
	gl.CompileShader(vertexShader)

	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	csource, free = gl.Strs(fragmentShaderSource)
	gl.ShaderSource(fragmentShader, 1, csource, nil)
	free()
	gl.CompileShader(fragmentShader)

	// Create program
	sr.shaderProgram = gl.CreateProgram()
	gl.AttachShader(sr.shaderProgram, vertexShader)
	gl.AttachShader(sr.shaderProgram, fragmentShader)
	gl.LinkProgram(sr.shaderProgram)

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	// Get uniform locations
	sr.modelLoc = gl.GetUniformLocation(sr.shaderProgram, gl.Str("model\x00"))
	sr.viewLoc = gl.GetUniformLocation(sr.shaderProgram, gl.Str("view\x00"))
	sr.projLoc = gl.GetUniformLocation(sr.shaderProgram, gl.Str("projection\x00"))
	sr.colorLoc = gl.GetUniformLocation(sr.shaderProgram, gl.Str("color\x00"))

	// Create cube wireframe vertices
	// This creates a wireframe cube outline
	vertices := []float32{
		// Front face
		-0.5, -0.5, 0.5, 0.5, -0.5, 0.5,
		0.5, -0.5, 0.5, 0.5, 0.5, 0.5,
		0.5, 0.5, 0.5, -0.5, 0.5, 0.5,
		-0.5, 0.5, 0.5, -0.5, -0.5, 0.5,

		// Back face
		-0.5, -0.5, -0.5, 0.5, -0.5, -0.5,
		0.5, -0.5, -0.5, 0.5, 0.5, -0.5,
		0.5, 0.5, -0.5, -0.5, 0.5, -0.5,
		-0.5, 0.5, -0.5, -0.5, -0.5, -0.5,
	}

	gl.GenVertexArrays(1, &sr.vao)
	gl.GenBuffers(1, &sr.vbo)

	gl.BindVertexArray(sr.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, sr.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 12, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	sr.initialized = true
}

// SetSelectedBlock sets the block that is currently selected
func (sr *SelectionRenderer) SetSelectedBlock(hit *world.RaycastHit) {
	sr.selectedBlock = hit
}

// SetSelectionColor sets the color of the selection highlight
func (sr *SelectionRenderer) SetSelectionColor(r, g, b, a float32) {
	sr.selectionColor = mgl32.Vec4{r, g, b, a}
}

// Render renders the selection highlight
func (sr *SelectionRenderer) Render(view, projection mgl32.Mat4) {
	if !sr.initialized || sr.selectedBlock == nil || !sr.selectedBlock.Hit {
		return
	}

	gl.UseProgram(sr.shaderProgram)
	gl.BindVertexArray(sr.vao)

	// Enable wireframe mode
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	gl.LineWidth(2.0)

	// Create model matrix for the selected block
	blockPos := sr.selectedBlock.BlockPos
	model := mgl32.Translate3D(blockPos.X+0.5, blockPos.Y, blockPos.Z+0.5)
	model = model.Mul4(mgl32.Scale3D(sr.outlineThickness, sr.outlineThickness, sr.outlineThickness))

	// Set uniforms
	gl.UniformMatrix4fv(sr.modelLoc, 1, false, &model[0])
	gl.UniformMatrix4fv(sr.viewLoc, 1, false, &view[0])
	gl.UniformMatrix4fv(sr.projLoc, 1, false, &projection[0])
	gl.Uniform4f(sr.colorLoc, sr.selectionColor[0], sr.selectionColor[1], sr.selectionColor[2], sr.selectionColor[3])

	// Disable depth test for outline
	gl.Disable(gl.DEPTH_TEST)

	// Draw wireframe cube (24 lines = 24 vertices)
	gl.DrawArrays(gl.LINES, 0, 24)

	// Re-enable depth test
	gl.Enable(gl.DEPTH_TEST)

	// Restore fill mode
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)

	gl.BindVertexArray(0)
	gl.UseProgram(0)
}

// Render3D renders a 3D wireframe cube at a specific position (for debugging)
func (sr *SelectionRenderer) RenderCubeAtPosition(pos types.Vec3, view, projection mgl32.Mat4) {
	if !sr.initialized {
		return
	}

	gl.UseProgram(sr.shaderProgram)
	gl.BindVertexArray(sr.vao)

	gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	gl.LineWidth(2.0)

	model := mgl32.Translate3D(pos.X, pos.Y, pos.Z)

	gl.UniformMatrix4fv(sr.modelLoc, 1, false, &model[0])
	gl.UniformMatrix4fv(sr.viewLoc, 1, false, &view[0])
	gl.UniformMatrix4fv(sr.projLoc, 1, false, &projection[0])
	gl.Uniform4f(sr.colorLoc, sr.selectionColor[0], sr.selectionColor[1], sr.selectionColor[2], sr.selectionColor[3])

	gl.Disable(gl.DEPTH_TEST)
	gl.DrawArrays(gl.LINES, 0, 24)
	gl.Enable(gl.DEPTH_TEST)

	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)

	gl.BindVertexArray(0)
	gl.UseProgram(0)
}

// Cleanup cleans up OpenGL resources
func (sr *SelectionRenderer) Cleanup() {
	if sr.initialized {
		gl.DeleteVertexArrays(1, &sr.vao)
		gl.DeleteBuffers(1, &sr.vbo)
		gl.DeleteProgram(sr.shaderProgram)
		sr.initialized = false
	}
}
