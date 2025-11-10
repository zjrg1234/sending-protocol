package repo

import (
	"megin/app/model"
)

type Goods struct {
	Repository[model.Goods]
}

func NewGoods() *Goods {
	repo := &Goods{}
	repo.initialize()
	return repo
}
