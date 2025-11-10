package service

import (
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"megin/app/consts"
	"megin/app/model"
	"megin/app/repo"
	"megin/app/vo"
	"megin/library/context/api"
	"time"
)

// TODO::这是个demo演示类
type Goods struct {
	Service
	repo *repo.Goods
}

func NewGoods(ctx *api.Context) *Goods {
	service := Goods{}
	service.initialize(ctx)
	service.repo = repo.NewGoods()
	return &service
}

// 保存商品
func (this *Goods) Save(req vo.Goods) error {
	this.Log.Info("Goods.save", zap.Any("GoodsSaveReq", req))
	model := this.repo.NewModel()
	err := copier.Copy(&model, &req)
	if err != nil {
		return this.error(err)
	}

	model.Status = consts.StatusEnable
	if req.ID == 0 {
		model.CreatedBy = this.ctx.JwtClaimData.UserId
	} else {
		old, err := this.repo.FindById(req.ID)
		if err != nil {
			return err
		}
		model.UpdatedBy = this.ctx.JwtClaimData.UserId
		model.UpdatedAt = time.Now()
		model.CreatedAt = old.CreatedAt
		model.CreatedBy = old.CreatedBy
	}

	err = this.repo.Save(&model).Error
	if err != nil {
		return this.error(err, "保存失败")
	}
	return nil
}

// 分页查询
func (this *Goods) PageList(req vo.GoodsPageQueryReq) (*vo.PageListResp[vo.Goods], error) {
	query := this.repo.NewQuery()
	if req.ID > 0 {
		query.Where("id", req.ID)
	}

	if req.CateId > 0 {
		query.Where("cate_id", req.CateId)
	}

	if len(req.GoodsName) > 0 {
		query.Like("goods_name", req.GoodsName)
	}
	pageList, err := query.Paginate(req.PageNo, req.PageSize)
	if err != nil {
		return nil, this.error(err)
	}

	resp := vo.PageListResp[vo.Goods]{}
	resp.PageNo = pageList.PageNo
	resp.TotalPage = pageList.TotalPage
	resp.PageSize = pageList.PageSize
	resp.TotalSize = pageList.TotalSize
	resp.List = []vo.Goods{}
	for _, item := range pageList.List {
		goods := vo.Goods{}
		err = copier.Copy(&goods, item)
		if err != nil {
			return &resp, this.error(err)
		}
		resp.List = append(resp.List, goods)
	}
	return &resp, err
}

// 修改状态
func (this *Goods) ChangeStatus(req vo.GoodsStatusReq) error {
	model, err := this.repo.FindById(req.ID)
	if err != nil {
		return this.error(err)
	}

	if model.ID == 0 {
		return this.errorMessage("商品不存在")
	}

	if req.Status == model.Status {
		return this.errorCodeMessage(4001, "无须修改")
	}

	model.Status = req.Status
	err = this.repo.Select("status").Save(model).Error
	if err != nil {
		return this.error(err)
	}
	return nil
}

// 获取详情
func (this *Goods) Detail(id uint) (model.Goods, error) {
	model, err := this.repo.FindById(id)
	return model, this.error(err)
}

// 删除多个ID
func (this *Goods) Delete(ids []string) error {
	err := this.repo.Delete(&model.Goods{}, ids).Error
	if err != nil {
		return this.error(err)
	}
	return nil
}
