<h1>TesselBox Improvement Plan</h1><p><strong>Date:</strong> January 2026<br><strong>Repository:</strong> tesselstudio/TesselBox<br><strong>Version:</strong> 0.1.0<br><strong>Analysis Date:</strong> January 2026</p><hr><h2>Executive Summary</h2><p>TesselBox is a voxel-based sandbox game built with Go and the Kaiju Engine. This comprehensive improvement plan addresses critical security vulnerabilities, code quality issues, testing gaps, and architectural concerns identified through thorough analysis of the codebase, existing documentation, and CI/CD workflows.</p><h3>Key Findings</h3><ul> <li><strong>Total Issues Identified:</strong> 86 (6 Critical, 18 High, 38 Medium, 24 Low)</li> <li><strong>Codebase Size:</strong> ~13,500 lines of Go code across 9 core packages</li> <li><strong>Test Coverage:</strong> Minimal (only 1 test file with basic audio tests)</li> <li><strong>CI/CD:</strong> Well-configured with GitHub Actions for testing, quality, and releases</li> <li><strong>Documentation:</strong> Good architecture docs but missing package-level documentation</li> </ul><h3>Priority Focus Areas</h3><ol> <li><strong>Critical Security &amp; Stability Fixes</strong> (Immediate)</li> <li><strong>Code Quality &amp; Standards</strong> (Short-term)</li> <li><strong>Testing Infrastructure</strong> (Medium-term)</li> <li><strong>Architecture Refactoring</strong> (Long-term)</li> </ol><hr><h2>1. Discovery &amp; Analysis</h2><h3>1.1 Project Overview</h3><p><strong>Tech Stack:</strong></p><ul> <li><strong>Language:</strong> Go 1.25+</li> <li><strong>Engine:</strong> Kaiju Engine (local submodule)</li> <li><strong>Graphics:</strong> Vulkan-based rendering</li> <li><strong>Platforms:</strong> Linux, Windows, macOS, Android, iOS</li> <li><strong>Build System:</strong> Make + GitHub Actions</li> </ul><p><strong>Core Packages:</strong></p><ul> <li><code>pkg/blocks</code> - Block registry and geometry</li> <li><code>pkg/crafting</code> - Items, inventory, recipes</li> <li><code>pkg/game</code> - Game controller and state</li> <li><code>pkg/network</code> - Client-server networking</li> <li><code>pkg/player</code> - Player state and management</li> <li><code>pkg/survival</code> - Health, environment</li> <li><code>pkg/ui</code> - HUD, menus, components</li> <li><code>pkg/world</code> - World generation, chunks, coordinates</li> <li><code>pkg/audio</code> - Audio management (only package with tests)</li> </ul><h3>1.2 Critical Issues Identified</h3><h4>Security Vulnerabilities</h4><ol> <li><strong>Unvalidated Network Message Size (Critical)</strong> <ul> <li><strong>Location:</strong> <code>pkg/network/protocol.go:399-405</code></li> <li><strong>Issue:</strong> Message length validated but still allows 64KB allocation per message</li> <li><strong>Impact:</strong> Potential DoS via memory exhaustion</li> <li><strong>Fix:</strong> Reduce max size, add rate limiting</li> </ul> </li> <li><strong>Unvalidated Player Name (Critical)</strong> <ul> <li><strong>Location:</strong> <code>pkg/network/server.go:229</code></li> <li><strong>Issue:</strong> Player name from network used without sanitization</li> <li><strong>Impact:</strong> Potential injection attacks, log poisoning</li> <li><strong>Fix:</strong> Validate length, sanitize characters</li> </ul> </li> <li><strong>Missing Rate Limiting (High)</strong> <ul> <li><strong>Location:</strong> <code>pkg/network/server.go:155-168</code></li> <li><strong>Issue:</strong> No throttling on connection attempts</li> <li><strong>Impact:</strong> Connection flood attacks</li> <li><strong>Fix:</strong> Implement connection rate limiting</li> </ul> </li> <li><strong>Path Traversal Risk (High)</strong> <ul> <li><strong>Location:</strong> <code>pkg/ui/menus.go:416, 647</code></li> <li><strong>Issue:</strong> World names used directly in file paths</li> <li><strong>Impact:</strong> Potential file system access outside intended directory</li> <li><strong>Fix:</strong> Sanitize world names, validate against whitelist</li> </ul> </li> </ol><h4>Code Quality Issues</h4><ol> <li><strong>Printf Format Bug (Critical)</strong> <ul> <li><strong>Location:</strong> <code>cmd/testapp/main.go:76</code>, <code>test_systems.go:76</code></li> <li><strong>Issue:</strong> <code>%.2f</code> format used with <code>int</code> type</li> <li><strong>Impact:</strong> Incorrect output, potential crashes</li> <li><strong>Fix:</strong> Use <code>%d</code> for integers</li> </ul> </li> <li><strong>String Conversion Bug (High)</strong> <ul> <li><strong>Location:</strong> <code>pkg/ui/hud.go:322, 422-449</code>, <code>pkg/ui/menus.go:102, 206, 407</code></li> <li><strong>Issue:</strong> <code>string(rune(int))</code> produces Unicode, not decimal strings</li> <li><strong>Impact:</strong> Incorrect UI display</li> <li><strong>Fix:</strong> Use <code>strconv.Itoa()</code> or <code>fmt.Sprintf()</code></li> </ul> </li> <li><strong>Missing Error Handling (High)</strong> <ul> <li><strong>Location:</strong> Multiple locations in network and UI code</li> <li><strong>Issue:</strong> Critical errors silently ignored</li> <li><strong>Impact:</strong> Silent failures, difficult debugging</li> <li><strong>Fix:</strong> Add proper error propagation</li> </ul> </li> <li><strong>Code Formatting (High)</strong> <ul> <li><strong>Location:</strong> Multiple files fail <code>gofmt -l</code></li> <li><strong>Issue:</strong> Inconsistent formatting</li> <li><strong>Impact:</strong> Code readability, CI failures</li> <li><strong>Fix:</strong> Run <code>gofmt -w .</code></li> </ul> </li> </ol><h4>Testing Gaps</h4><ol> <li><strong>No Unit Tests for Core Packages</strong> <ul> <li><strong>Affected:</strong> <code>pkg/blocks</code>, <code>pkg/crafting</code>, <code>pkg/world</code>, <code>pkg/network</code>, <code>pkg/player</code></li> <li><strong>Impact:</strong> No regression testing, high risk of bugs</li> <li><strong>Fix:</strong> Add comprehensive unit tests</li> </ul> </li> <li><strong>Duplicate Test Code</strong> <ul> <li><strong>Location:</strong> <code>test_systems.go</code> and <code>cmd/testapp/main.go</code></li> <li><strong>Issue:</strong> Nearly identical content, maintenance burden</li> <li><strong>Impact:</strong> Code duplication, drift risk</li> <li><strong>Fix:</strong> Consolidate into single test harness</li> </ul> </li> <li><strong>No Integration Tests</strong> <ul> <li><strong>Affected:</strong> Network protocol, world generation</li> <li><strong>Impact:</strong> End-to-end behavior untested</li> <li><strong>Fix:</strong> Add integration test suite</li> </ul> </li> </ol><h4>Architecture Concerns</h4><ol> <li><strong>Global Mutable State (Critical)</strong> <ul> <li><strong>Location:</strong> Global registries throughout codebase</li> <li><strong>Issue:</strong> Makes testing difficult, hides dependencies</li> <li><strong>Impact:</strong> Poor testability, tight coupling</li> <li><strong>Fix:</strong> Implement dependency injection</li> </ul> </li> <li><strong>Tight UI-Game Logic Coupling (High)</strong> <ul> <li><strong>Location:</strong> <code>pkg/ui/menus.go</code> imports <code>pkg/world</code></li> <li><strong>Issue:</strong> UI should be presentation layer only</li> <li><strong>Impact:</strong> Violates separation of concerns</li> <li><strong>Fix:</strong> Use callbacks/interfaces</li> </ul> </li> <li><strong>No Protocol Versioning (High)</strong> <ul> <li><strong>Location:</strong> <code>pkg/network/protocol.go</code></li> <li><strong>Issue:</strong> No mechanism for protocol evolution</li> <li><strong>Impact:</strong> Breaking changes difficult</li> <li><strong>Fix:</strong> Add protocol versioning scheme</li> </ul> </li> </ol><h3>1.3 Existing Issues &amp; PRs</h3><p><strong>Open Issues:</strong></p><ul> <li>#14: Game Wiki (enhancement, help wanted, good first issue)</li> <li>#10: New biomes suggestion (enhancement, help wanted, good first issue)</li> </ul><p><strong>Open PRs:</strong> None</p><p><strong>Observation:</strong> Low issue/PR activity suggests opportunity for community engagement.</p><hr><h2>2. Planning</h2><h3>2.1 Improvement Roadmap</h3><h4>Phase 1: Critical Fixes (Week 1-2)</h4><p><strong>Goal:</strong> Address all Critical and High-severity security/stability issues</p><table class="e-rte-table"> <thead> <tr> <th>Task</th> <th>Priority</th> <th>Effort</th> <th>Owner</th> </tr> </thead> <tbody><tr> <td>Fix printf format bugs</td> <td>Critical</td> <td>30 min</td> <td>-</td> </tr> <tr> <td>Add message size validation</td> <td>Critical</td> <td>1 hour</td> <td>-</td> </tr> <tr> <td>Fix string conversion bugs</td> <td>Critical</td> <td>2 hours</td> <td>-</td> </tr> <tr> <td>Add player name validation</td> <td>Critical</td> <td>1 hour</td> <td>-</td> </tr> <tr> <td>Add bounds checks to network handlers</td> <td>Critical</td> <td>2 hours</td> <td>-</td> </tr> <tr> <td>Fix missing error handling</td> <td>High</td> <td>3 hours</td> <td>-</td> </tr> <tr> <td>Add path traversal protection</td> <td>High</td> <td>1 hour</td> <td>-</td> </tr> <tr> <td>Run gofmt on all files</td> <td>High</td> <td>15 min</td> <td>-</td> </tr> <tr> <td>Add connection rate limiting</td> <td>High</td> <td>2 hours</td> <td>-</td> </tr> </tbody></table><p><strong>Total Effort:</strong> ~12.5 hours</p><h4>Phase 2: Code Quality &amp; Standards (Week 3-4)</h4><p><strong>Goal:</strong> Improve code quality, add documentation, establish standards</p><table class="e-rte-table"> <thead> <tr> <th>Task</th> <th>Priority</th> <th>Effort</th> <th>Owner</th> </tr> </thead> <tbody><tr> <td>Add package-level documentation</td> <td>High</td> <td>4 hours</td> <td>-</td> </tr> <tr> <td>Document all exported functions</td> <td>High</td> <td>8 hours</td> <td>-</td> </tr> <tr> <td>Fix magic numbers with constants</td> <td>Medium</td> <td>2 hours</td> <td>-</td> </tr> <tr> <td>Add input validation throughout</td> <td>High</td> <td>4 hours</td> <td>-</td> </tr> <tr> <td>Consolidate duplicate test code</td> <td>High</td> <td>2 hours</td> <td>-</td> </tr> <tr> <td>Add .gitignore and .editorconfig</td> <td>Medium</td> <td>30 min</td> <td>-</td> </tr> <tr> <td>Configure golangci-lint</td> <td>Medium</td> <td>1 hour</td> <td>-</td> </tr> <tr> <td>Fix long functions (refactor)</td> <td>Medium</td> <td>4 hours</td> <td>-</td> </tr> </tbody></table><p><strong>Total Effort:</strong> ~25.5 hours</p><h4>Phase 3: Testing Infrastructure (Month 2)</h4><p><strong>Goal:</strong> Establish comprehensive testing foundation</p><table class="e-rte-table"> <thead> <tr> <th>Task</th> <th>Priority</th> <th>Effort</th> <th>Owner</th> </tr> </thead> <tbody><tr> <td>Add unit tests for pkg/blocks</td> <td>High</td> <td>6 hours</td> <td>-</td> </tr> <tr> <td>Add unit tests for pkg/crafting</td> <td>High</td> <td>8 hours</td> <td>-</td> </tr> <tr> <td>Add unit tests for pkg/world</td> <td>High</td> <td>10 hours</td> <td>-</td> </tr> <tr> <td>Add unit tests for pkg/network</td> <td>High</td> <td>12 hours</td> <td>-</td> </tr> <tr> <td>Add unit tests for pkg/player</td> <td>Medium</td> <td>6 hours</td> <td>-</td> </tr> <tr> <td>Add integration tests for networking</td> <td>Medium</td> <td>8 hours</td> <td>-</td> </tr> <tr> <td>Add benchmarks for performance-critical code</td> <td>Medium</td> <td>6 hours</td> <td>-</td> </tr> <tr> <td>Set up coverage tracking</td> <td>Low</td> <td>2 hours</td> <td>-</td> </tr> </tbody></table><p><strong>Total Effort:</strong> ~58 hours</p><h4>Phase 4: Architecture Refactoring (Month 3-4)</h4><p><strong>Goal:</strong> Improve architecture for maintainability and testability</p><table class="e-rte-table"> <thead> <tr> <th>Task</th> <th>Priority</th> <th>Effort</th> <th>Owner</th> </tr> </thead> <tbody><tr> <td>Refactor global state to DI</td> <td>High</td> <td>20 hours</td> <td>-</td> </tr> <tr> <td>Decouple UI from game logic</td> <td>High</td> <td>12 hours</td> <td>-</td> </tr> <tr> <td>Add protocol versioning</td> <td>High</td> <td>8 hours</td> <td>-</td> </tr> <tr> <td>Split large controllers</td> <td>Medium</td> <td>16 hours</td> <td>-</td> </tr> <tr> <td>Add storage abstraction</td> <td>Medium</td> <td>8 hours</td> <td>-</td> </tr> <tr> <td>Implement proper error wrapping</td> <td>Medium</td> <td>6 hours</td> <td>-</td> </tr> </tbody></table><p><strong>Total Effort:</strong> ~70 hours</p><h3>2.2 Detailed Task Breakdown</h3><h4>Task 1: Fix Printf Format Bugs</h4><p><strong>Files:</strong> <code>cmd/testapp/main.go:76</code>, <code>test_systems.go:76</code></p><p><strong>Current Code:</strong></p><pre><code class="language-go">fmt.Printf("  ✓ Available block types: %.2f\n", len(allBlocks))
</code></pre><p><strong>Issue:</strong> Using <code>%.2f</code> (float) with <code>int</code> type</p><p><strong>Fix:</strong></p><pre><code class="language-go">fmt.Printf("  ✓ Available block types: %d\n", len(allBlocks))
</code></pre><p><strong>Testing:</strong> Run <code>go run test_systems.go</code> and verify output</p><hr><h4>Task 2: Add Message Size Validation</h4><p><strong>File:</strong> <code>pkg/network/protocol.go:399-405</code></p><p><strong>Current Code:</strong></p><pre><code class="language-go">length := binary.BigEndian.Uint32(lengthBytes)
if length &gt; 65536 { // Max message size
    return nil, fmt.Errorf("message too large: %d bytes", length)
}
messageBytes := make([]byte, length)
</code></pre><p><strong>Issue:</strong> 64KB still large, no rate limiting</p><p><strong>Fix:</strong></p><pre><code class="language-go">const MaxMessageSize = 4096 // 4KB max per message

