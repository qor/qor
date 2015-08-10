package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/qor/qor/responder"
	"github.com/qor/qor/roles"
	"github.com/qor/qor/validations"
)

type controller struct {
	*Admin
}

const HTTPUnprocessableEntity = 422

func (context *Context) checkResourcePermission(permission roles.PermissionMode) bool {
	if context.Resource == nil || context.Resource.HasPermission(permission, context.Context) {
		return true
	}
	context.Writer.Write([]byte("Permission denied"))
	return false
}

func (ac *controller) Dashboard(context *Context) {
	context.Execute("dashboard", nil)
}

func (ac *controller) Index(context *Context) {
	if context.checkResourcePermission(roles.Read) {
		if result, err := context.FindMany(); err == nil {
			responder.With("html", func() {
				context.Execute("index", result)
			}).With("json", func() {
				res := context.Resource
				js, _ := json.Marshal(res.convertObjectToMap(context, result, "index"))
				context.Writer.Write(js)
			}).Respond(context.Writer, context.Request)
		} else {
			http.NotFound(context.Writer, context.Request)
		}
	}
}

func (ac *controller) Show(context *Context) {
	if context.checkResourcePermission(roles.Read) {
		result, _ := context.FindOne()

		responder.With("html", func() {
			context.Execute("show", result)
		}).With("json", func() {
			res := context.Resource
			js, _ := json.Marshal(res.convertObjectToMap(context, result, "show"))
			context.Writer.Write(js)
		}).Respond(context.Writer, context.Request)
	}
}

func (ac *controller) New(context *Context) {
	if context.checkResourcePermission(roles.Create) {
		context.Execute("new", context.Resource.NewStruct())
	}
}

func (ac *controller) Create(context *Context) {
	if context.checkResourcePermission(roles.Create) {
		res := context.Resource

		result := res.NewStruct()
		if errs := res.Decode(context.Context, result); len(errs) == 0 {
			err := res.CallSaver(result, context.Context)
			responder.With("html", func() {
				if err == nil {
					context.Flash(context.dt("resource_successfully_created", "{{.Name}} was successfully created", res), "success")
					primaryKey := fmt.Sprintf("%v", context.GetDB().NewScope(result).PrimaryKeyValue())
					http.Redirect(context.Writer, context.Request, path.Join(context.Request.URL.Path, primaryKey), http.StatusFound)
				} else {
					context.Execute("new", result)
				}
			}).With("json", func() {
				if err == nil {
					js, _ := json.Marshal(context.Resource.convertObjectToMap(context, result, "show"))
					context.Writer.Write(js)
				} else {
					context.Writer.WriteHeader(HTTPUnprocessableEntity)
					data, _ := json.Marshal(map[string]interface{}{"errors": validations.GetErrors(context.GetDB())})
					context.Writer.Write(data)
				}
			}).Respond(context.Writer, context.Request)
		}
	}
}

func (ac *controller) Update(context *Context) {
	if context.checkResourcePermission(roles.Update) {
		res := context.Resource
		if result, err := context.FindOne(); err == nil {
			if errs := res.Decode(context.Context, result); len(errs) == 0 {
				err := res.CallSaver(result, context.Context)
				responder.With("html", func() {
					if err == nil {
						context.FlashNow(context.dt("resource_successfully_updated", "{{.Name}} was successfully updated", res), "success")
					}
					context.Execute("show", result)
				}).With("json", func() {
					if err == nil {
						js, _ := json.Marshal(context.Resource.convertObjectToMap(context, result, "show"))
						context.Writer.Write(js)
					} else {
						context.Writer.WriteHeader(HTTPUnprocessableEntity)
						data, _ := json.Marshal(map[string]interface{}{"errors": validations.GetErrors(context.GetDB())})
						context.Writer.Write(data)
					}
				}).Respond(context.Writer, context.Request)
			}
		}
	}
}

func (ac *controller) Delete(context *Context) {
	if context.checkResourcePermission(roles.Delete) {
		res := context.Resource
		status := http.StatusOK
		if err := res.CallDeleter(res.NewStruct(), context.Context); err != nil {
			status = http.StatusNotFound
		}

		responder.With("html", func() {
			http.Redirect(context.Writer, context.Request, path.Join(ac.GetRouter().Prefix, res.ToParam()), status)
		}).With("json", func() {
			context.Writer.WriteHeader(status)
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
