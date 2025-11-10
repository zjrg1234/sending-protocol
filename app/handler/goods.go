package handler

import (
	"github.com/spf13/cast"
	"megin/app/service"
	"megin/app/vo"
	"megin/library/context/api"
	"megin/library/context/result"
	"megin/library/validate"
	"strings"
)

// !规范的注释在生成文档时会被同步到接口文档中
// 商品
type Goods struct {
}

// 保存信息
func (this *Goods) Save(ctx *api.Context) *result.Result {
	var req vo.Goods
	validate.BindWithPanic(ctx, &req)
	err := service.NewGoods(ctx).Save(req)
	if err != nil {
		return result.Failed(err, "保存失败")
	}
	return result.Success()
}

// 查询列表
func (this *Goods) PageList(ctx *api.Context) *result.Result {
	var req vo.GoodsPageQueryReq
	validate.BindWithPanic(ctx, &req)
	list, err := service.NewGoods(ctx).PageList(req)
	if err != nil {
		return result.Failed(err, "查询失败")
	}
	return result.Success(list)
}

// 修改状态
func (this *Goods) ChangeStatus(ctx *api.Context) *result.Result {
	var req vo.GoodsStatusReq
	validate.BindWithPanic(ctx, &req)
	err := service.NewGoods(ctx).ChangeStatus(req)
	if err != nil {
		return result.Failed(err)
	}
	return result.Success()
}

// 获取详情
func (this *Goods) Detail(ctx *api.Context) *result.Result {
	reqId := ctx.Param("id")
	id := cast.ToUint(reqId)
	if id <= 0 {
		return result.Success()
	}
	model, err := service.NewGoods(ctx).Detail(id)
	if err != nil {
		return result.Failed(err)
	}
	return result.Success(model)
}

// 删除
func (g *Goods) Delete(ctx *api.Context) *result.Result {
	id := ctx.Param("id")
	ids := strings.Split(id, ",")
	err := service.NewGoods(ctx).Delete(ids)
	if err != nil {
		return result.Failed(err)
	}
	return result.Success()
}
