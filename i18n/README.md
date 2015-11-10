# I18n

I18n package support translations with different backends, like database, YAML.

## Usage

```go
import (
  "github.com/jinzhu/gorm"
  "github.com/qor/qor/i18n"
  "github.com/qor/qor/i18n/backends/database"
)

func main() {
  db, err := gorm.Open("mysql", "user:password@/dbname?charset=utf8&parseTime=True&loc=Local")

  // Using two backends, early backend has higher priority
  I18n = i18n.New(database.New(&db), yaml.New(filepath.Join(config.Root, "config/locales")))

  // Add Translation
  I18n.AddTranslation(&i18n.Translation{Key: "hello-world", Locale: "zh-CN", Value: "Hello World"})

  // Update Translation
  I18n.SaveTranslation(&i18n.Translation{Key: "hello-world", Locale: "zh-CN", Value: "Hello World"})

  // Delete Translation
  I18n.DeleteTranslation(&i18n.Translation{Key: "hello-world", Locale: "zh-CN", Value: "Hello World"})

  // Read transation with key `hello-world`
  I18n.T("zh-CN", "hello-world")

  // Read transation with `Scope`
  I18n.Scope("home-page").T("zh-CN", "hello-world") // read translation with translation key `home-page.hello-world`

  // Read transation with `Default Value`
  I18n.Scope("home-page").Default("Hello World").T("zh-CN", "not-existing") // Will return default value `Hello World`
}
```

## Advanced Usage

```go
// Using on frontend - you could define a T method then using in template
// <h2>{{T "home_page.how_it_works" "HOW DOES IT WORK? {{$1}}" "It is work" }}</h2>
func T(key string, value string, args ...interface{}) string {
	return config.Config.I18n.Default(value).T("en-US", key, args)
}

// Interpolation - i18n using golang template to parse translations with interpolation variable

I18n.AddTranslation(&i18n.Translation{Key: "hello", Locale: "en-US", Value: "Hello {{.Name}}"})
type User struct {
  Name string
}
I18n.T("en-US", "hello", User{Name: "jinzhu"}) //=> Hello jinzhu

// Pluralization - i18n is using [cldr](https://github.com/theplant/cldr) to do the job, it provide functions `p`, `zero`, `one`, `two`, `few`, `many`, `other` for pluralization, refer it for more details

I18n.AddTranslation(&i18n.Translation{Key: "count", Locale: "en-US", Value: "{{p "Count" (one "{{.Count}} item") (other "{{.Count}} items")}}"})
I18n.T("en-US", "count", map[string]int{"Count": 1}) //=> 1 item

// Ordered Params
I18n.AddTranslation(&i18n.Translation{Key: "ordered_params", Locale: "en-US", Value: "{{$1}} {{$2}} {{$1}}"})
I18n.T("en-US", "ordered_params", "string1", "string2") //=> string1 string2 string1
```
