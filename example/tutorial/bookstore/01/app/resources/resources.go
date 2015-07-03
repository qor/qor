package resources

import (
	"fmt"
	"log"
	"net/http"

	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/roles"

	. "github.com/qor/qor/example/tutorial/bookstore/01/app/models"
	"github.com/qor/qor/i18n"
	"github.com/qor/qor/i18n/backends/database"
)

var (
	Admin *admin.Admin
	I18n  *i18n.I18n
)

func init() {
	// setting up QOR admin
	// Admin := admin.New(&qor.Config{DB: &db})
	Admin = admin.New(&qor.Config{DB: Pub.DraftDB()})
	Admin.AddResource(Pub)
	Admin.SetAuth(&Auth{})

	I18n := i18n.New(database.New(StagingDB))
	// Admin.AddResource(I18n, &admin.Config{Name: "Translations", Menu: []string{"Site Management"}})
	Admin.AddResource(I18n)

	roles.Register("admin", func(req *http.Request, currentUser qor.CurrentUser) bool {
		if currentUser == nil {
			return false
		}

		if currentUser.(*User).Role == "admin" {
			return true
		}

		return false
	})

	roles.Register("user", func(req *http.Request, currentUser qor.CurrentUser) bool {
		if currentUser == nil {
			return false
		}

		if currentUser.(*User).Role == "user" {
			return true
		}

		return false
	})

	user := Admin.AddResource(
		&User{},
		&admin.Config{
			Menu: []string{"Users"},
			Name: "Users",
		},
	)

	user.Meta(&admin.Meta{
		Name:  "UserRole",
		Label: "Role",
		Type:  "select_one",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			return [][]string{
				{"admin", "admin"},
				{"user", "user"},
			}
		},
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			if value == nil {
				return value
			}
			user := value.(*User)
			log.Println("user", user.Role)
			return user.Role
		},
	})

	user.IndexAttrs("ID", "Name", "Role")
	user.EditAttrs("Name", "UserRole")

	author := Admin.AddResource(
		&Author{},
		&admin.Config{Menu: []string{
			"Authors"},
			Name: "Author",
		},
	)

	author.IndexAttrs("ID", "Name")
	author.SearchAttrs("ID", "Name")

	book := Admin.AddResource(
		&Book{},
		&admin.Config{
			Menu: []string{"Books"},
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
			// log.Println("ctxt", context)
			if value == nil {
				return value
			}
			book := value.(*Book)
			if err := context.GetDB().Model(&book).Related(&book.Authors, "Authors").Error; err != nil {
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
