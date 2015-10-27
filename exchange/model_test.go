package exchange_test

type Product struct {
	Code  string `gorm:"primary_key" sql:"size:100"`
	Name  string
	Price float64
}
