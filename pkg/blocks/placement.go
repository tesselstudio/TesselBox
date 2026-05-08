package blocks

import (
	"math"

	"github.com/tesselstudio/TesselBox/pkg/types"
)

// Block represents a placed block in the world
type Block struct {
	Type       BlockType
	Position   types.Vec3
	Rotation   int // 0-5 orientations
	Properties *BlockProperties
	Prism      *HexPrism
}

// NewBlock creates a new block instance
func NewBlock(blockType BlockType, position types.Vec3, rotation int) (*Block, error) {
	registry := GetGlobalRegistry()
	props, exists := registry.GetBlockByType(blockType)
	if !exists {
		return nil, ErrBlockNotFound
	}

	// Create hexagonal prism for this block
	prism := NewHexPrism(position, 1.0, 2.0) // Default dimensions
	prism.SetRotation(rotation)

	return &Block{
		Type:       blockType,
		Position:   position,
		Rotation:   rotation,
		Properties: props,
		Prism:      prism,
	}, nil
}

// CanAttach checks if another block can attach to this block on a specific face
func (b *Block) CanAttach(face AttachmentFace, otherBlock *Block) bool {
	// Different attachment rules based on block type
	switch b.Type {
	case BlockTypeFull:
		return b.canAttachFull(face, otherBlock)
	case BlockTypeHalfVertical:
		return b.canAttachHalfVertical(face, otherBlock)
	case BlockTypeHalfHorizontal:
		return b.canAttachHalfHorizontal(face, otherBlock)
	case BlockTypeCorner:
		return b.canAttachCorner(face, otherBlock)
	case BlockTypeStairs:
		return b.canAttachStairs(face, otherBlock)
	case BlockTypeSlab:
		return b.canAttachSlab(face, otherBlock)
	default:
		return false
	}
}

// canAttachFull handles attachment for full hexagonal prisms
func (b *Block) canAttachFull(face AttachmentFace, otherBlock *Block) bool {
	// Full blocks can attach on any face
	// Check if the other block is solid enough to support attachment
	if !otherBlock.Properties.Solid {
		return false
	}

	// Check for collision (simplified)
	dx := b.Position.X - otherBlock.Position.X
	dy := b.Position.Y - otherBlock.Position.Y
	dz := b.Position.Z - otherBlock.Position.Z
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
	if distance < 0.1 { // Blocks are too close
		return false
	}

	return true
}

// canAttachHalfVertical handles attachment for vertical half blocks
func (b *Block) canAttachHalfVertical(face AttachmentFace, otherBlock *Block) bool {
	// Vertical half blocks have limited attachment surfaces
	if face.Type == FaceTypeTop || face.Type == FaceTypeBottom {
		return false // Can't attach on flat ends of vertical half
	}

	return otherBlock.Properties.Solid
}

// canAttachHalfHorizontal handles attachment for horizontal half blocks
func (b *Block) canAttachHalfHorizontal(face AttachmentFace, otherBlock *Block) bool {
	// Horizontal half blocks can attach on sides and top
	if face.Type == FaceTypeBottom {
		return false // Can't attach on bottom
	}

	return otherBlock.Properties.Solid
}

// canAttachCorner handles attachment for corner blocks
func (b *Block) canAttachCorner(face AttachmentFace, otherBlock *Block) bool {
	// Corner blocks have very limited attachment points
	// Only allow attachment on specific faces based on rotation
	allowedFaces := b.getCornerAttachmentFaces()

	for _, allowedFace := range allowedFaces {
		if face.Index == allowedFace {
			return otherBlock.Properties.Solid
		}
	}

	return false
}

// canAttachStairs handles attachment for stair blocks
func (b *Block) canAttachStairs(face AttachmentFace, otherBlock *Block) bool {
	// Stairs have complex attachment rules based on rotation
	// Simplified: allow attachment on most faces except bottom
	if face.Type == FaceTypeBottom {
		return false
	}

	return otherBlock.Properties.Solid
}

// canAttachSlab handles attachment for slab blocks
func (b *Block) canAttachSlab(face AttachmentFace, otherBlock *Block) bool {
	// Slabs are thin and have limited attachment
	// Allow attachment on top and sides, but not bottom
	if face.Type == FaceTypeBottom {
		return false
	}

	return otherBlock.Properties.Solid
}

