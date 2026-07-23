package model

import (
	"time"

	"gorm.io/gorm"
)

type PaymentTransaction struct {
	gorm.Model
	TransactionNo   string `gorm:"size:64;not null;uniqueIndex"`
	OrderNum        string `gorm:"size:64;index"`
	UserID          uint   `gorm:"not null;index"`
	PayeeID         *uint  `gorm:"index"` // 充值交易无收款方
	Amount          float64
	PaymentMethod   string `gorm:"size:20"`
	TransactionType string `gorm:"size:20;not null"`
	Status          string `gorm:"size:20;not null"`
	ProviderTradeNo string `gorm:"size:128"`
	PaidAt          *time.Time
}