length := binary.BigEndian.Uint32(lengthBytes)
if length &gt; MaxMessageSize {
    return nil, fmt.Errorf("message too large: %d bytes (max %d)", length, MaxMessageSize)
}
messageBytes := make([]byte, length)
</code></pre><p><strong>Additional:</strong> Add rate limiting in server connection handler</p><hr><h4>Task 3: Fix String Conversion Bugs</h4><p><strong>Files:</strong> Multiple in <code>pkg/ui/hud.go</code> and <code>pkg/ui/menus.go</code></p><p><strong>Current Code:</strong></p><pre><code class="language-go">text := string(rune(slotIndex))
</code></pre><p><strong>Issue:</strong> Produces Unicode character, not decimal string</p><p><strong>Fix:</strong></p><pre><code class="language-go">text := strconv.Itoa(slotIndex)
// or
text := fmt.Sprintf("%d", slotIndex)
</code></pre><p><strong>Testing:</strong> Verify UI displays correct numbers</p><hr><h4>Task 4: Add Player Name Validation</h4><p><strong>File:</strong> <code>pkg/network/server.go:229</code></p><p><strong>Current Code:</strong></p><pre><code class="language-go">client.Name = handshake.PlayerName
</code></pre><p><strong>Fix:</strong></p><pre><code class="language-go">// Validate player name
name := strings.TrimSpace(handshake.PlayerName)
if len(name) == 0 || len(name) &gt; 32 {
    response := &amp;HandshakeResponseMessage{
        Success:  false,
        PlayerID: 0,
        Message:  "Player name must be 1-32 characters",
    }
    // Send response and close connection
    // ...
}

