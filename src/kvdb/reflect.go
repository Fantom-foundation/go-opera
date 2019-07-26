package kvdb

import (
	"reflect"
)

// MigrateTables sets tagget fields to database tables.
func MigrateTables(s interface{}, db Database) {
	value := reflect.ValueOf(s).Elem()
	for i := 0; i < value.NumField(); i++ {
		if prefix := value.Type().Field(i).Tag.Get("table"); prefix != "" && prefix != "-" {
			field := value.Field(i)
			var val reflect.Value
			if db != nil {
				table := db.NewTable([]byte(prefix))
				val = reflect.ValueOf(table)
			} else {
				val = reflect.Zero(field.Type())
			}
			field.Set(val)
		}
	}
}

// MigrateCaches sets tagget fields to get() result.
func MigrateCaches(c interface{}, get func() interface{}) {
	value := reflect.ValueOf(c).Elem()
	for i := 0; i < value.NumField(); i++ {
		if prefix := value.Type().Field(i).Tag.Get("cache"); prefix != "" {
			field := value.Field(i)
			var cache interface{}
			if get != nil {
				cache = get()
			}
			var val reflect.Value
			if cache != nil {
				val = reflect.ValueOf(cache)
			} else {
				val = reflect.Zero(field.Type())
			}
			field.Set(val)
		}
	}
}
