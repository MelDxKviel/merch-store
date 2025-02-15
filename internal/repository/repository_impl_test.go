package repository

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestBuyMerch_InsufficientFunds(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	employeeID := 1
	merchName := "t-shirt"
	quantity := 1
	price := 80
	totalCost := price * quantity

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT coin_balance FROM employees WHERE id = \$1`).
		WithArgs(employeeID).
		WillReturnRows(sqlmock.NewRows([]string{"coin_balance"}).AddRow(totalCost - 10))
	mock.ExpectRollback()

	err = repo.BuyMerch(employeeID, merchName, quantity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), ErrInsufficientFunds.Error())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuyMerch_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	employeeID := 1
	merchName := "t-shirt"
	quantity := 2
	price := 80
	totalCost := price * quantity

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT coin_balance FROM employees WHERE id = \$1`).
		WithArgs(employeeID).
		WillReturnRows(sqlmock.NewRows([]string{"coin_balance"}).AddRow(totalCost + 100))

	mock.ExpectExec(`UPDATE employees SET coin_balance = coin_balance - \$1 WHERE id = \$2`).
		WithArgs(totalCost, employeeID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`INSERT INTO purchases`).
		WithArgs(employeeID, merchName, price, quantity, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = repo.BuyMerch(employeeID, merchName, quantity)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransferCoins_InsufficientFunds(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	employeeID := 1
	toID := 2
	amount := 100

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT coin_balance FROM employees WHERE id = \$1`).
		WithArgs(employeeID).
		WillReturnRows(sqlmock.NewRows([]string{"coin_balance"}).AddRow(10))
	mock.ExpectRollback()

	err = repo.TransferCoins(employeeID, toID, amount)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), ErrInsufficientFunds.Error())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransferCoins_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	fromID, toID := 1, 2
	amount := 50

	mock.ExpectBegin()

	fromBalanceQuery := regexp.QuoteMeta(`SELECT coin_balance FROM employees WHERE id = $1`)
	mock.ExpectQuery(fromBalanceQuery).
		WithArgs(fromID).
		WillReturnRows(sqlmock.NewRows([]string{"coin_balance"}).AddRow(100))

	toBalanceQuery := regexp.QuoteMeta(`SELECT coin_balance FROM employees WHERE id = $1`)
	mock.ExpectQuery(toBalanceQuery).
		WithArgs(toID).
		WillReturnRows(sqlmock.NewRows([]string{"coin_balance"}).AddRow(200))

	updFrom := regexp.QuoteMeta(`UPDATE employees SET coin_balance = coin_balance - $1 WHERE id = $2`)
	mock.ExpectExec(updFrom).
		WithArgs(amount, fromID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	updTo := regexp.QuoteMeta(`UPDATE employees SET coin_balance = coin_balance + $1 WHERE id = $2`)
	mock.ExpectExec(updTo).
		WithArgs(amount, toID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	insTx := regexp.QuoteMeta(`INSERT INTO transactions (employee_id, counterparty_id, amount, transaction_type, created_at) VALUES ($1, $2, $3, $4, $5)`)
	mock.ExpectExec(insTx).
		WithArgs(fromID, toID, amount, "transfer", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = repo.TransferCoins(fromID, toID, amount)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestGetWalletInfo(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	employeeID := 1

	mock.ExpectQuery(`SELECT coin_balance FROM employees WHERE id = \$1`).
		WithArgs(employeeID).
		WillReturnRows(sqlmock.NewRows([]string{"coin_balance"}).AddRow(1000))

	createdAt := time.Now()
	rows := sqlmock.NewRows([]string{"id", "employee_id", "counterparty_id", "amount", "transaction_type", "created_at"}).
		AddRow(1, employeeID, 2, 50, "transfer", createdAt).
		AddRow(2, employeeID, 3, -30, "transfer", createdAt)
	mock.ExpectQuery(`SELECT id, employee_id, counterparty_id, amount, transaction_type, created_at FROM transactions WHERE employee_id = \$1 ORDER BY created_at DESC`).
		WithArgs(employeeID).
		WillReturnRows(rows)

	balance, transactions, err := repo.GetWalletInfo(employeeID)
	assert.NoError(t, err)
	assert.Equal(t, 1000, balance)
	assert.Len(t, transactions, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetInventory(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	employeeID := 1

	rows := sqlmock.NewRows([]string{"merch_name", "total_quantity"}).
		AddRow("t-shirt", 3).
		AddRow("pen", 5)
	mock.ExpectQuery(`SELECT merch_name, SUM\(quantity\) AS total_quantity FROM purchases WHERE employee_id = \$1 GROUP BY merch_name`).
		WithArgs(employeeID).
		WillReturnRows(rows)

	inventory, err := repo.GetInventory(employeeID)
	assert.NoError(t, err)
	assert.Len(t, inventory, 2)

	var foundTshirt, foundPen bool
	for _, item := range inventory {
		if item["type"] == "t-shirt" && item["quantity"] == 3 {
			foundTshirt = true
		}
		if item["type"] == "pen" && item["quantity"] == 5 {
			foundPen = true
		}
	}
	assert.True(t, foundTshirt, "inventory should contain t-shirt with quantity 3")
	assert.True(t, foundPen, "inventory should contain pen with quantity 5")

	assert.NoError(t, mock.ExpectationsWereMet())
}
