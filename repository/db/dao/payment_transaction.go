package dao

import (
	"context"

	"gorm.io/gorm"

	"e-mall/repository/db/model"
	"e-mall/types"
)

type PaymentTransactionDao struct {
	*gorm.DB
}

func NewPaymentTransactionDaoByDB(db *gorm.DB) *PaymentTransactionDao {
	return &PaymentTransactionDao{db}
}

func NewPaymentTransactionDao(ctx context.Context) *PaymentTransactionDao {
	return &PaymentTransactionDao{NewDBClient(ctx)}
}

func (dao *PaymentTransactionDao) CreatePaymentTransaction(tx *model.PaymentTransaction) error {
	return dao.DB.Create(tx).Error
}

// ListPaymentTransactionByUserID 买家查询自己的支付流水
func (dao *PaymentTransactionDao) ListPaymentTransactionByUserID(uid uint, page types.BasePage) (list []*types.UserPaymentTransactionResp, total int64, err error) {
	db := dao.DB.Model(&model.PaymentTransaction{}).Where("user_id = ?", uid)
	if err = db.Count(&total).Error; err != nil {
		return
	}
	err = db.Order("created_at DESC").
		Offset((page.PageNum - 1) * page.PageSize).
		Limit(page.PageSize).
		Find(&list).Error
	return
}

// ListPaymentTransactionByPayeeID 商家查询自己收到的支付流水
func (dao *PaymentTransactionDao) ListPaymentTransactionByPayeeID(payeeID uint, page types.BasePage) (list []*types.BossPaymentTransactionResp, total int64, err error) {
	db := dao.DB.Model(&model.PaymentTransaction{}).Where("payee_id = ?", payeeID)
	if err = db.Count(&total).Error; err != nil {
		return
	}
	err = db.Order("created_at DESC").
		Offset((page.PageNum - 1) * page.PageSize).
		Limit(page.PageSize).
		Find(&list).Error
	return
}

// ListPaymentTransactionAdmin 管理员查询所有支付流水
func (dao *PaymentTransactionDao) ListPaymentTransactionAdmin(req *types.AdminPaymentTransactionReq) (list []*types.AdminPaymentTransactionResp, total int64, err error) {
	applyFilters := func(db *gorm.DB) *gorm.DB {
		if req.TransactionType != "" {
			db = db.Where("transaction_type = ?", req.TransactionType)
		}
		if req.UserID != nil {
			db = db.Where("user_id = ?", *req.UserID)
		}
		return db
	}

	countDB := applyFilters(dao.DB.Model(&model.PaymentTransaction{}))
	if err = countDB.Count(&total).Error; err != nil {
		return
	}

	query := applyFilters(dao.DB.Model(&model.PaymentTransaction{}))
	err = query.Order("created_at DESC").
		Offset((req.PageNum - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&list).Error
	return
}
