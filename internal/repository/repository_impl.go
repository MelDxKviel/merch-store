package repository

import (
	"database/sql"
	"fmt"
	"merch-store/internal/config"
	"merch-store/internal/models"
	"time"
)

type repositoryImpl struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) CreateEmployee(username string) (models.Employee, error) {
	var emp models.Employee
	err := r.db.QueryRow(
		`INSERT INTO employees (username, coin_balance, created_at) VALUES ($1, $2, $3) RETURNING id, username, coin_balance, created_at`,
		username, 1000, time.Now(),
	).Scan(&emp.ID, &emp.Username, &emp.CoinBalance, &emp.CreatedAt)
	if err != nil {
		return emp, err
	}
	return emp, nil
}

func (r *repositoryImpl) BuyMerch(employeeID int, merchName string, quantity int) error {
	price, ok := config.MerchPrices[merchName]
	if !ok {
		return ErrInvalidMerch
	}
	totalCost := price * quantity

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var balance int
	err = tx.QueryRow(`SELECT coin_balance FROM employees WHERE id = $1`, employeeID).Scan(&balance)
	if err != nil {
		return err
	}
	if balance < totalCost {
		return ErrInsufficientFunds
	}

	_, err = tx.Exec(`UPDATE employees SET coin_balance = coin_balance - $1 WHERE id = $2`, totalCost, employeeID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO purchases (employee_id, merch_name, price, quantity, created_at) VALUES ($1, $2, $3, $4, $5)`,
		employeeID, merchName, price, quantity, time.Now(),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repositoryImpl) TransferCoins(fromID, toID, amount int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var fromBalance int
	err = tx.QueryRow(`SELECT coin_balance FROM employees WHERE id = $1`, fromID).Scan(&fromBalance)
	if err != nil {
		return err
	}
	if fromBalance < amount {
		return ErrInsufficientFunds
	}

	var toBalance int
	err = tx.QueryRow(`SELECT coin_balance FROM employees WHERE id = $1`, toID).Scan(&toBalance)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE employees SET coin_balance = coin_balance - $1 WHERE id = $2`, amount, fromID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE employees SET coin_balance = coin_balance + $1 WHERE id = $2`, amount, toID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO transactions (employee_id, counterparty_id, amount, transaction_type, created_at) VALUES ($1, $2, $3, $4, $5)`,
		fromID, toID, amount, "transfer", time.Now(),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repositoryImpl) GetEmployeeByID(id int) (models.Employee, error) {
	var emp models.Employee
	err := r.db.QueryRow(
		`SELECT id, username, coin_balance, created_at FROM employees WHERE id = $1`,
		id,
	).Scan(&emp.ID, &emp.Username, &emp.CoinBalance, &emp.CreatedAt)
	if err != nil {
		return emp, err
	}
	return emp, nil
}

func (r *repositoryImpl) GetEmployeeByUsername(username string) (models.Employee, error) {
	var emp models.Employee
	err := r.db.QueryRow(
		`SELECT id, username, coin_balance, created_at FROM employees WHERE username = $1`,
		username,
	).Scan(&emp.ID, &emp.Username, &emp.CoinBalance, &emp.CreatedAt)
	if err != nil {
		return emp, err
	}
	return emp, nil
}

func (r *repositoryImpl) GetWalletInfo(employeeID int) (int, []models.Transaction, error) {

	var balance int
	err := r.db.QueryRow(`SELECT coin_balance FROM employees WHERE id = $1`, employeeID).Scan(&balance)
	if err != nil {
		return 0, nil, err
	}

	rows, err := r.db.Query(`
        SELECT id, employee_id, counterparty_id, amount, transaction_type, created_at 
        FROM transactions 
        WHERE employee_id = $1 
        ORDER BY created_at DESC
    `, employeeID)
	if err != nil {
		return balance, nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.ID, &t.EmployeeID, &t.CounterpartyID, &t.Amount, &t.TransactionType, &t.CreatedAt); err != nil {
			return balance, nil, err
		}
		transactions = append(transactions, t)
	}

	return balance, transactions, nil
}

func (r *repositoryImpl) GetInventory(employeeID int) ([]map[string]interface{}, error) {

	query := `
		SELECT merch_name, SUM(quantity) AS total_quantity
		FROM purchases
		WHERE employee_id = $1
		GROUP BY merch_name
	`

	rows, err := r.db.Query(query, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query inventory: %w", err)
	}
	defer rows.Close()

	var inventory []map[string]interface{}
	for rows.Next() {
		var merchName string
		var totalQuantity int

		if err := rows.Scan(&merchName, &totalQuantity); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		item := map[string]interface{}{
			"type":     merchName,
			"quantity": totalQuantity,
		}
		inventory = append(inventory, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return inventory, nil
}
