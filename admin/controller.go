package admin

import (
	"database/sql"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/rules"
	"net/http"
	"strconv"
	"strings"

	"reflect"
)

func (admin *Admin) Dashboard(context *qor.Context) {
	content := Content{Admin: admin, Context: context, Action: "dashboard"}
	admin.Render("dashboard", content, rules.Read)
}

func (admin *Admin) Index(context *qor.Context) {
	resource := admin.Resources[context.ResourceName]
	sliceType := reflect.SliceOf(reflect.Indirect(reflect.ValueOf(resource.Model)).Type())
	slice := reflect.MakeSlice(sliceType, 0, 0)
	slicePtr := reflect.New(sliceType)
	slicePtr.Elem().Set(slice)
	result := slicePtr.Interface()
	admin.DB.Find(result)

	content := Content{Admin: admin, Context: context, Resource: resource, Result: result, Action: "index"}
	admin.Render("index", content, rules.Read)
}

func (admin *Admin) Show(context *qor.Context) {
	resource := admin.Resources[context.ResourceName]
	result := reflect.New(reflect.Indirect(reflect.ValueOf(resource.Model)).Type()).Interface()
	admin.DB.First(result, context.ResourceID)

	content := Content{Admin: admin, Context: context, Resource: resource, Result: result, Action: "show"}
	admin.Render("show", content, rules.Read, rules.Update)
}

func (admin *Admin) New(context *qor.Context) {
	resource := admin.Resources[context.ResourceName]
	content := Content{Admin: admin, Context: context, Resource: resource, Action: "new"}
	admin.Render("new", content, rules.Create)
}

func (admin *Admin) Create(context *qor.Context) {
}

func (admin *Admin) Update(context *qor.Context) {
	context.Request.ParseMultipartForm(32 << 22)

	var metas = []resource.Meta{}
	Resource := admin.Resources[context.ResourceName]
	attrs := Resource.EditAttrs()
	for _, meta := range attrs {
		if meta.HasPermission(rules.Update, context) {
			metas = append(metas, meta)
		}
	}

	result := reflect.New(reflect.Indirect(reflect.ValueOf(Resource.Model)).Type()).Interface()

	if !admin.DB.First(result, context.ResourceID).RecordNotFound() {
		for key, values := range context.Request.Form {
			value := values[0]

			if keys := strings.Split(key, "."); len(keys) >= 2 && keys[0] == "QorResource" {
				for _, meta := range metas {
					if meta.Name == keys[1] {
						// FIXME set value
						field := reflect.Indirect(reflect.ValueOf(result)).FieldByName(meta.Name)
						if field.IsValid() && field.CanAddr() {
							switch field.Kind() {
							case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
								if int, err := strconv.Atoi(value); err != nil {
									field.SetInt(reflect.ValueOf(int).Int())
								}
							default:
								if scanner, ok := field.Addr().Interface().(sql.Scanner); ok {
									scanner.Scan(value)
								} else {
									field.Set(reflect.ValueOf(value))
								}
							}
						}
						break
					}
				}
			}
		}
		admin.DB.Save(result)
		http.Redirect(context.Writer, context.Request, context.Request.RequestURI, http.StatusFound)
	}
}

func (admin *Admin) Delete(context *qor.Context) {
}
