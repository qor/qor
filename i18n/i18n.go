package i18n

import (
	"os"
	"path"
	"strings"

	"github.com/qor/qor/admin"
)

type I18n struct {
	Default string
}

func (i18n *I18n) InjectQorAdmin(res *admin.Resource) {
	res.Config.Theme = "i18n"

	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/qor/i18n/views"))
	}
}
