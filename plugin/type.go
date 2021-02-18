package plugin

import (
	"google.golang.org/protobuf/compiler/protogen"
)

type fieldType struct {
	name  string
	ident protogen.GoIdent
}

// func newFieldTypeFromField(f *protogen.Field) fieldType {
// 	var name string
// 	if f.Message == nil {

// 	}
// 	return fieldType{
// 		name: f.
// 	}
// }
