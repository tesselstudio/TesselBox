package game

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tesselstudio/TesselBox/pkg/crafting"
	"github.com/tesselstudio/TesselBox/pkg/player"
	"github.com/tesselstudio/TesselBox/pkg/survival"
)

// HUD represents heads-up display system
type HUD struct {
	// Player references
	player      *player.Player
	playerStats *survival.PlayerStats
	inventory   *crafting.Inventory

	// OpenGL resources
	crosshairVAO uint32
	crosshairVBO uint32
	hotbarVAO    uint32
	hotbarVBO    uint32
	healthBarVAO uint32
	healthBarVBO uint32
	hungerBarVAO uint32
	hungerBarVBO uint32

	// Shader programs
	hudShader uint32

	// HUD state
	visible       bool
	showInventory bool
	showCrafting  bool
	selectedSlot  int
}

// NewHUD creates a new HUD system
func NewHUD(p *player.Player) *HUD {
	hud := &HUD{
		player:       p,
		playerStats:  p.GetStats(),
		inventory:    p.GetInventory(),
		visible:      true,
		selectedSlot: 0,
	}

	// Initialize OpenGL resources
	hud.initOpenGL()

	return hud
}

// initOpenGL initializes OpenGL resources for HUD
func (h *HUD) initOpenGL() {
	// Create simple shader for HUD rendering
	vertexShader := `
		#version 410 core
		layout (location = 0) in vec2 aPos;
		layout (location = 1) in vec2 aTexCoord;
		out vec2 TexCoord;
		
		uniform mat4 projection;
		
		void main() {
			TexCoord = aTexCoord;
			gl_Position = projection * vec4(aPos, 0.0, 1.0);
		}
	`

	fragmentShader := `
		#version 410 core
		in vec2 TexCoord;
		out vec4 FragColor;
		
		uniform vec4 color;
		uniform sampler2D texture1;
		uniform bool useTexture;
		
		void main() {
			if (useTexture) {
				FragColor = texture(texture1, TexCoord) * color;
			} else {
				FragColor = color;
			}
		}
	`

	h.hudShader = h.createShaderProgram(vertexShader, fragmentShader)

	// Create crosshair
	h.createCrosshair()

	// Create hotbar
	h.createHotbar()

	// Create status bars
	h.createStatusBars()
}

// createShaderProgram compiles and links shaders
func (h *HUD) createShaderProgram(vertexSource, fragmentSource string) uint32 {
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	cstr, free := gl.Strs(vertexSource)
	gl.ShaderSource(vertexShader, 1, cstr, nil)
	gl.CompileShader(vertexShader)
	free()

	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	cstr, free = gl.Strs(fragmentSource)
	gl.ShaderSource(fragmentShader, 1, cstr, nil)
	gl.CompileShader(fragmentShader)
	free()

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	// Clean up shaders
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program
}

// createCrosshair creates the crosshair geometry
func (h *HUD) createCrosshair() {
	vertices := []float32{
		// Center lines
		-10.0, 0.0, // Left
		10.0, 0.0, // Right
		0.0, -10.0, // Top
		0.0, 10.0, // Bottom
	}

	gl.GenVertexArrays(1, &h.crosshairVAO)
	gl.GenBuffers(1, &h.crosshairVBO)

	gl.BindVertexArray(h.crosshairVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, h.crosshairVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Position attribute
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 2*4, gl.PtrOffset(0))

	gl.BindVertexArray(0)
}

