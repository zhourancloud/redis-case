package logic

import (
	"context"

	"greet/internal/svc"
	"greet/internal/types"

	"github.com/tal-tech/go-zero/core/logx"
)

type OrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) OrderLogic {
	return OrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderLogic) Order(req types.OrderReq) (resp *types.OrderResp, err error) {
	// todo: add your logic here and delete this line
	return &types.OrderResp{
		Message: "hello world",
	}, nil
}
