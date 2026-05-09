package opengl

import (
	"fmt"
	"log"
	"math"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tesselstudio/TesselBox/pkg/world"
)

// Engine represents the OpenGL rendering engine
type Engine struct {
	window        *glfw.Window
	shaderProgram uint32
	vao, vbo      uint32
	camera        *Camera
	initialized   bool
	meshRenderer  *MeshRenderer
	chunkMeshes   map[string]*ChunkMeshData
	frameCounter  int64
}

// Camera represents a 3D camera
type Camera struct {
	Position    mgl32.Vec3
	Target      mgl32.Vec3
	Up          mgl32.Vec3
	FOV         float32
	AspectRatio float32
	Near        float32
	Far         float32
}

// NewEngine creates a new OpenGL engine
func NewEngine(width, height int, title string) (*Engine, error) {
	println("🔧 Initializing OpenGL engine...")

	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		println("❌ Failed to initialize GLFW:", err.Error())
		return nil, err
	}
	println("✅ GLFW initialized successfully")

	// Configure GLFW
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// Create window
	window, err := glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		println("❌ Failed to create GLFW window:", err.Error())
		glfw.Terminate()
		return nil, err
	}
	println("✅ GLFW window created successfully")
	window.MakeContextCurrent()
	println("✅ OpenGL context made current")

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		println("❌ Failed to initialize OpenGL:", err.Error())
		return nil, err
	}
	println("✅ OpenGL initialized successfully")

	// Set viewport
	gl.Viewport(0, 0, int32(width), int32(height))
	println("✅ Viewport set to", width, "x", height)

	// Enable depth testing
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	// Set clear color
	gl.ClearColor(0.1, 0.1, 0.2, 1.0)
	println("✅ OpenGL state configured")

	engine := &Engine{
		window: window,
		camera: &Camera{
			Position:    mgl32.Vec3{0, 0, 5},
			Target:      mgl32.Vec3{0, 0, 0},
			Up:          mgl32.Vec3{0, 1, 0},
			FOV:         45.0,
			AspectRatio: float32(width) / float32(height),
			Near:        0.1,
			Far:         100.0,
		},
		meshRenderer: NewMeshRenderer(),
		chunkMeshes:  make(map[string]*ChunkMeshData),
	}

	// Initialize OpenGL resources
	if err := engine.initOpenGL(); err != nil {
		println("❌ Failed to initialize OpenGL resources:", err.Error())
		return nil, err
	}

	println("🎮 OpenGL engine initialized successfully!")
	return engine, nil
}