// createHotbar creates the hotbar geometry
func (h *HUD) createHotbar() {
	// Create hotbar background and slots
	vertices := make([]float32, 0)
	indices := make([]uint32, 0)

	slotWidth := float32(40.0)
	padding := float32(2.0)
	hotbarWidth := float32(10) * (slotWidth + padding)

	// Hotbar background
	bgVertices := []float32{
		float32(-10.0), float32(-50.0),
		hotbarWidth, float32(-50.0),
		hotbarWidth, float32(-10.0),
		float32(-10.0), float32(-10.0),
	}

	vertices = append(vertices, bgVertices...)

	// Create slot backgrounds
	for i := 0; i < 10; i++ {
		x := float32(i) * (slotWidth + padding)

		slotVertices := []float32{
			x, float32(-50.0),
			x + slotWidth, float32(-50.0),
			x + slotWidth, float32(-10.0),
			x, float32(-10.0),
		}

		vertices = append(vertices, slotVertices...)

		baseIndex := uint32(len(vertices)/8 - 4)
		slotIndices := []uint32{
			baseIndex, baseIndex + 1, baseIndex + 2,
			baseIndex, baseIndex + 2, baseIndex + 3,
		}
		indices = append(indices, slotIndices...)
	}

	gl.GenVertexArrays(1, &h.hotbarVAO)
	gl.GenBuffers(1, &h.hotbarVBO)

	gl.BindVertexArray(h.hotbarVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, h.hotbarVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 2*4, gl.PtrOffset(0))

	gl.BindVertexArray(0)
}

// createStatusBars creates health and hunger bars
func (h *HUD) createStatusBars() {
	// Health bar vertices
	healthVertices := []float32{
		10.0, 60.0, // Background
		210.0, 60.0,
		210.0, 80.0,
		10.0, 80.0,

		10.0, 60.0, // Health fill
		110.0, 60.0, // 50% health
		110.0, 80.0,
		10.0, 80.0,
	}

	// Hunger bar vertices
	hungerVertices := []float32{
		10.0, 90.0, // Background
		210.0, 90.0,
		210.0, 110.0,
		10.0, 110.0,

		10.0, 90.0, // Hunger fill
		160.0, 90.0, // 75% hunger
		160.0, 110.0,
		10.0, 110.0,
	}

	// Create health bar VAO/VBO
	gl.GenVertexArrays(1, &h.healthBarVAO)
	gl.GenBuffers(1, &h.healthBarVBO)

	gl.BindVertexArray(h.healthBarVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, h.healthBarVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(healthVertices)*4, gl.Ptr(healthVertices), gl.STATIC_DRAW)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 2*4, gl.PtrOffset(0))

	// Create hunger bar VAO/VBO
	gl.GenVertexArrays(1, &h.hungerBarVAO)
	gl.GenBuffers(1, &h.hungerBarVBO)

	gl.BindVertexArray(h.hungerBarVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, h.hungerBarVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(hungerVertices)*4, gl.Ptr(hungerVertices), gl.STATIC_DRAW)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 2*4, gl.PtrOffset(0))

	gl.BindVertexArray(0)
}

// Render renders the HUD
func (h *HUD) Render(width, height int) {
	if !h.visible {
		return
	}

	// Set up orthographic projection for 2D rendering
	projection := mgl32.Ortho(0, float32(width), float32(height), 0, -1, 1)

	gl.UseProgram(h.hudShader)
	projectionUniform := gl.GetUniformLocation(h.hudShader, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	// Disable depth testing for UI
	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// Render crosshair
	h.renderCrosshair()

	// Render hotbar
	h.renderHotbar()

	// Render status bars
	h.renderStatusBars()

	// Re-enable depth testing
	gl.Enable(gl.DEPTH_TEST)
	gl.Disable(gl.BLEND)
}

// renderCrosshair renders the crosshair
func (h *HUD) renderCrosshair() {
	colorUniform := gl.GetUniformLocation(h.hudShader, gl.Str("color\x00"))
	useTextureUniform := gl.GetUniformLocation(h.hudShader, gl.Str("useTexture\x00"))

	gl.Uniform4f(colorUniform, 1.0, 1.0, 1.0, 1.0) // White
	gl.Uniform1i(useTextureUniform, 0)             // No texture

	gl.BindVertexArray(h.crosshairVAO)
	gl.DrawArrays(gl.LINES, 0, 4)
	gl.BindVertexArray(0)
}

// renderHotbar renders the hotbar
func (h *HUD) renderHotbar() {
	colorUniform := gl.GetUniformLocation(h.hudShader, gl.Str("color\x00"))
	useTextureUniform := gl.GetUniformLocation(h.hudShader, gl.Str("useTexture\x00"))

	// Render hotbar background
	gl.Uniform4f(colorUniform, 0.2, 0.2, 0.2, 0.8) // Dark gray
	gl.Uniform1i(useTextureUniform, 0)             // No texture

	gl.BindVertexArray(h.hotbarVAO)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4) // Background

	// Render slot backgrounds
	for i := 0; i < 10; i++ {
		if i == h.selectedSlot {
			gl.Uniform4f(colorUniform, 0.8, 0.8, 0.2, 0.8) // Yellow for selected
		} else {
			gl.Uniform4f(colorUniform, 0.3, 0.3, 0.3, 0.8) // Gray for unselected
		}

		gl.DrawArrays(gl.TRIANGLE_FAN, int32(4+i*4), 4)
	}

	gl.BindVertexArray(0)
}

