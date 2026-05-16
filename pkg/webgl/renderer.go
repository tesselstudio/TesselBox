//go:build js && wasm
// +build js,wasm

package webgl

import (
	"fmt"
	"math"
	"syscall/js"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/tesselstudio/TesselBox/pkg/types"
	"github.com/tesselstudio/TesselBox/pkg/world"
)

// Renderer handles WebGL rendering for the web version
type Renderer struct {
	gl             js.Value
	shaderProgram  js.Value
	width          int
	height         int
	cameraPosition types.Vec3
	cameraRotation types.Vec3
	initialized    bool
	chunkMeshes    map[string]*ChunkMeshData
}

// ChunkMeshData stores mesh data for a chunk
type ChunkMeshData struct {
	Vertices    []float32
	Indices     []uint32
	VertexCount int32
	IndexCount  int32
	vao         js.Value
	vbo         js.Value
	ebo         js.Value
}

// NewRenderer creates a new WebGL renderer
func NewRenderer() *Renderer {
	return &Renderer{
		chunkMeshes: make(map[string]*ChunkMeshData),
	}
}

// Initialize initializes WebGL context and resources
func (r *Renderer) Initialize(canvas js.Value) error {
	fmt.Println("Initializing WebGL renderer...")

	// Get WebGL context
	r.gl = canvas.Call("getContext", "webgl2")
	if r.gl.IsNull() || r.gl.IsUndefined() {
		// Fallback to WebGL 1
		r.gl = canvas.Call("getContext", "webgl")
		if r.gl.IsNull() || r.gl.IsUndefined() {
			return fmt.Errorf("WebGL not supported")
		}
	}

	fmt.Println("WebGL context obtained")

	// Enable depth testing
	r.gl.Call("enable", r.gl.Get("DEPTH_TEST"))
	r.gl.Call("depthFunc", r.gl.Get("LEQUAL"))

	// Set clear color (sky blue)
	r.gl.Call("clearColor", 0.5, 0.7, 1.0, 1.0)

	// Initialize shaders
	if err := r.initShaders(); err != nil {
		return fmt.Errorf("failed to initialize shaders: %w", err)
	}

	r.initialized = true
	fmt.Println("WebGL renderer initialized successfully")
	return nil
}

// initShaders compiles and links shader programs
func (r *Renderer) initShaders() error {
	// Vertex shader source
	vertexShaderSource := `
		attribute vec3 aPos;
		attribute vec3 aColor;
		
		uniform mat4 model;
		uniform mat4 view;
		uniform mat4 projection;
		
		varying vec3 vertexColor;
		
		void main() {
			gl_Position = projection * view * model * vec4(aPos, 1.0);
			vertexColor = aColor;
		}
	`

	// Fragment shader source
	fragmentShaderSource := `
		precision mediump float;
		varying vec3 vertexColor;
		
		void main() {
			gl_FragColor = vec4(vertexColor, 1.0);
		}
	`

	// Compile vertex shader
	vertexShader := r.compileShader(vertexShaderSource, r.gl.Get("VERTEX_SHADER"))
	if vertexShader.IsNull() {
		return fmt.Errorf("failed to compile vertex shader")
	}

	// Compile fragment shader
	fragmentShader := r.compileShader(fragmentShaderSource, r.gl.Get("FRAGMENT_SHADER"))
	if fragmentShader.IsNull() {
		return fmt.Errorf("failed to compile fragment shader")
	}

	// Create shader program
	r.shaderProgram = r.gl.Call("createProgram")
	r.gl.Call("attachShader", r.shaderProgram, vertexShader)
	r.gl.Call("attachShader", r.shaderProgram, fragmentShader)
	r.gl.Call("linkProgram", r.shaderProgram)

	// Check link status
	linkStatus := r.gl.Call("getProgramParameter", r.shaderProgram, r.gl.Get("LINK_STATUS"))
	if !linkStatus.Bool() {
		infoLog := r.gl.Call("getProgramInfoLog", r.shaderProgram)
		return fmt.Errorf("shader program link failed: %s", infoLog.String())
	}

	// Clean up shaders
	r.gl.Call("deleteShader", vertexShader)
	r.gl.Call("deleteShader", fragmentShader)

	fmt.Println("Shaders compiled and linked successfully")
	return nil
}

// compileShader compiles a shader from source
func (r *Renderer) compileShader(source string, shaderType js.Value) js.Value {
	shader := r.gl.Call("createShader", shaderType)
	r.gl.Call("shaderSource", shader, source)
	r.gl.Call("compileShader", shader)

	// Check compilation status
	compileStatus := r.gl.Call("getShaderParameter", shader, r.gl.Get("COMPILE_STATUS"))
	if !compileStatus.Bool() {
		infoLog := r.gl.Call("getShaderInfoLog", shader)
		fmt.Printf("Shader compilation failed: %s\n", infoLog.String())
		r.gl.Call("deleteShader", shader)
		return js.Value{}
	}

	return shader
}

