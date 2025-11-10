package product_test

import (
	"megin/app"
	"megin/app/vo"
	"megin/system"
	"megin/tests"
	"testing"
)

func TestCreate(t *testing.T) {

	req := vo.Goods{}
	req.ID = 0
	req.Status = 1
	req.CateId = 100
	req.GoodsName = "GoodsName111"
	req.Price = 12
	req.Description = "test"
	req.Number = 100
	w := tests.Post("/api/goods/save", req)
	tests.Echo(w.Body.String())
}

func TestPageList(t *testing.T) {
	w := tests.Get("/api/goods/index", "page_size=10")
	tests.Echo(w.Body.String())
}

func TestChangeStatus(t *testing.T) {
	req := vo.GoodsStatusReq{
		ID:     1,
		Status: 1,
	}
	w := tests.Post("/api/goods/change_status", req)
	tests.Echo(w.Body.String())
}

func TestDel(t *testing.T) {
	w := tests.Delete("/api/goods/delete/1")
	tests.Echo(w.Body.String())
}

// ====================
func TestMain(m *testing.M) {
	system.ServerInit("../../resources/config-dev.yaml", app.OnAppInitialize)
	m.Run()
}
