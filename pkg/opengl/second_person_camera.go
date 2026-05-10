package opengl

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/tesselstudio/TesselBox/pkg/types"
)

// SecondPersonCamera represents an over-the-shoulder camera
type SecondPersonCamera struct {
	position       mgl32.Vec3
	front          mgl32.Vec3
	up             mgl32.Vec3
	right          mgl32.Vec3
	worldUp        mgl32.Vec3
	yaw            float32
	pitch          float32
	fov            float32
	aspectRatio    float32
	near           float32
	far            float32
	sensitivity    float32
	eyeHeight      float32
	shoulderOffset mgl32.Vec3
	distance       float32
	lastYaw        float32
	lastPitch      float32
}

// NewSecondPersonCamera creates a new second-person camera
func NewSecondPersonCamera(width, height int) *SecondPersonCamera {
	cam := &SecondPersonCamera{
		position:       mgl32.Vec3{0, 2, 0},
		front:          mgl32.Vec3{0, 0, -1},
		up:             mgl32.Vec3{0, 1, 0},
		right:          mgl32.Vec3{1, 0, 0},
		worldUp:        mgl32.Vec3{0, 1, 0},
		yaw:            -90.0,
		pitch:          0,
		fov:            45.0,
		aspectRatio:    float32(width) / float32(height),
		near:           0.1,
		far:            500.0,
		sensitivity:    0.1,
		eyeHeight:      1.62,
		shoulderOffset: mgl32.Vec3{0.5, -0.2, -0.8}, // Right shoulder offset
		distance:       2.0,                         // Distance behind player
	}
	return cam
}

// UpdateFromPlayer updates camera position and rotation from player position
func (c *SecondPersonCamera) UpdateFromPlayer(playerPos types.Vec3, playerYaw, playerPitch float32) {
	// Update rotation
	if c.yaw != playerYaw || c.pitch != playerPitch {
		c.yaw = playerYaw
		c.pitch = playerPitch
		c.updateCameraVectors()
	}

	// Calculate camera position (over-the-shoulder view)
	eyePos := mgl32.Vec3{
		playerPos.X,
		playerPos.Y + c.eyeHeight,
		playerPos.Z,
	}

	// Calculate shoulder offset in world space
	rightOffset := c.right.Mul(c.shoulderOffset[0])
	upOffset := c.up.Mul(c.shoulderOffset[1])
	backOffset := c.front.Mul(c.shoulderOffset[2])

	// Final camera position
	newPos := eyePos.Add(rightOffset).Add(upOffset).Add(backOffset)

	// Smooth camera movement
	smoothing := float32(0.9)
	c.position = c.position.Mul(smoothing).Add(newPos.Mul(1 - smoothing))
}

// updateCameraVectors updates the front, right, and up vectors
func (c *SecondPersonCamera) updateCameraVectors() {
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
func (c *SecondPersonCamera) GetViewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(
		c.position,              // Camera position
		c.position.Add(c.front), // Look at point
		c.up,                    // Up vector
	)
}

// GetProjectionMatrix returns the projection matrix
func (c *SecondPersonCamera) GetProjectionMatrix() mgl32.Mat4 {
	return mgl32.Perspective(
		mgl32.DegToRad(c.fov),
		c.aspectRatio,
		c.near,
		c.far,
	)
}

// GetPosition returns the camera position
func (c *SecondPersonCamera) GetPosition() mgl32.Vec3 {
	return c.position
}

// GetFront returns the front vector
func (c *SecondPersonCamera) GetFront() mgl32.Vec3 {
	return c.front
}

// GetRight returns the right vector
func (c *SecondPersonCamera) GetRight() mgl32.Vec3 {
	return c.right
}

// GetUp returns the up vector
func (c *SecondPersonCamera) GetUp() mgl32.Vec3 {
	return c.up
}

// SetFOV sets the field of view
func (c *SecondPersonCamera) SetFOV(fov float32) {
	c.fov = fov
	if c.fov < 1.0 {
		c.fov = 1.0
	}
	if c.fov > 120.0 {
		c.fov = 120.0
	}
}

// SetAspectRatio sets the aspect ratio
func (c *SecondPersonCamera) SetAspectRatio(width, height int) {
	c.aspectRatio = float32(width) / float32(height)
}

// SetClipPlanes sets the near and far clip planes
func (c *SecondPersonCamera) SetClipPlanes(near, far float32) {
	c.near = near
	c.far = far
}

// SetShoulderOffset sets the shoulder offset for over-the-shoulder view
func (c *SecondPersonCamera) SetShoulderOffset(offset mgl32.Vec3) {
	c.shoulderOffset = offset
}

// GetShoulderOffset returns the current shoulder offset
func (c *SecondPersonCamera) GetShoulderOffset() mgl32.Vec3 {
	return c.shoulderOffset
}
