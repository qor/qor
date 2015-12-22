package publish

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"
)

type publishController struct {
	*Publish
}

type visiblePublishResourceInterface interface {
	VisiblePublishResource() bool
}

func (db *publishController) Preview(context *admin.Context) {
	type resource struct {
		*admin.Resource
		Value interface{}
	}

	var drafts = []resource{}

	draftDB := context.GetDB().Set(publishDraftMode, true).Unscoped()
	for _, res := range context.Admin.GetResources() {
		if visibleInterface, ok := res.Value.(visiblePublishResourceInterface); ok {
			if !visibleInterface.VisiblePublishResource() {
				continue
			}
		} else if res.Config.Invisible {
			continue
		}

		results := res.NewSlice()
		if IsPublishableModel(res.Value) || IsPublishEvent(res.Value) {
			if draftDB.Unscoped().Where("publish_status = ?", DIRTY).Find(results).RowsAffected > 0 {
				drafts = append(drafts, resource{
					Resource: res,
					Value:    results,
				})
			}
		}
	}
	context.Execute("publish_drafts", drafts)
}

func (db *publishController) Diff(context *admin.Context) {
	resourceID := context.Request.URL.Query().Get(":publish_unique_key")
	params := strings.Split(resourceID, "__")
	name, id := params[0], params[1]
	res := context.Admin.GetResource(name)

	draft := res.NewStruct()
	context.GetDB().Set(publishDraftMode, true).Unscoped().First(draft, id)

	production := res.NewStruct()
	context.GetDB().Set(publishDraftMode, false).Unscoped().First(production, id)

	results := map[string]interface{}{"Production": production, "Draft": draft, "Resource": res}

	fmt.Fprintf(context.Writer, string(context.Render("publish_diff", results)))
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

	draftDB := context.GetDB().Set(publishDraftMode, true).Unscoped()
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

func (publish *Publish) ConfigureQorResource(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
			admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/qor/publish/views"))
		}
		res.UseTheme("publish")

		if event := res.GetAdmin().GetResource("PublishEvent"); event == nil {
			eventResource := res.GetAdmin().AddResource(&PublishEvent{}, &admin.Config{Invisible: true})
			eventResource.IndexAttrs("Name", "Description", "CreatedAt")
		}

		controller := publishController{publish}
		router := res.GetAdmin().GetRouter()
		router.Get(fmt.Sprintf("/%v/diff/:publish_unique_key", res.ToParam()), controller.Diff)
		router.Get(res.ToParam(), controller.Preview)
		router.Post(res.ToParam(), controller.PublishOrDiscard)

		res.GetAdmin().RegisterFuncMap("publish_unique_key", func(res *admin.Resource, record interface{}, context *admin.Context) string {
			return fmt.Sprintf("%s__%v", res.ToParam(), context.GetDB().NewScope(record).PrimaryKeyValue())
		})

		res.GetAdmin().RegisterFuncMap("is_publish_event_resource", func(res *admin.Resource) bool {
			return IsPublishEvent(res.Value)
		})
	}
}