// BeginFrame clears the screen for a new frame
func (r *Renderer) BeginFrame() {
	if !r.initialized {
		return
	}
	colorBit := r.gl.Get("COLOR_BUFFER_BIT").Int()
	depthBit := r.gl.Get("DEPTH_BUFFER_BIT").Int()
	r.gl.Call("clear", colorBit|depthBit)
}

// EndFrame completes the frame (no-op for WebGL, buffers swap automatically)
func (r *Renderer) EndFrame() {
	// WebGL automatically swaps buffers
}

// Render renders the scene
func (r *Renderer) Render() {
	if !r.initialized {
		return
	}

	// Use shader program
	r.gl.Call("useProgram", r.shaderProgram)

	// Create transformation matrices
	model := mgl32.Ident4()
	view := r.getViewMatrix()
	projection := r.getProjectionMatrix()

	// Set uniforms
	modelLoc := r.gl.Call("getUniformLocation", r.shaderProgram, "model")
	viewLoc := r.gl.Call("getUniformLocation", r.shaderProgram, "view")
	projLoc := r.gl.Call("getUniformLocation", r.shaderProgram, "projection")

	r.setUniformMatrix4fv(modelLoc, model)
	r.setUniformMatrix4fv(viewLoc, view)
	r.setUniformMatrix4fv(projLoc, projection)

	// Render all chunk meshes
	for _, mesh := range r.chunkMeshes {
		r.renderChunkMesh(mesh)
	}

	// Unbind
	r.gl.Call("useProgram", js.Value{})
}

// renderChunkMesh renders a single chunk mesh
func (r *Renderer) renderChunkMesh(mesh *ChunkMeshData) {
	if mesh == nil || mesh.vao.IsNull() {
		return
	}

	r.gl.Call("bindVertexArray", mesh.vao)
	r.gl.Call("drawElements", r.gl.Get("TRIANGLES"), mesh.IndexCount, r.gl.Get("UNSIGNED_INT"), 0)
	r.gl.Call("bindVertexArray", js.Value{})
}

// AddChunkMesh adds a chunk mesh to the renderer
func (r *Renderer) AddChunkMesh(coord world.ChunkCoord, vertices []float32, indices []uint32) {
	key := fmt.Sprintf("%d,%d", coord.X, coord.Z)

	// Create mesh data
	mesh := &ChunkMeshData{
		Vertices:    vertices,
		Indices:     indices,
		VertexCount: int32(len(vertices) / 6), // 6 floats per vertex (pos + color)
		IndexCount:  int32(len(indices)),
	}

	// Create WebGL buffers
	r.createChunkBuffers(mesh)

	r.chunkMeshes[key] = mesh
	fmt.Printf("Added chunk mesh for %s: %d vertices, %d indices\n", key, mesh.VertexCount, mesh.IndexCount)
}

// createChunkBuffers creates WebGL buffers for a chunk mesh
func (r *Renderer) createChunkBuffers(mesh *ChunkMeshData) {
	// Create VAO
	mesh.vao = r.gl.Call("createVertexArray")
	r.gl.Call("bindVertexArray", mesh.vao)

	// Create VBO
	mesh.vbo = r.gl.Call("createBuffer")
	r.gl.Call("bindBuffer", r.gl.Get("ARRAY_BUFFER"), mesh.vbo)

	// Upload vertex data
	vertexData := r.float32ArrayToJS(mesh.Vertices)
	r.gl.Call("bufferData", r.gl.Get("ARRAY_BUFFER"), vertexData, r.gl.Get("STATIC_DRAW"))

	// Set vertex attributes
	// Position attribute (location 0)
	posLoc := r.gl.Call("getAttribLocation", r.shaderProgram, "aPos")
	r.gl.Call("enableVertexAttribArray", posLoc)
	r.gl.Call("vertexAttribPointer", posLoc, 3, r.gl.Get("FLOAT"), false, 6*4, 0)

	// Color attribute (location 1)
	colorLoc := r.gl.Call("getAttribLocation", r.shaderProgram, "aColor")
	r.gl.Call("enableVertexAttribArray", colorLoc)
	r.gl.Call("vertexAttribPointer", colorLoc, 3, r.gl.Get("FLOAT"), false, 6*4, 3*4)

	// Create EBO for indices
	mesh.ebo = r.gl.Call("createBuffer")
	r.gl.Call("bindBuffer", r.gl.Get("ELEMENT_ARRAY_BUFFER"), mesh.ebo)

	// Upload index data
	indexData := r.uint32ArrayToJS(mesh.Indices)
	r.gl.Call("bufferData", r.gl.Get("ELEMENT_ARRAY_BUFFER"), indexData, r.gl.Get("STATIC_DRAW"))

	// Unbind
	r.gl.Call("bindBuffer", r.gl.Get("ARRAY_BUFFER"), js.Value{})
	r.gl.Call("bindVertexArray", js.Value{})
}

