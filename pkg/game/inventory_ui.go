package game

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tesselstudio/TesselBox/pkg/crafting"
	"github.com/tesselstudio/TesselBox/pkg/player"
)

// InventoryUI handles the in-game inventory interface
type InventoryUI struct {
	// Player reference
	player *player.Player

	// OpenGL resources
	inventoryVAO    uint32
	inventoryVBO    uint32
	inventoryShader uint32

	// UI state
	visible      bool
	selectedSlot int
	draggedSlot  int
	dragActive   bool
}

// NewInventoryUI creates a new inventory UI system
func NewInventoryUI(p *player.Player) *InventoryUI {
	ui := &InventoryUI{
		player:       p,
		visible:      false,
		selectedSlot: 0,
		draggedSlot:  -1,
		dragActive:   false,
	}

	ui.initOpenGL()
	return ui
}

// initOpenGL initializes OpenGL resources for inventory UI
func (ui *InventoryUI) initOpenGL() {
	// Create inventory shader
	vertexShader := `
		#version 410 core
		layout (location = 0) in vec2 aPos;
		layout (location = 1) in vec2 aTexCoord;
		uniform mat4 projection;
		out vec2 TexCoord;
		
		void main() {
			TexCoord = aTexCoord;
			gl_Position = projection * vec4(aPos, 0.0, 1.0);
		}
	`

	fragmentShader := `
		#version 410 core
		in vec2 TexCoord;
		out vec4 FragColor;
		uniform sampler2D texture1;
		uniform vec4 color;
		uniform bool hasTexture;
		
		void main() {
			if (hasTexture) {
				FragColor = texture(texture1, TexCoord) * color;
			} else {
				FragColor = color;
			}
		}
	`

	ui.inventoryShader = ui.createShaderProgram(vertexShader, fragmentShader)

	// Create inventory VAO/VBO
	ui.createInventoryGeometry()
}

// createShaderProgram creates and compiles shaders
func (ui *InventoryUI) createShaderProgram(vertexSource, fragmentSource string) uint32 {
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

	// Check linking status
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status != gl.TRUE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		logBytes := make([]byte, logLength+1)
		gl.GetProgramInfoLog(program, logLength, nil, &logBytes[0])
		gl.DeleteProgram(program)
		return 0
	}

	// Clean up shaders
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program
}

