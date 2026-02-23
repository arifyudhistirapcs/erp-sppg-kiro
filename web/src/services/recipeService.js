import api from './api'

const recipeService = {
  // Get all recipes with optional filters
  getRecipes(params = {}) {
    return api.get('/recipes', { params })
  },

  // Get single recipe by ID
  getRecipe(id) {
    return api.get(`/recipes/${id}`)
  },

  // Create new recipe
  createRecipe(data) {
    return api.post('/recipes', data)
  },

  // Update existing recipe
  updateRecipe(id, data) {
    return api.put(`/recipes/${id}`, data)
  },

  // Delete recipe
  deleteRecipe(id) {
    return api.delete(`/recipes/${id}`)
  },

  // Get recipe nutrition info
  getRecipeNutrition(id) {
    return api.get(`/recipes/${id}/nutrition`)
  },

  // Get recipe version history
  getRecipeHistory(id) {
    return api.get(`/recipes/${id}/history`)
  },

  // Get all ingredients for recipe form
  getIngredients(params = {}) {
    return api.get('/ingredients', { params })
  }
}

export default recipeService
