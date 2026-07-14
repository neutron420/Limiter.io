package handlers

import (
	"log"
	"net/http"

	"limiter.io/internal/dto"
	"limiter.io/internal/repository"
	internalws "limiter.io/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WSHandler struct {
	hub         *internalws.Hub
	projectRepo repository.ProjectRepository
}

func NewWSHandler(hub *internalws.Hub, projectRepo repository.ProjectRepository) *WSHandler {
	return &WSHandler{
		hub:         hub,
		projectRepo: projectRepo,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow cross-origin connection for developers dashboard
	},
}

func (h *WSHandler) Connect(c *gin.Context) {
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

	// Verify project ownership before upgrading HTTP connection
	userID := uuid.MustParse(userIDStr.(string))
	project, err := h.projectRepo.GetByID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "project not found"})
		return
	}

	if project.UserID != userID {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "unauthorized to stream analytics for this project"})
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade HTTP to WebSocket: %v", err)
		return
	}

	client := &internalws.Client{
		Hub:       h.hub,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		ProjectID: projectIDStr,
	}

	h.hub.RegisterClient(client)

	// Start write pump in a separate goroutine
	go client.WritePump()

	// Read loop to detect disconnects
	defer func() {
		h.hub.UnregisterClient(client)
		conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			
		break
		}
	}
}