// Sanitize name (allow only alphanumeric, spaces, underscores)
validName := regexp.MustCompile(`^[a-zA-Z0-9_ ]+$`).MatchString(name)
if !validName {
    response := &amp;HandshakeResponseMessage{
        Success:  false,
        PlayerID: 0,
        Message:  "Player name contains invalid characters",
    }
    // Send response and close connection
    // ...
}

client.Name = name
</code></pre><hr><h4>Task 5: Add Path Traversal Protection</h4><p><strong>File:</strong> <code>pkg/ui/menus.go:416, 647</code></p><p><strong>Current Code:</strong></p><pre><code class="language-go">worldPath := filepath.Join(saveDir, worldName)
</code></pre><p><strong>Fix:</strong></p><pre><code class="language-go">// Sanitize world name
sanitizedName := filepath.Base(filepath.Clean(worldName))
if sanitizedName != worldName || sanitizedName == "." || sanitizedName == ".." {
    log.Printf("Invalid world name: %s", worldName)
    return nil, fmt.Errorf("invalid world name")
}

worldPath := filepath.Join(saveDir, sanitizedName)
</code></pre><hr><h4>Task 6: Add Unit Tests for pkg/blocks</h4><p><strong>New File:</strong> <code>pkg/blocks/registry_test.go</code></p><pre><code class="language-go">package blocks

import (
    "testing"
)

func TestNewBlockRegistry(t *testing.T) {
    registry := NewBlockRegistry()
    if registry == nil {
        t.Fatal("NewBlockRegistry() returned nil")
    }
}

func TestGetBlock(t *testing.T) {
    registry := NewBlockRegistry()
    
    // Test getting valid block
    block := registry.GetBlock(BlockTypeStone)
    if block == nil {
        t.Error("GetBlock(BlockTypeStone) returned nil")
    }
    if block.Name != "Stone" {
        t.Errorf("Expected block name 'Stone', got '%s'", block.Name)
    }
    
    // Test getting invalid block
    block = registry.GetBlock(BlockType(999))
    if block != nil {
        t.Error("GetBlock(invalid) should return nil")
    }
}

func TestGetAllBlocks(t *testing.T) {
    registry := NewBlockRegistry()
    blocks := registry.GetAllBlocks()
    
    if len(blocks) == 0 {
        t.Error("GetAllBlocks() returned empty slice")
    }
    
    // Verify no duplicates
    seen := make(map[BlockType]bool)
    for _, block := range blocks {
        if seen[block.Type] {
            t.Errorf("Duplicate block type: %d", block.Type)
        }
        seen[block.Type] = true
    }
}

func TestBlockProperties(t *testing.T) {
    registry := NewBlockRegistry()
    
    // Test stone properties
    stone := registry.GetBlock(BlockTypeStone)
    if !stone.Solid {
        t.Error("Stone should be solid")
    }
    if stone.Transparent {
        t.Error("Stone should not be transparent")
    }
    
    // Test air properties
    air := registry.GetBlock(BlockTypeAir)
    if air.Solid {
        t.Error("Air should not be solid")
    }
    if !air.Transparent {
        t.Error("Air should be transparent")
    }
}
</code></pre><hr><h4>Task 7: Add Unit Tests for pkg/world</h4><p><strong>New File:</strong> <code>pkg/world/coordinates_test.go</code></p><pre><code class="language-go">package world

import (
    "testing"
    "kaijuengine.com/matrix"
)

func TestHexCoordToWorld(t *testing.T) {
    tests := []struct {
        name     string
        coord    HexCoord
        scale    float32
        expected matrix.Vec3
    }{
        {
            name:  "Origin",
            coord: HexCoord{Q: 0, R: 0},
            scale: 1.0,
            expected: matrix.NewVec3(0, 0, 0),
        },
        {
            name:  "Positive Q",
            coord: HexCoord{Q: 1, R: 0},
            scale: 1.0,
            expected: matrix.NewVec3(1.732, 0, 0),
        },
        {
            name:  "Positive R",
            coord: HexCoord{Q: 0, R: 1},
            scale: 1.0,
            expected: matrix.NewVec3(0.866, 0, 1.5),
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.coord.ToWorld(tt.scale)
            
            // Allow small floating point errors
            tolerance := float32(0.01)
            if abs(result.X()-tt.expected.X()) &gt; tolerance {
                t.Errorf("X: expected %.3f, got %.3f", tt.expected.X(), result.X())
            }
            if abs(result.Z()-tt.expected.Z()) &gt; tolerance {
                t.Errorf("Z: expected %.3f, got %.3f", tt.expected.Z(), result.Z())
            }
        })
    }
}

func TestHexCoordDistance(t *testing.T) {
    tests := []struct {
        name     string
        a        HexCoord
        b        HexCoord
        expected int
    }{
        {
            name:     "Same coordinate",
            a:        HexCoord{Q: 0, R: 0},
            b:        HexCoord{Q: 0, R: 0},
            expected: 0,
        },
        {
            name:     "Adjacent",
            a:        HexCoord{Q: 0, R: 0},
            b:        HexCoord{Q: 1, R: 0},
            expected: 1,
        },
        {
            name:     "Distance 2",
            a:        HexCoord{Q: 0, R: 0},
            b:        HexCoord{Q: 2, R: 0},
            expected: 2,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.a.Distance(tt.b)
            if result != tt.expected {
                t.Errorf("Expected distance %d, got %d", tt.expected, result)
            }
        })
    }
}

func TestHexCoordNeighbor(t *testing.T) {
    center := HexCoord{Q: 0, R: 0}
    
    // Test all 6 directions
    directions := []HexDirection{
        HexDirectionEast,
        HexDirectionNorthEast,
        HexDirectionNorthWest,
        HexDirectionWest,
        HexDirectionSouthWest,
        HexDirectionSouthEast,
    }
    
    for _, dir := range directions {
        neighbor := center.Neighbor(dir)
        distance := center.Distance(neighbor)
        if distance != 1 {
            t.Errorf("Neighbor distance should be 1, got %d", distance)
        }
    }
}

