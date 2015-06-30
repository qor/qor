package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
