package opengl_ui

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// UIRenderer handles rendering of UI elements using OpenGL
type UIRenderer struct {
	window           *glfw.Window
	shaderProgram    uint32
	vao, vbo         uint32
	projection       mgl32.Mat4
	width, height    int
	initialized      bool
	charVAO, charVBO uint32
	fontTexture      uint32
}

// NewUIRenderer creates a new UI renderer
func NewUIRenderer(window *glfw.Window) (*UIRenderer, error) {
	renderer := &UIRenderer{
		window: window,
	}

	if err := renderer.init(); err != nil {
		return nil, err
	}

	return renderer, nil
}

// init initializes the OpenGL UI renderer
func (r *UIRenderer) init() error {
	// Get window dimensions
	r.width, r.height = r.window.GetSize()

	// Set up orthographic projection for 2D UI
	r.projection = mgl32.Ortho(0, float32(r.width), float32(r.height), 0, -1, 1)

	// Create and compile shaders
	vertexShader, err := r.compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return fmt.Errorf("failed to compile vertex shader: %w", err)
	}

	fragmentShader, err := r.compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return fmt.Errorf("failed to compile fragment shader: %w", err)
	}

	// Create shader program
	r.shaderProgram = gl.CreateProgram()
	gl.AttachShader(r.shaderProgram, vertexShader)
	gl.AttachShader(r.shaderProgram, fragmentShader)
	gl.LinkProgram(r.shaderProgram)

	// Check linking errors
	var status int32
	gl.GetProgramiv(r.shaderProgram, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(r.shaderProgram, gl.INFO_LOG_LENGTH, &logLength)

		logMsg := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(r.shaderProgram, logLength, nil, gl.Str(logMsg))

		return fmt.Errorf("failed to link shader program: %v", logMsg)
	}

	// Clean up shaders
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	// Set up vertex data
	gl.GenVertexArrays(1, &r.vao)
	gl.GenBuffers(1, &r.vbo)

	gl.BindVertexArray(r.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)

	// Set up vertex attributes
	// Position attribute (x, y)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, nil)

	// Texture coordinate attribute (u, v)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))

	gl.BindVertexArray(0)

	// Set up character rendering
	gl.GenVertexArrays(1, &r.charVAO)
	gl.GenBuffers(1, &r.charVBO)
	gl.BindVertexArray(r.charVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.charVBO)
	gl.BufferData(gl.ARRAY_BUFFER, 6*4*4, nil, gl.DYNAMIC_DRAW)

	// Position attribute (x, y)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, nil)

	// Texture coordinate attribute (u, v)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))

	gl.BindVertexArray(0)

	// Create font texture
	r.createFontTexture()

	r.initialized = true
	return nil
}

// createFontTexture creates a simple bitmap font texture
func (r *UIRenderer) createFontTexture() {
	// Create a simple 8x16 bitmap font texture
	// This is a very basic font with ASCII characters 32-126
	textureWidth := 128
	textureHeight := 64

	// Create texture data
	pixels := make([]byte, textureWidth*textureHeight*4)

	// Generate simple character patterns
	for y := 0; y < textureHeight; y++ {
		for x := 0; x < textureWidth; x++ {
			charX := x / 8
			charY := y / 16
			if charX < 16 && charY < 4 {
				// Simple character pattern - create a basic readable font
				pixelX := x % 8
				pixelY := y % 16

				// Create a simple pattern for characters
				isChar := r.isCharacterPixel(charX+charY*16, pixelX, pixelY)

				idx := (y*textureWidth + x) * 4
				if isChar {
					pixels[idx] = 255   // R
					pixels[idx+1] = 255 // G
					pixels[idx+2] = 255 // B
					pixels[idx+3] = 255 // A
				} else {
					pixels[idx] = 0   // R
					pixels[idx+1] = 0 // G
					pixels[idx+2] = 0 // B
					pixels[idx+3] = 0 // A
				}
			}
		}
	}

	// Create OpenGL texture
	gl.GenTextures(1, &r.fontTexture)
	gl.BindTexture(gl.TEXTURE_2D, r.fontTexture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(textureWidth), int32(textureHeight), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pixels))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