func TestWorldToHexCoord(t *testing.T) {
    tests := []struct {
        name     string
        worldPos matrix.Vec3
        scale    float32
        expected HexCoord
    }{
        {
            name:     "Origin",
            worldPos: matrix.NewVec3(0, 0, 0),
            scale:    1.0,
            expected: HexCoord{Q: 0, R: 0},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := FromWorld(tt.worldPos, tt.scale)
            if result != tt.expected {
                t.Errorf("Expected %+v, got %+v", tt.expected, result)
            }
        })
    }
}

func abs(x float32) float32 {
    if x &lt; 0 {
        return -x
    }
    return x
}
</code></pre><hr><h4>Task 8: Add Package Documentation</h4><p><strong>Example for pkg/blocks/registry.go:</strong></p><pre><code class="language-go">// Package blocks provides block type definitions, registry management,
// and geometry generation for hexagonal prism blocks.
//
// The block system is responsible for:
//   - Defining block types and their properties (solid, transparent, etc.)
//   - Managing the global block registry
//   - Generating mesh geometry for rendering
//   - Handling block placement and removal
//
// Block Types
//
// The following block types are defined:
//   - BlockTypeAir: Empty space, not solid, transparent
//   - BlockTypeStone: Solid, opaque block
//   - BlockTypeDirt: Solid, opaque block
//   - BlockTypeGrass: Solid, opaque block with grass top
//   - BlockTypeSand: Solid, opaque block
//   - BlockTypeWater: Not solid, transparent
//   - BlockTypeWood: Solid, opaque block
//   - BlockTypeLeaves: Solid, transparent block
//   - BlockTypeBrick: Solid, opaque block
//   - BlockTypeGlass: Not solid, transparent block
//   - BlockTypeBedrock: Solid, indestructible block
//
// Usage
//
//   registry := blocks.NewBlockRegistry()
//   block := registry.GetBlock(blocks.BlockTypeStone)
//   fmt.Println(block.Name) // "Stone"
//
// Thread Safety
//
// The block registry is thread-safe for concurrent reads after initialization.
// All modifications should be done during initialization before the game starts.
package blocks
</code></pre><hr><h4>Task 9: Configure golangci-lint</h4><p><strong>New File:</strong> <code>.golangci.yml</code></p><pre><code class="language-yaml">run:
  timeout: 5m
  tests: true
  modules-download-mode: readonly

linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - varcheck
    - ineffassign
    - deadcode
    - typecheck
    - gosec
    - gocyclo
    - dupl
    - goconst
    - misspell
    - lll
    - goimports
    - prealloc

linters-settings:
  govet:
    check-shadowing: true
  errcheck:
    check-type-assertions: true
    check-blank: true
  gocyclo:
    min-complexity: 15
  dupl:
    threshold: 100
  goconst:
    min-len: 3
    min-occurrences: 3
  lll:
    line-length: 120
  gosec:
    excludes:
      - G104 # Errors unhandled

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
</code></pre><hr><h4>Task 10: Add .gitignore</h4><p><strong>New File:</strong> <code>.gitignore</code></p><pre><code class="language-gitignore"># Binaries
dist/
*.exe
*.exe~
*.dll
*.so
*.dylib
tesselbox
tesselbox-*

# Test binary, built with `go test -c`
*.test

# Output of the go coverage tool
*.out
coverage.html
coverage.out

# Go workspace file
go.work

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Build artifacts
*.o
*.a

# Temporary files
*.tmp
*.log

# Game saves (optional - uncomment if you don't want to track saves)
# saves/
# worlds/
</code></pre><hr><h4>Task 11: Add .editorconfig</h4><p><strong>New File:</strong> <code>.editorconfig</code></p><pre><code class="language-ini">root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true

[*.{go,mod,sum}]
indent_style = tab
indent_size = 4

[*.{yml,yaml}]
indent_style = space
indent_size = 2

[*.{md}]
indent_style = space
indent_size = 2
trim_trailing_whitespace = false

[Makefile]
indent_style = tab
</code></pre><hr><h2>3. Implementation Guidance</h2><h3>3.1 Top 5 Priority Improvements</h3><h4>Priority 1: Fix Critical Security Vulnerabilities</h4><p><strong>Implementation Steps:</strong></p><ol> <li><p><strong>Create feature branch:</strong></p> <pre><code class="language-bash">git checkout -b fix/critical-security-vulnerabilities
</code></pre> </li> <li><strong>Fix message size validation:</strong> <ul> <li>Edit <code>pkg/network/protocol.go</code></li> <li>Reduce <code>MaxMessageSize</code> from 65536 to 4096</li> <li>Add constant definition at top of file</li> <li>Test with various message sizes</li> </ul> </li> <li><strong>Add player name validation:</strong> <ul> <li>Edit <code>pkg/network/server.go</code></li> <li>Add validation function</li> <li>Test with empty, too long, and invalid character names</li> </ul> </li> <li><strong>Add path traversal protection:</strong> <ul> <li>Edit <code>pkg/ui/menus.go</code></li> <li>Add sanitization function</li> <li>Test with malicious path names</li> </ul> </li> <li><strong>Add connection rate limiting:</strong> <ul> <li>Edit <code>pkg/network/server.go</code></li> <li>Implement rate limiter using <code>golang.org/x/time/rate</code></li> <li>Test with rapid connection attempts</li> </ul> </li> </ol><p><strong>Testing Strategy:</strong></p><pre><code class="language-bash"># Run all tests
make test

# Run security-focused tests
go test -v ./pkg/network/...

# Manual testing
# 1. Start server
make run-server

# 2. Test with various inputs
# - Try connecting with invalid names
# - Try sending oversized messages
# - Try rapid connections
</code></pre><p><strong>Potential Pitfalls:</strong></p><ul> <li>Breaking existing client connections when changing protocol</li> <li>Rate limiting too aggressive for legitimate users</li> <li>Validation too strict for valid use cases</li> </ul><p><strong>Best Practices:</strong></p><ul> <li>Add unit tests for validation functions</li> <li>Document validation rules in protocol docs</li> <li>Consider adding configuration options for limits</li> <li>Log rejected connections for monitoring</li> </ul><hr><h4>Priority 2: Fix Code Quality Issues</h4><p><strong>Implementation Steps:</strong></p><ol> <li><p><strong>Create feature branch:</strong></p> <pre><code class="language-bash">git checkout -b fix/code-quality-issues
</code></pre> </li> <li><p><strong>Fix printf format bugs:</strong></p> <pre><code class="language-bash"># Find all instances
grep -r "%.2f" --include="*.go" | grep -v "_test.go"

# Fix each instance
# Change %.2f to %d for integers
</code></pre> </li> <li><p><strong>Fix string conversion bugs:</strong></p> <pre><code class="language-bash"># Find all instances
grep -r "string(rune(" --include="*.go"

# Replace with strconv.Itoa()
# Add import "strconv" if needed
</code></pre> </li> <li><p><strong>Run gofmt:</strong></p> <pre><code class="language-bash"># Check which files need formatting
gofmt -l .

# Format all files
gofmt -w .

# Verify
gofmt -l .  # Should return empty
</code></pre> </li> <li><strong>Fix missing error handling:</strong> <ul> <li>Review all <code>err</code> variables</li> <li>Add proper error propagation</li> <li>Use <code>fmt.Errorf("context: %w", err)</code> for wrapping</li> </ul> </li> </ol><p><strong>Testing Strategy:</strong></p><pre><code class="language-bash"># Run linter
make lint

# Run tests
make test

# Build to verify no compilation errors
make dev

# Run go vet
go vet ./...
</code></pre><p><strong>Code Example - Error Handling:</strong></p><pre><code class="language-go">// Before
func LoadWorld(name string) (*World, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    // ...
}