// getCornerAttachmentFaces returns the valid attachment faces for a corner block
func (b *Block) getCornerAttachmentFaces() []int {
	// Corner blocks can only attach on 3 faces based on rotation
	// This is a simplified implementation
	switch b.Rotation {
	case 0:
		return []int{0, 1, 6} // Side 0, Side 1, Top
	case 1:
		return []int{1, 2, 6} // Side 1, Side 2, Top
	case 2:
		return []int{2, 3, 6} // Side 2, Side 3, Top
	case 3:
		return []int{3, 4, 6} // Side 3, Side 4, Top
	case 4:
		return []int{4, 5, 6} // Side 4, Side 5, Top
	case 5:
		return []int{5, 0, 6} // Side 5, Side 0, Top
	default:
		return []int{0, 1, 6}
	}
}

// GetAttachmentPoints returns all valid attachment points for this block
func (b *Block) GetAttachmentPoints() []AttachmentPoint {
	faces := b.Prism.GetAttachmentFaces()
	points := make([]AttachmentPoint, 0, len(faces))

	for _, face := range faces {
		// Calculate attachment point position
		scaledNormal := types.NewVec3(face.Normal.X*0.1, face.Normal.Y*0.1, face.Normal.Z*0.1)
		pointPos := types.NewVec3(face.Center.X+scaledNormal.X, face.Center.Y+scaledNormal.Y, face.Center.Z+scaledNormal.Z)

		points = append(points, AttachmentPoint{
			Position: pointPos,
			Face:     face,
			Block:    b,
		})
	}

	return points
}

// Rotate rotates the block by the specified number of 60-degree increments
func (b *Block) Rotate(increments int) {
	b.Rotation = (b.Rotation + increments) % 6
	b.Prism.SetRotation(b.Rotation)
}

// GetBounds returns the bounding box of this block
func (b *Block) GetBounds() BlockBounds {
	// Calculate bounds based on hexagonal prism
	radius := b.Prism.Radius
	height := b.Prism.Height

	return BlockBounds{
		Min: types.NewVec3(
			b.Position.X-float32(radius),
			b.Position.Y-float32(height)/2,
			b.Position.Z-float32(radius),
		),
		Max: types.NewVec3(
			b.Position.X+float32(radius),
			b.Position.Y+float32(height)/2,
			b.Position.Z+float32(radius),
		),
	}
}

// Intersects checks if this block intersects with another block's bounds
func (b *Block) Intersects(other *Block) bool {
	bounds1 := b.GetBounds()
	bounds2 := other.GetBounds()

	return bounds1.Intersects(bounds2)
}

// AttachmentPoint represents a point where another block can be attached
type AttachmentPoint struct {
	Position types.Vec3
	Face     AttachmentFace
	Block    *Block
}

// BlockBounds represents the axis-aligned bounding box of a block
type BlockBounds struct {
	Min types.Vec3
	Max types.Vec3
}

// Intersects checks if this bounding box intersects with another
func (bb BlockBounds) Intersects(other BlockBounds) bool {
	return (bb.Min.X <= other.Max.X && bb.Max.X >= other.Min.X) &&
		(bb.Min.Y <= other.Max.Y && bb.Max.Y >= other.Min.Y) &&
		(bb.Min.Z <= other.Max.Z && bb.Max.Z >= other.Min.Z)
}

// Contains checks if a point is inside this bounding box
func (bb BlockBounds) Contains(point types.Vec3) bool {
	return point.X >= bb.Min.X && point.X <= bb.Max.X &&
		point.Y >= bb.Min.Y && point.Y <= bb.Max.Y &&
		point.Z >= bb.Min.Z && point.Z <= bb.Max.Z
}

// GetCenter returns the center point of this bounding box
func (bb BlockBounds) GetCenter() types.Vec3 {
	return types.NewVec3(
		(bb.Min.X+bb.Max.X)/2,
		(bb.Min.Y+bb.Max.Y)/2,
		(bb.Min.Z+bb.Max.Z)/2,
	)
}

// GetSize returns the size of this bounding box
func (bb BlockBounds) GetSize() types.Vec3 {
	return types.NewVec3(
		bb.Max.X-bb.Min.X,
		bb.Max.Y-bb.Min.Y,
		bb.Max.Z-bb.Min.Z,
	)
}
