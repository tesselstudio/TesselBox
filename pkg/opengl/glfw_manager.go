package opengl

import (
	"sync"
	
	"github.com/go-gl/glfw/v3.3/glfw"
)

// GLFWManager manages the global GLFW state
type GLFWManager struct {
	mu     sync.RWMutex
	initialized bool
}

var (
	globalGLFWManager = &GLFWManager{}
)

// InitializeGLFW initializes GLFW if not already initialized
func InitializeGLFW() error {
	globalGLFWManager.mu.Lock()
	defer globalGLFWManager.mu.Unlock()
	
	if !globalGLFWManager.initialized {
		if err := glfw.Init(); err != nil {
			return err
		}
		globalGLFWManager.initialized = true
	}
	return nil
}

// TerminateGLFW terminates GLFW if it's initialized
func TerminateGLFW() {
	globalGLFWManager.mu.Lock()
	defer globalGLFWManager.mu.Unlock()
	
	if globalGLFWManager.initialized {
		glfw.Terminate()
		globalGLFWManager.initialized = false
	}
}

// IsGLFWInitialized returns whether GLFW is initialized
func IsGLFWInitialized() bool {
	globalGLFWManager.mu.RLock()
	defer globalGLFWManager.mu.RUnlock()
	return globalGLFWManager.initialized
}
