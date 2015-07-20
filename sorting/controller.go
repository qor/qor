package sorting

import (
	"fmt"
	"strconv"

	"github.com/qor/qor/admin"
)

func updatePosition(context *admin.Context) {
	if result, err := context.FindOne(); err == nil {
		if position, ok := result.(positionInterface); ok {
			if pos, err := strconv.Atoi(context.Request.URL.Query().Get("pos")); err == nil {
				if MoveTo(context.GetDB(), position, pos) == nil {
					context.Writer.Write([]byte("OK"))
					return
				}
			}
		}
	}
	context.Writer.Write([]byte("Error"))
}

func (s *Sorting) InjectQorAdmin(res *admin.Resource) {
	Admin := res.GetAdmin()
	res.UseTheme("sorting")

	router := Admin.GetRouter()
	router.Get(fmt.Sprintf("^/%v/sorting/update_position?pos=\\d+$", res.ToParam()), updatePosition)
}
