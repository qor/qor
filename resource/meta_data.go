package resource

import "errors"

type MetaData struct {
	Name      string
	Value     interface{}
	MetaDatas MetaDatas
	Meta      Metaor
}

type MetaDatas []MetaData

func (metaDatas *MetaDatas) Get(name string) (MetaData, error) {
	for _, metaData := range *metaDatas {
		if metaData.Name == name {
			return metaData, nil
		}
	}
	return MetaData{}, errors.New("meta not found")
}
