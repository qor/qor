package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/qor/qor/example/tutorial/bookstore/01/app/controllers"
	. "github.com/qor/qor/example/tutorial/bookstore/01/app/resources"
)

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
		bookRoutes.GET("", controllers.ListBooksHandler)
		bookRoutes.GET("/", controllers.ListBooksHandler) // really? i need both of those?...
		// single book - product page
		bookRoutes.GET("/:id", controllers.ViewBookHandler)
	}

	mux.Handle("/", router)

	// handle login and logout of users
	mux.HandleFunc("/login", controllers.LoginHandler)
	mux.HandleFunc("/logout", controllers.LogoutHandler)

	// start the server
	http.ListenAndServe(":9000", mux)
}
