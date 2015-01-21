package publish

import (
	"os"
	"path"
	"strings"

	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"
)

func (db *DB) PreviewAction(context *admin.Context) {
	draftDB := db.DraftMode()
	drafts := make(map[*resource.Resource]interface{})
	for _, model := range db.SupportedModels {
		var res *resource.Resource
		var name = modelType(model).Name()

		if r := context.Admin.GetResource(strings.ToLower(name)); r != nil {
			res = &r.Resource
		} else {
			res = resource.New(model)
		}

		results := res.NewSlice()
		draftDB.Unscoped().Where("publish_status = ?", DIRTY).Find(results)
		drafts[res] = results
	}
	context.Execute("publish/drafts", drafts)
}

func (db *DB) PublishAction(context *admin.Context) {
}

func (db *DB) InjectQorAdmin(web *admin.Admin) {
	router := web.GetRouter()
	router.Get("/publish", db.PreviewAction)
	router.Post("/publish", db.PublishAction)

	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/qor/publish/views"))
	}
}