// isCharacterPixel determines if a pixel should be drawn for a character
func (r *UIRenderer) isCharacterPixel(charCode int, x, y int) bool {
	// Simple patterns for common characters
	if charCode == 32 { // Space
		return false
	}

	// For simplicity, create a basic rectangle for most characters
	// This will at least show text outlines
	if x == 0 || x == 7 || y == 0 || y == 7 {
		return charCode >= 33 && charCode <= 126 // Printable ASCII range
	}

	// Add some internal patterns for common letters
	switch charCode {
	case 65, 97: // A, a
		return (y == 0 || y == 7) && (x >= 1 && x <= 6) ||
			(x == 0 || x == 7) && (y >= 1 && y <= 6)
	case 66, 98: // B, b
		return (y == 0 || y == 7) && (x >= 1 && x <= 6) ||
			(x == 0 || x == 7) && (y >= 1 && y <= 6)
	case 69, 101: // E, e
		return y == 0 || y == 3 || y == 7 ||
			x == 0 && (y >= 1 && y <= 6)
	case 84, 116: // T, t
		return y == 0 || x == 3
	}

	return false
}

// compileShader compiles a shader from source
func (r *UIRenderer) compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()

	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		logMsg := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(logMsg))

		return 0, fmt.Errorf("failed to compile %v: %v", source, logMsg)
	}

	return shader, nil
}

// Begin prepares for UI rendering
func (r *UIRenderer) Begin() {
	if !r.initialized {
		return
	}

	// Enable blending for transparency
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// Use UI shader program
	gl.UseProgram(r.shaderProgram)

	// Set projection matrix
	projectionLoc := gl.GetUniformLocation(r.shaderProgram, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionLoc, 1, false, &r.projection[0])

	// Bind vertex array
	gl.BindVertexArray(r.vao)
}

// End finishes UI rendering
func (r *UIRenderer) End() {
	if !r.initialized {
		return
	}

	gl.BindVertexArray(0)
	gl.Disable(gl.BLEND)
}

// RenderQuad renders a colored rectangle
func (r *UIRenderer) RenderQuad(x, y, width, height float32, color mgl32.Vec4) {
	if !r.initialized {
		return
	}

	// Debug output
	fmt.Printf("Rendering quad at (%.1f, %.1f) size (%.1f, %.1f) color (%.1f, %.1f, %.1f, %.1f)\n",
		x, y, width, height, color[0], color[1], color[2], color[3])

	// Bind the correct VAO for quads
	gl.BindVertexArray(r.vao)

	// Set up vertex data for quad
	vertices := []float32{
		// positions      // texture coords
		x, y, 0.0, 0.0, // bottom-left
		x + width, y, 1.0, 0.0, // bottom-right
		x + width, y + height, 1.0, 1.0, // top-right
		x, y + height, 0.0, 1.0, // top-left
		x, y, 0.0, 0.0, // top-left
		x + width, y, 1.0, 0.0, // top-right
		x + width, y + height, 1.0, 1.0, // bottom-right
		x, y + height, 0.0, 1.0, // bottom-left
	}

	// Update vertex buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Set color uniform
	colorLoc := gl.GetUniformLocation(r.shaderProgram, gl.Str("color\x00"))
	gl.Uniform4f(colorLoc, color[0], color[1], color[2], color[3])

	// Set texture uniform to false (using color only)
	useTextureLoc := gl.GetUniformLocation(r.shaderProgram, gl.Str("useTexture\x00"))
	gl.Uniform1i(useTextureLoc, 0)

	// Draw quad
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
}

// RenderTexturedQuad renders a textured rectangle
func (r *UIRenderer) RenderTexturedQuad(x, y, width, height float32, textureID uint32) {
	if !r.initialized {
		return
	}

	// Bind the correct VAO for quads
	gl.BindVertexArray(r.vao)

	// Set up vertex data for quad with full texture coordinates
	vertices := []float32{
		// positions      // texture coords
		x, y, 0.0, 0.0, // bottom-left
		x + width, y, 1.0, 0.0, // bottom-right
		x + width, y + height, 1.0, 1.0, // top-right
		x, y + height, 0.0, 1.0, // top-left
	}

	// Update vertex buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Bind texture
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, textureID)

	// Set color to white (full color)
	colorLoc := gl.GetUniformLocation(r.shaderProgram, gl.Str("color\x00"))
	gl.Uniform4f(colorLoc, 1.0, 1.0, 1.0, 1.0)

	// Set texture uniform to true (using texture)
	useTextureLoc := gl.GetUniformLocation(r.shaderProgram, gl.Str("useTexture\x00"))
	gl.Uniform1i(useTextureLoc, 1)

	// Draw quad
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

	// Unbind texture
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