// createInventoryGeometry creates the inventory UI geometry
func (ui *InventoryUI) createInventoryGeometry() {
	// Inventory grid: 9 rows x 9 columns = 81 slots
	slotWidth := float32(40.0)
	slotHeight := float32(40.0)
	padding := float32(2.0)

	vertices := make([]float32, 0)
	indices := make([]uint32, 0)

	// Generate inventory slots
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			x := float32(col) * (slotWidth + padding)
			y := float32(row) * (slotHeight + padding)

			// Create quad for slot
			slotVertices := []float32{
				x, y,
				x + slotWidth, y,
				x + slotWidth, y + slotHeight,
				x, y + slotHeight,
			}

			// Add texture coordinates
			texCoords := []float32{
				0.0, 0.0, // Top-left
				1.0, 0.0, // Top-right
				1.0, 1.0, // Bottom-right
				0.0, 1.0, // Bottom-left
			}

			// Add vertices with texture coordinates
			baseIndex := len(vertices) / 6
			for i := 0; i < 4; i++ {
				vertices = append(vertices, slotVertices[i*2], slotVertices[i*2+1])
				vertices = append(vertices, texCoords[i*2], texCoords[i*2+1])
			}

			// Add indices for this quad
			quadIndices := []uint32{
				uint32(baseIndex + 0),
				uint32(baseIndex + 1),
				uint32(baseIndex + 2),
				uint32(baseIndex + 3),
			}

			// Offset indices by current vertex count
			for _, idx := range quadIndices {
				indices = append(indices, idx+uint32(baseIndex))
			}
		}
	}

	// Create VAO and VBO
	gl.GenVertexArrays(1, &ui.inventoryVAO)
	gl.GenBuffers(1, &ui.inventoryVBO)

	// Bind and fill VBO
	gl.BindVertexArray(ui.inventoryVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, ui.inventoryVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Set vertex attributes
	// Position (location 0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	// Texture coordinates (location 1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)

	// Create and fill EBO
	var ebo uint32
	gl.GenBuffers(1, &ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// Unbind
	gl.BindVertexArray(0)
}

// Show displays the inventory UI
func (ui *InventoryUI) Show() {
	ui.visible = true
}

// Hide hides the inventory UI
func (ui *InventoryUI) Hide() {
	ui.visible = false
}

// IsVisible returns whether the inventory UI is visible
func (ui *InventoryUI) IsVisible() bool {
	return ui.visible
}

// Render renders the inventory UI
func (ui *InventoryUI) Render(width, height int) {
	if !ui.visible || ui.player == nil {
		return
	}

	// Use inventory shader
	gl.UseProgram(ui.inventoryShader)

	// Set up orthographic projection for UI
	widthf := float32(width)
	heightf := float32(height)
	projection := mgl32.Ortho(0, widthf, heightf, -1, 1, 1)
	projectionUniform := gl.GetUniformLocation(ui.inventoryShader, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	// Enable transparency
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// Bind inventory VAO
	gl.BindVertexArray(ui.inventoryVAO)

	// Render inventory background
	ui.renderInventoryBackground(width, height)

	// Render inventory slots
	ui.renderInventorySlots()

	// Render selected item
	ui.renderSelectedItem()

	// Disable transparency
	gl.Disable(gl.BLEND)
}

// renderInventoryBackground renders the inventory background
func (ui *InventoryUI) renderInventoryBackground(width, height int) {
	// Background color
	colorUniform := gl.GetUniformLocation(ui.inventoryShader, gl.Str("color\x00"))
	hasTextureUniform := gl.GetUniformLocation(ui.inventoryShader, gl.Str("hasTexture\x00"))

	gl.Uniform4f(colorUniform, 0.1, 0.1, 0.1, 0.9) // Dark gray
	gl.Uniform1i(hasTextureUniform, 0)             // No texture

	// Background quad covering most of screen
	bgVertices := []float32{
		50.0, 50.0, // Top-left
		float32(width) - 50.0, 50.0, // Top-right
		float32(width) - 50.0, float32(height) - 50.0, // Bottom-right
		50.0, float32(height) - 50.0, // Bottom-left
	}

	// Temporary VAO/VBO for background
	var bgVAO, bgVBO uint32
	gl.GenVertexArrays(1, &bgVAO)
	gl.GenBuffers(1, &bgVBO)

	gl.BindVertexArray(bgVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, bgVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(bgVertices)*4, gl.Ptr(bgVertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

	// Cleanup
	gl.DeleteVertexArrays(1, &bgVAO)
	gl.DeleteBuffers(1, &bgVBO)
}

// ...
func (ui *InventoryUI) renderInventorySlots() {
	inventory := ui.player.GetInventory()
	if inventory == nil {
		return
	}

	// Render each slot
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			slotIndex := row*9 + col
			item := inventory.GetSlot(slotIndex)

			if item != nil {
				ui.renderItem(col, row, item)
			} else {
				ui.renderEmptySlot(col, row)
			}
		}
	}
}

// renderItem renders an item in an inventory slot
func (ui *InventoryUI) renderItem(col, row int, itemStack *crafting.ItemStack) {
	// This would render item sprite
	// For now, render a colored rectangle
	colorUniform := gl.GetUniformLocation(ui.inventoryShader, gl.Str("color\x00"))
	hasTextureUniform := gl.GetUniformLocation(ui.inventoryShader, gl.Str("hasTexture\x00"))

	// Item color based on type
	if itemStack != nil {
		switch itemStack.Item.ID {
		case "wood":
			gl.Uniform4f(colorUniform, 0.6, 0.4, 0.2, 1.0) // Brown
		case "stone":
			gl.Uniform4f(colorUniform, 0.5, 0.5, 0.5, 1.0) // Gray
		case "dirt":
			gl.Uniform4f(colorUniform, 0.4, 0.3, 0.1, 1.0) // Brown
		default:
			gl.Uniform4f(colorUniform, 0.8, 0.8, 0.2, 1.0) // Light green
		}
	} else {
		gl.Uniform4f(colorUniform, 0.2, 0.2, 0.2, 0.5) // Dark gray for empty
	}

	gl.Uniform1i(hasTextureUniform, 0) // No texture

	// Calculate vertex offset (6 vertices per slot: pos + texcoord)
	vertexOffset := (col + row*9) * 6 * 4

	gl.DrawArrays(gl.TRIANGLE_FAN, int32(vertexOffset), 4)
}

// renderEmptySlot renders an empty inventory slot
func (ui *InventoryUI) renderEmptySlot(col, row int) {
	colorUniform := gl.GetUniformLocation(ui.inventoryShader, gl.Str("color\x00"))
	hasTextureUniform := gl.GetUniformLocation(ui.inventoryShader, gl.Str("hasTexture\x00"))

	gl.Uniform4f(colorUniform, 0.2, 0.2, 0.2, 0.3) // Dark gray
	gl.Uniform1i(hasTextureUniform, 0)             // No texture

	// Calculate vertex offset
	vertexOffset := (col + row*9) * 6 * 4

	gl.DrawArrays(gl.TRIANGLE_FAN, int32(vertexOffset), 4)
}

// renderSelectedItem renders highlight around selected slot
func (ui *InventoryUI) renderSelectedItem() {
	hotbarSlot := ui.player.GetHotbarSlot()
	if hotbarSlot < 0 || hotbarSlot > 8 {
		return
	}

	row := hotbarSlot / 9
	col := hotbarSlot % 9

	colorUniform := gl.GetUniformLocation(ui.inventoryShader, gl.Str("color\x00"))
	hasTextureUniform := gl.GetUniformLocation(ui.inventoryShader, gl.Str("hasTexture\x00"))

	gl.Uniform4f(colorUniform, 1.0, 1.0, 0.0, 0.5) // Yellow highlight
	gl.Uniform1i(hasTextureUniform, 0)             // No texture

	// Calculate slot position
	slotWidth := float32(40.0)
	slotHeight := float32(40.0)
	padding := float32(2.0)
	posX := 100.0 + float32(col)*(slotWidth+padding)
	posY := 100.0 + float32(row)*(slotHeight+padding)

	// Render selection border (slightly larger)
	borderWidth := slotWidth + 4.0
	borderHeight := slotHeight + 4.0
	borderVertices := []float32{
		posX - 2.0, posY - 2.0,
		posX + borderWidth, posY - 2.0,
		posX + borderWidth, posY + borderHeight - 2.0,
		posX - 2.0, posY + borderHeight - 2.0,
	}

	// Temporary VAO/VBO for border
	var borderVAO, borderVBO uint32
	gl.GenVertexArrays(1, &borderVAO)
	gl.GenBuffers(1, &borderVBO)

	gl.BindVertexArray(borderVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, borderVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(borderVertices)*4, gl.Ptr(borderVertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

	// Cleanup
	gl.DeleteVertexArrays(1, &borderVAO)
	gl.DeleteBuffers(1, &borderVBO)
}

// HandleInput handles input when inventory is open
func (ui *InventoryUI) HandleInput(keyCode int) bool {
	if !ui.visible {
		return false
	}

	// Handle ESC to close inventory
	if keyCode == 256 { // ESC key
		ui.Hide()
		return true
	}

	// Handle number keys for hotbar selection
	if keyCode >= 49 && keyCode <= 57 { // 1-9 keys
		newSlot := keyCode - 49
		if newSlot >= 0 && newSlot <= 8 {
			ui.player.SetHotbarSlot(newSlot)
			ui.selectedSlot = newSlot
		}
	}

	return false
}

// Cleanup cleans up OpenGL resources
func (ui *InventoryUI) Cleanup() {
	gl.DeleteVertexArrays(1, &ui.inventoryVAO)
	gl.DeleteBuffers(1, &ui.inventoryVBO)
	gl.DeleteProgram(ui.inventoryShader)
}