// After
func LoadWorld(name string) (*World, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read world file %s: %w", name, err)
    }
    // ...
}
</code></pre><hr><h4>Priority 3: Add Unit Tests</h4><p><strong>Implementation Steps:</strong></p><ol> <li><p><strong>Create feature branch:</strong></p> <pre><code class="language-bash">git checkout -b feature/add-unit-tests
</code></pre> </li> <li><strong>Start with pkg/blocks:</strong> <ul> <li>Create <code>pkg/blocks/registry_test.go</code></li> <li>Test all exported functions</li> <li>Aim for &gt;80% coverage</li> </ul> </li> <li><strong>Move to pkg/world:</strong> <ul> <li>Create <code>pkg/world/coordinates_test.go</code></li> <li>Create <code>pkg/world/chunk_test.go</code></li> <li>Test coordinate conversions and chunk operations</li> </ul> </li> <li><strong>Add pkg/crafting tests:</strong> <ul> <li>Create <code>pkg/crafting/inventory_test.go</code></li> <li>Create <code>pkg/crafting/recipes_test.go</code></li> <li>Test inventory operations and recipe matching</li> </ul> </li> <li><strong>Add pkg/network tests:</strong> <ul> <li>Create <code>pkg/network/protocol_test.go</code></li> <li>Test message encoding/decoding</li> <li>Mock connections for testing</li> </ul> </li> </ol><p><strong>Testing Strategy:</strong></p><pre><code class="language-bash"># Run tests with coverage
go test -v -race -coverprofile=coverage.out ./pkg/...

# View coverage report
go tool cover -html=coverage.out -o coverage.html

# Run specific package tests
go test -v ./pkg/blocks/...

# Run with race detector
go test -race ./pkg/...
</code></pre><p><strong>Best Practices:</strong></p><ul> <li>Use table-driven tests for multiple cases</li> <li>Test both success and error paths</li> <li>Use <code>t.Run()</code> for subtests</li> <li>Keep tests fast (avoid sleep, use mocks)</li> <li>Test edge cases (nil, empty, max values)</li> </ul><p><strong>Example - Table-Driven Test:</strong></p><pre><code class="language-go">func TestGetBlock(t *testing.T) {
    registry := NewBlockRegistry()
    
    tests := []struct {
        name     string
        blockType BlockType
        wantName string
        wantSolid bool
    }{
        {
            name:      "Stone",
            blockType: BlockTypeStone,
            wantName:  "Stone",
            wantSolid: true,
        },
        {
            name:      "Air",
            blockType: BlockTypeAir,
            wantName:  "Air",
            wantSolid: false,
        },
        {
            name:      "Invalid",
            blockType: BlockType(999),
            wantName:  "",
            wantSolid: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := registry.GetBlock(tt.blockType)
            
            if tt.wantName == "" {
                if got != nil {
                    t.Errorf("GetBlock() = %v, want nil", got)
                }
                return
            }
            
            if got == nil {
                t.Fatalf("GetBlock() = nil, want non-nil")
            }
            
            if got.Name != tt.wantName {
                t.Errorf("GetBlock().Name = %v, want %v", got.Name, tt.wantName)
            }
            
            if got.Solid != tt.wantSolid {
                t.Errorf("GetBlock().Solid = %v, want %v", got.Solid, tt.wantSolid)
            }
        })
    }
}
</code></pre><hr><h4>Priority 4: Add Documentation</h4><p><strong>Implementation Steps:</strong></p><ol> <li><p><strong>Create feature branch:</strong></p> <pre><code class="language-bash">git checkout -b docs/add-package-documentation
</code></pre> </li> <li><strong>Add package docs:</strong> <ul> <li>Start with <code>pkg/blocks/registry.go</code></li> <li>Add <code>// Package blocks ...</code> comment</li> <li>Document all exported types</li> <li>Document all exported functions</li> </ul> </li> <li><strong>Add function docs:</strong> <ul> <li>Follow Go doc conventions</li> <li>Include parameters, returns, examples</li> <li>Add <code>Example*</code> functions for complex APIs</li> </ul> </li> <li><strong>Update README:</strong> <ul> <li>Add development section</li> <li>Add testing instructions</li> <li>Add contribution guidelines link</li> </ul> </li> </ol><p><strong>Documentation Template:</strong></p><pre><code class="language-go">// FunctionName does X and returns Y.
//
// The function performs the following steps:
//   1. Validate inputs
//   2. Process data
//   3. Return result
//
// Parameters:
//   - param1: Description of parameter 1
//   - param2: Description of parameter 2
//
// Returns:
//   - result1: Description of result 1
//   - error: Description of error conditions
//
// Example:
//   result, err := FunctionName(arg1, arg2)
//   if err != nil {
//       log.Fatal(err)
//   }
//   fmt.Println(result)
func FunctionName(param1 Type1, param2 Type2) (ResultType, error) {
    // ...
}
</code></pre><p><strong>Testing Strategy:</strong></p><pre><code class="language-bash"># Generate documentation
go doc ./pkg/blocks

# View in browser
godoc -http=:6060

# Check for missing docs
go doc -all ./pkg/... | grep "NO DOCUMENTATION"
</code></pre><hr><h4>Priority 5: Configure Development Tools</h4><p><strong>Implementation Steps:</strong></p><ol> <li><p><strong>Create feature branch:</strong></p> <pre><code class="language-bash">git checkout -b config/dev-tools
</code></pre> </li> <li><strong>Add .golangci.yml:</strong> <ul> <li>Copy configuration from Task 9</li> <li>Customize for project needs</li> <li>Test with <code>golangci-lint run</code></li> </ul> </li> <li><strong>Add .gitignore:</strong> <ul> <li>Copy from Task 10</li> <li>Add project-specific ignores</li> <li>Verify with <code>git status</code></li> </ul> </li> <li><strong>Add .editorconfig:</strong> <ul> <li>Copy from Task 11</li> <li>Test with editor</li> </ul> </li> <li><strong>Update Makefile:</strong> <ul> <li>Add <code>make lint</code> target</li> <li>Add <code>make coverage</code> target</li> <li>Add <code>make docs</code> target</li> </ul> </li> </ol><p><strong>Updated Makefile:</strong></p><pre><code class="language-makefile"># Test
test:
	go test -v -race -coverprofile=coverage.out ./pkg/...

# Coverage
coverage:
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Lint
lint:
	golangci-lint run

# Format
fmt:
	gofmt -w .
	goimports -w .

# Vet
vet:
	go vet ./...

# Docs
docs:
	godoc -http=:6060

# All checks
check: fmt vet lint test
	@echo "All checks passed!"

# Clean
clean:
	rm -rf dist/
	rm -f coverage.out coverage.html
	rm -f tesselbox tesselbox-*