// initOpenGL initializes OpenGL resources
func (e *Engine) initOpenGL() error {
	// Vertex shader source
	vertexShaderSource := `
		#version 410 core
		layout (location = 0) in vec3 aPos;
		layout (location = 1) in vec3 aColor;
		
		uniform mat4 model;
		uniform mat4 view;
		uniform mat4 projection;
		
		out vec3 vertexColor;
		
		void main() {
			gl_Position = projection * view * model * vec4(aPos, 1.0);
			vertexColor = aColor;
		}
	`

	// Fragment shader source
	fragmentShaderSource := `
		#version 410 core
		in vec3 vertexColor;
		out vec4 FragColor;
		
		void main() {
			FragColor = vec4(vertexColor, 1.0);
		}
	`

	// Compile shaders
	vertexShader, err := e.compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	fragmentShader, err := e.compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	// Create shader program
	e.shaderProgram = gl.CreateProgram()
	gl.AttachShader(e.shaderProgram, vertexShader)
	gl.AttachShader(e.shaderProgram, fragmentShader)
	gl.LinkProgram(e.shaderProgram)

	// Check for linking errors
	var success int32
	gl.GetProgramiv(e.shaderProgram, gl.LINK_STATUS, &success)
	if success == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(e.shaderProgram, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength)
		gl.GetProgramInfoLog(e.shaderProgram, logLength, nil, &logMsg[0])
		errorMsg := string(logMsg)
		log.Printf("❌ Shader program linking failed: %s", errorMsg)
		println("❌ CRITICAL: Shader program linking failed:", errorMsg)
		return fmt.Errorf("shader program linking failed: %s", errorMsg)
	}

	// Clean up shaders
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	// Create cube vertices with colors
	vertices := []float32{
		// Front face - Red
		-0.5, -0.5, 0.5, 1.0, 0.0, 0.0,
		0.5, -0.5, 0.5, 1.0, 0.0, 0.0,
		0.5, 0.5, 0.5, 1.0, 0.0, 0.0,
		-0.5, 0.5, 0.5, 1.0, 0.0, 0.0,
		// Back face - Green
		-0.5, -0.5, -0.5, 0.0, 1.0, 0.0,
		-0.5, 0.5, -0.5, 0.0, 1.0, 0.0,
		0.5, 0.5, -0.5, 0.0, 1.0, 0.0,
		0.5, -0.5, -0.5, 0.0, 1.0, 0.0,
		// Top face - Blue
		-0.5, 0.5, -0.5, 0.0, 0.0, 1.0,
		-0.5, 0.5, 0.5, 0.0, 0.0, 1.0,
		0.5, 0.5, 0.5, 0.0, 0.0, 1.0,
		0.5, 0.5, -0.5, 0.0, 0.0, 1.0,
		// Bottom face - Yellow
		-0.5, -0.5, -0.5, 1.0, 1.0, 0.0,
		0.5, -0.5, -0.5, 1.0, 1.0, 0.0,
		0.5, -0.5, 0.5, 1.0, 1.0, 0.0,
		-0.5, -0.5, 0.5, 1.0, 1.0, 0.0,
		// Right face - Magenta
		0.5, -0.5, -0.5, 1.0, 0.0, 1.0,
		0.5, 0.5, -0.5, 1.0, 0.0, 1.0,
		0.5, 0.5, 0.5, 1.0, 0.0, 1.0,
		0.5, -0.5, 0.5, 1.0, 0.0, 1.0,
		// Left face - Cyan
		-0.5, -0.5, -0.5, 0.0, 1.0, 1.0,
		-0.5, -0.5, 0.5, 0.0, 1.0, 1.0,
		-0.5, 0.5, 0.5, 0.0, 1.0, 1.0,
		-0.5, 0.5, -0.5, 0.0, 1.0, 1.0,
	}

	// Create VAO and VBO
	gl.GenVertexArrays(1, &e.vao)
	gl.GenBuffers(1, &e.vbo)

	// Bind VAO
	gl.BindVertexArray(e.vao)

	// Bind and fill VBO
	gl.BindBuffer(gl.ARRAY_BUFFER, e.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Set vertex attributes
	// Position attribute
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, nil)
	gl.EnableVertexAttribArray(0)
	// Color attribute
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	// Unbind
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	e.initialized = true
	return nil
}

// compileShader compiles a shader from source
func (e *Engine) compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csource, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csource, nil)
	free()
	gl.CompileShader(shader)

	// Check for compilation errors
	var success int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &success)
	if success == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength)
		gl.GetShaderInfoLog(shader, logLength, nil, &logMsg[0])
		log.Printf("Shader compilation failed: %s", string(logMsg))
		return 0, nil
	}

	return shader, nil
}

// BeginFrame prepares for rendering a new frame
func (e *Engine) BeginFrame() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

// EndFrame completes the frame and swaps buffers
func (e *Engine) EndFrame() {
	e.window.SwapBuffers()
}

