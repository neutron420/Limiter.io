package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"limiter.io/internal/services"
)

type PasskeyHandler struct {
	svc *services.PasskeyService
}

func NewPasskeyHandler(db *gorm.DB) *PasskeyHandler {
	return &PasskeyHandler{
		svc: services.NewPasskeyService(db),
	}
}

func (h *PasskeyHandler) BeginRegistration(c *gin.Context) {
	userID := c.GetString("user_id")
	session, err := h.svc.BeginRegistration(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"challenge":  session.Challenge,
		"rp":         map[string]string{"name": "Limiter.io"},
		"user":       map[string]string{"id": userID, "name": c.GetString("email")},
		"pubKeyCredParams": []map[string]interface{}{
			{"type": "public-key", "alg": -7},
			{"type": "public-key", "alg": -257},
		},
		"timeout": 60000,
	})
}

func (h *PasskeyHandler) CompleteRegistration(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		CredentialID    string `json:"credential_id"`
		PublicKey       string `json:"public_key"`
		AttestationType string `json:"attestation_type"`
		AAGUID          string `json:"aaguid"`
		Nickname        string `json:"nickname"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	passkey, err := h.svc.CompleteRegistration(userID, req.CredentialID, req.PublicKey, req.AttestationType, req.AAGUID, req.Nickname)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, passkey)
}

func (h *PasskeyHandler) BeginLogin(c *gin.Context) {
	challenge := h.svc.BeginLogin()
	c.JSON(http.StatusOK, gin.H{
		"challenge": challenge,
		"timeout":   60000,
	})
}

func (h *PasskeyHandler) CompleteLogin(c *gin.Context) {
	var req struct {
		CredentialID string `json:"credential_id"`
		Signature    string `json:"signature"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.svc.CompleteLogin(req.CredentialID, req.Signature)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "login successful", "user_id": user.ID})
}

func (h *PasskeyHandler) ListPasskeys(c *gin.Context) {
	passkeys, err := h.svc.ListPasskeys(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, passkeys)
}

func (h *PasskeyHandler) DeletePasskey(c *gin.Context) {
	if err := h.svc.DeletePasskey(c.Param("id"), c.GetString("user_id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "passkey not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "passkey deleted"})
}
