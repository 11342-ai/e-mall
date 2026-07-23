package types

import "time"

// ===== 买家查询 =====

type UserPaymentTransactionReq struct {
	BasePage
}

type UserPaymentTransactionResp struct {
	ID              uint       `json:"id"`
	TransactionNo   string     `json:"transaction_no"`
	OrderNum        string     `json:"order_num"`
	Amount          float64    `json:"amount"`
	PaymentMethod   string     `json:"payment_method"`
	TransactionType string     `json:"transaction_type"`
	Status          string     `json:"status"`
	PaidAt          *time.Time `json:"paid_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ===== 商家查询 =====

type BossPaymentTransactionReq struct {
	BasePage
}

type BossPaymentTransactionResp struct {
	ID              uint       `json:"id"`
	TransactionNo   string     `json:"transaction_no"`
	OrderNum        string     `json:"order_num"`
	UserID          uint       `json:"user_id"`
	Amount          float64    `json:"amount"`
	PaymentMethod   string     `json:"payment_method"`
	TransactionType string     `json:"transaction_type"`
	Status          string     `json:"status"`
	PaidAt          *time.Time `json:"paid_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ===== 管理员查询 =====

type AdminPaymentTransactionReq struct {
	TransactionType string `form:"transaction_type" json:"transaction_type"`
	UserID          *uint  `form:"user_id" json:"user_id"`
	BasePage
}

type AdminPaymentTransactionResp struct {
	ID              uint       `json:"id"`
	TransactionNo   string     `json:"transaction_no"`
	OrderNum        string     `json:"order_num"`
	UserID          uint       `json:"user_id"`
	PayeeID         *uint      `json:"payee_id"`
	Amount          float64    `json:"amount"`
	PaymentMethod   string     `json:"payment_method"`
	TransactionType string     `json:"transaction_type"`
	Status          string     `json:"status"`
	ProviderTradeNo string     `json:"provider_trade_no"`
	PaidAt          *time.Time `json:"paid_at"`
	CreatedAt       time.Time  `json:"created_at"`
}
