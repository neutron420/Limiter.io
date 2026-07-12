package handlers

import (
	"net/http"

	"limiter.io/internal/dto"
	"limiter.io/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type APIKeyHandler struct {
	apiKeyService services.APIKeyService
}

func NewAPIKeyHandler(apiKeyService services.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{apiKeyService: apiKeyService}
}

func (h *APIKeyHandler) Create(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	var req dto.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	apiKey, plainKey, err := h.apiKeyService.CreateAPIKey(c.Request.Context(), userID, projectID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.APIKeyResponse{
		ID:        apiKey.ID,
		ProjectID: apiKey.ProjectID,
		Name:      apiKey.Name,
		Prefix:    apiKey.Prefix,
		PlainKey:  plainKey, // visible ONLY on creation response
		ExpiresAt: apiKey.ExpiresAt,
		CreatedAt: apiKey.CreatedAt,
	})
}

func (h *APIKeyHandler) List(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	keys, err := h.apiKeyService.ListAPIKeys(c.Request.Context(), userID, projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	resp := make([]dto.APIKeyResponse, len(keys))
	for i, key := range keys {
		resp[i] = dto.APIKeyResponse{
			ID:         key.ID,
			ProjectID:  key.ProjectID,
			Name:       key.Name,
			Prefix:     key.Prefix,
			ExpiresAt:  key.ExpiresAt,
			RevokedAt:  key.RevokedAt,
			LastUsedAt: key.LastUsedAt,
			CreatedAt:  key.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *APIKeyHandler) Rotate(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	keyIDStr := c.Param("keyId")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid key ID format"})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	apiKey, plainKey, err := h.apiKeyService.RotateAPIKey(c.Request.Context(), userID, projectID, keyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.APIKeyResponse{
		ID:        apiKey.ID,
		ProjectID: apiKey.ProjectID,
		Name:      apiKey.Name,
		Prefix:    apiKey.Prefix,
		PlainKey:  plainKey, // visible ONLY on rotation response
		ExpiresAt: apiKey.ExpiresAt,
		CreatedAt: apiKey.CreatedAt,
	})
}

func (h *APIKeyHandler) Revoke(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	keyIDStr := c.Param("keyId")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid key ID format"})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	err = h.apiKeyService.RevokeAPIKey(c.Request.Context(), userID, projectID, keyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API Key revoked successfully"})
}

func (h *APIKeyHandler) Delete(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID"})
		return
	}

	keyIDStr := c.Param("keyId")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid key ID format"})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	err = h.apiKeyService.DeleteAPIKey(c.Request.Context(), userID, projectID, keyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API Key deleted successfully"})
}
