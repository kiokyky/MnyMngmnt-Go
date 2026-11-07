package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-jwt-secret"
	expectedUserID := uuid.New()

	r := gin.New()
	r.Use(AuthMiddleware(secret))
	r.GET("/protected", func(c *gin.Context) {
		uid, ok := GetUserID(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "missing user id in context"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": uid.String()})
	})

	t.Run("Rejects missing Authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d", w.Code)
		}
	})

	t.Run("Rejects invalid token signed with wrong secret", func(t *testing.T) {
		badToken, _ := GenerateToken(expectedUserID, "user@example.com", "wrong-secret", 24)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+badToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d", w.Code)
		}
	})

	t.Run("Allows valid Bearer token and injects user_id", func(t *testing.T) {
		validToken, err := GenerateToken(expectedUserID, "user@example.com", secret, 24)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d (%s)", w.Code, w.Body.String())
		}
	})
}

