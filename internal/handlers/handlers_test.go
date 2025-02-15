package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"merch-store/internal/models"
)

// fakeRepo – минимальная реализация интерфейса repository.Repository для unit тестов
type fakeRepo struct{}

func (f *fakeRepo) CreateEmployee(username string) (models.Employee, error) {
	return models.Employee{
		ID:          1,
		Username:    username,
		CoinBalance: 1000,
		CreatedAt:   time.Now(),
	}, nil
}

func (f *fakeRepo) GetEmployeeByUsername(username string) (models.Employee, error) {
	return models.Employee{
		ID:          1,
		Username:    username,
		CoinBalance: 1000,
		CreatedAt:   time.Now(),
	}, nil
}

func (f *fakeRepo) GetEmployeeByID(id int) (models.Employee, error) {
	return models.Employee{
		ID:          id,
		Username:    "test",
		CoinBalance: 1000,
		CreatedAt:   time.Now(),
	}, nil
}

func (f *fakeRepo) BuyMerch(employeeID int, merchName string, quantity int) error {
	return nil
}

func (f *fakeRepo) TransferCoins(fromID, toID, amount int) error {
	return nil
}

func (f *fakeRepo) GetWalletInfo(employeeID int) (int, []models.Transaction, error) {
	return 1000, []models.Transaction{}, nil
}

func (f *fakeRepo) GetInventory(employeeID int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func TestHandler_Auth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeRepo{}
	handler := NewHandler(repo, "test_secret")

	router := gin.New()
	router.POST("/api/auth", handler.Auth)

	payload := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/auth", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	_, exists := resp["token"]
	assert.True(t, exists, "token должен присутствовать в ответе")
}

func TestHandler_BuyItem(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeRepo{}
	handler := NewHandler(repo, "test_secret")

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", float64(1))
		c.Next()
	})
	router.GET("/api/buy/:item", handler.BuyItem)

	req, _ := http.NewRequest("GET", "/api/buy/t-shirt", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "purchase successful", resp["message"])
}

func TestHandler_SendCoin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeRepo{}
	handler := NewHandler(repo, "test_secret")

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", float64(1))
		c.Next()
	})
	router.POST("/api/sendCoin", handler.SendCoin)

	payload := map[string]interface{}{
		"toUser": "recipientUser",
		"amount": 100,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/sendCoin", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "transfer successful", resp["message"])
}

func TestHandler_GetInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeRepo{}
	handler := NewHandler(repo, "test_secret")

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", float64(1))
		c.Next()
	})
	router.GET("/api/info", handler.GetInfo)

	req, _ := http.NewRequest("GET", "/api/info", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	_, ok := resp["coins"]
	assert.True(t, ok, "coins должны присутствовать в ответе")
	_, ok = resp["inventory"]
	assert.True(t, ok, "inventory должен присутствовать в ответе")
	_, ok = resp["coinHistory"]
	assert.True(t, ok, "coinHistory должен присутствовать в ответе")
}
