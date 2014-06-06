package resource

type attrs struct {
	index []string
	new   []string
	edit  []string
	show  []string
}

func (a *attrs) Index(columns ...string) {
}

func (a *attrs) New(columns ...string) {
}

func (a *attrs) Edit(columns ...string) {
}

func (a *attrs) Show(columns ...string) {
}
