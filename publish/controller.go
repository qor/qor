package publish

import (
	"os"
	"path"
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
		var res *admin.Resource
		var name = modelType(model).Name()

		if r := context.Admin.GetResource(strings.ToLower(name)); r != nil {
			res = r
		} else {
			res = admin.NewResource(model)
		}

		results := res.NewSlice()
		draftDB.Unscoped().Where("publish_status = ?", DIRTY).Find(results)
		drafts[res] = results
	}
	context.Execute("publish/drafts", drafts)
}

func (db *PublishController) Diff(context *admin.Context) {
	resourceID := strings.Split(context.Request.URL.Path, "/")[4]
	params := strings.Split(resourceID, "__")
	name, id := params[0], params[1]
	res := context.Admin.GetResource(name)

	draft := res.NewStruct()
	db.DraftMode().First(draft, id)

	production := res.NewStruct()
	db.ProductionMode().First(production, id)
}

func (db *PublishController) Publish(context *admin.Context) {
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