// Render renders chunk meshes and game elements
func (e *Engine) Render(gameController interface{}) {
	if !e.initialized {
		return
	}

	// Debug: Print camera position every 60 frames
	if e.frameCounter%60 == 0 {
		println("🔍 DEBUG: Camera Position:", e.camera.Position[0], e.camera.Position[1], e.camera.Position[2])
		println("🔍 DEBUG: Camera Target:", e.camera.Target[0], e.camera.Target[1], e.camera.Target[2])
	}
	e.frameCounter++

	// Check for OpenGL errors
	if err := gl.GetError(); err != gl.NO_ERROR {
		println("❌ OpenGL Error before render:", err)
		return
	}

	// Use shader program
	gl.UseProgram(e.shaderProgram)

	// Create transformation matrices
	model := mgl32.Ident4()
	view := mgl32.LookAtV(e.camera.Position, e.camera.Target, e.camera.Up)
	projection := mgl32.Perspective(mgl32.DegToRad(e.camera.FOV), e.camera.AspectRatio, e.camera.Near, e.camera.Far)

	// Set uniforms
	modelLoc := gl.GetUniformLocation(e.shaderProgram, gl.Str("model\x00"))
	viewLoc := gl.GetUniformLocation(e.shaderProgram, gl.Str("view\x00"))
	projLoc := gl.GetUniformLocation(e.shaderProgram, gl.Str("projection\x00"))

	gl.UniformMatrix4fv(modelLoc, 1, false, &model[0])
	gl.UniformMatrix4fv(viewLoc, 1, false, &view[0])
	gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])

	// Test cube rendering disabled due to crash
	// TODO: Fix test cube rendering later

	// Render all chunk meshes with error checking
	if err := gl.GetError(); err != gl.NO_ERROR {
		println("❌ OpenGL Error before mesh render:", err)
		gl.UseProgram(0)
		return
	}

	e.meshRenderer.RenderAllMeshes()

	// Check for errors after rendering
	if err := gl.GetError(); err != gl.NO_ERROR {
		println("❌ OpenGL Error after mesh render:", err)
	}

	// Unbind
	gl.UseProgram(0)

	// Render HUD if game controller is provided
	if gameController != nil {
		// Get window size for HUD rendering
		width, height := e.window.GetSize()

		// Type assert to get controller methods
		type Renderer interface {
			Render(int, int)
		}
		if controller, ok := gameController.(Renderer); ok {
			controller.Render(int(width), int(height))
		}
	}
}

// renderTestCube renders a simple colored cube at the origin for debugging
func (e *Engine) renderTestCube() {
	// Use the existing VAO/VBO from initialization
	gl.BindVertexArray(e.vao)

	// Create a simple transformation matrix for the test cube
	model := mgl32.Translate3D(0, 0, -10) // Place 10 units in front of camera

	// Update the model uniform
	modelLoc := gl.GetUniformLocation(e.shaderProgram, gl.Str("model\x00"))
	gl.UniformMatrix4fv(modelLoc, 1, false, &model[0])

	// Draw the test cube (36 indices for 12 triangles)
	gl.DrawElements(gl.TRIANGLES, 36, gl.UNSIGNED_INT, nil)

	println("🔍 DEBUG: Rendered test cube at origin")

	// Reset model matrix to identity
	identityMatrix := mgl32.Ident4()
	gl.UniformMatrix4fv(modelLoc, 1, false, &identityMatrix[0])
}

// ShouldClose returns true if the window should close
func (e *Engine) ShouldClose() bool {
	return e.window.ShouldClose()
}

// PollEvents processes pending events
func (e *Engine) PollEvents() {
	glfw.PollEvents()
}

// Cleanup cleans up OpenGL resources
func (e *Engine) Cleanup() {
	if e.initialized {
		gl.DeleteVertexArrays(1, &e.vao)
		gl.DeleteBuffers(1, &e.vbo)
		gl.DeleteProgram(e.shaderProgram)
	}
	if e.window != nil {
		e.window.Destroy()
	}
	glfw.Terminate()
}

// GetWindow returns the GLFW window
func (e *Engine) GetWindow() *glfw.Window {
	return e.window
}

// GetLoadedMeshCount returns the number of loaded chunk meshes
func (e *Engine) GetLoadedMeshCount() int {
	return e.meshRenderer.GetLoadedMeshCount()
}

