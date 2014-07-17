package admin

import (
	"bytes"
	"fmt"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/rules"
	"io"
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

func (content *Content) AllowedMetas(modes ...rules.PermissionMode) func() []resource.Meta {
	return func() []resource.Meta {
		var attrs []resource.Meta
		switch content.Action {
		case "index":
			attrs = content.Resource.IndexAttrs()
		case "show":
			attrs = content.Resource.ShowAttrs()
		case "edit":
			attrs = content.Resource.EditAttrs()
		case "new":
			attrs = content.Resource.NewAttrs()
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
	for _, meta := range metas {
		content.RenderMeta(result, meta, value)
	}
	return result.String()
}

func (content *Content) RenderMeta(writer io.Writer, meta resource.Meta, value interface{}) {
	var tmpl *template.Template
	tmpl = content.getTemplate(tmpl, "forms/string.tmpl")

	data := map[string]string{}
	data["InputId"] = strings.Join([]string{"QorResource", meta.Name}, "")
	data["Label"] = meta.Label
	data["InputName"] = fmt.Sprintf("QorResource[%v]", meta.Name)
	data["Value"] = fmt.Sprintf("%v", meta.GetValue(value, content.Context))

	if err := tmpl.Execute(writer, data); err != nil {
		fmt.Println(err)
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