// RenderText renders text at the specified position with the given color and scale
func (r *UIRenderer) RenderText(text string, x, y float32, scale float32, color mgl32.Vec4) {
	if !r.initialized || text == "" {
		return
	}

	// Enable blending for text
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// Bind font texture
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, r.fontTexture)

	// Set color uniform
	colorLoc := gl.GetUniformLocation(r.shaderProgram, gl.Str("color\x00"))
	gl.Uniform4f(colorLoc, color[0], color[1], color[2], color[3])

	// Use texture for text rendering
	useTextureLoc := gl.GetUniformLocation(r.shaderProgram, gl.Str("useTexture\x00"))
	gl.Uniform1i(useTextureLoc, 1)

	// Bind character VAO
	gl.BindVertexArray(r.charVAO)

	// Character dimensions
	charWidth := 8.0 * scale
	charHeight := 16.0 * scale
	lineHeight := charHeight

	// Texture dimensions
	texCharWidth := float32(8.0) / float32(128.0)
	texCharHeight := float32(16.0) / float32(64.0)

	// Render each character
	currentX := x
	currentY := y

	for _, char := range text {
		if char == '\n' {
			currentX = x
			currentY += lineHeight
			continue
		}

		if char == ' ' {
			currentX += charWidth * 0.5 // Space is half width
			continue
		}

		// Calculate texture coordinates for this character
		charCode := int(char)
		if charCode < 32 || charCode > 126 {
			charCode = 63 // Use '?' for invalid characters
		}

		// Calculate position in texture atlas
		charIndex := charCode - 32
		atlasX := float32(charIndex%16) * texCharWidth
		atlasY := float32(charIndex/16) * texCharHeight

		// Create quad for this character with proper texture coordinates
		vertices := []float32{
			// positions      // texture coords
			currentX, currentY, atlasX, atlasY, // bottom-left
			currentX + charWidth, currentY, atlasX + texCharWidth, atlasY, // bottom-right
			currentX + charWidth, currentY + charHeight, atlasX + texCharWidth, atlasY + texCharHeight, // top-right
			currentX, currentY + charHeight, atlasX, atlasY + texCharHeight, // top-left
		}

		// Update VBO data
		gl.BindBuffer(gl.ARRAY_BUFFER, r.charVBO)
		gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(vertices)*4, gl.Ptr(vertices))

		// Draw character quad
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

		// Move to next character position
		currentX += charWidth * 0.8 // Slight spacing between characters
	}

	// Clean up
	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindTexture(gl.TEXTURE_2D, 0)

	// Reset texture usage
	gl.Uniform1i(useTextureLoc, 0)
}

// Resize updates the renderer for new window dimensions
func (r *UIRenderer) Resize(width, height int) {
	r.width = width
	r.height = height
	r.projection = mgl32.Ortho(0, float32(width), float32(height), 0, -1, 1)
}

// Cleanup releases OpenGL resources
func (r *UIRenderer) Cleanup() {
	if r.initialized {
		gl.DeleteVertexArrays(1, &r.vao)
		gl.DeleteBuffers(1, &r.vbo)
		gl.DeleteVertexArrays(1, &r.charVAO)
		gl.DeleteBuffers(1, &r.charVBO)
		gl.DeleteTextures(1, &r.fontTexture)
		gl.DeleteProgram(r.shaderProgram)
		r.initialized = false
	}
}

// Vertex shader source for UI rendering
const vertexShaderSource = `#version 330 core
layout (location = 0) in vec2 aPos;
layout (location = 1) in vec2 aTexCoord;

uniform mat4 projection;

out vec2 TexCoord;

void main() {
    gl_Position = projection * vec4(aPos, 0.0, 1.0);
    TexCoord = aTexCoord;
}
`

// Fragment shader source for UI rendering
const fragmentShaderSource = `#version 330 core
in vec2 TexCoord;

uniform vec4 color;
uniform sampler2D texture1;
uniform bool useTexture;

out vec4 FragColor;

void main() {
    if (useTexture) {
        FragColor = texture(texture1, TexCoord) * color;
    } else {
        FragColor = color;
    }
}
`
