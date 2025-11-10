package repo

import (
	"github.com/micrease/gorme"
	"megin/system/datasource"
)

// 常用的获取数据方法进行封装,继承于gorm
type Repository[T gorme.Model] struct {
	gorme.Repository[T]
}

func (this *Repository[T]) initialize() *Repository[T] {
	this.SetDB(datasource.GetDB())
	this.NewQuery()
	return this
}

func (this *Repository[T]) FindById(id uint) (T, error) {
	return this.Where("id=?", id).First()
}

func (this *Repository[T]) DeleteById(id uint) error {
	return this.Where("id", id).Delete().Error
}

func (this *Repository[T]) DeleteByIds(ids []uint) error {
	return this.WhereIn("id", ids).Delete().Error
}