</code></pre><p><strong>Testing Strategy:</strong></p><pre><code class="language-bash"># Test all new targets
make fmt
make vet
make lint
make test
make coverage
make check
</code></pre><hr><h3>3.2 Testing Strategies</h3><h4>Unit Testing</h4><p><strong>Approach:</strong></p><ul> <li>Test each function in isolation</li> <li>Use table-driven tests for multiple cases</li> <li>Mock external dependencies</li> <li>Test both success and error paths</li> </ul><p><strong>Tools:</strong></p><ul> <li><code>testing</code> package (built-in)</li> <li><code>testify/assert</code> for assertions (optional)</li> <li><code>gomock</code> for mocking (optional)</li> </ul><p><strong>Example:</strong></p><pre><code class="language-go">func TestInventoryAddItem(t *testing.T) {
    tests := []struct {
        name      string
        item      *Item
        quantity  int
        wantError bool
    }{
        {
            name:      "Add valid item",
            item:      &amp;Item{Type: ItemTypeStone},
            quantity:  10,
            wantError: false,
        },
        {
            name:      "Add nil item",
            item:      nil,
            quantity:  10,
            wantError: true,
        },
        {
            name:      "Add negative quantity",
            item:      &amp;Item{Type: ItemTypeStone},
            quantity:  -5,
            wantError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            inv := NewInventory()
            err := inv.AddItem(tt.item, tt.quantity)
            
            if (err != nil) != tt.wantError {
                t.Errorf("AddItem() error = %v, wantError %v", err, tt.wantError)
            }
        })
    }
}
</code></pre><h4>Integration Testing</h4><p><strong>Approach:</strong></p><ul> <li>Test interactions between components</li> <li>Use test servers/clients</li> <li>Test end-to-end flows</li> </ul><p><strong>Example:</strong></p><pre><code class="language-go">func TestNetworkHandshake(t *testing.T) {
    // Start test server
    server := NewTestServer(t, ":0")
    defer server.Close()
    
    // Connect client
    client := NewTestClient(t, server.Addr())
    defer client.Close()
    
    // Send handshake
    handshake := &amp;HandshakeMessage{
        Version:    "0.1.0",
        PlayerName: "TestPlayer",
    }
    err := client.SendMessage(handshake)
    if err != nil {
        t.Fatalf("Failed to send handshake: %v", err)
    }
    
    // Receive response
    msg, err := client.ReadMessage()
    if err != nil {
        t.Fatalf("Failed to read message: %v", err)
    }
    
    if msg.Type != MessageTypeHandshakeResponse {
        t.Errorf("Expected HandshakeResponse, got %v", msg.Type)
    }
    
    // Verify response
    response := DecodeHandshakeResponse(msg.Data)
    if !response.Success {
        t.Errorf("Handshake failed: %s", response.Message)
    }
}
</code></pre><h4>Benchmark Testing</h4><p><strong>Approach:</strong></p><ul> <li>Measure performance of critical paths</li> <li>Identify bottlenecks</li> <li>Track performance over time</li> </ul><p><strong>Example:</strong></p><pre><code class="language-go">func BenchmarkHexCoordDistance(b *testing.B) {
    a := HexCoord{Q: 100, R: 100}
    c := HexCoord{Q: 105, R: 105}
    
    b.ResetTimer()
    for i := 0; i &lt; b.N; i++ {
        a.Distance(c)
    }
}

func BenchmarkChunkMeshGeneration(b *testing.B) {
    chunk := NewChunk()
    // Fill chunk with blocks
    
    b.ResetTimer()
    for i := 0; i &lt; b.N; i++ {
        chunk.GenerateMesh()
    }
}
</code></pre><h4>Race Condition Testing</h4><p><strong>Approach:</strong></p><ul> <li>Run tests with race detector</li> <li>Identify concurrent access issues</li> <li>Fix with proper synchronization</li> </ul><p><strong>Command:</strong></p><pre><code class="language-bash">go test -race ./pkg/...
</code></pre><hr><h2>4. Git Workflow</h2><h3>4.1 Branching Strategy</h3><p><strong>Recommended Strategy: Git Flow</strong></p><pre><code>main (production)
  ↑
  develop (integration)
    ↑
    feature/fix-security-vulnerabilities
    feature/add-unit-tests
    feature/refactor-global-state
    fix/code-quality-issues
    docs/add-documentation
</code></pre><p><strong>Branch Types:</strong></p><ol> <li><strong>main</strong> <ul> <li>Production-ready code</li> <li>Tags for releases (v0.1.0, v0.2.0, etc.)</li> <li>Protected branch (requires PR review)</li> </ul> </li> <li><strong>develop</strong> <ul> <li>Integration branch</li> <li>All features merge here</li> <li>Nightly builds run from here</li> </ul> </li> <li><strong>feature/</strong> <ul> <li>New features</li> <li>Named: <code>feature/description</code></li> <li>Example: <code>feature/add-biome-system</code></li> </ul> </li> <li><strong>fix/</strong> <ul> <li>Bug fixes</li> <li>Named: <code>fix/description</code></li> <li>Example: <code>fix/memory-leak-in-chunk-loading</code></li> </ul> </li> <li><strong>docs/</strong> <ul> <li>Documentation changes</li> <li>Named: <code>docs/description</code></li> <li>Example: <code>docs/update-api-documentation</code></li> </ul> </li> <li><strong>refactor/</strong> <ul> <li>Code refactoring</li> <li>Named: <code>refactor/description</code></li> <li>Example: <code>refactor-dependency-injection</code></li> </ul> </li> <li><strong>hotfix/</strong> <ul> <li>Emergency fixes to main</li> <li>Named: <code>hotfix/description</code></li> <li>Example: <code>hotfix/critical-security-patch</code></li> </ul> </li> </ol><h3>4.2 Git Commands</h3><h4>Starting Work</h4><pre><code class="language-bash"># Ensure you're on develop and up to date
git checkout develop
git pull origin develop

# Create feature branch
git checkout -b feature/add-unit-tests

# Or for a fix
git checkout -b fix/critical-security-vulnerabilities
</code></pre><h4>Making Changes</h4><pre><code class="language-bash"># Stage changes
git add pkg/blocks/registry_test.go

# Commit with conventional commit message
git commit -m "test: add unit tests for block registry

- Add tests for NewBlockRegistry
- Add tests for GetBlock
- Add tests for GetAllBlocks
- Achieve 85% coverage for pkg/blocks

Closes #123"
</code></pre><h4>Commit Message Conventions</h4><p><strong>Format:</strong></p><pre><code>&lt;type&gt;(&lt;scope&gt;): &lt;subject&gt;

&lt;body&gt;

&lt;footer&gt;
</code></pre><p><strong>Types:</strong></p><ul> <li><code>feat</code>: New feature</li> <li><code>fix</code>: Bug fix</li> <li><code>docs</code>: Documentation changes</li> <li><code>style</code>: Code style changes (formatting, etc.)</li> <li><code>refactor</code>: Code refactoring</li> <li><code>test</code>: Adding or updating tests</li> <li><code>chore</code>: Maintenance tasks</li> <li><code>perf</code>: Performance improvements</li> <li><code>ci</code>: CI/CD changes</li> <li><code>build</code>: Build system changes</li> </ul><p><strong>Examples:</strong></p><pre><code>feat(network): add connection rate limiting

Implement rate limiting for incoming connections to prevent
DoS attacks. Uses golang.org/x/time/rate with configurable
limits.

- Add rate limiter to server
- Add configuration options
- Add tests for rate limiting

Closes #45
</code></pre><pre><code>fix(protocol): reduce max message size to 4KB

Reduce maximum message size from 64KB to 4KB to prevent
memory exhaustion attacks. Also adds proper error handling
for oversized messages.

Fixes #67
</code></pre><pre><code>test(world): add unit tests for coordinate system

Add comprehensive unit tests for hexagonal coordinate
conversions and distance calculations. Achieves 90%
coverage for pkg/world/coordinates.

- Test ToWorld conversion
- Test FromWorld conversion
- Test Distance calculation
- Test Neighbor calculation
</code></pre><h4>Pushing Changes</h4><pre><code class="language-bash"># Push to remote
git push origin feature/add-unit-tests

# If branch doesn't exist on remote
git push -u origin feature/add-unit-tests
</code></pre><h4>Creating Pull Request</h4><pre><code class="language-bash"># Using GitHub CLI
gh pr create \
  --title "test: add unit tests for block registry" \
  --body "Add comprehensive unit tests for the block registry system.

## Changes
- Add tests for NewBlockRegistry
- Add tests for GetBlock
- Add tests for GetAllBlocks
- Achieve 85% coverage for pkg/blocks

