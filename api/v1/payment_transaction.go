package v1

import (
	"e-mall/consts"
	"e-mall/service"
	"e-mall/types"
	"e-mall/utils/ctl"
	"e-mall/utils/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminPaymentTransactionList 管理员查流水
func AdminPaymentTransactionList() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req types.AdminPaymentTransactionReq

		if err := ctx.ShouldBind(&req); err != nil {
			log.LogrusObj.Infoln(err)
			ctx.JSON(http.StatusOK, ErrorResponse(ctx, err))
			return
		}

		if req.PageSize == 0 {
			req.PageSize = consts.BasePageSize
		}
		if req.PageNum == 0 {
			req.PageNum = 1
		}

		l := service.GetPaymentTransactionSrv()
		resp, err := l.AdminPaymentTransactionList(ctx.Request.Context(), &req)
		if err != nil {
			log.LogrusObj.Infoln(err)
			ctx.JSON(http.StatusOK, ErrorResponse(ctx, err))
			return
		}

		ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, resp))
		return
	}
}

// ListPaymentTransactionByUserID 买家查自己的流水
func ListPaymentTransactionByUserID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req types.UserPaymentTransactionReq
		var userID uint

		if err := ctx.ShouldBind(&req); err != nil {
			log.LogrusObj.Infoln(err)
			ctx.JSON(http.StatusOK, ErrorResponse(ctx, err))
			return
		}

		if req.PageSize == 0 {
			req.PageSize = consts.BasePageSize
		}
		if req.PageNum == 0 {
			req.PageNum = 1
		}

		u, err := ctl.GetUserInfo(ctx.Request.Context())
		if err != nil {
			ctx.JSON(http.StatusOK, ErrorResponse(ctx, err))
			return
		}
		userID = u.Id

		l := service.GetPaymentTransactionSrv()
		resp, err := l.ListPaymentTransactionByUserID(ctx, userID, &req)
		if err != nil {
			log.LogrusObj.Infoln(err)
			ctx.JSON(http.StatusOK, ErrorResponse(ctx, err))
			return
		}
		ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, resp))
		return
	}
}

func ListPaymentTransactionByPayeeID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req types.UserPaymentTransactionReq
		var payee uint

		if err := ctx.ShouldBind(&req); err != nil {
			log.LogrusObj.Infoln(err)
			ctx.JSON(http.StatusOK, ErrorResponse(ctx, err))
			return
		}

		u, err := ctl.GetUserInfo(ctx.Request.Context())
		if err != nil {
			log.LogrusObj.Infoln(err)
			ctx.JSON(http.StatusOK, ErrorResponse(ctx, err))
			return
		}
		payee = u.Id

		l := service.GetPaymentTransactionSrv()
		resp, err := l.ListPaymentTransactionByPayeeID(ctx, payee, &req)
		if err != nil {
			log.LogrusObj.Infoln(err)
			ctx.JSON(http.StatusOK, ErrorResponse(ctx, err))
			return
		}
		ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, resp))
		return
	}
}
