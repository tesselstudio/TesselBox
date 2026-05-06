package crafting

import (
	"strconv"
	"sync"
)

// Recipe represents a crafting recipe
type Recipe struct {
	ID          string
	Name        string
	Description string
	Ingredients []*RecipeIngredient
	Result      *RecipeResult
	Shapeless   bool
	GridSize    int
	Category    string
	Tags        []string
}

// RecipeIngredient represents an ingredient in a recipe
type RecipeIngredient struct {
	ItemID   string
	Quantity int
	Optional bool
}

// RecipeResult represents the result of a recipe
type RecipeResult struct {
	ItemID   string
	Quantity int
	Damage   int
}

// RecipePattern represents a shaped recipe pattern
type RecipePattern [][]string

// RecipeRegistry manages all crafting recipes
type RecipeRegistry struct {
	mu         sync.RWMutex
	recipes    map[string]*Recipe
	byCategory map[string][]*Recipe
	shaped     map[string]*Recipe // Shaped recipes by pattern
}

// NewRecipeRegistry creates a new recipe registry
func NewRecipeRegistry() *RecipeRegistry {
	registry := &RecipeRegistry{
		recipes:    make(map[string]*Recipe),
		byCategory: make(map[string][]*Recipe),
		shaped:     make(map[string]*Recipe),
	}

	// Register default recipes
	registry.registerDefaults()

	return registry
}

// RegisterRecipe registers a new recipe
func (rr *RecipeRegistry) RegisterRecipe(recipe *Recipe) error {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	if _, exists := rr.recipes[recipe.ID]; exists {
		return ErrRecipeAlreadyExists
	}

	rr.recipes[recipe.ID] = recipe
	rr.byCategory[recipe.Category] = append(rr.byCategory[recipe.Category], recipe)

	if !recipe.Shapeless {
		// Register shaped recipe pattern
		pattern := rr.generatePattern(recipe)
		rr.shaped[pattern] = recipe
	}

	return nil
}

// GetRecipe retrieves a recipe by ID
func (rr *RecipeRegistry) GetRecipe(id string) (*Recipe, bool) {
	rr.mu.RLock()
	defer rr.mu.RUnlock()

	recipe, exists := rr.recipes[id]
	return recipe, exists
}

// GetRecipesByCategory retrieves all recipes in a category
func (rr *RecipeRegistry) GetRecipesByCategory(category string) []*Recipe {
	rr.mu.RLock()
	defer rr.mu.RUnlock()

	recipes := make([]*Recipe, len(rr.byCategory[category]))
	copy(recipes, rr.byCategory[category])

	return recipes
}

// GetAllRecipes returns all registered recipes
func (rr *RecipeRegistry) GetAllRecipes() []*Recipe {
	rr.mu.RLock()
	defer rr.mu.RUnlock()

	recipes := make([]*Recipe, 0, len(rr.recipes))
	for _, recipe := range rr.recipes {
		recipes = append(recipes, recipe)
	}

	return recipes
}

// FindRecipe finds a recipe that matches the given ingredients
func (rr *RecipeRegistry) FindRecipe(grid []*ItemStack, gridSize int) *Recipe {
	rr.mu.RLock()
	defer rr.mu.RUnlock()

	// Try shaped recipes first
	pattern := rr.generateGridPattern(grid, gridSize)
	if recipe, exists := rr.shaped[pattern]; exists {
		return recipe
	}

	// Try shapeless recipes
	for _, recipe := range rr.recipes {
		if recipe.Shapeless && rr.matchesShapelessRecipe(recipe, grid) {
			return recipe
		}
	}

	return nil
}

