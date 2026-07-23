package service

import (
	"context"
	"e-mall/consts"
	"e-mall/repository/db/dao"
	"e-mall/types"
	util "e-mall/utils/log"
	"sync"
)

var PaymentTransactionSrvIns *PaymentTransactionSrv
var PaymentTransactionSrvOnce sync.Once

type PaymentTransactionSrv struct{}

func GetPaymentTransactionSrv() *PaymentTransactionSrv {
	PaymentTransactionSrvOnce.Do(func() {
		PaymentTransactionSrvIns = &PaymentTransactionSrv{}
	})
	return PaymentTransactionSrvIns
}

// AdminPaymentTransactionList 管理员查流水
func (s *PaymentTransactionSrv) AdminPaymentTransactionList(ctx context.Context, req *types.AdminPaymentTransactionReq) (resp interface{}, err error) {
	if req.PageSize == 0 {
		req.PageSize = consts.BasePageSize
	}
	if req.PageNum == 0 {
		req.PageNum = 1
	}

	paymentTransaction, total, err := dao.NewPaymentTransactionDao(ctx).ListPaymentTransactionAdmin(req)
	if err != nil {
		util.LogrusObj.Error(err)
		return
	}
	resp = types.DataListResp{
		Item:  paymentTransaction,
		Total: total,
	}
	return
}

// ListPaymentTransactionByUserID 买家查自己的流水
func (s *PaymentTransactionSrv) ListPaymentTransactionByUserID(ctx context.Context, userID uint, req *types.UserPaymentTransactionReq) (resp interface{}, err error) {
	if req.PageSize == 0 {
		req.PageSize = consts.BasePageSize
	}
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	paymentTransaction, total, err := dao.NewPaymentTransactionDao(ctx).ListPaymentTransactionByUserID(userID, req.BasePage)
	if err != nil {
		util.LogrusObj.Error(err)
		return
	}
	resp = types.DataListResp{
		Item:  paymentTransaction,
		Total: total,
	}
	return
}

// ListPaymentTransactionByPayeeID 商家查自己的流水
func (s *PaymentTransactionSrv) ListPaymentTransactionByPayeeID(ctx context.Context, payeeID uint, req *types.UserPaymentTransactionReq) (resp interface{}, err error) {
	if req.PageSize == 0 {
		req.PageSize = consts.BasePageSize
	}
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	paymentTransaction, total, err := dao.NewPaymentTransactionDao(ctx).ListPaymentTransactionByPayeeID(payeeID, req.BasePage)
	if err != nil {
		util.LogrusObj.Error(err)
		return
	}

	resp = types.DataListResp{
		Item:  paymentTransaction,
		Total: total,
	}
	return
}