// renderStatusBars renders health and hunger bars
func (h *HUD) renderStatusBars() {
	colorUniform := gl.GetUniformLocation(h.hudShader, gl.Str("color\x00"))
	useTextureUniform := gl.GetUniformLocation(h.hudShader, gl.Str("useTexture\x00"))

	if h.playerStats != nil {
		// Render health bar background
		gl.Uniform4f(colorUniform, 0.2, 0.2, 0.2, 0.8) // Dark gray background
		gl.Uniform1i(useTextureUniform, 0)

		gl.BindVertexArray(h.healthBarVAO)
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

		// Render health bar fill
		healthPercent := h.playerStats.Health.GetHealthPercentage()
		if healthPercent > 0 {
			if healthPercent > 0.6 {
				gl.Uniform4f(colorUniform, 0.2, 0.8, 0.2, 0.9) // Green
			} else if healthPercent > 0.3 {
				gl.Uniform4f(colorUniform, 0.8, 0.8, 0.2, 0.9) // Yellow
			} else {
				gl.Uniform4f(colorUniform, 0.8, 0.2, 0.2, 0.9) // Red
			}

			gl.DrawArrays(gl.TRIANGLE_FAN, 4, 4)
		}

		// Render hunger bar background
		gl.Uniform4f(colorUniform, 0.2, 0.2, 0.2, 0.8) // Dark gray background
		gl.BindVertexArray(h.hungerBarVAO)
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

		// Render hunger bar fill
		hungerPercent := h.playerStats.Hunger.GetHungerPercentage()
		if hungerPercent > 0 {
			gl.Uniform4f(colorUniform, 0.8, 0.6, 0.2, 0.9) // Orange
			gl.DrawArrays(gl.TRIANGLE_FAN, 4, 4)
		}

		gl.BindVertexArray(0)
	}
}

// Update updates HUD state
func (h *HUD) Update() {
	if h.playerStats != nil {
		// Update selected slot from player
		h.selectedSlot = h.player.GetHotbarSlot()
	}
}

// SetVisible sets HUD visibility
func (h *HUD) SetVisible(visible bool) {
	h.visible = visible
}

// IsVisible returns HUD visibility
func (h *HUD) IsVisible() bool {
	return h.visible
}

// Cleanup cleans up OpenGL resources
func (h *HUD) Cleanup() {
	gl.DeleteVertexArrays(1, &h.crosshairVAO)
	gl.DeleteBuffers(1, &h.crosshairVBO)
	gl.DeleteVertexArrays(1, &h.hotbarVAO)
	gl.DeleteBuffers(1, &h.hotbarVBO)
	gl.DeleteVertexArrays(1, &h.healthBarVAO)
	gl.DeleteBuffers(1, &h.healthBarVBO)
	gl.DeleteVertexArrays(1, &h.hungerBarVAO)
	gl.DeleteBuffers(1, &h.hungerBarVBO)
	gl.DeleteProgram(h.hudShader)
}
