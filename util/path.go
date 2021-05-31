package util

import "reflect"

func PackagePath(i interface{}) string {
	return reflect.TypeOf(i).PkgPath()
}
