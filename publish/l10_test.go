package publish_test

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/l10n"
	"github.com/qor/qor/publish"
)

type Book struct {
	gorm.Model
	l10n.Locale
	publish.Publish
	Name       string
	CategoryID uint
	Category   Category
	Publisher  Publisher
	Comments   []Comment
	Authors    []Author `gorm:"many2many:author_books"`
}

type Publisher struct {
	gorm.Model
	l10n.Locale
	publish.Publish
	Name string
}

type Comment struct {
	gorm.Model
	l10n.Locale
	publish.Publish
	Content string
	BookID  uint
}

type Author struct {
	gorm.Model
	l10n.Locale
	publish.Publish
	Name string
}

func TestPublishL10nRecords(t *testing.T) {
	book := Book{
		Name: "l10n-book1",
		Category: Category{
			Name: "l10n-category1",
		},
		Publisher: Publisher{
			Name: "l10n-publisher",
		},
		Comments: []Comment{
			{Content: "l10n-content1"},
			{Content: "l10n-content2"},
		},
		Authors: []Author{
			{Name: "l10n-author1"},
			{Name: "l10n-author2"},
		},
	}

	pbdraft.Debug().Save(&book)
}
