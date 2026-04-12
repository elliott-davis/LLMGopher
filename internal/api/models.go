package api

import (
	"net/http"
	"sort"

	"github.com/ed007183/llmgopher/internal/storage"
)

const modelsEndpointOwner = "llmgopher"

type modelsListResponse struct {
	Object string            `json:"object"`
	Data   []modelsListEntry `json:"data"`
}

type modelsListEntry struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Owner   string `json:"owned_by"`
}

// HandleListModels returns the OpenAI-compatible model list response.
func HandleListModels(cache *storage.StateCache) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		resp := modelsListResponse{
			Object: "list",
			Data:   make([]modelsListEntry, 0),
		}

		if cache != nil {
			if state := cache.Load(); state != nil {
				resp.Data = make([]modelsListEntry, 0, len(state.Models))
				for _, model := range state.Models {
					resp.Data = append(resp.Data, modelsListEntry{
						ID:      model.Alias,
						Object:  "model",
						Created: model.CreatedAt.Unix(),
						Owner:   modelsEndpointOwner,
					})
				}
				sort.Slice(resp.Data, func(i, j int) bool {
					return resp.Data[i].ID < resp.Data[j].ID
				})
			}
		}

		WriteJSON(w, http.StatusOK, resp)
	}
}
