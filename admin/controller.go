package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/qor/qor/responder"
	"github.com/qor/qor/roles"
)

type controller struct {
	*Admin
}

func renderError(context *Context, err error) {
	responder.With("html", func() {
		context.Writer.WriteHeader(http.StatusNotAcceptable)
		if _, er := context.Writer.Write([]byte(err.Error())); er != nil {
			println("failed to write response", er.Error())
		}
	}).With("json", func() {
		data, er := json.Marshal(map[string]string{"error": err.Error()})
		if er != nil {
			println("failed to marshal error json")
		}
		context.Writer.WriteHeader(http.StatusNotAcceptable)
		if _, er := context.Writer.Write(data); er != nil {
			println("failed to write reponse", er.Error())
		}
	}).Respond(context.Writer, context.Request)
}

func (context *Context) CheckResourcePermission(permission roles.PermissionMode) bool {
	if context.Resource.HasPermission(permission, context.Context) {
		return true
	}
	context.Writer.Write([]byte("Permission denied"))
	return false
}

func (ac *controller) Dashboard(context *Context) {
	context.Execute("dashboard", nil)
}

func (ac *controller) Index(context *Context) {
	if context.CheckResourcePermission(roles.Read) {
		if result, err := context.FindMany(); err == nil {
			responder.With("html", func() {
				context.Execute("index", result)
			}).With("json", func() {
				res := context.Resource
				js, _ := json.Marshal(res.ConvertObjectToMap(context, result, res.IndexMetas()))
				context.Writer.Write(js)
			}).Respond(context.Writer, context.Request)
		} else {
			http.NotFound(context.Writer, context.Request)
		}
	}
}

func (ac *controller) Show(context *Context) {
	if context.CheckResourcePermission(roles.Read) {
		result, _ := context.FindOne()

		responder.With("html", func() {
			context.Execute("show", result)
		}).With("json", func() {
			res := context.Resource
			js, _ := json.Marshal(res.ConvertObjectToMap(context, result, res.ShowMetas()))
			context.Writer.Write(js)
		}).Respond(context.Writer, context.Request)
	}
}

func (ac *controller) New(context *Context) {
	if context.CheckResourcePermission(roles.Create) {
		context.Execute("new", nil)
	}
}

func (ac *controller) Create(context *Context) {
	if context.CheckResourcePermission(roles.Create) {
		res := context.Resource

		result := res.NewStruct()
		if errs := res.Decode(context, result); len(errs) == 0 {
			res.CallSaver(result, context.Context)
			responder.With("html", func() {
				primaryKey := fmt.Sprintf("%v", context.GetDB().NewScope(result).PrimaryKeyValue())
				http.Redirect(context.Writer, context.Request, path.Join(context.Request.RequestURI, primaryKey), http.StatusFound)
			}).With("json", func() {
				res := context.Resource
				js, _ := json.Marshal(res.ConvertObjectToMap(context, result, res.ShowMetas()))
				context.Writer.Write(js)
			}).Respond(context.Writer, context.Request)
		}
	}
}

func (ac *controller) Update(context *Context) {
	if context.CheckResourcePermission(roles.Update) {
		if result, err := context.FindOne(); err == nil {
			if errs := context.Resource.Decode(context, result); len(errs) == 0 {
				if err := context.Resource.CallSaver(result, context.Context); err != nil {
					renderError(context, err)
					return
				}
				responder.With("html", func() {
					http.Redirect(context.Writer, context.Request, context.Request.RequestURI, http.StatusFound)
				}).With("json", func() {
					res := context.Resource
					js, _ := json.Marshal(res.ConvertObjectToMap(context, result, res.ShowMetas()))
					context.Writer.Write(js)
				}).Respond(context.Writer, context.Request)
			}
		} else {
			renderError(context, err)
		}
	}
}

func (ac *controller) Delete(context *Context) {
	if context.CheckResourcePermission(roles.Delete) {
		res := context.Resource

		responder.With("html", func() {
			if res.CallDeleter(res.NewStruct(), context.Context) == nil {
				http.Redirect(context.Writer, context.Request, path.Join(ac.GetRouter().Prefix, res.ToParam()), http.StatusFound)
			} else {
				http.Redirect(context.Writer, context.Request, path.Join(ac.GetRouter().Prefix, res.ToParam()), http.StatusNotFound)
			}
		}).With("json", func() {
			if res.CallDeleter(res.NewStruct(), context.Context) == nil {
				context.Writer.WriteHeader(http.StatusOK)
			} else {
				context.Writer.WriteHeader(http.StatusNotFound)
			}
		}).Respond(context.Writer, context.Request)
	}
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
		http.Redirect(context.Writer, context.Request, context.Request.Referer(), http.StatusFound)
	}).With("json", func() {
		if err == nil {
			context.Writer.Write([]byte("OK"))
		} else {
			context.Writer.Write([]byte(err.Error()))
		}
	}).Respond(context.Writer, context.Request)
}

func (ac *controller) Asset(context *Context) {
	file := strings.TrimPrefix(context.Request.URL.Path, ac.GetRouter().Prefix)
	if filename, err := context.findFile(file); err == nil {
		http.ServeFile(context.Writer, context.Request, filename)
	} else {
		http.NotFound(context.Writer, context.Request)
	}
}
