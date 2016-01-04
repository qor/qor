package publish_test

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/qor/l10n"
	"github.com/qor/qor/publish"
)

type Book struct {
	gorm.Model
	l10n.Locale
	publish.Status
	Name        string
	CategoryID  uint
	Category    Category
	PublisherID uint
	Publisher   Publisher
	Comments    []Comment
	Authors     []Author `gorm:"many2many:author_books;ForeignKey:ID;AssociationForeignKey:ID"`
}

type Publisher struct {
	gorm.Model
	publish.Status
	Name string
}

type Comment struct {
	gorm.Model
	l10n.Locale
	publish.Status
	Content string
	BookID  uint
}

type Author struct {
	gorm.Model
	l10n.Locale
	Name string
}

func generateBook(name string) *Book {
	book := Book{
		Name: name,
		Category: Category{
			Name: name + "_category",
		},
		Publisher: Publisher{
			Name: name + "_publisher",
		},
		Comments: []Comment{
			{Content: name + "_comment1"},
			{Content: name + "_comment2"},
		},
		Authors: []Author{
			{Name: name + "_author1"},
			{Name: name + "_author2"},
		},
	}
	return &book
}

func TestBelongsToForL10nResource(t *testing.T) {
	name := "belongs_to_for_l10n"
	book := generateBook(name)
	pbdraft.Save(book)

	pb.Publish(book)

	if pbprod.Where("id = ?", book.ID).First(&Book{}).RecordNotFound() {
		t.Errorf("should find book from production db")
	}

	if pbprod.Where("name LIKE ?", name+"%").First(&Publisher{}).RecordNotFound() {
		t.Errorf("should find publisher from production db")
	}

	if pbprod.Where("name LIKE ?", name+"%").First(&Category{}).RecordNotFound() {
		t.Errorf("should find category from production db")
	}
}

func TestMany2ManyForL10nResource(t *testing.T) {
	name := "many2many_for_l10n"
	book := generateBook(name)
	pbdraft.Save(book)

	if pbdraft.Model(book).Association("Authors").Count() != 2 {
		t.Errorf("should find two authors from draft db before publish")
	}

	if pbprod.Model(book).Association("Authors").Count() != 0 {
		t.Errorf("should find none author from production db before publish")
	}

	pb.Publish(book)

	if pbprod.Where("id = ?", book.ID).First(&Book{}).RecordNotFound() {
		t.Errorf("should find book from production db")
	}

	if pbdraft.Model(book).Association("Authors").Count() != 2 {
		t.Errorf("should find two authors from draft db after publish")
	}

	if pbprod.Model(book).Association("Authors").Count() != 2 {
		t.Errorf("should find two authors from draft db after publish")
	}
}
