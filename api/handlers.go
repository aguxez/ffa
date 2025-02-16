package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aguxez/ffa/agent"
)

type MealPlanner interface {
	GenerateMealPlan(ctx context.Context) (agent.MealPlanResponse, error)
}

// HTTP handlers
func HandleMealPlanRequest(planner MealPlanner, w http.ResponseWriter, r *http.Request) {
	log.Println("Generating meal plan...")

	planResponse, err := planner.GenerateMealPlan(r.Context())

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(planResponse)
}
