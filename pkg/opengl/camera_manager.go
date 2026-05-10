package opengl

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tesselstudio/TesselBox/pkg/types"
)

// CameraMode represents different camera perspectives
type CameraMode int

const (
	FirstPerson CameraMode = iota
	SecondPerson
	ThirdPerson
)

// String returns the string representation of the camera mode
func (cm CameraMode) String() string {
	switch cm {
	case FirstPerson:
		return "FirstPerson"
	case SecondPerson:
		return "SecondPerson"
	case ThirdPerson:
		return "ThirdPerson"
	default:
		return "Unknown"
	}
}

// CameraInterface defines common camera methods
type CameraInterface interface {
	UpdateFromPlayer(playerPos types.Vec3, playerYaw, playerPitch float32)
	GetViewMatrix() mgl32.Mat4
	GetProjectionMatrix() mgl32.Mat4
	GetPosition() mgl32.Vec3
	GetFront() mgl32.Vec3
	GetRight() mgl32.Vec3
	GetUp() mgl32.Vec3
	SetAspectRatio(width, height int)
	SetFOV(fov float32)
	SetClipPlanes(near, far float32)
}

// CameraManager manages different camera modes and transitions
type CameraManager struct {
	currentMode CameraMode

	// Camera instances
	firstPerson  *FirstPersonCamera
	secondPerson *SecondPersonCamera
	thirdPerson  *ThirdPersonCamera

	// Transition smoothing
	transitionProgress float32
	transitioning      bool
	transitionSpeed    float32

	// Current active camera (for interface)
	activeCamera CameraInterface
}

// NewCameraManager creates a new camera manager with all camera types
func NewCameraManager(width, height int) *CameraManager {
	cm := &CameraManager{
		currentMode:        FirstPerson,
		transitionSpeed:    5.0, // transitions per second
		transitioning:      false,
		transitionProgress: 0,
	}

	// Initialize all cameras
	cm.firstPerson = NewFirstPersonCamera(width, height)
	cm.secondPerson = NewSecondPersonCamera(width, height)
	cm.thirdPerson = NewThirdPersonCamera(width, height)

	// Set default active camera
	cm.activeCamera = cm.firstPerson

	return cm
}

// SetMode sets the camera mode directly
func (cm *CameraManager) SetMode(mode CameraMode) {
	if cm.currentMode != mode {
		cm.currentMode = mode
		cm.transitioning = true
		cm.transitionProgress = 0

		// Update active camera
		switch mode {
		case FirstPerson:
			cm.activeCamera = cm.firstPerson
		case SecondPerson:
			cm.activeCamera = cm.secondPerson
		case ThirdPerson:
			cm.activeCamera = cm.thirdPerson
		}
	}
}

// CycleMode cycles to the next camera mode
func (cm *CameraManager) CycleMode() {
	switch cm.currentMode {
	case FirstPerson:
		cm.SetMode(SecondPerson)
	case SecondPerson:
		cm.SetMode(ThirdPerson)
	case ThirdPerson:
		cm.SetMode(FirstPerson)
	}
}

// GetMode returns the current camera mode as interface
func (cm *CameraManager) GetMode() interface{ String() string } {
	return cm.currentMode
}

// UpdateFromPlayer updates the current camera from player position and rotation
func (cm *CameraManager) UpdateFromPlayer(playerPos types.Vec3, playerYaw, playerPitch float32) {
	// Update all cameras to keep them in sync
	cm.firstPerson.UpdateFromPlayer(playerPos, playerYaw, playerPitch)
	cm.secondPerson.UpdateFromPlayer(playerPos, playerYaw, playerPitch)
	cm.thirdPerson.UpdateFromPlayer(playerPos, playerYaw, playerPitch)

	// Handle transitions
	if cm.transitioning {
		cm.transitionProgress += cm.transitionSpeed * 0.016 // assuming 60 FPS
		if cm.transitionProgress >= 1.0 {
			cm.transitionProgress = 1.0
			cm.transitioning = false
		}
	}
}

// GetViewMatrix returns the view matrix of the current camera
func (cm *CameraManager) GetViewMatrix() mgl32.Mat4 {
	return cm.activeCamera.GetViewMatrix()
}

// GetProjectionMatrix returns the projection matrix of the current camera
func (cm *CameraManager) GetProjectionMatrix() mgl32.Mat4 {
	return cm.activeCamera.GetProjectionMatrix()
}

// GetPosition returns the position of the current camera
func (cm *CameraManager) GetPosition() mgl32.Vec3 {
	return cm.activeCamera.GetPosition()
}

// GetFront returns the front vector of the current camera
func (cm *CameraManager) GetFront() mgl32.Vec3 {
	return cm.activeCamera.GetFront()
}

// GetRight returns the right vector of the current camera
func (cm *CameraManager) GetRight() mgl32.Vec3 {
	return cm.activeCamera.GetRight()
}

// GetUp returns the up vector of the current camera
func (cm *CameraManager) GetUp() mgl32.Vec3 {
	return cm.activeCamera.GetUp()
}

// SetAspectRatio sets the aspect ratio for all cameras
func (cm *CameraManager) SetAspectRatio(width, height int) {
	cm.firstPerson.SetAspectRatio(width, height)
	cm.secondPerson.SetAspectRatio(width, height)
	cm.thirdPerson.SetAspectRatio(width, height)
}

// SetFOV sets the field of view for all cameras
func (cm *CameraManager) SetFOV(fov float32) {
	cm.firstPerson.SetFOV(fov)
	cm.secondPerson.SetFOV(fov)
	cm.thirdPerson.SetFOV(fov)
}

// SetClipPlanes sets the near and far clip planes for all cameras
func (cm *CameraManager) SetClipPlanes(near, far float32) {
	cm.firstPerson.SetClipPlanes(near, far)
	cm.secondPerson.SetClipPlanes(near, far)
	cm.thirdPerson.SetClipPlanes(near, far)
}

// IsTransitioning returns true if currently transitioning between camera modes
func (cm *CameraManager) IsTransitioning() bool {
	return cm.transitioning
}

// GetTransitionProgress returns the progress of the current transition (0-1)
func (cm *CameraManager) GetTransitionProgress() float32 {
	return cm.transitionProgress
}