// matchesShapelessRecipe checks if ingredients match a shapeless recipe
func (rr *RecipeRegistry) matchesShapelessRecipe(recipe *Recipe, grid []*ItemStack) bool {
	// Count ingredients in grid
	gridIngredients := make(map[string]int)
	for _, stack := range grid {
		if stack != nil && !stack.IsEmpty() {
			gridIngredients[stack.Item.ID] += stack.Quantity
		}
	}

	// Check recipe ingredients
	for _, ingredient := range recipe.Ingredients {
		if !ingredient.Optional {
			if gridIngredients[ingredient.ItemID] < ingredient.Quantity {
				return false
			}
			gridIngredients[ingredient.ItemID] -= ingredient.Quantity
		}
	}

	// Check for extra ingredients (should be none for exact match)
	for _, count := range gridIngredients {
		if count > 0 {
			return false
		}
	}

	return true
}

// generatePattern generates a pattern string for a shaped recipe
func (rr *RecipeRegistry) generatePattern(recipe *Recipe) string {
	// Simplified pattern generation
	// In a full implementation, this would handle the actual recipe pattern
	pattern := recipe.ID
	for _, ingredient := range recipe.Ingredients {
		pattern += ":" + ingredient.ItemID + ":" + strconv.Itoa(ingredient.Quantity)
	}
	return pattern
}

// generateGridPattern generates a pattern from the crafting grid
func (rr *RecipeRegistry) generateGridPattern(grid []*ItemStack, gridSize int) string {
	pattern := ""
	for i := 0; i < gridSize*gridSize; i++ {
		if i < len(grid) && grid[i] != nil && !grid[i].IsEmpty() {
			pattern += grid[i].Item.ID
		} else {
			pattern += "_"
		}
	}
	return pattern
}

