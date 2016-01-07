package admin

import (
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
			http.Redirect(context.Writer, context.Request, context.UrlFor(result, res), http.StatusFound)
		}).With("json", func() {
			context.JSON("show", result)
		}).Respond(context.Request)
	}
}

func (ac *controller) Show(context *Context) {
	var result interface{}
	var err error

	// If singleton Resource
	if context.Resource.Config.Singleton {
		result = context.Resource.NewStruct()
		if err = context.Resource.CallFindMany(result, context.Context); err == gorm.RecordNotFound {
			context.Execute("new", result)
			return
		}
	} else {
		result, err = context.FindOne()
	}
	context.AddError(err)

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
	var result interface{}
	var err error

	// If singleton Resource
	if context.Resource.Config.Singleton {
		result = context.Resource.NewStruct()
		context.Resource.CallFindMany(result, context.Context)
	} else {
		result, err = context.FindOne()
		context.AddError(err)
	}

	res := context.Resource
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
	paths := strings.Split(context.Request.URL.Path, "/")
	name := paths[len(paths)-1]

	for _, action := range context.Resource.actions {
		if action.Name == name {
			err = action.Handle(&ActionArgument{
				IDs:     context.Request.Form["IDs[]"],
				Context: context,
			})
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

func (ac *controller) ActionForm(context *Context) {
	paths := strings.Split(context.Request.URL.Path, "/")
	name := paths[len(paths)-2]

	var action *Action
	for _, act := range context.Resource.actions {
		if act.Name == name {
			action = act
		}
	}

	responder.With("html", func() {
		context.Execute("actions/shared/form", action)
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
