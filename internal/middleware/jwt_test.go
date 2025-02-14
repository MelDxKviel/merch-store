package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)


func TestJWTAuthMiddleware_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(JWTAuthMiddleware("mysecret"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]string
	_ = json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "Authorization header is missing", resp["error"])
}


func TestJWTAuthMiddleware_InvalidHeaderFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(JWTAuthMiddleware("mysecret"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)

	req.Header.Set("Authorization", "InvalidToken")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]string
	_ = json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "Authorization header format must be Bearer {token}", resp["error"])
}


func TestJWTAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(JWTAuthMiddleware("mysecret"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)

	req.Header.Set("Authorization", "Bearer invalidtoken")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]string
	_ = json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "Invalid token", resp["error"])
}


func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenStr, err := GenerateJWT(123, "mysecret")
	assert.NoError(t, err)

	router := gin.New()
	router.Use(JWTAuthMiddleware("mysecret"))

	router.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "userID not set"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"userID": userID})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)

	idFloat, ok := resp["userID"].(float64)
	assert.True(t, ok)
	assert.Equal(t, 123, int(idFloat))
}


func TestGenerateJWT(t *testing.T) {
	secret := "mysecret"
	userID := 456
	tokenStr, err := GenerateJWT(userID, secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	assert.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(jwt.MapClaims)
	assert.True(t, ok)

	uid, ok := claims["userID"].(float64)
	assert.True(t, ok)
	assert.Equal(t, float64(userID), uid)

	exp, ok := claims["exp"].(float64)
	assert.True(t, ok)
	now := time.Now().Unix()
	assert.True(t, int64(exp) > now)
}