## Testing
- All tests pass
- Coverage report included
- No race conditions detected

## Checklist
- [x] Tests added
- [x] Documentation updated
- [x] Code follows style guidelines
- [x] No linting errors

Closes #123" \
  --base develop \
  --head feature/add-unit-tests
</code></pre><h4>Merging PR</h4><pre><code class="language-bash"># After PR is approved and tests pass
gh pr merge 123 --squash

# Or merge without squash
gh pr merge 123 --merge

# Or rebase and merge
gh pr merge 123 --rebase
</code></pre><h4>Updating Branch</h4><pre><code class="language-bash"># Sync with develop
git checkout develop
git pull origin develop

# Rebase feature branch
git checkout feature/add-unit-tests
git rebase develop

# Resolve conflicts if any
# git add &lt;resolved-files&gt;
# git rebase --continue

# Force push (careful!)
git push --force-with-lease origin feature/add-unit-tests
</code></pre><h4>Cleaning Up</h4><pre><code class="language-bash"># After merge, delete local branch
git branch -d feature/add-unit-tests

# Delete remote branch
git push origin --delete feature/add-unit-tests

# Or using GitHub CLI
gh pr view 123 --json url
gh pr close 123 --delete-branch
</code></pre><h3>4.3 Release Workflow</h3><pre><code class="language-bash"># 1. Merge all features to develop
# (via PRs)

# 2. Create release branch from develop
git checkout develop
git pull origin develop
git checkout -b release/v0.2.0

# 3. Update version numbers
# (in go.mod, README, etc.)

# 4. Finalize release
git commit -m "chore: release v0.2.0"

# 5. Merge to main
git checkout main
git merge --no-ff release/v0.2.0
git tag -a v0.2.0 -m "Release v0.2.0"

# 6. Push
git push origin main
git push origin v0.2.0

# 7. Merge back to develop
git checkout develop
git merge --no-ff release/v0.2.0
git push origin develop

# 8. Delete release branch
git branch -d release/v0.2.0
git push origin --delete release/v0.2.0

# 9. GitHub Actions will create release automatically
</code></pre><h3>4.4 Best Practices</h3><ol> <li><strong>Keep branches small and focused</strong> <ul> <li>One feature/fix per branch</li> <li>Short-lived branches (days, not weeks)</li> </ul> </li> <li><strong>Commit frequently</strong> <ul> <li>Small, atomic commits</li> <li>Clear commit messages</li> </ul> </li> <li><strong>Write tests before merging</strong> <ul> <li>All tests must pass</li> <li>Aim for good coverage</li> </ul> </li> <li><strong>Code review required</strong> <ul> <li>At least one approval</li> <li>Address all review comments</li> </ul> </li> <li><strong>Keep main clean</strong> <ul> <li>Only merge tested code</li> <li>Use tags for releases</li> </ul> </li> <li><strong>Use meaningful branch names</strong> <ul> <li><code>feature/add-biome-system</code> ✓</li> <li><code>fix/memory-leak</code> ✓</li> <li><code>stuff</code> ✗</li> <li><code>wip</code> ✗</li> </ul> </li> <li><strong>Resolve conflicts locally</strong> <ul> <li>Don't push conflicts</li> <li>Test after resolving</li> </ul> </li> <li><strong>Protect important branches</strong> <ul> <li>Require PR for main/develop</li> <li>Require status checks to pass</li> <li>Require code review</li> </ul> </li> </ol><hr><h2>5. Summary &amp; Next Steps</h2><h3>5.1 Quick Start Guide</h3><p><strong>For Immediate Action (This Week):</strong></p><pre><code class="language-bash"># 1. Clone and setup
git clone https://github.com/tesselstudio/TesselBox.git
cd TesselBox
git checkout -b fix/critical-security-vulnerabilities

# 2. Fix critical issues
# - Edit pkg/network/protocol.go (message size)
# - Edit pkg/network/server.go (player name validation)
# - Edit pkg/ui/menus.go (path traversal)

# 3. Test changes
make test
make lint

# 4. Commit and push
git add .
git commit -m "fix: address critical security vulnerabilities

- Reduce max message size to 4KB
- Add player name validation
- Add path traversal protection

Fixes #67, #68, #69"
git push -u origin fix/critical-security-vulnerabilities

