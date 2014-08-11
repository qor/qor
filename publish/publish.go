package publish

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

type Publish struct {
	*gorm.DB
}

func (publish *Publish) Support(models ...interface{}) {
}

func (publish *Publish) ProductionMode() {
}

func (publish *Publish) DraftMode() {
}

func SetTable(scope *gorm.Scope) {
	tableName := scope.TableName()
	inDraft := true

	if inDraft {
		tableName = fmt.Sprintf("%v_draft", tableName)
		scope.Search.TableName = tableName
	}
}

// Qor::Publish.original_s3_bucket = "lacostedev"
// Qor::Publish.draft_s3_bucket    = "lacostedevdraft"

// Qor::Publish.ignore_models = [Cart, CartItem, Order, OrderItem, ::Qor::Job::Worker, FilterItem, Point]

// Qor::Publish.original_cache_store = YAML.load_file("#{Rails.root}/config/memcached.yml").unshift(:dalli_store)
// Qor::Publish.draft_cache_store = YAML.load_file("#{Rails.root}/config/memcached.yml").unshift(:dalli_store)
// Auto Migration
