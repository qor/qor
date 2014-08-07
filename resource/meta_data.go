package resource

type MetaData struct {
	Name  string
	Value interface{}
	Metaor
}

type MetaDatas []MetaData

func (metaDatas *MetaDatas) Get(name string) Metaor {
	for _, metaData := range *metaDatas {
		if metaData.Name == name {
			return metaData.GetMeta()
		}
	}
	return nil
}
