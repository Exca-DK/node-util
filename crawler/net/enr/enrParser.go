package enr

import (
	"reflect"

	"github.com/ethereum/go-ethereum/p2p/enr"
)

func RetrieveFieldNames(enr *enr.Record) []string {
	val := reflect.ValueOf(enr).Elem()
	rpairs := val.FieldByName("pairs")

	names := make([]string, 0)
	switch rpairs.Kind() {
	case reflect.Slice:
		for i := 0; i < rpairs.Len(); i++ {
			rs := rpairs.Index(i)
			names = append(names, rs.Field(0).String())
		}
	}
	return names
}
