package repository

import "merch-store/internal/models"

type Repository interface {
	CreateEmployee(username string) (models.Employee, error)
	GetEmployeeByID(id int) (models.Employee, error)
	GetEmployeeByUsername(username string) (models.Employee, error)
	BuyMerch(employeeID int, merchName string, quantity int) error
	TransferCoins(fromID, toID, amount int) error
	GetWalletInfo(employeeID int) (int, []models.Transaction, error)
	GetInventory(employeeID int) ([]map[string]interface{}, error)
}
