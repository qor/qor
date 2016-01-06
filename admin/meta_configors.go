package admin

import (
	"time"

	"github.com/qor/qor"
)

var metaConfigorMaps = map[string]func(*Meta){
	"date": func(meta *Meta) {
		if meta.FormattedValuer == nil {
			meta.SetFormattedValuer(func(value interface{}, context *qor.Context) interface{} {
				switch date := meta.GetValuer()(value, context).(type) {
				case time.Time, *time.Time:
					return date.Format("2006-01-02")
				case **time.Time:
					if *date == nil {
						return ""
					}
					return (*date).Format("2006-01-02")
				default:
					return date
				}
			})
		}
	},

	"datetime": func(meta *Meta) {
		if meta.FormattedValuer == nil {
			meta.SetFormattedValuer(func(value interface{}, context *qor.Context) interface{} {
				switch date := meta.GetValuer()(value, context).(type) {
				case time.Time, *time.Time:
					return date.Format("2006-01-02 15:04")
				case **time.Time:
					if *date == nil {
						return ""
					}
					return (*date).Format("2006-01-02 15:04")
				default:
					return date
				}
			})
		}
	},

	"string": func(meta *Meta) {
		if meta.FormattedValuer == nil {
			meta.SetFormattedValuer(func(value interface{}, context *qor.Context) interface{} {
				if str := meta.GetValuer()(value, context).(type); str == nil {
					return ""
				} else {
					return str
				}
			})
		}
	},

	"text": func(meta *Meta) {
		if meta.FormattedValuer == nil {
			meta.SetFormattedValuer(func(value interface{}, context *qor.Context) interface{} {
				if str := meta.GetValuer()(value, context).(type); str == nil {
					return ""
				} else {
					return str
				}
			})
		}
	},
}
