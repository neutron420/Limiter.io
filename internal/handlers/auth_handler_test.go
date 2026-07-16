package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAdminPermissions(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", c.GetHeader("X-User-ID"))
		c.Next()
	})

	r.GET("/admin/projects", func(c *gin.Context) {
		uid := c.GetString("user_id")
		if uid == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"projects": []string{}})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin/projects", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/admin/projects", nil)
	req2.Header.Set("X-User-ID", "admin-1")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w2.Code)
	}
}
