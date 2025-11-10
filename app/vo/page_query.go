package vo

import "megin/library/logger"

// 所有的请求类对象，都要继承一下这个
type ReqHeader struct {
	TraceId string     `json:"-"`
	Log     logger.Log `json:"-"`
}

//声明一个接口类型
type IPageQuery interface {
	GetPageSize() int
	GetPageNo() int
}

// 分页类对象继承
type PageQuery struct {
	PageSize int `json:"size" form:"size" `   //分页:每页条数
	PageNo   int `json:"page"   form:"page" ` //分页:当前页码,起始页为1
}

func (this PageQuery) GetPageSize() int {
	if this.PageSize == 0 {
		return 10
	}
	return this.PageSize
}

func (this PageQuery) GetPageNo() int {
	if this.PageNo == 0 {
		return 1
	}
	return this.PageNo
}

// 分页返回数据类型
type PageListResp[T any] struct {
	PageSize  int   `json:"page_size"`  //每页条数
	PageNo    int   `json:"page_no"`    //当前页码
	TotalPage int64 `json:"total_page"` //总页数
	TotalSize int64 `json:"total_size"` //总条数
	List      []T   `json:"list"`       //数据列表
}

type PageInfo struct {
	PageSize  int `json:"page_size"`  //每页条数
	PageNo    int `json:"page_no"`    //当前页码
	TotalPage int `json:"total_page"` //总页数
	TotalSize int `json:"total_size"` //总条数
	LastPage  int `json:"last_page"`  //最后一页
}

type PageSummary[T any] struct {
	PageInfo PageInfo `json:"page_info"`
	Summary  T        `json:"summary"`
}