// registerDefaults registers the default crafting recipes
func (rr *RecipeRegistry) registerDefaults() {
	defaultRecipes := []*Recipe{
		// Tools
		{
			ID:          "wooden_pickaxe",
			Name:        "Wooden Pickaxe",
			Description: "A basic wooden pickaxe",
			Ingredients: []*RecipeIngredient{
				{ItemID: "wood", Quantity: 3},
				{ItemID: "wood", Quantity: 2},
			},
			Result: &RecipeResult{
				ItemID:   "wooden_pickaxe",
				Quantity: 1,
			},
			Shapeless: false,
			GridSize:  3,
			Category:  "tools",
			Tags:      []string{"tool", "pickaxe", "wooden"},
		},
		{
			ID:          "stone_pickaxe",
			Name:        "Stone Pickaxe",
			Description: "A stone pickaxe",
			Ingredients: []*RecipeIngredient{
				{ItemID: "stone", Quantity: 3},
				{ItemID: "wood", Quantity: 2},
			},
			Result: &RecipeResult{
				ItemID:   "stone_pickaxe",
				Quantity: 1,
			},
			Shapeless: false,
			GridSize:  3,
			Category:  "tools",
			Tags:      []string{"tool", "pickaxe", "stone"},
		},
		{
			ID:          "iron_pickaxe",
			Name:        "Iron Pickaxe",
			Description: "An iron pickaxe",
			Ingredients: []*RecipeIngredient{
				{ItemID: "iron_ingot", Quantity: 3},
				{ItemID: "wood", Quantity: 2},
			},
			Result: &RecipeResult{
				ItemID:   "iron_pickaxe",
				Quantity: 1,
			},
			Shapeless: false,
			GridSize:  3,
			Category:  "tools",
			Tags:      []string{"tool", "pickaxe", "iron"},
		},
		{
			ID:          "diamond_pickaxe",
			Name:        "Diamond Pickaxe",
			Description: "A diamond pickaxe",
			Ingredients: []*RecipeIngredient{
				{ItemID: "diamond", Quantity: 3},
				{ItemID: "wood", Quantity: 2},
			},
			Result: &RecipeResult{
				ItemID:   "diamond_pickaxe",
				Quantity: 1,
			},
			Shapeless: false,
			GridSize:  3,
			Category:  "tools",
			Tags:      []string{"tool", "pickaxe", "diamond"},
		},

		// Weapons
		{
			ID:          "wooden_sword",
			Name:        "Wooden Sword",
			Description: "A basic wooden sword",
			Ingredients: []*RecipeIngredient{
				{ItemID: "wood", Quantity: 2},
				{ItemID: "wood", Quantity: 1},
			},
			Result: &RecipeResult{
				ItemID:   "wooden_sword",
				Quantity: 1,
			},
			Shapeless: false,
			GridSize:  3,
			Category:  "weapons",
			Tags:      []string{"weapon", "sword", "wooden"},
		},
		{
			ID:          "iron_sword",
			Name:        "Iron Sword",
			Description: "An iron sword",
			Ingredients: []*RecipeIngredient{
				{ItemID: "iron_ingot", Quantity: 2},
				{ItemID: "wood", Quantity: 1},
			},
			Result: &RecipeResult{
				ItemID:   "iron_sword",
				Quantity: 1,
			},
			Shapeless: false,
			GridSize:  3,
			Category:  "weapons",
			Tags:      []string{"weapon", "sword", "iron"},
		},

		// Smelting (shapeless recipes for simplicity)
		{
			ID:          "iron_ingot_from_ore",
			Name:        "Iron Ingot",
			Description: "Smelt iron ore into iron ingot",
			Ingredients: []*RecipeIngredient{
				{ItemID: "iron_ore", Quantity: 1},
			},
			Result: &RecipeResult{
				ItemID:   "iron_ingot",
				Quantity: 1,
			},
			Shapeless: true,
			GridSize:  1,
			Category:  "smelting",
			Tags:      []string{"smelting", "metal"},
		},

		// Building blocks
		{
			ID:          "wood_planks",
			Name:        "Wood Planks",
			Description: "Craft wood into planks",
			Ingredients: []*RecipeIngredient{
				{ItemID: "wood", Quantity: 1},
			},
			Result: &RecipeResult{
				ItemID:   "wood",
				Quantity: 4,
			},
			Shapeless: true,
			GridSize:  1,
			Category:  "building",
			Tags:      []string{"building", "wood"},
		},

		// Food
		{
			ID:          "bread",
			Name:        "Bread",
			Description: "Craft bread from wheat",
			Ingredients: []*RecipeIngredient{
				{ItemID: "wheat", Quantity: 3},
			},
			Result: &RecipeResult{
				ItemID:   "bread",
				Quantity: 1,
			},
			Shapeless: true,
			GridSize:  3,
			Category:  "food",
			Tags:      []string{"food", "prepared"},
		},

		// Armor
		{
			ID:          "leather_helmet",
			Name:        "Leather Helmet",
			Description: "Craft a leather helmet",
			Ingredients: []*RecipeIngredient{
				{ItemID: "leather", Quantity: 5},
			},
			Result: &RecipeResult{
				ItemID:   "leather_helmet",
				Quantity: 1,
			},
			Shapeless: false,
			GridSize:  3,
			Category:  "armor",
			Tags:      []string{"armor", "helmet", "leather"},
		},
		{
			ID:          "iron_chestplate",
			Name:        "Iron Chestplate",
			Description: "Craft an iron chestplate",
			Ingredients: []*RecipeIngredient{
				{ItemID: "iron_ingot", Quantity: 8},
			},
			Result: &RecipeResult{
				ItemID:   "iron_chestplate",
				Quantity: 1,
			},
			Shapeless: false,
			GridSize:  3,
			Category:  "armor",
			Tags:      []string{"armor", "chestplate", "iron"},
		},
	}

	for _, recipe := range defaultRecipes {
		rr.recipes[recipe.ID] = recipe
		rr.byCategory[recipe.Category] = append(rr.byCategory[recipe.Category], recipe)

		if !recipe.Shapeless {
			pattern := rr.generatePattern(recipe)
			rr.shaped[pattern] = recipe
		}
	}
}

// CraftingManager manages the crafting system
type CraftingManager struct {
	recipeRegistry *RecipeRegistry
	itemRegistry   *ItemRegistry
}

