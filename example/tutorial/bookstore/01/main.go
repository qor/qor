package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/qor/qor/l10n"

	"github.com/qor/qor/example/tutorial/bookstore/01/app/handlers"
	. "github.com/qor/qor/example/tutorial/bookstore/01/app/resources"
)

const (
	ENV_STAGING = iota
	ENV_PRODUCTION
)

func init() {
	l10n.Global = "en-US"
}

func main() {
	mux := http.NewServeMux()
	Admin.MountTo("/admin", mux)

	// frontend routes
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	// serve static files
	router.StaticFS("/system/", http.Dir("public/system"))
	router.StaticFS("/assets/", http.Dir("public/assets"))

	// books
	bookRoutes := router.Group("/books")
	{
		// listing
		bookRoutes.GET("", handlers.ListBooksHandler)
		bookRoutes.GET("/", handlers.ListBooksHandler) // really? i need both of those?...
		// single book - product page
		bookRoutes.GET("/:id", handlers.ViewBookHandler)
	}

	mux.Handle("/", router)

	// handle login and logout of users
	mux.HandleFunc("/login", handlers.LoginHandler)
	mux.HandleFunc("/logout", handlers.LogoutHandler)

	// start the server
	http.ListenAndServe(":9000", mux)
}
