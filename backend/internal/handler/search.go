package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mino/backend/internal/middleware"
	resp "github.com/mino/backend/internal/pkg/response"
	"github.com/mino/backend/internal/service"
)

type SearchHandler struct {
	searchService *service.SearchService
}

func NewSearchHandler(searchService *service.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

// Search handles GET /v1/search?q=...&limit=10
func (h *SearchHandler) Search(c *gin.Context) {
	userID := middleware.GetUserID(c)
	query := c.Query("q")
	if query == "" {
		resp.BadRequest(c, "query parameter 'q' is required")
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	result, err := h.searchService.Search(c.Request.Context(), userID, query, limit)
	if err != nil {
		resp.InternalError(c, "search failed: "+err.Error())
		return
	}
	resp.OK(c, result)
}

// Reindex handles POST /v1/search/reindex — re-syncs all user data to Typesense.
func (h *SearchHandler) Reindex(c *gin.Context) {
	userID := middleware.GetUserID(c)

	count, err := h.searchService.SyncAllFromDB(c.Request.Context(), userID)
	if err != nil {
		resp.InternalError(c, "reindex failed: "+err.Error())
		return
	}
	resp.OK(c, gin.H{"indexed": count})
}
