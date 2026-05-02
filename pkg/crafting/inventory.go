package crafting

import (
	"sync"
)

// Inventory represents a player's inventory
type Inventory struct {
	mu       sync.RWMutex
	slots    []*ItemStack
	capacity int
	hotbar   int // Number of hotbar slots
}

// NewInventory creates a new inventory with the specified capacity
func NewInventory(capacity, hotbar int) *Inventory {
	return &Inventory{
		slots:    make([]*ItemStack, capacity),
		capacity: capacity,
		hotbar:   hotbar,
	}
}

// AddItem adds an item to the inventory
func (inv *Inventory) AddItem(stack *ItemStack) error {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	// Try to stack with existing items first
	for _, existing := range inv.slots {
		if existing != nil && existing.CanStack(stack) {
			remaining := existing.Item.MaxStack - existing.Quantity
			if remaining >= stack.Quantity {
				// Can add entire stack
				existing.Quantity += stack.Quantity
				return nil
			} else {
				// Add what we can and continue
				existing.Quantity = existing.Item.MaxStack
				stack.Quantity -= remaining
			}
		}
	}

	// Find empty slot for remaining items
	for i, slot := range inv.slots {
		if slot == nil || slot.IsEmpty() {
			inv.slots[i] = stack
			return nil
		}
	}

	return ErrInventoryFull
}

// RemoveItem removes items from the inventory
func (inv *Inventory) RemoveItem(itemID string, quantity int) (*ItemStack, error) {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	var removed *ItemStack
	remaining := quantity

	// Find and remove items from slots
	for i, slot := range inv.slots {
		if slot != nil && slot.Item.ID == itemID && !slot.IsEmpty() {
			if slot.Quantity >= remaining {
				// Remove from this slot and we're done
				removedStack := slot.Split(remaining)
				if removed == nil {
					removedStack = NewItemStack(slot.Item, remaining)
					removedStack.Damage = slot.Damage
				}

				if removed != nil {
					if removed == nil {
						removed = removedStack
					} else {
						removed.Stack(removedStack)
					}
				} else {
					removed = removedStack
				}

				if slot.IsEmpty() {
					inv.slots[i] = nil
				}

				return removed, nil
			} else {
				// Remove all from this slot and continue
				if removed == nil {
					removed = NewItemStack(slot.Item, slot.Quantity)
					removed.Damage = slot.Damage
				} else {
					toAdd := NewItemStack(slot.Item, slot.Quantity)
					toAdd.Damage = slot.Damage
					removed.Stack(toAdd)
				}

				remaining -= slot.Quantity
				inv.slots[i] = nil
			}
		}
	}

	if remaining > 0 {
		return nil, ErrInsufficientItems
	}

	return removed, nil
}

// GetSlot returns the item stack in the specified slot
func (inv *Inventory) GetSlot(index int) *ItemStack {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	if index < 0 || index >= len(inv.slots) {
		return nil
	}

	return inv.slots[index]
}

// SetSlot sets the item stack in the specified slot
func (inv *Inventory) SetSlot(index int, stack *ItemStack) error {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	if index < 0 || index >= len(inv.slots) {
		return ErrInvalidSlot
	}

	inv.slots[index] = stack
	return nil
}

// SwapSlots swaps two slots in the inventory
func (inv *Inventory) SwapSlots(slot1, slot2 int) error {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	if slot1 < 0 || slot1 >= len(inv.slots) || slot2 < 0 || slot2 >= len(inv.slots) {
		return ErrInvalidSlot
	}

	inv.slots[slot1], inv.slots[slot2] = inv.slots[slot2], inv.slots[slot1]
	return nil
}

// MoveItem moves an item from one slot to another
func (inv *Inventory) MoveItem(fromSlot, toSlot int, quantity int) error {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	if fromSlot < 0 || fromSlot >= len(inv.slots) || toSlot < 0 || toSlot >= len(inv.slots) {
		return ErrInvalidSlot
	}

	from := inv.slots[fromSlot]
	to := inv.slots[toSlot]

	if from == nil || from.IsEmpty() {
		return ErrEmptySlot
	}

	if quantity <= 0 || quantity > from.Quantity {
		return ErrInvalidQuantity
	}

	// If destination is empty or same item, try to stack
	if to == nil || to.IsEmpty() {
		if from.Item.ID == to.Item.ID && from.CanStack(to) {
			// Stack items
			space := to.Item.MaxStack - to.Quantity
			if space >= quantity {
				// Move all to destination
				to.Quantity += quantity
				from.Quantity -= quantity
				if from.IsEmpty() {
					inv.slots[fromSlot] = nil
				}
			} else {
				// Move what fits
				to.Quantity = to.Item.MaxStack
				from.Quantity -= space
				if from.IsEmpty() {
					inv.slots[fromSlot] = nil
				}
				// Create new stack for remaining
				remaining := NewItemStack(from.Item, from.Quantity)
				remaining.Damage = from.Damage
				inv.slots[toSlot] = remaining
				inv.slots[fromSlot] = nil
			}
		} else {
			// Move to empty slot
			if quantity == from.Quantity {
				// Move entire stack
				inv.slots[toSlot] = from
				inv.slots[fromSlot] = nil
			} else {
				// Split stack
				split := from.Split(quantity)
				inv.slots[toSlot] = split
				if from.IsEmpty() {
					inv.slots[fromSlot] = nil
				}
			}
		}
	} else {
		// Destination has different item, swap
		inv.slots[fromSlot], inv.slots[toSlot] = inv.slots[toSlot], inv.slots[fromSlot]
	}

	return nil
}

