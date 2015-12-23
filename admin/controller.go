package admin

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/responder"
)

type controller struct {
	*Admin
}

const HTTPUnprocessableEntity = 422

func (ac *controller) Dashboard(context *Context) {
	context.Execute("dashboard", nil)
}

func (ac *controller) Index(context *Context) {
	result, err := context.FindMany()
	context.AddError(err)

	if context.HasError() {
		http.NotFound(context.Writer, context.Request)
	} else {
		responder.With("html", func() {
			context.Execute("index", result)
		}).With("json", func() {
			context.JSON("index", result)
		}).Respond(context.Request)
	}
}

func (ac *controller) SearchCenter(context *Context) {
	type searchResult struct {
		Context  *Context
		Resource *Resource
		Results  interface{}
	}
	var searchResults []searchResult
	for _, res := range context.Admin.searchResources {
		resourceName := context.Request.URL.Query().Get("resource_name")
		if resourceName == "" || res.ToParam() == resourceName {
			ctx := context.clone().setResource(res)
			if results, err := ctx.FindMany(); err == nil {
				searchResults = append(searchResults, searchResult{
					Context:  ctx,
					Resource: res,
					Results:  results,
				})
			}
		}
	}
	context.Execute("search_center", searchResults)
}

func (ac *controller) New(context *Context) {
	context.Execute("new", context.Resource.NewStruct())
}

func (ac *controller) Create(context *Context) {
	res := context.Resource
	result := res.NewStruct()
	if context.AddError(res.Decode(context.Context, result)); !context.HasError() {
		context.AddError(res.CallSave(result, context.Context))
	}

	if context.HasError() {
		responder.With("html", func() {
			context.Writer.WriteHeader(HTTPUnprocessableEntity)
			context.Execute("new", result)
		}).With("json", func() {
			context.Writer.WriteHeader(HTTPUnprocessableEntity)
			context.JSON("index", map[string]interface{}{"errors": context.GetErrors()})
		}).Respond(context.Request)
	} else {
		responder.With("html", func() {
			context.Flash(string(context.dt("resource_successfully_created", "{{.Name}} was successfully created", res)), "success")
			http.Redirect(context.Writer, context.Request, context.showResourcePath(result, res), http.StatusFound)
		}).With("json", func() {
			context.JSON("show", result)
		}).Respond(context.Request)
	}
}

func (ac *controller) Show(context *Context) {
	result, err := context.FindOne()

	// Singleton Resource & Failed to find record
	if context.Resource.Config.Singleton && err == gorm.RecordNotFound {
		context.Execute("new", result)
		return
	}

	responder.With("html", func() {
		context.Execute("show", result)
	}).With("json", func() {
		context.JSON("show", result)
	}).Respond(context.Request)
}

func (ac *controller) Edit(context *Context) {
	result, err := context.FindOne()
	context.AddError(err)

	responder.With("html", func() {
		context.Execute("edit", result)
	}).With("json", func() {
		context.JSON("edit", result)
	}).Respond(context.Request)
}

func (ac *controller) Update(context *Context) {
	res := context.Resource
	result, err := context.FindOne()
	context.AddError(err)
	if !context.HasError() {
		if context.AddError(res.Decode(context.Context, result)); !context.HasError() {
			context.AddError(res.CallSave(result, context.Context))
		}
	}

	if context.HasError() {
		context.Writer.WriteHeader(HTTPUnprocessableEntity)
		responder.With("html", func() {
			context.Execute("edit", result)
		}).With("json", func() {
			context.JSON("edit", map[string]interface{}{"errors": context.GetErrors()})
		}).Respond(context.Request)
	} else {
		responder.With("html", func() {
			context.FlashNow(string(context.dt("resource_successfully_updated", "{{.Name}} was successfully updated", res)), "success")
			context.Execute("show", result)
		}).With("json", func() {
			context.JSON("show", result)
		}).Respond(context.Request)
	}
}

func (ac *controller) Delete(context *Context) {
	res := context.Resource
	status := http.StatusOK
	if context.AddError(res.CallDelete(res.NewStruct(), context.Context)); context.HasError() {
		status = http.StatusNotFound
	}

	responder.With("html", func() {
		http.Redirect(context.Writer, context.Request, path.Join(ac.GetRouter().Prefix, res.ToParam()), status)
	}).With("json", func() {
		context.Writer.WriteHeader(status)
	}).Respond(context.Request)
}

func (ac *controller) Action(context *Context) {
	var err error
	name := strings.Split(context.Request.URL.Path, "/")[4]

	for _, action := range context.Resource.actions {
		if action.Name == name {
			ids := context.Request.Form.Get("ids")
			scope := context.GetDB().Where(fmt.Sprintf("%v IN (?)", context.Resource.PrimaryField().DBName), ids)
			err = action.Handle(scope, context.Context)
		}
	}

	responder.With("html", func() {
		if err == nil {
			http.Redirect(context.Writer, context.Request, context.Request.Referer(), http.StatusFound)
		} else {
			context.Writer.WriteHeader(HTTPUnprocessableEntity)
			context.Writer.Write([]byte(err.Error()))
		}
	}).With("json", func() {
		if err == nil {
			context.Writer.Write([]byte("OK"))
		} else {
			context.Writer.WriteHeader(HTTPUnprocessableEntity)
			context.Writer.Write([]byte(err.Error()))
		}
	}).Respond(context.Request)
}

func (ac *controller) Asset(context *Context) {
	file := strings.TrimPrefix(context.Request.URL.Path, ac.GetRouter().Prefix)
	if filename, err := context.findFile(file); err == nil {
		http.ServeFile(context.Writer, context.Request, filename)
	} else {
		http.NotFound(context.Writer, context.Request)
	}
}
