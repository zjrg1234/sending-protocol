package vo

import "time"

// 请求参数
type Goods struct {
	ID          uint      `json:"id"`
	GoodsName   string    `json:"product_name"` //商品名称
	Price       float64   `json:"price"`        //价格
	Number      int       `json:"number"`       //数量
	CateId      int       `json:"cate_id"`      //分类ID
	Status      int       `json:"status"`       //状态1开启,2关闭
	Description string    `json:"description"`  //描述
	CreatedBy   uint64    `json:"created_by"`   // 创建者
	UpdatedBy   uint64    `json:"updated_by"`   // 更新者
	CreatedAt   time.Time `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`   // 更新时间
	ExpireTime  time.Time `json:"expire_time"`  // 过期时间
}

// 分页查询请求
type GoodsPageQueryReq struct {
	PageQuery        //分页参数
	ID        uint   `json:"id"`           //ID
	GoodsName string `json:"product_name"` //商品名称
	CateId    int    `json:"cate_id"`      //分类ID
}

// 修改状态
type GoodsStatusReq struct {
	ID     uint `json:"id"`     //ID
	Status int  `json:"status"` //状态1开启,2关闭
}
