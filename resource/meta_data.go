package resource

type MetaData struct {
	Name      string
	Value     interface{}
	MetaDatas MetaDatas
	Meta      Metaor
}

type MetaDatas []*MetaData

func (mds MetaDatas) Get(name string) *MetaData {
	for _, md := range mds {
		if md.Name == name {
			return md
		}
	}

	return nil
}
