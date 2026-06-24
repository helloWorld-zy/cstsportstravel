import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useAuth } from '@acme/auth';
import { useToast } from '@/hooks/use-toast';
import {
  Plus,
  Calendar,
  Utensils,
  Clock,
  Users,
  ChefHat,
  ShoppingCart,
  Trash2,
  Edit,
  RefreshCw,
} from 'lucide-react';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:4002';

interface MealPlan {
  id: string;
  weekStart: string;
  meals: Array<{
    day: string;
    meal: string;
    recipe?: {
      id: string;
      name: string;
      category: string;
      prepTime: number;
      cookTime: number;
      servings: number;
    };
  }>;
}

interface Recipe {
  id: string;
  name: string;
  description: string;
  ingredients: string[];
  instructions: string[];
  prepTime: number;
  cookTime: number;
  servings: number;
  category: string;
  tags: string[];
  userId: string;
}

export default function Home() {
  const { user, logout } = useAuth();
  const queryClient = useQueryClient();
  const { toast } = useToast();
  const [activeTab, setActiveTab] = useState('meal-planner');
  const [isAddRecipeOpen, setIsAddRecipeOpen] = useState(false);
  const [newRecipe, setNewRecipe] = useState({
    name: '',
    description: '',
    ingredients: '',
    instructions: '',
    prepTime: 30,
    cookTime: 30,
    servings: 4,
    category: 'DINNER',
    tags: '',
  });

  const daysOfWeek = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'];
  const mealTypes = ['Breakfast', 'Lunch', 'Dinner', 'Snack'];

  const { data: mealPlans, isLoading: mealPlansLoading } = useQuery<MealPlan[]>({
    queryKey: ['meal-plans'],
    queryFn: async () => {
      const res = await fetch(`${API_URL}/api/meal-plans`, {
        headers: { 'x-user-id': user?.id || '' },
      });
      if (!res.ok) throw new Error('Failed to fetch meal plans');
      return res.json();
    },
  });

  const { data: recipes, isLoading: recipesLoading } = useQuery<Recipe[]>({
    queryKey: ['recipes'],
    queryFn: async () => {
      const res = await fetch(`${API_URL}/api/recipes`, {
        headers: { 'x-user-id': user?.id || '' },
      });
      if (!res.ok) throw new Error('Failed to fetch recipes');
      return res.json();
    },
  });

  const createRecipeMutation = useMutation({
    mutationFn: async (recipe: Omit<Recipe, 'id' | 'userId'>) => {
      const res = await fetch(`${API_URL}/api/recipes`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'x-user-id': user?.id || '',
        },
        body: JSON.stringify(recipe),
      });
      if (!res.ok) throw new Error('Failed to create recipe');
      return res.json();
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['recipes'] });
      setIsAddRecipeOpen(false);
      resetRecipeForm();
      toast({ title: 'Recipe created successfully!' });
    },
    onError: () => {
      toast({ title: 'Failed to create recipe', variant: 'destructive' });
    },
  });

  const generateMealPlanMutation = useMutation({
    mutationFn: async () => {
      const res = await fetch(`${API_URL}/api/meal-plans/generate`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'x-user-id': user?.id || '',
        },
        body: JSON.stringify({ preferences: { dietaryRestrictions: [] } }),
      });
      if (!res.ok) throw new Error('Failed to generate meal plan');
      return res.json();
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['meal-plans'] });
      toast({ title: 'Meal plan generated!' });
    },
  });

  const deleteRecipeMutation = useMutation({
    mutationFn: async (id: string) => {
      const res = await fetch(`${API_URL}/api/recipes/${id}`, {
        method: 'DELETE',
        headers: { 'x-user-id': user?.id || '' },
      });
      if (!res.ok) throw new Error('Failed to delete recipe');
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['recipes'] });
      toast({ title: 'Recipe deleted' });
    },
  });

  const resetRecipeForm = () => {
    setNewRecipe({
      name: '',
      description: '',
      ingredients: '',
      instructions: '',
      prepTime: 30,
      cookTime: 30,
      servings: 4,
      category: 'DINNER',
      tags: '',
    });
  };

  const handleCreateRecipe = () => {
    createRecipeMutation.mutate({
      ...newRecipe,
      ingredients: newRecipe.ingredients.split('\n').filter(Boolean),
      instructions: newRecipe.instructions.split('\n').filter(Boolean),
      tags: newRecipe.tags.split(',').map((t) => t.trim()).filter(Boolean),
    });
  };

  const getCategoryColor = (category: string) => {
    const colors: Record<string, string> = {
      BREAKFAST: 'bg-yellow-100 text-yellow-800',
      LUNCH: 'bg-green-100 text-green-800',
      DINNER: 'bg-blue-100 text-blue-800',
      SNACK: 'bg-purple-100 text-purple-800',
      DESSERT: 'bg-pink-100 text-pink-800',
    };
    return colors[category] || 'bg-gray-100 text-gray-800';
  };

  const currentMealPlan = mealPlans?.[0];

  const getMealForDayAndType = (day: string, mealType: string) => {
    return currentMealPlan?.meals.find(
      (m) => m.day === day && m.meal === mealType
    );
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-orange-50 via-white to-red-50">
      {/* Header */}
      <header className="bg-white/80 backdrop-blur-sm border-b border-orange-100 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-gradient-to-br from-orange-500 to-red-500 rounded-xl flex items-center justify-center shadow-lg">
                <Utensils className="w-5 h-5 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-gray-900">Meal Planner</h1>
                <p className="text-xs text-gray-500">Healthy eating made easy</p>
              </div>
            </div>
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <div className="w-8 h-8 bg-orange-100 rounded-full flex items-center justify-center">
                  <Users className="w-4 h-4 text-orange-600" />
                </div>
                <span className="text-sm font-medium text-gray-700">
                  {user?.name || 'User'}
                </span>
              </div>
              <Button variant="outline" size="sm" onClick={logout}>
                Logout
              </Button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
          <TabsList className="bg-white border border-orange-200 shadow-sm">
            <TabsTrigger
              value="meal-planner"
              className="data-[state=active]:bg-orange-500 data-[state=active]:text-white"
            >
              <Calendar className="w-4 h-4 mr-2" />
              Meal Planner
            </TabsTrigger>
            <TabsTrigger
              value="recipes"
              className="data-[state=active]:bg-orange-500 data-[state=active]:text-white"
            >
              <ChefHat className="w-4 h-4 mr-2" />
              My Recipes
            </TabsTrigger>
          </TabsList>

          {/* Meal Planner Tab */}
          <TabsContent value="meal-planner" className="space-y-6">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-2xl font-bold text-gray-900">
                  Weekly Meal Plan
                </h2>
                <p className="text-gray-600">
                  {currentMealPlan
                    ? `Week of ${new Date(currentMealPlan.weekStart).toLocaleDateString()}`
                    : 'No meal plan yet'}
                </p>
              </div>
              <Button
                onClick={() => generateMealPlanMutation.mutate()}
                disabled={generateMealPlanMutation.isPending}
                className="bg-gradient-to-r from-orange-500 to-red-500 hover:from-orange-600 hover:to-red-600 shadow-lg"
              >
                <RefreshCw className={`w-4 h-4 mr-2 ${generateMealPlanMutation.isPending ? 'animate-spin' : ''}`} />
                Generate New Plan
              </Button>
            </div>

            {mealPlansLoading ? (
              <div className="grid grid-cols-7 gap-4">
                {daysOfWeek.map((day) => (
                  <Card key={day} className="border-orange-100">
                    <CardHeader className="pb-2">
                      <Skeleton className="h-4 w-20" />
                    </CardHeader>
                    <CardContent className="space-y-2">
                      {[1, 2, 3].map((i) => (
                        <Skeleton key={i} className="h-16 w-full" />
                      ))}
                    </CardContent>
                  </Card>
                ))}
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-7 gap-4">
                {daysOfWeek.map((day) => (
                  <Card key={day} className="border-orange-100 hover:shadow-lg transition-shadow">
                    <CardHeader className="pb-2 bg-gradient-to-r from-orange-50 to-red-50 rounded-t-lg">
                      <CardTitle className="text-sm font-semibold text-gray-700">
                        {day}
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-2 p-3">
                      {mealTypes.map((mealType) => {
                        const meal = getMealForDayAndType(day, mealType);
                        return (
                          <div
                            key={mealType}
                            className={`p-2 rounded-lg border ${
                              meal
                                ? 'border-orange-200 bg-white hover:bg-orange-50 cursor-pointer'
                                : 'border-dashed border-gray-300 bg-gray-50'
                            } transition-colors`}
                          >
                            <p className="text-xs font-medium text-gray-500 mb-1">
                              {mealType}
                            </p>
                            {meal?.recipe ? (
                              <div>
                                <p className="text-sm font-semibold text-gray-800 truncate">
                                  {meal.recipe.name}
                                </p>
                                <div className="flex items-center gap-1 mt-1">
                                  <Clock className="w-3 h-3 text-gray-400" />
                                  <span className="text-xs text-gray-500">
                                    {meal.recipe.prepTime + meal.recipe.cookTime}min
                                  </span>
                                </div>
                              </div>
                            ) : (
                              <p className="text-xs text-gray-400 italic">
                                + Add meal
                              </p>
                            )}
                          </div>
                        );
                      })}
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}

            {/* Shopping List Preview */}
            {currentMealPlan && (
              <Card className="border-orange-100">
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <CardTitle className="flex items-center gap-2">
                      <ShoppingCart className="w-5 h-5 text-orange-500" />
                      Shopping List
                    </CardTitle>
                    <Button variant="outline" size="sm">
                      View Full List
                    </Button>
                  </div>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-gray-600">
                    Generate a shopping list from your meal plan with one click.
                    Ingredients will be consolidated and organized by store section.
                  </p>
                </CardContent>
              </Card>
            )}
          </TabsContent>

          {/* Recipes Tab */}
          <TabsContent value="recipes" className="space-y-6">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-2xl font-bold text-gray-900">My Recipes</h2>
                <p className="text-gray-600">
                  {recipes?.length || 0} recipes in your collection
                </p>
              </div>
              <Dialog open={isAddRecipeOpen} onOpenChange={setIsAddRecipeOpen}>
                <DialogTrigger asChild>
                  <Button className="bg-gradient-to-r from-orange-500 to-red-500 hover:from-orange-600 hover:to-red-600 shadow-lg">
                    <Plus className="w-4 h-4 mr-2" />
                    Add Recipe
                  </Button>
                </DialogTrigger>
                <DialogContent className="max-w-md max-h-[80vh] overflow-y-auto">
                  <DialogHeader>
                    <DialogTitle>Add New Recipe</DialogTitle>
                  </DialogHeader>
                  <div className="space-y-4">
                    <div>
                      <Label htmlFor="name">Recipe Name</Label>
                      <Input
                        id="name"
                        value={newRecipe.name}
                        onChange={(e) => setNewRecipe({ ...newRecipe, name: e.target.value })}
                        placeholder="e.g., Spaghetti Bolognese"
                      />
                    </div>
                    <div>
                      <Label htmlFor="description">Description</Label>
                      <Textarea
                        id="description"
                        value={newRecipe.description}
                        onChange={(e) => setNewRecipe({ ...newRecipe, description: e.target.value })}
                        placeholder="A brief description..."
                      />
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <Label htmlFor="prepTime">Prep Time (min)</Label>
                        <Input
                          id="prepTime"
                          type="number"
                          value={newRecipe.prepTime}
                          onChange={(e) => setNewRecipe({ ...newRecipe, prepTime: parseInt(e.target.value) || 0 })}
                        />
                      </div>
                      <div>
                        <Label htmlFor="cookTime">Cook Time (min)</Label>
                        <Input
                          id="cookTime"
                          type="number"
                          value={newRecipe.cookTime}
                          onChange={(e) => setNewRecipe({ ...newRecipe, cookTime: parseInt(e.target.value) || 0 })}
                        />
                      </div>
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <Label htmlFor="servings">Servings</Label>
                        <Input
                          id="servings"
                          type="number"
                          value={newRecipe.servings}
                          onChange={(e) => setNewRecipe({ ...newRecipe, servings: parseInt(e.target.value) || 0 })}
                        />
                      </div>
                      <div>
                        <Label htmlFor="category">Category</Label>
                        <Select
                          value={newRecipe.category}
                          onValueChange={(value) => setNewRecipe({ ...newRecipe, category: value })}
                        >
                          <SelectTrigger>
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="BREAKFAST">Breakfast</SelectItem>
                            <SelectItem value="LUNCH">Lunch</SelectItem>
                            <SelectItem value="DINNER">Dinner</SelectItem>
                            <SelectItem value="SNACK">Snack</SelectItem>
                            <SelectItem value="DESSERT">Dessert</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                    </div>
                    <div>
                      <Label htmlFor="ingredients">Ingredients (one per line)</Label>
                      <Textarea
                        id="ingredients"
                        value={newRecipe.ingredients}
                        onChange={(e) => setNewRecipe({ ...newRecipe, ingredients: e.target.value })}
                        placeholder="2 cups flour&#10;1 cup sugar&#10;3 eggs"
                        rows={4}
                      />
                    </div>
                    <div>
                      <Label htmlFor="instructions">Instructions (one per line)</Label>
                      <Textarea
                        id="instructions"
                        value={newRecipe.instructions}
                        onChange={(e) => setNewRecipe({ ...newRecipe, instructions: e.target.value })}
                        placeholder="Preheat oven to 350°F&#10;Mix dry ingredients&#10;Add wet ingredients"
                        rows={4}
                      />
                    </div>
                    <div>
                      <Label htmlFor="tags">Tags (comma separated)</Label>
                      <Input
                        id="tags"
                        value={newRecipe.tags}
                        onChange={(e) => setNewRecipe({ ...newRecipe, tags: e.target.value })}
                        placeholder="vegetarian, quick, healthy"
                      />
                    </div>
                    <Button
                      className="w-full bg-gradient-to-r from-orange-500 to-red-500"
                      onClick={handleCreateRecipe}
                      disabled={createRecipeMutation.isPending || !newRecipe.name}
                    >
                      {createRecipeMutation.isPending ? 'Creating...' : 'Create Recipe'}
                    </Button>
                  </div>
                </DialogContent>
              </Dialog>
            </div>

            {recipesLoading ? (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {[1, 2, 3, 4, 5, 6].map((i) => (
                  <Card key={i} className="border-orange-100">
                    <CardContent className="p-4 space-y-3">
                      <Skeleton className="h-4 w-3/4" />
                      <Skeleton className="h-3 w-full" />
                      <Skeleton className="h-3 w-1/2" />
                    </CardContent>
                  </Card>
                ))}
              </div>
            ) : recipes && recipes.length > 0 ? (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {recipes.map((recipe) => (
                  <Card
                    key={recipe.id}
                    className="border-orange-100 hover:shadow-lg transition-all hover:-translate-y-1"
                  >
                    <CardContent className="p-4">
                      <div className="flex items-start justify-between mb-2">
                        <div>
                          <h3 className="font-semibold text-gray-900">{recipe.name}</h3>
                          <p className="text-sm text-gray-600 line-clamp-2 mt-1">
                            {recipe.description}
                          </p>
                        </div>
                        <Badge className={getCategoryColor(recipe.category)}>
                          {recipe.category}
                        </Badge>
                      </div>
                      <div className="flex items-center gap-4 mt-3 text-sm text-gray-500">
                        <div className="flex items-center gap-1">
                          <Clock className="w-3 h-3" />
                          {recipe.prepTime + recipe.cookTime}min
                        </div>
                        <div className="flex items-center gap-1">
                          <Users className="w-3 h-3" />
                          {recipe.servings} servings
                        </div>
                      </div>
                      {recipe.tags && recipe.tags.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-3">
                          {recipe.tags.slice(0, 3).map((tag) => (
                            <Badge
                              key={tag}
                              variant="secondary"
                              className="text-xs bg-orange-50 text-orange-700"
                            >
                              {tag}
                            </Badge>
                          ))}
                        </div>
                      )}
                      <div className="flex items-center gap-2 mt-4">
                        <Button
                          variant="outline"
                          size="sm"
                          className="flex-1 border-orange-200 hover:bg-orange-50"
                        >
                          <Edit className="w-3 h-3 mr-1" />
                          Edit
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          className="border-red-200 hover:bg-red-50 text-red-600"
                          onClick={() => deleteRecipeMutation.mutate(recipe.id)}
                        >
                          <Trash2 className="w-3 h-3" />
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            ) : (
              <Card className="border-orange-100 border-dashed">
                <CardContent className="flex flex-col items-center justify-center py-12">
                  <ChefHat className="w-12 h-12 text-orange-300 mb-4" />
                  <h3 className="text-lg font-semibold text-gray-700 mb-2">
                    No recipes yet
                  </h3>
                  <p className="text-gray-500 text-center mb-4">
                    Start building your recipe collection to create personalized meal plans.
                  </p>
                  <Button
                    onClick={() => setIsAddRecipeOpen(true)}
                    className="bg-gradient-to-r from-orange-500 to-red-500"
                  >
                    <Plus className="w-4 h-4 mr-2" />
                    Add Your First Recipe
                  </Button>
                </CardContent>
              </Card>
            )}
          </TabsContent>
        </Tabs>
      </main>
    </div>
  );
}
