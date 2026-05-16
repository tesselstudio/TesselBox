// TesselBox Web Game Loader

const WebGame = {
    wasm: null,
    canvas: null,
    gl: null,
    running: false,
    animationId: null,

    async init() {
        try {
            console.log('Initializing TesselBox Web...');
            
            // Load WebAssembly support
            if (!WebAssembly.instantiateStreaming) {
                console.log('Browser does not support WebAssembly streaming, falling back to instantiate');
                WebAssembly.instantiateStreaming = async (resp, importObject) => {
                    return (await WebAssembly.instantiate(await resp.arrayBuffer(), importObject));
                };
            }

            // Load WASM file
            const go = new Go();
            const response = await fetch('main.wasm');
            
            if (!response.ok) {
                throw new Error(`Failed to load WASM: ${response.status} ${response.statusText}`);
            }

            const buffer = await response.arrayBuffer();
            const result = await WebAssembly.instantiate(buffer, go.importObject);
            
            this.wasm = result.instance;
            go.run(this.wasm);
            
            console.log('WASM loaded successfully');
            
            // Hide loading, show start screen
            document.getElementById('loading').style.display = 'none';
            document.getElementById('start-screen').style.display = 'flex';
            
            // Set up start button
            document.getElementById('start-btn').addEventListener('click', () => {
                this.startGame();
            });
            
        } catch (error) {
            console.error('Failed to initialize game:', error);
            this.showError(error.message);
        }
    },

    startGame() {
        try {
            console.log('Starting game...');
            
            // Hide start screen
            document.getElementById('start-screen').style.display = 'none';
            
            // Show canvas and UI
            this.canvas = document.getElementById('game-canvas');
            this.canvas.style.display = 'block';
            document.getElementById('ui-overlay').style.display = 'block';
            document.getElementById('controls-info').style.display = 'block';
            
            // Set canvas size
            this.canvas.width = window.innerWidth;
            this.canvas.height = window.innerHeight;
            
            // Request pointer lock synchronously (same call stack as the button click user gesture)
            if (this.canvas.requestPointerLock) {
                this.canvas.requestPointerLock();
            }

            // Initialize WebGL via WASM
            this.setupRenderer();

            // Set up input handling
            this.setupInput();
            
            // Start game loop
            this.running = true;
            this.gameLoop();
            
            // Call Go function to start world
            if (typeof startWorld !== 'undefined') {
                startWorld('browser_world', Date.now());
            }
            
            console.log('Game started');
            
        } catch (error) {
            console.error('Failed to start game:', error);
            this.showError(error.message);
        }
    },

    setupRenderer() {
        try {
            if (typeof globalThis.initWebGL !== 'function') {
                throw new Error('WASM initWebGL not available');
            }

            const ok = globalThis.initWebGL(this.canvas);
            if (!ok) {
                throw new Error('Failed to initialize WebGL renderer (WASM)');
            }

            if (typeof resizeCanvas !== 'undefined') {
                resizeCanvas(this.canvas.width, this.canvas.height);
            }

            console.log('WebGL initialized via WASM');
        } catch (error) {
            console.error('Failed to initialize WebGL:', error);
            throw error;
        }
    },

    setupInput() {
        // Keyboard input
        const keys = {};
        
        window.addEventListener('keydown', (e) => {
            keys[e.code] = true;
            this.handleKeyInput(e.code, true);
        });
        
        window.addEventListener('keyup', (e) => {
            keys[e.code] = false;
            this.handleKeyInput(e.code, false);
        });
        
        // Mouse movement
        document.addEventListener('mousemove', (e) => {
            if (document.pointerLockElement === this.canvas) {
                this.handleMouseMove(e.movementX, e.movementY);
            }
        });
        
        // Mouse buttons
        this.canvas.addEventListener('mousedown', (e) => {
            // Re-acquire pointer lock if lost (no permission prompt needed after initial request)
            if (document.pointerLockElement !== this.canvas) {
                this.canvas.requestPointerLock();
            }
            this.handleMouseInput(e.button, true);
        });
        
        this.canvas.addEventListener('mouseup', (e) => {
            this.handleMouseInput(e.button, false);
        });
        
        // Handle window resize
        window.addEventListener('resize', () => {
            this.canvas.width = window.innerWidth;
            this.canvas.height = window.innerHeight;
            if (typeof resizeCanvas !== 'undefined') {
                resizeCanvas(this.canvas.width, this.canvas.height);
            }
        });
        
        // Pointer lock change
        document.addEventListener('pointerlockchange', () => {
            if (document.pointerLockElement !== this.canvas) {
                // Pointer lock lost - game paused
                console.log('Game paused (pointer lock lost)');
            }
        });
    },

    handleKeyInput(code, pressed) {
        // Map keyboard codes to game input
        const keyMap = {
            'KeyW': 87,
            'KeyS': 83,
            'KeyA': 65,
            'KeyD': 68,
            'Space': 32,
            'KeyE': 69,
            'Escape': 27,
        };
        
        const keyCode = keyMap[code];
        if (keyCode && typeof handleKeyInput !== 'undefined') {
            handleKeyInput(keyCode, pressed ? 1 : 0);
        }
    },

    handleMouseMove(dx, dy) {
        if (typeof handleMouseMove !== 'undefined') {
            handleMouseMove(dx, dy);
        }
    },

    handleMouseInput(button, pressed) {
        if (typeof handleMouseInput !== 'undefined') {
            handleMouseInput(button, pressed ? 1 : 0);
        }
    },

    gameLoop() {
        if (!this.running) return;
        
        try {
            if (typeof update !== 'undefined') {
                update();
            }
            if (typeof render !== 'undefined') {
                render();
            }
            
            this.animationId = requestAnimationFrame(() => this.gameLoop());
            
        } catch (error) {
            console.error('Game loop error:', error);
            this.running = false;
            this.showError(error.message);
        }
    },

    showError(message) {
        document.getElementById('loading').style.display = 'none';
        document.getElementById('start-screen').style.display = 'none';
        document.getElementById('error-message').style.display = 'block';
        document.getElementById('error-text').textContent = message;
    },

    stop() {
        console.log('Stopping game...');
        this.running = false;
        
        if (this.animationId) {
            cancelAnimationFrame(this.animationId);
        }
        
        if (document.pointerLockElement === this.canvas) {
            document.exitPointerLock();
        }
        
        // Call Go cleanup function
        if (typeof cleanup !== 'undefined') {
            cleanup();
        }
    }
};

// Global functions for Go to call
window.registerUpdateLoop = function() {
    console.log('Update loop registered');
};

window.handleKeyInput = function(keyCode, keyState) {
    // This will be called from Go
    WebGame.handleKeyInputFromGo(keyCode, keyState);
};

window.handleMouseMove = function(dx, dy) {
    // This will be called from Go
    console.log('Mouse move:', dx, dy);
};

window.handleMouseInput = function(button, buttonState) {
    // This will be called from Go
    console.log('Mouse input:', button, buttonState);
};

// Initialize game when page loads
window.addEventListener('load', () => {
    WebGame.init();
});

// Handle page unload
window.addEventListener('beforeunload', () => {
    WebGame.stop();
});
