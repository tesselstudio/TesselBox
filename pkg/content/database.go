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
	"path/filepath"
	"strings"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/platform/profiler/tracing"
)

// GameContentDatabase wraps a FileDatabase and handles path transformations
// for game content, similar to how EditorContent works for editor content.
type GameContentDatabase struct {
	inner assets.Database
}

// NewGameContentDatabase creates a new game content database with path transformation
func NewGameContentDatabase(root string) (assets.Database, error) {
	inner, err := assets.NewFileDatabase(root)
	if err != nil {
		return nil, err
	}
	return &GameContentDatabase{inner: inner}, nil
}

// transformPath converts relative content paths to full paths based on file extension
func transformPath(key string) string {
	// If already has directory separators, assume it's already a full path
	if strings.Contains(key, "/") || strings.Contains(key, "\\") {
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

func (g *GameContentDatabase) PostWindowCreate(handle assets.PostWindowCreateHandle) error {
	return g.inner.PostWindowCreate(handle)
}

func (g *GameContentDatabase) Cache(key string, data []byte) {
	g.inner.Cache(transformPath(key), data)
}

func (g *GameContentDatabase) CacheRemove(key string) {
	g.inner.CacheRemove(transformPath(key))
}

func (g *GameContentDatabase) CacheClear() {
	g.inner.CacheClear()
}

func (g *GameContentDatabase) Read(key string) ([]byte, error) {
	defer tracing.NewRegion("GameContentDatabase.Read: " + key).End()
	// For PNGs, try fonts first (for font textures), then textures
	if ext := strings.ToLower(filepath.Ext(key)); ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
		// Try fonts folder first
		if data, err := g.inner.Read(filepath.Join("fonts", key)); err == nil {
			return data, nil
		}
		// Fall back to textures folder
		return g.inner.Read(filepath.Join("textures", key))
	}
	return g.inner.Read(transformPath(key))
}

func (g *GameContentDatabase) ReadText(key string) (string, error) {
	defer tracing.NewRegion("GameContentDatabase.ReadText: " + key).End()
	// For PNGs, try fonts first (for font textures), then textures
	if ext := strings.ToLower(filepath.Ext(key)); ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
		if data, err := g.inner.ReadText(filepath.Join("fonts", key)); err == nil {
			return data, nil
		}
		return g.inner.ReadText(filepath.Join("textures", key))
	}
	return g.inner.ReadText(transformPath(key))
}

func (g *GameContentDatabase) Exists(key string) bool {
	defer tracing.NewRegion("GameContentDatabase.Exists: " + key).End()
	// For PNGs, check fonts first, then textures
	if ext := strings.ToLower(filepath.Ext(key)); ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
		if g.inner.Exists(filepath.Join("fonts", key)) {
			return true
		}
		return g.inner.Exists(filepath.Join("textures", key))
	}
	return g.inner.Exists(transformPath(key))
}

func (g *GameContentDatabase) Close() {
	g.inner.Close()
}
