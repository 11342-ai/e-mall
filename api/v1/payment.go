package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"e-mall/service"
	"e-mall/types"
	"e-mall/utils/ctl"
	"e-mall/utils/log"
)

func OrderPaymentHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req types.PaymentDownReq
		if err := ctx.ShouldBind(&req); err != nil {
			// 参数校验
			log.LogrusObj.Infoln(err)
			ctx.JSON(http.StatusOK, ErrorResponse(ctx, err))
			return
		}

		l := service.GetPaymentSrv()
		resp, err := l.PayDown(ctx.Request.Context(), &req)
		if err != nil {
			log.LogrusObj.Infoln(err)
			ctx.JSON(http.StatusOK, ErrorResponse(ctx, err))
			return
		}
		ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, resp))
	}
}
