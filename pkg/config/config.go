package config

import (
	"reflect"

	"github.com/ksysoev/deriv-api-bff/pkg/api"
	"github.com/ksysoev/deriv-api-bff/pkg/prov/deriv"
	"github.com/ksysoev/deriv-api-bff/pkg/repo"
)

type Config struct {
	Server api.Config       `mapstructure:"server"`
	Deriv  deriv.Config     `mapstructure:"deriv"`
	API    repo.CallsConfig `mapstructure:"api"`
	Etcd   repo.EtcdConfig  `mapstructure:"etcd"`
}

// TODO: add godoc
func Compare(_old, _new interface{}, path string) []string {
	var diffs []string

	oldMeta := reflect.ValueOf(_old)
	newMeta := reflect.ValueOf(_new)

	for i := 0; i < oldMeta.NumField(); i++ {
		oldField := oldMeta.Field(i)
		newField := newMeta.Field(i)
		fieldName := oldMeta.Type().Field(i).Name

		currentPath := path + "." + fieldName

		if oldField.Kind() == reflect.Struct {
			nestedDiffs := Compare(oldField.Interface(), newField.Interface(), currentPath)
			diffs = append(diffs, nestedDiffs...)
		} else if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
			diffs = append(diffs, currentPath)
		}
	}

	return diffs
}
