package router

import (
	"megin/app/handler"
	"megin/library/context/router"
)

func GoodsRouter(apiGroup *router.Router) *router.RouterGroup {
	goodsGroup := apiGroup.Group("api/goods")
	{
		goods := handler.Goods{}
		goodsGroup.GET("/index", goods.PageList)
		goodsGroup.POST("/save", goods.Save)
		goodsGroup.POST("/change_status", goods.ChangeStatus)
		goodsGroup.GET("/detail/:id", goods.Detail)
		goodsGroup.DELETE("/delete/:id", goods.Delete)
	}
	return goodsGroup
}