// SetCameraPosition updates the camera position
func (e *Engine) SetCameraPosition(pos mgl32.Vec3) {
	e.camera.Position = pos
}

// SetCameraTarget updates the camera target
func (e *Engine) SetCameraTarget(target mgl32.Vec3) {
	e.camera.Target = target
}

// AddChunkMesh adds a chunk mesh to rendering system (accepts interface for compatibility)
func (e *Engine) AddChunkMesh(coord world.ChunkCoord, meshData interface{}) {
	println("🔍 DEBUG: AddChunkMesh called for chunk", coord.X, coord.Z)

	if meshData == nil {
		println("❌ DEBUG: meshData is nil for chunk", coord.X, coord.Z)
		return
	}

	// Type assert to get mesh data
	if data, ok := meshData.(map[string]interface{}); ok {
		vertices, verticesOk := data["Vertices"].([]float32)
		indices, indicesOk := data["Indices"].([]uint32)
		vertexCount, vertexOk := data["VertexCount"].(int32)
		indexCount, indexOk := data["IndexCount"].(int32)

		println("🔍 DEBUG: Mesh data extraction - Vertices OK:", verticesOk,
			"Indices OK:", indicesOk, "VertexCount OK:", vertexOk, "IndexCount OK:", indexOk)

		if vertices != nil && indices != nil {
			println("🔍 DEBUG: Creating ChunkMeshData - Vertices:", len(vertices),
				"Indices:", len(indices), "VertexCount:", vertexCount, "IndexCount:", indexCount)

			chunkMesh := &ChunkMeshData{
				Vertices:    vertices,
				Indices:     indices,
				VertexCount: vertexCount,
				IndexCount:  indexCount,
			}
			e.meshRenderer.AddMesh(coord, chunkMesh)
			println("✅ DEBUG: Successfully added mesh to renderer for chunk", coord.X, coord.Z)
		} else {
			println("❌ DEBUG: Failed to extract vertex/index data for chunk", coord.X, coord.Z)
			if vertices == nil {
				println("  - Vertices is nil")
			}
			if indices == nil {
				println("  - Indices is nil")
			}
		}
	} else {
		println("❌ DEBUG: Type assertion failed for chunk", coord.X, coord.Z)
	}
}

// RemoveChunkMesh removes a chunk mesh from rendering
func (e *Engine) RemoveChunkMesh(coord world.ChunkCoord) {
	e.meshRenderer.RemoveMesh(coord)
}

// UpdateCameraFromPlayer updates camera position/rotation from player
func (e *Engine) UpdateCameraFromPlayer(playerPos, playerRot mgl32.Vec3) {
	// Camera is at player position + eye height, offset backwards for first-person view
	eyeHeight := float32(1.62) // Player eye height
	e.camera.Position = mgl32.Vec3{playerPos.X(), playerPos.Y() + eyeHeight, playerPos.Z()}
	e.camera.Target = mgl32.Vec3{playerPos.X(), playerPos.Y() + eyeHeight, playerPos.Z()}

	// Apply pitch and yaw rotation
	pitchRad := float32(playerRot.X() * 3.14159 / 180.0)
	yawRad := float32(playerRot.Y() * 3.14159 / 180.0)

	// Calculate look direction
	forwardX := float32(-math.Sin(float64(yawRad)) * math.Cos(float64(pitchRad)))
	forwardY := float32(math.Sin(float64(pitchRad)))
	forwardZ := float32(-math.Cos(float64(yawRad)) * math.Cos(float64(pitchRad)))

	e.camera.Target = mgl32.Vec3{
		e.camera.Position.X() + forwardX,
		e.camera.Position.Y() + forwardY,
		e.camera.Position.Z() + forwardZ,
	}
}

func init() {
	// This is needed to ensure that main runs on the main thread
	runtime.LockOSThread()
}
