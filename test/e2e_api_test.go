package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"merch-store/internal/handlers"
	"merch-store/internal/middleware"
	"merch-store/internal/models"
	"merch-store/internal/repository"
)

// TestRepo – in‑memory реализация интерфейса для e2e тестов
type TestRepo struct {
	mu        sync.Mutex
	employees map[int]models.Employee
	nextID    int
}

func NewTestRepo() *TestRepo {
	return &TestRepo{
		employees: make(map[int]models.Employee),
		nextID:    1,
	}
}

func (r *TestRepo) CreateEmployee(username string) (models.Employee, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	emp := models.Employee{
		ID:          r.nextID,
		Username:    username,
		CoinBalance: 1000,
		CreatedAt:   time.Now(),
	}
	r.employees[r.nextID] = emp
	r.nextID++
	return emp, nil
}

func (r *TestRepo) GetEmployeeByUsername(username string) (models.Employee, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, emp := range r.employees {
		if emp.Username == username {
			return emp, nil
		}
	}
	return models.Employee{}, repository.ErrNotFound
}

func (r *TestRepo) GetEmployeeByID(id int) (models.Employee, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	emp, ok := r.employees[id]
	if !ok {
		return models.Employee{}, repository.ErrNotFound
	}
	return emp, nil
}

func (r *TestRepo) BuyMerch(employeeID int, merchName string, quantity int) error {
	const price = 80
	if merchName != "t-shirt" {
		return repository.ErrInvalidMerch
	}
	total := price * quantity

	r.mu.Lock()
	defer r.mu.Unlock()
	emp, ok := r.employees[employeeID]
	if !ok {
		return repository.ErrNotFound
	}
	if emp.CoinBalance < total {
		return repository.ErrInsufficientFunds
	}
	emp.CoinBalance -= total
	r.employees[employeeID] = emp
	return nil
}

func (r *TestRepo) TransferCoins(fromID, toID, amount int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	from, ok := r.employees[fromID]
	if !ok {
		return repository.ErrNotFound
	}
	to, ok := r.employees[toID]
	if !ok {
		return repository.ErrNotFound
	}
	if from.CoinBalance < amount {
		return repository.ErrInsufficientFunds
	}
	from.CoinBalance -= amount
	to.CoinBalance += amount
	r.employees[fromID] = from
	r.employees[toID] = to
	return nil
}

func (r *TestRepo) GetWalletInfo(employeeID int) (int, []models.Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	emp, ok := r.employees[employeeID]
	if !ok {
		return 0, nil, repository.ErrNotFound
	}

	return emp.CoinBalance, []models.Transaction{}, nil
}

func (r *TestRepo) GetInventory(employeeID int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

// Сценарий покупки мерча
func TestE2E_BuyMerch(t *testing.T) {
	repo := NewTestRepo()
	handler := handlers.NewHandler(repo, "test_secret")

	router := gin.Default()
	router.POST("/api/auth", handler.Auth)
	apiGroup := router.Group("/api")
	apiGroup.Use(middleware.JWTAuthMiddleware("test_secret"))
	{
		apiGroup.GET("/buy/:item", handler.BuyItem)
	}
	ts := httptest.NewServer(router)
	defer ts.Close()

	authPayload := map[string]string{
		"username": "buyer",
		"password": "pass",
	}
	authBody, _ := json.Marshal(authPayload)
	resp, err := http.Post(ts.URL+"/api/auth", "application/json", bytes.NewBuffer(authBody))
	assert.NoError(t, err)
	defer resp.Body.Close()

	var authResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	assert.NoError(t, err)
	token, ok := authResp["token"]
	assert.True(t, ok, "JWT token должен присутствовать в ответе")

	client := &http.Client{}
	req, err := http.NewRequest("GET", ts.URL+"/api/buy/t-shirt", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var buyResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&buyResp)
	assert.NoError(t, err)
	assert.Equal(t, "purchase successful", buyResp["message"])

	buyer, err := repo.GetEmployeeByUsername("buyer")
	assert.NoError(t, err)
	assert.Equal(t, 920, buyer.CoinBalance, "Баланс должен уменьшиться на стоимость покупки")
}

// Сценарий передачи монет
func TestE2E_SendCoin(t *testing.T) {
	repo := NewTestRepo()
	handler := handlers.NewHandler(repo, "test_secret")

	router := gin.Default()
	router.POST("/api/auth", handler.Auth)
	apiGroup := router.Group("/api")
	apiGroup.Use(middleware.JWTAuthMiddleware("test_secret"))
	{
		apiGroup.POST("/sendCoin", handler.SendCoin)
	}
	ts := httptest.NewServer(router)
	defer ts.Close()

	senderPayload := map[string]string{
		"username": "sender",
		"password": "pass",
	}
	senderBody, _ := json.Marshal(senderPayload)
	resp, err := http.Post(ts.URL+"/api/auth", "application/json", bytes.NewBuffer(senderBody))
	assert.NoError(t, err)
	defer resp.Body.Close()
	var senderResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&senderResp)
	assert.NoError(t, err)
	senderToken, ok := senderResp["token"]
	assert.True(t, ok, "sender token должен присутствовать")

	recipientPayload := map[string]string{
		"username": "recipient",
		"password": "pass",
	}
	recipientBody, _ := json.Marshal(recipientPayload)
	resp, err = http.Post(ts.URL+"/api/auth", "application/json", bytes.NewBuffer(recipientBody))
	assert.NoError(t, err)
	defer resp.Body.Close()
	var recipientResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&recipientResp)
	assert.NoError(t, err)
	_, err = repo.GetEmployeeByUsername("recipient")
	assert.NoError(t, err)

	sendPayload := map[string]interface{}{
		"toUser": "recipient",
		"amount": 100,
	}
	sendBody, _ := json.Marshal(sendPayload)
	req, err := http.NewRequest("POST", ts.URL+"/api/sendCoin", bytes.NewBuffer(sendBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+senderToken)
	client := &http.Client{}
	resp, err = client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var sendResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&sendResp)
	assert.NoError(t, err)
	assert.Equal(t, "transfer successful", sendResp["message"])

	sender, err := repo.GetEmployeeByUsername("sender")
	assert.NoError(t, err)
	recipient, err := repo.GetEmployeeByUsername("recipient")
	assert.NoError(t, err)
	assert.Equal(t, 900, sender.CoinBalance, "Баланс отправителя должен уменьшиться на сумму перевода")
	assert.Equal(t, 1100, recipient.CoinBalance, "Баланс получателя должен увеличиться на сумму перевода")
}
