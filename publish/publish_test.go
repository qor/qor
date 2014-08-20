package publish_test

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor/publish"

	"testing"
)

var pb *publish.Publish

func init() {
	pb, _ = publish.Open("sqlite3", "/tmp/qor_publish_test.db")
	pb.Support(&Product{})

	pb.AutoMigrate(&Product{})
	pb.AutoMigrateDrafts()
}

type Product struct {
	Name string
}

func TestPublishStruct(t *testing.T) {
	pb.Save(Product{Name: "product"})
	var product Product
	pb.First(&product)
}
