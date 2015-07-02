package resources

import (
	"fmt"
	"log"

	"github.com/qor/qor"
	"github.com/qor/qor/admin"

	. "github.com/qor/qor/example/tutorial/bookstore/01/app/models"
)

var (
	Admin *admin.Admin
)

func init() {
	// setting up QOR admin
	// Admin := admin.New(&qor.Config{DB: &db})
	Admin = admin.New(&qor.Config{DB: Pub.DraftDB()})
	Admin.AddResource(Pub)

	Admin.AddResource(
		&User{},
		&admin.Config{
			Menu: []string{"User Management"},
			Name: "Users",
		},
	)

	author := Admin.AddResource(
		&Author{},
		&admin.Config{Menu: []string{
			"Author Management"},
			Name: "Author",
		},
	)

	author.IndexAttrs("ID", "Name")

	book := Admin.AddResource(
		&Book{},
		&admin.Config{
			Menu: []string{"Book Management"},
			Name: "Books",
		},
	)

	// alternate price display
	book.Meta(&admin.Meta{
		Name:  "DisplayPrice",
		Label: "Price",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			book := value.(*Book)
			return fmt.Sprintf("¥%v", book.Price)
		},
	})

	book.Meta(&admin.Meta{
		Name:  "FormattedDate",
		Label: "Release Date",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			book := value.(*Book)
			return book.ReleaseDate.Format("Jan 2, 2006")
		},
	})

	// defines the display field for authors in the product list
	book.Meta(&admin.Meta{
		Name:  "AuthorNames",
		Label: "Authors",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			// log.Println("LOCALE:", context.MustGet("locale"))
			log.Println("ctxt", context)
			if value == nil {
				return value
			}
			book := value.(*Book)
			if err := Db.Model(&book).Related(&book.Authors, "Authors").Error; err != nil {
				panic(err)
			}

			var authors string
			for i, author := range book.Authors {
				if i >= 1 {
					authors += ", "
				}
				authors += author.Name
			}
			return authors
		},
	})

	book.Meta(&admin.Meta{
		Name: "Synopsis",
		Type: "rich_editor",
	})

	// book.Meta(&admin.Meta{
	// 	Name:  "Authors",
	// 	Label: "Authors",
	// 	Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
	// 		if authors := []Author{}; !context.GetDB().Find(&authors).RecordNotFound() {
	// 			for _, author := range authors {
	// 				results = append(results, []string{fmt.Sprintf("%v", author.ID), author.Name})
	// 			}
	// 		}
	// 		return
	// 	},
	// })

	// what fields should be displayed in the books list on admin
	book.IndexAttrs("ID", "Title", "AuthorNames", "FormattedDate", "DisplayPrice")
	// what fields should be editable in the book esit interface
	book.EditAttrs("Title", "Authors", "Synopsis", "ReleaseDate", "Price", "CoverImage")
	book.SearchAttrs("ID", "Title")
}

type Auth struct{}

func (Auth) LoginURL(c *admin.Context) string {
	return "/login"
}

func (Auth) LogoutURL(c *admin.Context) string {
	return "/logout"
}

func (Auth) GetCurrentUser(c *admin.Context) qor.CurrentUser {
	if userid, err := c.Request.Cookie("userid"); err == nil {
		var user User
		if !Db.First(&user, "id = ?", userid.Value).RecordNotFound() {
			return &user
		}
	}
	return nil
}
