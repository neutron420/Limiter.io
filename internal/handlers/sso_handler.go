package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"limiter.io/internal/sso"
)

type SSOHandler struct {
	svc *sso.SSOService
}

func NewSSOHandler() *SSOHandler {
	return &SSOHandler{svc: sso.NewSSOService()}
}

func (h *SSOHandler) SetSAMLConfig(c *gin.Context) {
	var cfg sso.SAMLConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.svc.SetSAMLConfig(&cfg)
	c.JSON(http.StatusOK, gin.H{"message": "SAML config saved"})
}

func (h *SSOHandler) GetSAMLConfig(c *gin.Context) {
	orgID := c.Param("orgId")
	cfg, err := h.svc.GetSAMLConfig(orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *SSOHandler) SetOIDCConfig(c *gin.Context) {
	var cfg sso.OIDCConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.svc.SetOIDCConfig(&cfg)
	c.JSON(http.StatusOK, gin.H{"message": "OIDC config saved"})
}

func (h *SSOHandler) GetOIDCConfig(c *gin.Context) {
	orgID := c.Param("orgId")
	cfg, err := h.svc.GetOIDCConfig(orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}
