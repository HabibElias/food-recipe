package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"go/food-recipe/clients/hasura"

	"github.com/google/uuid"
)

type likeEventPayload struct {
	Event struct {
		Op   string `json:"op"`
		Data struct {
			New struct {
				UserID    uuid.UUID `json:"user_id"`
				RecipeId  int       `json:"recipe_id"`
				CreatedAt string    `json:"created_at"`
				ID        uuid.UUID `json:"id"`
			} `json:"new"`
		} `json:"data"`
	} `json:"event"`
}

func LikeNotificationHandler(w http.ResponseWriter, r *http.Request) {
	var payload likeEventPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	// only handle inserts
	if payload.Event.Op != "INSERT" && payload.Event.Op != "insert" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true,"skipped":true}`))
		return
	}

	client := hasura.Get()
	if client == nil {
		log.Println("hasura client not initialized")
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	// like ID is already decoded as uuid.UUID by the payload struct
	likeID := payload.Event.Data.New.ID
	// Query to get recipe owner and liker username
	vars := map[string]interface{}{"id": hasura.UUID(likeID)}
	if err := client.Query(context.Background(), &hasura.GetRecipeCreatorFromLikeIdQuery, vars); err != nil {
		log.Printf("failed to fetch recipe creator: %v", err)
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	like := hasura.GetRecipeCreatorFromLikeIdQuery.Like
	ownerIDStr := like.Recipe.UserID
	likerUsername := like.User.Username

	ownerUUID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		log.Printf("invalid owner id from query: %v", err)
		http.Error(w, "invalid owner id", http.StatusInternalServerError)
		return
	}

	variables := map[string]interface{}{
		"user_id":      hasura.UUID(ownerUUID),
		"message":      "user @" + likerUsername + " liked your recipe",
		"initiator_id": hasura.UUID(payload.Event.Data.New.UserID),
	}

	if err := client.Mutate(context.Background(), &hasura.CreateNotificationMutation, variables); err != nil {
		log.Printf("failed to create notification: %v", err)
		http.Error(w, "failed to create notification", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}
