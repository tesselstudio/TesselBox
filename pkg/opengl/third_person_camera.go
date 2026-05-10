package opengl

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/tesselstudio/TesselBox/pkg/types"
)

// ThirdPersonCamera represents a follow-behind camera
type ThirdPersonCamera struct {
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
	distance      float32
	minDistance   float32
	maxDistance   float32
	lastYaw       float32
	lastPitch     float32
}

// NewThirdPersonCamera creates a new third-person camera
func NewThirdPersonCamera(width, height int) *ThirdPersonCamera {
	cam := &ThirdPersonCamera{
		position:      mgl32.Vec3{0, 2, 0},
		front:         mgl32.Vec3{0, 0, -1},
		up:            mgl32.Vec3{0, 1, 0},
		right:         mgl32.Vec3{1, 0, 0},
		worldUp:       mgl32.Vec3{0, 1, 0},
		yaw:           -90.0,
		pitch:         0,
		fov:           45.0,
		aspectRatio:   float32(width) / float32(height),
		near:          0.1,
		far:           500.0,
		sensitivity:   0.1,
		eyeHeight:     1.62,
		distance:      5.0, // Distance behind player
		minDistance:   2.0,
		maxDistance:   20.0,
	}
	return cam
}

// UpdateFromPlayer updates camera position and rotation from player position
func (c *ThirdPersonCamera) UpdateFromPlayer(playerPos types.Vec3, playerYaw, playerPitch float32) {
	// Update rotation
	if c.yaw != playerYaw || c.pitch != playerPitch {
		c.yaw = playerYaw
		c.pitch = playerPitch
		c.updateCameraVectors()
	}

	// Calculate player eye position
	eyePos := mgl32.Vec3{
		playerPos.X,
		playerPos.Y + c.eyeHeight,
		playerPos.Z,
	}

	// Calculate camera position (behind player)
	cameraOffset := c.front.Mul(-c.distance)
	newPos := eyePos.Add(cameraOffset)

	// Smooth camera movement
	smoothing := float32(0.9)
	c.position = c.position.Mul(smoothing).Add(newPos.Mul(1 - smoothing))
}

// updateCameraVectors updates the front, right, and up vectors
func (c *ThirdPersonCamera) updateCameraVectors() {
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
func (c *ThirdPersonCamera) GetViewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(
		c.position,                    // Camera position
		c.position.Add(c.front),       // Look at point (player)
		c.up,                          // Up vector
	)
}

// GetProjectionMatrix returns the projection matrix
func (c *ThirdPersonCamera) GetProjectionMatrix() mgl32.Mat4 {
	return mgl32.Perspective(
		mgl32.DegToRad(c.fov),
		c.aspectRatio,
		c.near,
		c.far,
	)
}

// GetPosition returns the camera position
func (c *ThirdPersonCamera) GetPosition() mgl32.Vec3 {
	return c.position
}

// GetFront returns the front vector
func (c *ThirdPersonCamera) GetFront() mgl32.Vec3 {
	return c.front
}

// GetRight returns the right vector
func (c *ThirdPersonCamera) GetRight() mgl32.Vec3 {
	return c.right
}

// GetUp returns the up vector
func (c *ThirdPersonCamera) GetUp() mgl32.Vec3 {
	return c.up
}

// SetFOV sets the field of view
func (c *ThirdPersonCamera) SetFOV(fov float32) {
	c.fov = fov
	if c.fov < 1.0 {
		c.fov = 1.0
	}
	if c.fov > 120.0 {
		c.fov = 120.0
	}
}

// SetAspectRatio sets the aspect ratio
func (c *ThirdPersonCamera) SetAspectRatio(width, height int) {
	c.aspectRatio = float32(width) / float32(height)
}

// SetClipPlanes sets the near and far clip planes
func (c *ThirdPersonCamera) SetClipPlanes(near, far float32) {
	c.near = near
	c.far = far
}

// SetDistance sets the camera distance from player
func (c *ThirdPersonCamera) SetDistance(distance float32) {
	c.distance = distance
	if c.distance < c.minDistance {
		c.distance = c.minDistance
	}
	if c.distance > c.maxDistance {
		c.distance = c.maxDistance
	}
}

// GetDistance returns the current camera distance
func (c *ThirdPersonCamera) GetDistance() float32 {
	return c.distance
}

// AdjustDistance adjusts the camera distance by the given amount
func (c *ThirdPersonCamera) AdjustDistance(delta float32) {
	c.SetDistance(c.distance + delta)
}

// SetDistanceLimits sets the minimum and maximum distance limits
func (c *ThirdPersonCamera) SetDistanceLimits(min, max float32) {
	c.minDistance = min
	c.maxDistance = max
}
