package publish

import "github.com/qor/qor/admin"

func (db *DB) PreviewAction(context *admin.Context) {
}

func (db *DB) PublishAction(context *admin.Context) {
}

func (db *DB) InjectQorAdmin(admin *admin.Admin) {
	router := admin.GetRouter()
	router.Get("/publish", db.PreviewAction)
	router.Post("/publish", db.PublishAction)
}
