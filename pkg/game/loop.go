package game

import (
	"fmt"
	"sync"
	"time"

	"github.com/tesselstudio/TesselBox/pkg/opengl"
	"github.com/tesselstudio/TesselBox/pkg/world"
)

// GameLoop manages the main game loop with frame synchronization
type GameLoop struct {
	mu sync.RWMutex

	// Target FPS
	targetFPS   int
	frameTime   time.Duration
	deltaTime   float64
	accumulator float64

	// Controllers
	gameController *Controller
	renderEngine   *opengl.Engine

	// Timing
	lastFrameTime time.Time
	currentTime   time.Time

	// Running state
	isRunning bool
	isPaused  bool

	// Frame statistics
	frameCount    int64
	totalTime     float64
	averageFPS    float64
	lastStatsTime time.Time
}

// NewGameLoop creates a new game loop
func NewGameLoop(controller *Controller, engine *opengl.Engine, targetFPS int) *GameLoop {
	return &GameLoop{
		targetFPS:      targetFPS,
		frameTime:      time.Duration(1000000000 / targetFPS), // nanoseconds
		gameController: controller,
		renderEngine:   engine,
		lastFrameTime:  time.Now(),
		currentTime:    time.Now(),
		lastStatsTime:  time.Now(),
	}
}

// Run starts the game loop
func (gl *GameLoop) Run() error {
	gl.mu.Lock()
	if gl.isRunning {
		gl.mu.Unlock()
		return fmt.Errorf("game loop already running")
	}
	gl.isRunning = true
	gl.mu.Unlock()

	defer func() {
		gl.mu.Lock()
		gl.isRunning = false
		gl.mu.Unlock()
	}()

	fmt.Println("🎮 Game Loop Started (60 FPS target)")

	for !gl.renderEngine.ShouldClose() && gl.isRunning {
		frameStart := time.Now()

		// Calculate delta time
		gl.updateDeltaTime()

		// Handle events
		gl.renderEngine.PollEvents()

		// Update game logic
		if !gl.isPaused {
			gl.update()
		}

		// Render frame
		gl.render()

		// Frame rate limiting
		gl.limitFrameRate(frameStart)

		// Update statistics
		gl.updateStats()
	}

	fmt.Println("🎮 Game Loop Stopped")
	return nil
}

// updateDeltaTime calculates delta time for this frame
func (gl *GameLoop) updateDeltaTime() {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	now := time.Now()
	gl.deltaTime = now.Sub(gl.lastFrameTime).Seconds()
	gl.lastFrameTime = now
	gl.currentTime = now

	// Cap delta time to prevent large jumps (0.1 second max)
	if gl.deltaTime > 0.1 {
		gl.deltaTime = 0.1
	}
}

// update updates game logic
func (gl *GameLoop) update() {
	gl.mu.RLock()
	dt := gl.deltaTime
	controller := gl.gameController
	gl.mu.RUnlock()

	if controller == nil {
		return
	}

	// Update game controller (includes camera switching and input processing)
	controller.Update()

	// Don't update game logic if paused
	if controller.IsPaused() {
		return
	}

	// Update player
	player := controller.GetPlayer()
	if player != nil {
		player.Update(dt)
	}

	// Update world
	w := controller.GetWorld()
	if w != nil {
		if player != nil {
			playerPos := player.GetPosition()
			w.Update(dt, world.NewVec3(playerPos.X, playerPos.Y, playerPos.Z))
		}
	}

	// Check for auto-save
	controller.checkAutoSave(dt)
}

// render renders the current frame
func (gl *GameLoop) render() {
	gl.mu.RLock()
	engine := gl.renderEngine
	controller := gl.gameController
	gl.mu.RUnlock()

	if engine == nil || controller == nil {
		return
	}

	// Begin frame
	engine.BeginFrame()

	// Render game
	engine.Render(controller)

	// End frame
	engine.EndFrame()
}

// limitFrameRate enforces target frame rate
func (gl *GameLoop) limitFrameRate(frameStart time.Time) {
	gl.mu.RLock()
	frameTime := gl.frameTime
	gl.mu.RUnlock()

	elapsed := time.Since(frameStart)
	if elapsed < frameTime {
		time.Sleep(frameTime - elapsed)
	}
}

// updateStats updates frame rate statistics
func (gl *GameLoop) updateStats() {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	gl.frameCount++
	gl.totalTime += gl.deltaTime

	// Update stats every second
	if time.Since(gl.lastStatsTime) >= time.Second {
		gl.averageFPS = float64(gl.frameCount) / time.Since(gl.lastStatsTime).Seconds()

		// Print stats periodically
		if gl.frameCount%60 == 0 {
			fmt.Printf("FPS: %.1f | Frames: %d\n",
				gl.averageFPS,
				gl.frameCount)
		}

		gl.frameCount = 0
		gl.lastStatsTime = time.Now()
	}
}

// Pause pauses the game loop (updates stop, rendering continues)
func (gl *GameLoop) Pause() {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.isPaused = true
	fmt.Println("⏸️  Game Paused")
}

// Resume resumes the game loop
func (gl *GameLoop) Resume() {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.isPaused = false
	gl.lastFrameTime = time.Now() // Reset frame time to avoid large delta
	fmt.Println("▶️  Game Resumed")
}

// IsPaused returns whether the game is paused
func (gl *GameLoop) IsPaused() bool {
	gl.mu.RLock()
	defer gl.mu.RUnlock()
	return gl.isPaused
}

// GetDeltaTime returns the delta time for the current frame
func (gl *GameLoop) GetDeltaTime() float64 {
	gl.mu.RLock()
	defer gl.mu.RUnlock()
	return gl.deltaTime
}

// GetAverageFPS returns the average FPS
func (gl *GameLoop) GetAverageFPS() float64 {
	gl.mu.RLock()
	defer gl.mu.RUnlock()
	return gl.averageFPS
}

// Stop stops the game loop
func (gl *GameLoop) Stop() {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.isRunning = false
}

// IsRunning returns whether the game loop is running
func (gl *GameLoop) IsRunning() bool {
	gl.mu.RLock()
	defer gl.mu.RUnlock()
	return gl.isRunning
}
