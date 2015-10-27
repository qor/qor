package exchange_test

import "github.com/jinzhu/gorm"

type Product struct {
	gorm.Model
	Code  string
	Name  string
	Price float64
}
