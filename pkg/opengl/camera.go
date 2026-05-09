package opengl

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/tesselstudio/TesselBox/pkg/types"
)

// FirstPersonCamera represents a first-person camera
type FirstPersonCamera struct {
	position      mgl32.Vec3
	front         mgl32.Vec3
	up            mgl32.Vec3
	right         mgl32.Vec3
	worldUp       mgl32.Vec3
	yaw           float32
	pitch         float32
	fov           float32
	aspectRatio   float32
	near          float32
	far           float32
	sensitivity   float32
	eyeHeight     float32
	lastYaw       float32
	lastPitch     float32
}

// NewFirstPersonCamera creates a new first-person camera
func NewFirstPersonCamera(width, height int) *FirstPersonCamera {
	cam := &FirstPersonCamera{
		position:    mgl32.Vec3{0, 2, 0},
		front:       mgl32.Vec3{0, 0, -1},
		up:          mgl32.Vec3{0, 1, 0},
		right:       mgl32.Vec3{1, 0, 0},
		worldUp:     mgl32.Vec3{0, 1, 0},
		yaw:         -90.0,
		pitch:       0,
		fov:         45.0,
		aspectRatio: float32(width) / float32(height),
		near:        0.1,
		far:         500.0,
		sensitivity: 0.1,
		eyeHeight:   1.62, // Eye height above player feet
	}
	return cam
}

// UpdateFromPlayer updates camera position and rotation from player position
func (c *FirstPersonCamera) UpdateFromPlayer(playerPos types.Vec3, playerYaw, playerPitch float32) {
	// Update position (add eye height)
	newPos := mgl32.Vec3{
		playerPos.X,
		playerPos.Y + c.eyeHeight,
		playerPos.Z,
	}

	// Smooth camera movement
	smoothing := float32(0.9)
	c.position = c.position.Mul(smoothing).Add(newPos.Mul(1 - smoothing))

	// Update rotation
	if c.yaw != playerYaw || c.pitch != playerPitch {
		c.yaw = playerYaw
		c.pitch = playerPitch
		c.updateCameraVectors()
	}
}

// updateCameraVectors updates the front, right, and up vectors
func (c *FirstPersonCamera) updateCameraVectors() {
	// Convert degrees to radians
	yawRad := mgl32.DegToRad(c.yaw)
	pitchRad := mgl32.DegToRad(c.pitch)

	// Calculate front vector
	front := mgl32.Vec3{
		float32(math.Cos(float64(yawRad)) * math.Cos(float64(pitchRad))),
		float32(math.Sin(float64(pitchRad))),
		float32(math.Sin(float64(yawRad)) * math.Cos(float64(pitchRad))),
	}

	c.front = front.Normalize()

	// Recalculate right vector
	c.right = c.front.Cross(c.worldUp).Normalize()

	// Recalculate up vector
	c.up = c.right.Cross(c.front).Normalize()
}

// GetViewMatrix returns the view matrix
func (c *FirstPersonCamera) GetViewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(
		c.position,                    // Camera position
		c.position.Add(c.front),       // Look at point
		c.up,                          // Up vector
	)
}

// GetProjectionMatrix returns the projection matrix
func (c *FirstPersonCamera) GetProjectionMatrix() mgl32.Mat4 {
	return mgl32.Perspective(
		mgl32.DegToRad(c.fov),
		c.aspectRatio,
		c.near,
		c.far,
	)
}

// SetPosition sets the camera position
func (c *FirstPersonCamera) SetPosition(pos mgl32.Vec3) {
	c.position = pos
}

// GetPosition returns the camera position
func (c *FirstPersonCamera) GetPosition() mgl32.Vec3 {
	return c.position
}

// GetFront returns the front vector
func (c *FirstPersonCamera) GetFront() mgl32.Vec3 {
	return c.front
}

// GetRight returns the right vector
func (c *FirstPersonCamera) GetRight() mgl32.Vec3 {
	return c.right
}

// GetUp returns the up vector
func (c *FirstPersonCamera) GetUp() mgl32.Vec3 {
	return c.up
}

// SetFOV sets the field of view
func (c *FirstPersonCamera) SetFOV(fov float32) {
	c.fov = fov
	if c.fov < 1.0 {
		c.fov = 1.0
	}
	if c.fov > 120.0 {
		c.fov = 120.0
	}
}

// SetAspectRatio sets the aspect ratio
func (c *FirstPersonCamera) SetAspectRatio(width, height int) {
	c.aspectRatio = float32(width) / float32(height)
}

// SetEyeHeight sets the eye height above the player
func (c *FirstPersonCamera) SetEyeHeight(height float32) {
	c.eyeHeight = height
}

// GetEyeHeight returns the eye height
func (c *FirstPersonCamera) GetEyeHeight() float32 {
	return c.eyeHeight
}

// SetClipPlanes sets the near and far clip planes
func (c *FirstPersonCamera) SetClipPlanes(near, far float32) {
	c.near = near
	c.far = far
}

// GetDirection returns the camera direction in world coordinates
func (c *FirstPersonCamera) GetDirection(screenX, screenY, screenWidth, screenHeight float32) mgl32.Vec3 {
	// Normalize screen coordinates to -1.0 to 1.0
	ndc := mgl32.Vec2{
		(2.0 * screenX) / screenWidth - 1.0,
		1.0 - (2.0 * screenY) / screenHeight,
	}

	// Build a ray in clip coordinates
	clipCoords := mgl32.Vec4{ndc[0], ndc[1], -1.0, 1.0}

	// Apply inverse projection matrix
	projMatrix := c.GetProjectionMatrix()
	projMatrixInv := projMatrix.Inv()
	eyeCoords := projMatrixInv.Mul4x1(clipCoords)
	eyeCoords = mgl32.Vec4{eyeCoords[0], eyeCoords[1], -1.0, 0.0}

	// Apply inverse view matrix
	viewMatrix := c.GetViewMatrix()
	viewMatrixInv := viewMatrix.Inv()
	worldCoords := viewMatrixInv.Mul4x1(eyeCoords)
	direction := mgl32.Vec3{worldCoords[0], worldCoords[1], worldCoords[2]}

	return direction.Normalize()
}

// GetScreenSpaceRay returns the normalized direction vector for a screen position
// where screen center is (0,0) and range is -1 to 1
func (c *FirstPersonCamera) GetScreenSpaceRay(screenX, screenY float32) mgl32.Vec3 {
	// Screen center ray
	if screenX == 0 && screenY == 0 {
		return c.front
	}

	// Calculate direction based on screen offset
	// This is used for raycasting from screen center
	horizontalComponent := c.right.Mul(screenX * 0.5)
	verticalComponent := c.up.Mul(screenY * 0.5)

	direction := c.front.Add(horizontalComponent).Add(verticalComponent)
	return direction.Normalize()
}