# 5. Create PR
gh pr create --base develop
</code></pre><h3>5.2 Priority Matrix</h3><table class="e-rte-table"> <thead> <tr> <th>Priority</th> <th>Tasks</th> <th>Timeframe</th> <th>Impact</th> </tr> </thead> <tbody><tr> <td><strong>P0</strong></td> <td>Critical security fixes</td> <td>Week 1-2</td> <td>Prevents exploits</td> </tr> <tr> <td><strong>P1</strong></td> <td>Code quality issues</td> <td>Week 2-3</td> <td>Improves stability</td> </tr> <tr> <td><strong>P2</strong></td> <td>Unit tests</td> <td>Month 2</td> <td>Enables safe refactoring</td> </tr> <tr> <td><strong>P3</strong></td> <td>Documentation</td> <td>Month 2-3</td> <td>Improves onboarding</td> </tr> <tr> <td><strong>P4</strong></td> <td>Architecture refactoring</td> <td>Month 3-4</td> <td>Long-term maintainability</td> </tr> </tbody></table><h3>5.3 Success Metrics</h3><p><strong>Phase 1 (Critical Fixes):</strong></p><ul> <li><input disabled="" type="checkbox"> All 6 critical issues resolved</li> <li><input disabled="" type="checkbox"> All 18 high-priority issues resolved</li> <li><input disabled="" type="checkbox"> Zero security vulnerabilities</li> <li><input disabled="" type="checkbox"> All tests passing</li> </ul><p><strong>Phase 2 (Code Quality):</strong></p><ul> <li><input disabled="" type="checkbox"> 100% files pass gofmt</li> <li><input disabled="" type="checkbox"> Zero linting errors</li> <li><input disabled="" type="checkbox"> All exported functions documented</li> <li><input disabled="" type="checkbox"> CI/CD green on all branches</li> </ul><p><strong>Phase 3 (Testing):</strong></p><ul> <li><input disabled="" type="checkbox"> <blockquote> <p>80% code coverage</p> </blockquote> </li> <li><input disabled="" type="checkbox"> Unit tests for all core packages</li> <li><input disabled="" type="checkbox"> Integration tests for networking</li> <li><input disabled="" type="checkbox"> Benchmarks for critical paths</li> </ul><p><strong>Phase 4 (Architecture):</strong></p><ul> <li><input disabled="" type="checkbox"> No global mutable state</li> <li><input disabled="" type="checkbox"> Clear separation of concerns</li> <li><input disabled="" type="checkbox"> Protocol versioning implemented</li> <li><input disabled="" type="checkbox"> Dependency injection throughout</li> </ul><h3>5.4 Resources</h3><p><strong>Documentation:</strong></p><ul> <li>Go Documentation: <a href="https://golang.org/doc/">https://golang.org/doc/</a></li> <li>Effective Go: <a href="https://golang.org/doc/effective_go">https://golang.org/doc/effective_go</a></li> <li>Go Code Review Comments: <a href="https://github.com/golang/go/wiki/CodeReviewComments">https://github.com/golang/go/wiki/CodeReviewComments</a></li> </ul><p><strong>Tools:</strong></p><ul> <li>golangci-lint: <a href="https://golangci-lint.run/">https://golangci-lint.run/</a></li> <li>gofmt: <a href="https://golang.org/cmd/gofmt/">https://golang.org/cmd/gofmt/</a></li> <li>go vet: <a href="https://golang.org/cmd/vet/">https://golang.org/cmd/vet/</a></li> </ul><p><strong>Testing:</strong></p><ul> <li>Testing Package: <a href="https://golang.org/pkg/testing/">https://golang.org/pkg/testing/</a></li> <li>Testify: <a href="https://github.com/stretchr/testify">https://github.com/stretchr/testify</a></li> <li>Gomock: <a href="https://github.com/golang/mock">https://github.com/golang/mock</a></li> </ul><p><strong>CI/CD:</strong></p><ul> <li>GitHub Actions: <a href="https://docs.github.com/en/actions">https://docs.github.com/en/actions</a></li> <li>Conventional Commits: <a href="https://www.conventionalcommits.org/">https://www.conventionalcommits.org/</a></li> </ul><hr><h2>Appendix A: File-by-File Priority</h2><table class="e-rte-table"> <thead> <tr> <th>File</th> <th>Lines</th> <th>Critical</th> <th>High</th> <th>Medium</th> <th>Low</th> <th>Total</th> <th>Priority</th> </tr> </thead> <tbody><tr> <td><code>pkg/network/server.go</code></td> <td>651</td> <td>2</td> <td>3</td> <td>4</td> <td>0</td> <td>9</td> <td><strong>CRITICAL</strong></td> </tr> <tr> <td><code>pkg/ui/menus.go</code></td> <td>901</td> <td>2</td> <td>3</td> <td>5</td> <td>4</td> <td>14</td> <td><strong>CRITICAL</strong></td> </tr> <tr> <td><code>pkg/network/protocol.go</code></td> <td>439</td> <td>1</td> <td>2</td> <td>3</td> <td>0</td> <td>6</td> <td><strong>CRITICAL</strong></td> </tr> <tr> <td><code>pkg/world/coordinates.go</code></td> <td>331</td> <td>1</td> <td>0</td> <td>2</td> <td>2</td> <td>5</td> <td><strong>HIGH</strong></td> </tr> <tr> <td><code>pkg/world/world.go</code></td> <td>222</td> <td>0</td> <td>1</td> <td>2</td> <td>2</td> <td>5</td> <td><strong>HIGH</strong></td> </tr> <tr> <td><code>pkg/crafting/inventory.go</code></td> <td>441</td> <td>0</td> <td>1</td> <td>3</td> <td>0</td> <td>4</td> <td><strong>HIGH</strong></td> </tr> <tr> <td><code>pkg/game/controller.go</code></td> <td>575</td> <td>0</td> <td>1</td> <td>2</td> <td>3</td> <td>6</td> <td><strong>HIGH</strong></td> </tr> <tr> <td><code>pkg/network/client.go</code></td> <td>571</td> <td>0</td> <td>1</td> <td>3</td> <td>2</td> <td>6</td> <td><strong>HIGH</strong></td> </tr> <tr> <td><code>pkg/ui/components.go</code></td> <td>528</td> <td>0</td> <td>1</td> <td>4</td> <td>2</td> <td>7</td> <td><strong>HIGH</strong></td> </tr> <tr> <td><code>pkg/ui/hud.go</code></td> <td>469</td> <td>0</td> <td>1</td> <td>3</td> <td>2</td> <td>6</td> <td><strong>HIGH</strong></td> </tr> <tr> <td><code>pkg/blocks/geometry.go</code></td> <td>290</td> <td>0</td> <td>0</td> <td>2</td> <td>3</td> <td>5</td> <td><strong>MEDIUM</strong></td> </tr> <tr> <td><code>pkg/blocks/placement.go</code></td> <td>251</td> <td>0</td> <td>0</td> <td>3</td> <td>0</td> <td>3</td> <td><strong>MEDIUM</strong></td> </tr> <tr> <td><code>pkg/blocks/registry.go</code></td> <td>298</td> <td>0</td> <td>0</td> <td>2</td> <td>4</td> <td>6</td> <td><strong>MEDIUM</strong></td> </tr> <tr> <td><code>pkg/crafting/items.go</code></td> <td>413</td> <td>0</td> <td>0</td> <td>2</td> <td>3</td> <td>5</td> <td><strong>MEDIUM</strong></td> </tr> <tr> <td><code>pkg/crafting/recipes.go</code></td> <td>498</td> <td>0</td> <td>0</td> <td>2</td> <td>2</td> <td>4</td> <td><strong>MEDIUM</strong></td> </tr> <tr> <td><code>pkg/player/player.go</code></td> <td>488</td> <td>0</td> <td>0</td> <td>2</td> <td>4</td> <td>6</td> <td><strong>MEDIUM</strong></td> </tr> <tr> <td><code>pkg/survival/environment.go</code></td> <td>496</td> <td>0</td> <td>0</td> <td>2</td> <td>4</td> <td>6</td> <td><strong>MEDIUM</strong></td> </tr> <tr> <td><code>pkg/survival/health.go</code></td> <td>638</td> <td>0</td> <td>0</td> <td>2</td> <td>3</td> <td>5</td> <td><strong>MEDIUM</strong></td> </tr> <tr> <td><code>cmd/main.go</code></td> <td>662</td> <td>0</td> <td>0</td> <td>0</td> <td>3</td> <td>3</td> <td><strong>LOW</strong></td> </tr> <tr> <td><code>cmd/testapp/main.go</code></td> <td>177</td> <td>1</td> <td>0</td> <td>0</td> <td>2</td> <td>3</td> <td><strong>HIGH</strong></td> </tr> <tr> <td><code>test_systems.go</code></td> <td>177</td> <td>1</td> <td>0</td> <td>0</td> <td>2</td> <td>3</td> <td><strong>HIGH</strong></td> </tr> </tbody></table><hr><h2>Appendix B: Checklist Templates</h2><h3>Pre-Commit Checklist</h3><ul> <li><input disabled="" type="checkbox"> Code follows project style guide</li> <li><input disabled="" type="checkbox"> All tests pass (<code>make test</code>)</li> <li><input disabled="" type="checkbox"> No linting errors (<code>make lint</code>)</li> <li><input disabled="" type="checkbox"> No vet errors (<code>go vet ./...</code>)</li> <li><input disabled="" type="checkbox"> Code is formatted (<code>gofmt -w .</code>)</li> <li><input disabled="" type="checkbox"> Documentation updated (if needed)</li> <li><input disabled="" type="checkbox"> Commit message follows conventions</li> <li><input disabled="" type="checkbox"> Changes are focused and atomic</li> </ul><h3>Pre-PR Checklist</h3><ul> <li><input disabled="" type="checkbox"> Branch is up to date with develop</li> <li><input disabled="" type="checkbox"> All commits are squashed/clean</li> <li><input disabled="" type="checkbox"> PR description is clear</li> <li><input disabled="" type="checkbox"> Related issues referenced</li> <li><input disabled="" type="checkbox"> Tests added for new features</li> <li><input disabled="" type="checkbox"> Breaking changes documented</li> <li><input disabled="" type="checkbox"> CI/CD passes on all platforms</li> </ul><h3>Code Review Checklist</h3><ul> <li><input disabled="" type="checkbox"> Code is readable and maintainable</li> <li><input disabled="" type="checkbox"> No security vulnerabilities</li> <li><input disabled="" type="checkbox"> Error handling is proper</li> <li><input disabled="" type="checkbox"> Tests are comprehensive</li> <li><input disabled="" type="checkbox"> Documentation is accurate</li> <li><input disabled="" type="checkbox"> Performance is acceptable</li> <li><input disabled="" type="checkbox"> No unnecessary complexity</li> <li><input disabled="" type="checkbox"> Follows project conventions</li> </ul><hr><p><strong>End of Improvement Plan</strong></p><p><em>This document provides a comprehensive roadmap for improving TesselBox. Start with Phase 1 (Critical Fixes) and work through each phase systematically. Regular communication with the team and continuous feedback will ensure successful implementation.</em></p>