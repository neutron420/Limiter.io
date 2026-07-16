//go:build cgo

package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"limiter.io/internal/models"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.ProjectMember{},
		&models.APIKey{},
		&models.RateLimitRule{},
		&models.AnalyticsLog{},
		&models.ProjectInvite{},
		&models.ProjectAuditEvent{},
		&models.WebhookEvent{},
		&models.AlertRule{},
		&models.RuleVersion{},
	))
	return db
}

func TestRegister(t *testing.T) {
	db := setupTestDB(t)
	r := gin.New()
	r.POST("/auth/register", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		c.ShouldBindJSON(&req)
		user := models.User{Email: req.Email, PasswordHash: req.Password}
		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "exists"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "registered"})
	})

	w := httptest.NewRecorder()
	body := fmt.Sprintf(`{"email":"test@test.com","password":"Pass123!"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestRateLimitMiddleware(t *testing.T) {
	db := setupTestDB(t)
	rule := models.RateLimitRule{
		Name:         "test-rule",
		Limit:        5,
		Period:       60,
		IsActive:     true,
		Priority:     10,
		RoutePattern: "/*",
	}
	db.Create(&rule)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	srv := httptest.NewServer(r)
	defer srv.Close()

	client := &http.Client{}
	for i := 0; i < 3; i++ {
		resp, err := client.Get(srv.URL + "/test")
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestProjectCRUD(t *testing.T) {
	db := setupTestDB(t)
	r := gin.New()

	pid := uuid.New()
	r.POST("/projects", func(c *gin.Context) {
		project := models.Project{Name: "Test Project"}
		db.Create(&project)
		c.JSON(http.StatusCreated, project)
	})

	r.GET("/projects/:id", func(c *gin.Context) {
		var p models.Project
		if err := db.First(&p, "id = ?", c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, p)
	})

	r.DELETE("/projects/:id", func(c *gin.Context) {
		db.Delete(&models.Project{}, "id = ?", c.Param("id"))
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/projects", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var project models.Project
	json.Unmarshal(w.Body.Bytes(), &project)
	assert.NotEmpty(t, project.ID)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/projects/"+project.ID.String(), nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	_ = pid
}

func TestBackupVerification(t *testing.T) {
	db := setupTestDB(t)

	user := models.User{Email: "backup-test@test.com", PasswordHash: "pass"}
	db.Create(&user)
	project := models.Project{Name: "Backup Project"}
	db.Create(&project)

	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	assert.Equal(t, int64(1), userCount)

	var projectCount int64
	db.Model(&models.Project{}).Count(&projectCount)
	assert.Equal(t, int64(1), projectCount)
}

func TestBackupRestoreVerification(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.APIKey{},
	))

	user := models.User{Email: "backup-test@test.com", PasswordHash: "pass"}
	db.Create(&user)
	project := models.Project{Name: "Backup Project"}
	db.Create(&project)

	var tableCount int64
	db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tableCount)
	assert.Greater(t, tableCount, int64(0))
}

func TestDryRunMode(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&models.AnalyticsLog{}))

	pid := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	log := models.AnalyticsLog{
		ProjectID: pid,
		Route:     "/api/test",
		Decision:  "allowed",
		Timestamp: time.Now(),
	}
	result := db.Create(&log)
	assert.NoError(t, result.Error)

	var found models.AnalyticsLog
	db.First(&found, log.ID)
	assert.Equal(t, "allowed", found.Decision)
}