// RemoveChunkMesh removes a chunk mesh from the renderer
func (r *Renderer) RemoveChunkMesh(coord world.ChunkCoord) {
	key := fmt.Sprintf("%d,%d", coord.X, coord.Z)
	if mesh, exists := r.chunkMeshes[key]; exists {
		// Clean up WebGL resources
		if !mesh.vao.IsNull() {
			r.gl.Call("deleteVertexArray", mesh.vao)
		}
		if !mesh.vbo.IsNull() {
			r.gl.Call("deleteBuffer", mesh.vbo)
		}
		if !mesh.ebo.IsNull() {
			r.gl.Call("deleteBuffer", mesh.ebo)
		}
		delete(r.chunkMeshes, key)
		fmt.Printf("Removed chunk mesh for %s\n", key)
	}
}

// UpdateCamera updates the camera position and rotation
func (r *Renderer) UpdateCamera(position types.Vec3, rotation types.Vec3) {
	r.cameraPosition = position
	r.cameraRotation = rotation
}

// Resize handles window resize
func (r *Renderer) Resize(width, height int) {
	r.width = width
	r.height = height
	if r.initialized {
		r.gl.Call("viewport", 0, 0, width, height)
	}
}

// Cleanup cleans up WebGL resources
func (r *Renderer) Cleanup() {
	if !r.initialized {
		return
	}

	// Clean up all chunk meshes
	for _, mesh := range r.chunkMeshes {
		if !mesh.vao.IsNull() {
			r.gl.Call("deleteVertexArray", mesh.vao)
		}
		if !mesh.vbo.IsNull() {
			r.gl.Call("deleteBuffer", mesh.vbo)
		}
		if !mesh.ebo.IsNull() {
			r.gl.Call("deleteBuffer", mesh.ebo)
		}
	}
	r.chunkMeshes = make(map[string]*ChunkMeshData)

	// Clean up shader program
	if !r.shaderProgram.IsNull() {
		r.gl.Call("deleteProgram", r.shaderProgram)
	}

	r.initialized = false
	fmt.Println("WebGL renderer cleaned up")
}

// getViewMatrix returns the view matrix
func (r *Renderer) getViewMatrix() mgl32.Mat4 {
	// Simple view matrix looking from camera position
	eyeHeight := float32(1.62)
	eyePos := mgl32.Vec3{r.cameraPosition.X, r.cameraPosition.Y + eyeHeight, r.cameraPosition.Z}

	// Calculate look direction from rotation
	pitchRad := r.cameraRotation.X * 3.14159 / 180.0
	yawRad := r.cameraRotation.Y * 3.14159 / 180.0

	forwardX := float32(-math.Sin(float64(yawRad)) * math.Cos(float64(pitchRad)))
	forwardY := float32(math.Sin(float64(pitchRad)))
	forwardZ := float32(-math.Cos(float64(yawRad)) * math.Cos(float64(pitchRad)))

	target := mgl32.Vec3{
		eyePos.X() + forwardX,
		eyePos.Y() + forwardY,
		eyePos.Z() + forwardZ,
	}

	return mgl32.LookAtV(eyePos, target, mgl32.Vec3{0, 1, 0})
}

// getProjectionMatrix returns the projection matrix
func (r *Renderer) getProjectionMatrix() mgl32.Mat4 {
	if r.width == 0 || r.height == 0 {
		r.width, r.height = 1280, 720
	}
	aspectRatio := float32(r.width) / float32(r.height)
	return mgl32.Perspective(mgl32.DegToRad(45.0), aspectRatio, 0.1, 1000.0)
}

// setUniformMatrix4fv sets a 4x4 matrix uniform
func (r *Renderer) setUniformMatrix4fv(loc js.Value, matrix mgl32.Mat4) {
	data := r.float32ArrayToJS(matrix[:])
	r.gl.Call("uniformMatrix4fv", loc, false, data)
}

// float32ArrayToJS converts a []float32 to a JavaScript Float32Array
func (r *Renderer) float32ArrayToJS(data []float32) js.Value {
	jsArray := js.Global().Get("Float32Array").New(len(data))
	for i, v := range data {
		jsArray.SetIndex(i, v)
	}
	return jsArray
}

// uint32ArrayToJS converts a []uint32 to a JavaScript Uint32Array
func (r *Renderer) uint32ArrayToJS(data []uint32) js.Value {
	jsArray := js.Global().Get("Uint32Array").New(len(data))
	for i, v := range data {
		jsArray.SetIndex(i, v)
	}
	return jsArray
}
