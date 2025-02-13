package handlers

import (
	"merch-store/internal/middleware"
	"merch-store/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo      repository.Repository
	jwtSecret string
}

func NewHandler(repo repository.Repository, jwtSecret string) *Handler {
	return &Handler{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (h *Handler) Auth(c *gin.Context) {
	type AuthRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": err.Error()})
		return
	}

	employee, err := h.repo.GetEmployeeByUsername(req.Username)
	if err != nil {
		employee, err = h.repo.CreateEmployee(req.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"errors": "cannot create employee"})
			return
		}
	}

	token, err := middleware.GenerateJWT(employee.ID, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"errors": "cannot generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handler) BuyItem(c *gin.Context) {
	item := c.Param("item")
	if item == "" {
		c.JSON(http.StatusBadRequest, gin.H{"errors": "item is required"})
		return
	}
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"errors": "user not found in context"})
		return
	}
	userID := int(userIDInterface.(float64))

	if err := h.repo.BuyMerch(userID, item, 1); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "purchase successful"})
}

func (h *Handler) SendCoin(c *gin.Context) {
	type SendCoinRequest struct {
		ToUser string `json:"toUser" binding:"required"`
		Amount int    `json:"amount" binding:"required,gt=0"`
	}
	var req SendCoinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": err.Error()})
		return
	}

	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"errors": "user not found in context"})
		return
	}
	fromUserID := int(userIDInterface.(float64))

	recipient, err := h.repo.GetEmployeeByUsername(req.ToUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": "recipient not found"})
		return
	}

	if err := h.repo.TransferCoins(fromUserID, recipient.ID, req.Amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "transfer successful"})
}

func (h *Handler) GetInfo(c *gin.Context) {

	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"errors": "user not found in context"})
		return
	}
	userID := int(userIDInterface.(float64))

	balance, transactions, err := h.repo.GetWalletInfo(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"errors": err.Error()})
		return
	}

	var received []map[string]interface{}
	var sent []map[string]interface{}
	for _, t := range transactions {
		if t.Amount > 0 {
			received = append(received, map[string]interface{}{
				"fromUser": t.CounterpartyID,
				"amount":   t.Amount,
			})
		} else {
			sent = append(sent, map[string]interface{}{
				"toUser": t.CounterpartyID,
				"amount": -t.Amount,
			})
		}
	}

	inventory, err := h.repo.GetInventory(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"errors": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"coins":     balance,
		"inventory": inventory,
		"coinHistory": map[string]interface{}{
			"received": received,
			"sent":     sent,
		},
	})
}
