package service

import (
	"context"
	"sync"

	"e-mall/repository/db/dao"
	"e-mall/types"
	util "e-mall/utils/log"
)

var CarouselSrvIns *CarouselSrv
var CarouselSrvOnce sync.Once

type CarouselSrv struct {
}

func GetCarouselSrv() *CarouselSrv {
	CarouselSrvOnce.Do(func() {
		CarouselSrvIns = &CarouselSrv{}
	})
	return CarouselSrvIns
}

// ListCarousel 列表
func (s *CarouselSrv) ListCarousel(ctx context.Context, req *types.ListCarouselReq) (resp interface{}, err error) {
	carousels, err := dao.NewCarouselDao(ctx).ListCarousel()
	if err != nil {
		util.LogrusObj.Error(err)
		return
	}

	resp = &types.DataListResp{
		Item:  carousels,
		Total: int64(len(carousels)),
	}

	return
}
