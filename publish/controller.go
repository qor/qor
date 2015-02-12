package publish

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/qor/qor/admin"
)

type PublishController struct {
	*DB
}

func (db *PublishController) Preview(context *admin.Context) {
	draftDB := db.DraftMode()
	drafts := make(map[*admin.Resource]interface{})
	for _, model := range db.SupportedModels {
		var name = modelType(model).Name()
		if res := context.Admin.GetResource(strings.ToLower(name)); res != nil {
			results := res.NewSlice()
			if draftDB.Unscoped().Where("publish_status = ?", DIRTY).Find(results).RowsAffected > 0 {
				drafts[res] = results
			}
		}
	}
	context.Execute("publish/drafts", drafts)
}

func (db *PublishController) Diff(context *admin.Context) {
	resourceID := strings.Split(context.Request.URL.Path, "/")[4]
	params := strings.Split(resourceID, "__")
	name, id := params[0], params[1]
	res := context.Admin.GetResource(name)

	draft := res.NewStruct()
	db.DraftMode().Unscoped().First(draft, id)

	production := res.NewStruct()
	db.ProductionMode().Unscoped().First(production, id)

	results := map[string]interface{}{"Production": production, "Draft": draft, "Resource": res}

	fmt.Fprintf(context.Writer, context.Render("publish/diff", results))
}

func (db *PublishController) Publish(context *admin.Context) {
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

	for name, value := range values {
		res := context.Admin.GetResource(name)
		results := res.NewSlice()
		if db.DraftMode().Unscoped().Find(results, fmt.Sprintf("%v IN (?)", res.PrimaryFieldDBName()), value).Error == nil {
			resultValues := reflect.Indirect(reflect.ValueOf(results))
			for i := 0; i < resultValues.Len(); i++ {
				records = append(records, resultValues.Index(i).Interface())
			}
		}
	}

	if request.Form.Get("publish_type") == "publish" {
		db.DB.Publish(records...)
	} else if request.Form.Get("publish_type") == "discard" {
		db.DB.Discard(records...)
	}
	http.Redirect(context.Writer, context.Request, context.Request.RequestURI, http.StatusFound)
}

func (db *DB) InjectQorAdmin(web *admin.Admin) {
	controller := PublishController{db}
	router := web.GetRouter()
	router.Get("^/publish/diff/", controller.Diff)
	router.Get("^/publish", controller.Preview)
	router.Post("^/publish", controller.Publish)

	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/qor/publish/views"))
	}
}
