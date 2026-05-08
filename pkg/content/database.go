/******************************************************************************/
/* database.go                                                                */
/******************************************************************************/
/*                            TesselBox Game Content                          */
/*                      https://github.com/tesselstudio                       */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2025 TesselStudio                                              */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                          */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package content

import (
	"os"
	"path/filepath"
	"strings"
)

// GameContentDatabase handles path transformations for game content
type GameContentDatabase struct {
	rootPath string
}

// NewGameContentDatabase creates a new game content database with path transformation
func NewGameContentDatabase(root string) (*GameContentDatabase, error) {
	return &GameContentDatabase{rootPath: root}, nil
}

// transformPath converts relative content paths to full paths based on file extension
func transformPath(key string) string {
	// If already has directory separators, assume it's already a full path
	if strings.Contains(key, "/") || strings.Contains(key, "\\") {
		return key
	}
	// If key already starts with known folder prefix, don't transform
	if strings.HasPrefix(key, "fonts/") ||
		strings.HasPrefix(key, "textures/") ||
		strings.HasPrefix(key, "meshes/") ||
		strings.HasPrefix(key, "ui/") ||
		strings.HasPrefix(key, "renderer/") {
		return key
	}

	ext := strings.ToLower(filepath.Ext(key))
	switch ext {
	case ".bin":
		return filepath.Join("fonts", key)
	case ".gltf", ".glb":
		return filepath.Join("meshes", key)
	case ".png", ".jpg", ".jpeg":
		// Try textures first, then fonts
		return filepath.Join("textures", key)
	case ".css", ".html":
		return filepath.Join("ui", key)
	case ".material":
		return filepath.Join("renderer", "materials", key)
	case ".renderpass":
		return filepath.Join("renderer", "passes", key)
	case ".shaderpipeline":
		return filepath.Join("renderer", "pipelines", key)
	case ".shader":
		return filepath.Join("renderer", "shaders", key)
	case ".spv":
		return filepath.Join("renderer", "spv", key)
	default:
		return key
	}
}

func (g *GameContentDatabase) PostWindowCreate(handle interface{}) error {
	// No-op for simplified implementation
	return nil
}

func (g *GameContentDatabase) Cache(key string, data []byte) {
	// No-op for simplified implementation
}

func (g *GameContentDatabase) CacheRemove(key string) {
	// No-op for simplified implementation
}

func (g *GameContentDatabase) CacheClear() {
	// No-op for simplified implementation
}

func (g *GameContentDatabase) Read(key string) ([]byte, error) {
	// For PNGs without path prefix, try fonts first (for font textures), then textures
	if ext := strings.ToLower(filepath.Ext(key)); ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
		// Only transform if it's just a filename without folder prefix
		if !strings.Contains(key, "/") && !strings.Contains(key, "\\") {
			// Try fonts folder first
			if data, err := os.ReadFile(filepath.Join(g.rootPath, "fonts", key)); err == nil {
				return data, nil
			}
			// Fall back to textures folder
			return os.ReadFile(filepath.Join(g.rootPath, "textures", key))
		}
	}
	return os.ReadFile(filepath.Join(g.rootPath, transformPath(key)))
}

func (g *GameContentDatabase) ReadText(key string) (string, error) {
	// For PNGs without path prefix, try fonts first (for font textures), then textures
	if ext := strings.ToLower(filepath.Ext(key)); ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
		// Only transform if it's just a filename without folder prefix
		if !strings.Contains(key, "/") && !strings.Contains(key, "\\") {
			if data, err := os.ReadFile(filepath.Join(g.rootPath, "fonts", key)); err == nil {
				return string(data), nil
			}
			data, err := os.ReadFile(filepath.Join(g.rootPath, "textures", key))
			return string(data), err
		}
	}
	data, err := os.ReadFile(filepath.Join(g.rootPath, transformPath(key)))
	return string(data), err
}

func (g *GameContentDatabase) Exists(key string) bool {
	// For PNGs without path prefix, check fonts first, then textures
	if ext := strings.ToLower(filepath.Ext(key)); ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
		// Only transform if it's just a filename without folder prefix
		if !strings.Contains(key, "/") && !strings.Contains(key, "\\") {
			if _, err := os.Stat(filepath.Join(g.rootPath, "fonts", key)); err == nil {
				return true
			}
			_, err := os.Stat(filepath.Join(g.rootPath, "textures", key))
			return err == nil
		}
	}
	_, err := os.Stat(filepath.Join(g.rootPath, transformPath(key)))
	return err == nil
}

func (g *GameContentDatabase) Close() {
	// No-op for simplified implementation
}