// HasItem checks if the inventory contains at least the specified quantity of an item
func (inv *Inventory) HasItem(itemID string, quantity int) bool {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	total := 0
	for _, slot := range inv.slots {
		if slot != nil && slot.Item.ID == itemID {
			total += slot.Quantity
			if total >= quantity {
				return true
			}
		}
	}

	return false
}

// CountItem returns the total quantity of an item in the inventory
func (inv *Inventory) CountItem(itemID string) int {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	total := 0
	for _, slot := range inv.slots {
		if slot != nil && slot.Item.ID == itemID {
			total += slot.Quantity
		}
	}

	return total
}

// GetItems returns all non-empty item stacks in the inventory
func (inv *Inventory) GetItems() []*ItemStack {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	items := make([]*ItemStack, 0)
	for _, slot := range inv.slots {
		if slot != nil && !slot.IsEmpty() {
			items = append(items, slot)
		}
	}

	return items
}

// GetHotbarItems returns the items in the hotbar
func (inv *Inventory) GetHotbarItems() []*ItemStack {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	items := make([]*ItemStack, inv.hotbar)
	for i := 0; i < inv.hotbar && i < len(inv.slots); i++ {
		items[i] = inv.slots[i]
	}

	return items
}

// IsEmpty returns true if the inventory is empty
func (inv *Inventory) IsEmpty() bool {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	for _, slot := range inv.slots {
		if slot != nil && !slot.IsEmpty() {
			return false
		}
	}

	return true
}

// IsFull returns true if the inventory is full
func (inv *Inventory) IsFull() bool {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	for _, slot := range inv.slots {
		if slot == nil || slot.IsEmpty() || !slot.IsFull() {
			return false
		}
	}

	return true
}

// GetEmptySlots returns the number of empty slots
func (inv *Inventory) GetEmptySlots() int {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	empty := 0
	for _, slot := range inv.slots {
		if slot == nil || slot.IsEmpty() {
			empty++
		}
	}

	return empty
}

// Clear clears the entire inventory
func (inv *Inventory) Clear() {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	for i := range inv.slots {
		inv.slots[i] = nil
	}
}

// Compact compacts the inventory by moving items to empty slots
func (inv *Inventory) Compact() {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	// Move all non-empty items to the beginning
	items := make([]*ItemStack, 0)
	for _, slot := range inv.slots {
		if slot != nil && !slot.IsEmpty() {
			items = append(items, slot)
		}
	}

	// Clear all slots
	for i := range inv.slots {
		inv.slots[i] = nil
	}

	// Place items back in order
	for i, item := range items {
		if i < len(inv.slots) {
			inv.slots[i] = item
		}
	}
}

// CraftingInventory represents a crafting inventory interface
type CraftingInventory struct {
	*Inventory
	craftingGrid []*ItemStack
	gridSize     int
}

// NewCraftingInventory creates a new crafting inventory
func NewCraftingInventory(inventoryCapacity, hotbar, gridSize int) *CraftingInventory {
	craftingGrid := make([]*ItemStack, gridSize*gridSize)

	return &CraftingInventory{
		Inventory:    NewInventory(inventoryCapacity, hotbar),
		craftingGrid: craftingGrid,
		gridSize:     gridSize,
	}
}

// GetCraftingSlot returns the item in the specified crafting slot
func (ci *CraftingInventory) GetCraftingSlot(index int) *ItemStack {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	if index < 0 || index >= len(ci.craftingGrid) {
		return nil
	}

	return ci.craftingGrid[index]
}

// SetCraftingSlot sets the item in the specified crafting slot
func (ci *CraftingInventory) SetCraftingSlot(index int, stack *ItemStack) error {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	if index < 0 || index >= len(ci.craftingGrid) {
		return ErrInvalidSlot
	}

	ci.craftingGrid[index] = stack
	return nil
}

// GetCraftingGrid returns the entire crafting grid
func (ci *CraftingInventory) GetCraftingGrid() []*ItemStack {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	grid := make([]*ItemStack, len(ci.craftingGrid))
	copy(grid, ci.craftingGrid)

	return grid
}

// ClearCraftingGrid clears the crafting grid
func (ci *CraftingInventory) ClearCraftingGrid() {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	for i := range ci.craftingGrid {
		ci.craftingGrid[i] = nil
	}
}

// Errors
var (
	ErrInventoryFull     = &InventoryError{Message: "inventory is full"}
	ErrInsufficientItems = &InventoryError{Message: "insufficient items"}
	ErrInvalidSlot       = &InventoryError{Message: "invalid slot"}
	ErrEmptySlot         = &InventoryError{Message: "slot is empty"}
	ErrInvalidQuantity   = &InventoryError{Message: "invalid quantity"}
)

// InventoryError represents an inventory-related error
type InventoryError struct {
	Message string
}

func (e *InventoryError) Error() string {
	return e.Message
}
