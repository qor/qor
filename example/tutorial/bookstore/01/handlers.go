package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// books
func listBooksHandler(ctx *gin.Context) {
	var books []*Book

	if err := db.Find(&books).Error; err != nil {
		panic(err)
	}

	ctx.HTML(
		http.StatusOK,
		"list.tmpl",
		gin.H{
			"title": "List of Books",
			"books": books,
		},
	)
}

func viewBookHandler(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Params.ByName("id"), 10, 64)
	if err != nil {
		panic(err)
	}
	var book = &Book{}
	if err := db.Find(&book, id).Error; err != nil {
		panic(err)
	}

	if err := db.Model(&book).Related(&book.Authors, "Authors").Error; err != nil {
		panic(err)
	}

	ctx.HTML(
		http.StatusOK,
		"book.tmpl",
		gin.H{
			"book": book,
		},
	)
}

// simple login - username only/no passwords yet
// TODO: switch this to gin too?
func loginHandler(writer http.ResponseWriter, request *http.Request) {
	var user User

	if request.Method == "POST" {
		request.ParseForm()
		if !db.First(&user, "name = ?", request.Form.Get("username")).RecordNotFound() {
			cookie := http.Cookie{Name: "userid", Value: fmt.Sprintf("%v", user.ID), Expires: time.Now().AddDate(1, 0, 0)}
			http.SetCookie(writer, &cookie)
			writer.Write([]byte("<html><body>logged in as `" + user.Name + "`, go to <a href='/admin'>admin</a></body></html>"))
		} else {
			http.Redirect(writer, request, "/login?failed_to_login", 301)
		}
	} else if userid, err := request.Cookie("userid"); err == nil {
		if !db.First(&user, "id = ?", userid.Value).RecordNotFound() {
			writer.Write([]byte("<html><body>already logged as `" + user.Name + "`, go <a href='/admin'>admin</a></body></html>"))
		} else {
			http.Redirect(writer, request, "/logout", http.StatusSeeOther)
		}
	} else {
		writer.Write([]byte(`<html><form action="/login" method="POST"><input name="username" value="" placeholder="username"><input type=submit value="Login"></form></html>`))
	}
}

func logoutHandler(writer http.ResponseWriter, request *http.Request) {
	cookie := http.Cookie{Name: "userid", MaxAge: -1}
	http.SetCookie(writer, &cookie)
	http.Redirect(writer, request, "/login?logged_out", http.StatusSeeOther)
}
