package publish

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/qor/qor/admin"
	"github.com/qor/qor/utils"
)

type publishController struct {
	*Publish
}

func (db *publishController) Preview(context *admin.Context) {
	drafts := make(map[*admin.Resource]interface{})
	draftDB := context.GetDB().Set("publish:draft_mode", true).Unscoped()
	for _, res := range context.Admin.GetResources() {
		results := res.NewSlice()
		if isPublishableModel(res.Value) {
			if draftDB.Where("publish_status = ?", DIRTY).Find(results).RowsAffected > 0 {
				drafts[res] = results
			}
		}
	}
	context.Execute("publish/drafts", drafts)
}

func (db *publishController) Diff(context *admin.Context) {
	resourceID := strings.Split(context.Request.URL.Path, "/")[4]
	params := strings.Split(resourceID, "__")
	name, id := params[0], params[1]
	res := context.Admin.GetResource(name)

	draft := res.NewStruct()
	context.GetDB().Set("publish:draft_mode", true).Unscoped().First(draft, id)

	production := res.NewStruct()
	context.GetDB().Set("publish:draft_mode", false).Unscoped().First(production, id)

	results := map[string]interface{}{"Production": production, "Draft": draft, "Resource": res}

	fmt.Fprintf(context.Writer, string(context.Render("publish/diff", results)))
}

func (db *publishController) PublishOrDiscard(context *admin.Context) {
	var request = context.Request
	var ids = request.Form["checked_ids[]"]
	var records = []interface{}{}
	var values = map[string][]string{}

	for _, id := range ids {
		if keys := strings.Split(id, "__"); len(keys) == 2 {
			name, id := keys[0], keys[1]
			values[name] = append(values[name], id)
		}
	}

	draftDB := context.GetDB().Set("publish:draft_mode", true).Unscoped()
	for name, value := range values {
		res := context.Admin.GetResource(name)
		results := res.NewSlice()
		if draftDB.Find(results, fmt.Sprintf("%v IN (?)", res.PrimaryDBName()), value).Error == nil {
			resultValues := reflect.Indirect(reflect.ValueOf(results))
			for i := 0; i < resultValues.Len(); i++ {
				records = append(records, resultValues.Index(i).Interface())
			}
		}
	}

	if request.Form.Get("publish_type") == "publish" {
		Publish{DB: draftDB}.Publish(records...)
	} else if request.Form.Get("publish_type") == "discard" {
		Publish{DB: draftDB}.Discard(records...)
	}
	http.Redirect(context.Writer, context.Request, context.Request.RequestURI, http.StatusFound)
}

var injected bool

func (publish *Publish) InjectQorAdmin(res *admin.Resource) {
	if !injected {
		injected = true
		for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
			admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/qor/publish/views"))
		}
	}
	res.UseTheme("publish")

	controller := publishController{publish}
	router := res.GetAdmin().GetRouter()
	router.Get(fmt.Sprintf("^/%v/diff/", res.ToParam()), controller.Diff)
	router.Get(fmt.Sprintf("^/%v", res.ToParam()), controller.Preview)
	router.Post(fmt.Sprintf("^/%v", res.ToParam()), controller.PublishOrDiscard)

	res.GetAdmin().RegisterFuncMap("render_publish_meta", func(value interface{}, meta *admin.Meta, context *admin.Context) template.HTML {
		var err error
		var result = bytes.NewBufferString("")
		var tmpl = template.New(meta.Type + ".tmpl").Funcs(context.FuncMap())

		if tmpl, err = context.FindTemplate(tmpl, fmt.Sprintf("metas/publish/%v.tmpl", meta.Type)); err != nil {
			if tmpl, err = context.FindTemplate(tmpl, fmt.Sprintf("metas/index/%v.tmpl", meta.Type)); err != nil {
				tmpl, _ = tmpl.Parse("{{.Value}}")
			}
		}

		data := map[string]interface{}{"Value": context.ValueOf(value, meta), "Meta": meta}
		if err := tmpl.Execute(result, data); err != nil {
			utils.ExitWithMsg(err.Error())
		}
		return template.HTML(result.String())
	})
}