// NewCraftingManager creates a new crafting manager
func NewCraftingManager() *CraftingManager {
	return &CraftingManager{
		recipeRegistry: NewRecipeRegistry(),
		itemRegistry:   GetGlobalItemRegistry(),
	}
}

// GetRecipeRegistry returns the recipe registry
func (cm *CraftingManager) GetRecipeRegistry() *RecipeRegistry {
	return cm.recipeRegistry
}

// GetItemRegistry returns the item registry
func (cm *CraftingManager) GetItemRegistry() *ItemRegistry {
	return cm.itemRegistry
}

// Craft attempts to craft an item using the given ingredients
// Returns the result item and the modified grid (with consumed ingredients removed)
func (cm *CraftingManager) Craft(grid []*ItemStack, gridSize int) (*ItemStack, []*ItemStack, error) {
	recipe := cm.recipeRegistry.FindRecipe(grid, gridSize)
	if recipe == nil {
		return nil, grid, ErrNoRecipeFound
	}

	// Check if we have all required ingredients
	if !cm.hasIngredients(recipe, grid) {
		return nil, grid, ErrInsufficientIngredients
	}

	// Create result item
	item, exists := cm.itemRegistry.GetItem(recipe.Result.ItemID)
	if !exists {
		return nil, grid, ErrItemNotFound
	}

	result := NewItemStack(item, recipe.Result.Quantity)
	result.Damage = recipe.Result.Damage

	// Consume ingredients
	grid = cm.ConsumeIngredients(recipe, grid)

	return result, grid, nil
}

// hasIngredients checks if the grid contains all required ingredients
func (cm *CraftingManager) hasIngredients(recipe *Recipe, grid []*ItemStack) bool {
	// Count ingredients in grid
	gridIngredients := make(map[string]int)
	for _, stack := range grid {
		if stack != nil && !stack.IsEmpty() {
			gridIngredients[stack.Item.ID] += stack.Quantity
		}
	}

	// Check recipe ingredients
	for _, ingredient := range recipe.Ingredients {
		if !ingredient.Optional {
			if gridIngredients[ingredient.ItemID] < ingredient.Quantity {
				return false
			}
		}
	}

	return true
}

// ConsumeIngredients consumes the ingredients for a recipe from the crafting grid
func (cm *CraftingManager) ConsumeIngredients(recipe *Recipe, grid []*ItemStack) []*ItemStack {
	// Track what we need to consume
	remainingIngredients := make(map[string]int)
	for _, ingredient := range recipe.Ingredients {
		if !ingredient.Optional {
			remainingIngredients[ingredient.ItemID] += ingredient.Quantity
		}
	}

	// Consume from each slot
	for i, stack := range grid {
		if stack == nil || stack.IsEmpty() {
			continue
		}

		needed, exists := remainingIngredients[stack.Item.ID]
		if !exists || needed <= 0 {
			continue
		}

		// Consume from this stack
		if stack.Quantity >= needed {
			stack.Quantity -= needed
			remainingIngredients[stack.Item.ID] = 0
			if stack.Quantity <= 0 {
				grid[i] = nil
			}
		} else {
			remainingIngredients[stack.Item.ID] -= stack.Quantity
			stack.Quantity = 0
			grid[i] = nil
		}
	}

	return grid
}

// Global crafting manager instance
var GlobalCraftingManager = NewCraftingManager()

// GetGlobalCraftingManager returns the global crafting manager
func GetGlobalCraftingManager() *CraftingManager {
	return GlobalCraftingManager
}

// Errors
var (
	ErrRecipeAlreadyExists     = &RecipeError{Message: "recipe already exists"}
	ErrNoRecipeFound           = &RecipeError{Message: "no recipe found"}
	ErrInsufficientIngredients = &RecipeError{Message: "insufficient ingredients"}
	ErrInvalidRecipe           = &RecipeError{Message: "invalid recipe"}
)

// RecipeError represents a recipe-related error
type RecipeError struct {
	Message string
}

func (e *RecipeError) Error() string {
	return e.Message
}
