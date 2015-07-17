package controllers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/qor/qor/example/tutorial/bookstore/01/app/models"
	"github.com/qor/qor/example/tutorial/bookstore/01/app/resources"
	"github.com/qor/qor/i18n"
)

// books
func ListBooksHandler(ctx *gin.Context) {
	var books []*Book

	if err := Db.Find(&books).Error; err != nil {
		panic(err)
	}
	for _, book := range books {
		if err := Db.Model(&book).Related(&book.Authors, "Authors").Error; err != nil {
			panic(err)
		}
	}

	ctx.HTML(
		http.StatusOK,
		"list.tmpl",
		gin.H{
			"books": books,
			"t": func(key string, args ...interface{}) template.HTML {
				return template.HTML(resources.I18n.T(retrieveLocale(ctx), key, args...))
			},
		},
	)
}

func retrieveLocale(ctx *gin.Context) string {
	if l := ctx.Query("locale"); l != "" {
		return l
	}

	return i18n.Default
}

func ViewBookHandler(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Params.ByName("id"), 10, 64)
	if err != nil {
		panic(err)
	}
	var book = &Book{}
	if err := Db.Find(&book, id).Error; err != nil {
		panic(err)
	}

	if err := Db.Model(&book).Related(&book.Authors, "Authors").Error; err != nil {
		panic(err)
	}

	ctx.HTML(
		http.StatusOK,
		"book.tmpl",
		gin.H{
			"book": book,
			"t": func(key string, args ...interface{}) template.HTML {
				return template.HTML(resources.I18n.T(retrieveLocale(ctx), key, args...))
			},
		},
	)
}

// simple login - username only/no passwords yet
// TODO: switch this to gin too?
func LoginHandler(writer http.ResponseWriter, request *http.Request) {
	var user User

	if request.Method == "POST" {
		request.ParseForm()
		if !Db.First(&user, "name = ?", request.Form.Get("username")).RecordNotFound() {
			cookie := http.Cookie{Name: "userid", Value: fmt.Sprintf("%v", user.ID), Expires: time.Now().AddDate(1, 0, 0)}
			http.SetCookie(writer, &cookie)
			writer.Write([]byte("<html><body>logged in as `" + user.Name + "`, go to <a href='/admin'>admin</a></body></html>"))
		} else {
			http.Redirect(writer, request, "/login?failed_to_login", 301)
		}
	} else if userid, err := request.Cookie("userid"); err == nil {
		if !Db.First(&user, "id = ?", userid.Value).RecordNotFound() {
			writer.Write([]byte("<html><body>already logged as `" + user.Name + "`, go <a href='/admin'>admin</a></body></html>"))
		} else {
			http.Redirect(writer, request, "/logout", http.StatusSeeOther)
		}
	} else {
		writer.Write([]byte(`<html><form action="/login" method="POST"><input name="username" value="" placeholder="username"><input type=submit value="Login"></form></html>`))
	}
}

func LogoutHandler(writer http.ResponseWriter, request *http.Request) {
	cookie := http.Cookie{Name: "userid", MaxAge: -1}
	http.SetCookie(writer, &cookie)
	http.Redirect(writer, request, "/login?logged_out", http.StatusSeeOther)
}
