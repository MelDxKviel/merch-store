package models

import "time"

type Employee struct {
	ID          int       `json:"id"`
	Username    string    `json:"username"`
	CoinBalance int       `json:"coin_balance"`
	CreatedAt   time.Time `json:"created_at"`
}

type Purchase struct {
	ID         int       `json:"id"`
	EmployeeID int       `json:"employee_id"`
	MerchName  string    `json:"merch_name"`
	Price      int       `json:"price"`
	Quantity   int       `json:"quantity"`
	CreatedAt  time.Time `json:"created_at"`
}

type Transaction struct {
	ID              int       `json:"id"`
	EmployeeID      int       `json:"employee_id"`
	CounterpartyID  int       `json:"counterparty_id"`
	Amount          int       `json:"amount"`
	TransactionType string    `json:"transaction_type"`
	CreatedAt       time.Time `json:"created_at"`
}
