package admin

import (
	"bytes"
	"fmt"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/rules"
	"path"
	"strings"

	"text/template"
)

type Content struct {
	Admin    *Admin
	Context  *qor.Context
	Resource *resource.Resource
	Result   interface{}
	Action   string
}

func (content *Content) AllowedMetas(modes ...rules.PermissionMode) func(reses ...*resource.Resource) []resource.Meta {
	return func(reses ...*resource.Resource) []resource.Meta {
		var res = content.Resource
		if len(reses) > 0 {
			res = reses[0]
		}

		var attrs []resource.Meta
		switch content.Action {
		case "index":
			attrs = res.IndexAttrs()
		case "show":
			attrs = res.ShowAttrs()
		case "edit":
			attrs = res.EditAttrs()
		case "new":
			attrs = res.NewAttrs()
		}

		var metas = []resource.Meta{}
		for _, meta := range attrs {
			for _, mode := range modes {
				if meta.HasPermission(mode, content.Context) {
					metas = append(metas, meta)
					break
				}
			}
		}
		return metas
	}
}

func (content *Content) ValueOf(value interface{}, meta resource.Meta) interface{} {
	return meta.GetValue(value, content.Context)
}

func (content *Content) UrlFor(value interface{}) string {
	var url string
	if admin, ok := value.(*Admin); ok {
		url = admin.Prefix
	} else if res, ok := value.(*resource.Resource); ok {
		url = path.Join(content.Admin.Prefix, res.RelativePath())
	} else {
		primaryKey := content.Admin.DB.NewScope(value).PrimaryKeyValue()
		url = path.Join(content.Admin.Prefix, content.Resource.RelativePath(), fmt.Sprintf("%v", primaryKey))
	}
	return url
}

func (content *Content) LinkTo(text interface{}, value interface{}) string {
	return fmt.Sprintf(`<a href="%v">%v</a>`, content.UrlFor(value), text)
}

func (content *Content) RenderForm(value interface{}, metas []resource.Meta) string {
	var result = bytes.NewBufferString("")
	content.renderForm(result, value, metas, []string{"QorResource"})
	return result.String()
}

func (content *Content) renderForm(result *bytes.Buffer, value interface{}, metas []resource.Meta, prefix []string) {
	for _, meta := range metas {
		content.RenderMeta(result, meta, value, prefix)
	}
}

func (content *Content) RenderMeta(writer *bytes.Buffer, meta resource.Meta, value interface{}, prefix []string) {
	var tmpl = template.New(meta.Type + ".tmpl").Funcs(content.funcMap(rules.Read, rules.Update))
	prefix = append(prefix, meta.Name)

	if tmpl, err := content.getTemplate(tmpl, fmt.Sprintf("forms/%v.tmpl", meta.Type)); err == nil {
		data := map[string]interface{}{}
		data["InputId"] = strings.Join(prefix, "")
		data["Label"] = meta.Label
		data["InputName"] = strings.Join(prefix, ".")
		data["Value"] = meta.GetValue(value, content.Context)
		data["Meta"] = meta

		// QorResource.Name => // jinzhu
		// QorResource.Role => // admin
		// QorResource.Address[0].Id -> if slice
		// QorResource.Address[0].Address1
		// QorResource.Address[0].Address2
		// QorResource.Address[1].Address1
		// QorResource.CreditCard.Number // if struct
		// AllowedMetas

		if err := tmpl.Execute(writer, data); err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Printf("Form type %v not supported\n", meta.Type)
	}
}

func (content *Content) funcMap(modes ...rules.PermissionMode) template.FuncMap {
	return template.FuncMap{
		"allowed_metas": content.AllowedMetas(modes...),
		"value_of":      content.ValueOf,
		"url_for":       content.UrlFor,
		"link_to":       content.LinkTo,
		"render_form":   content.RenderForm,
	}
}
