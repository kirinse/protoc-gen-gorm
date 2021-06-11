package plugin

import (
	"fmt"
	"strings"

	gorm "github.com/kirinse/protoc-gen-gorm/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

func ormIdent(ident protogen.GoIdent) protogen.GoIdent {
	ident.GoName += "ORM"
	return ident
}

// IsAbleToMakePQArray tells us if the specific field-type can automatically be turned into a PQ array:
func (p *OrmPlugin) IsAbleToMakePQArray(fieldType string) bool {
	switch fieldType {
	case "[]bool", "[]float32", "[]float64", "[]int32", "[]int64", "[]string":
		return true
	default:
		return false
	}
}

// IsAbleToMakePQArray tells us if the specific field-type can automatically be turned into a PQ array:
func (p *OrmPlugin) fieldToPQArrayIdent(field *protogen.Field) (i protogen.GoIdent, t string, err error) {
	fieldType := p.fieldType(field)
	switch fieldType {
	case "[]bool":
		return identpqBoolArray, "bool[]", nil
	case "[]float32":
		return identpqFloat32Array, "float[]", nil
	case "[]float64":
		return identpqFloat64Array, "float[]", nil
	case "[]int32":
		return identpqInt32Array, "integer[]", nil
	case "[]int64":
		return identpqInt64Array, "integer[]", nil
	case "[]string":
		return identpqStringArray, "text[]", nil
	default:
		return protogen.GoIdent{}, "", fmt.Errorf("invalid fieldtype")
	}
}

func (p *OrmPlugin) qualifiedGoIdent(ident protogen.GoIdent) string {
	isPointer := strings.Contains(ident.GoName, "*")
	isList := strings.Contains(ident.GoName, "[]")
	var result = ""
	tmpIdent := protogen.GoIdent{
		GoName:       strings.TrimLeft(ident.GoName, "*[]"),
		GoImportPath: ident.GoImportPath,
	}
	if ident.GoImportPath != "" {
		result = p.currentFile.QualifiedGoIdent(tmpIdent)
	} else {
		result = tmpIdent.GoName
	}
	if isPointer {
		result = "*" + result
	}
	if isList {
		result = "[]" + result
	}
	return result
}

func (p *OrmPlugin) qualifiedGoIdentPtr(ident protogen.GoIdent) string {
	return fmt.Sprintf("*%s", p.qualifiedGoIdent(ident))
}

func (p *OrmPlugin) identFnCall(funcName protogen.GoIdent, args ...string) string {
	return p.fnCall(p.qualifiedGoIdent(funcName), args...)
}

func (p *OrmPlugin) fnCall(funcName string, args ...string) string {
	return fmt.Sprint(funcName, `(`+strings.Join(args, ",")+`)`)
}

// retrieves the GormMessageOptions from a message
func getMessageOptions(message *protogen.Message) *gorm.GormMessageOptions {
	if message.Desc.Options() == nil {
		return nil
	}
	v := proto.GetExtension(message.Desc.Options(), gorm.E_Opts)
	opts, ok := v.(*gorm.GormMessageOptions)
	if !ok {
		return nil
	}
	return opts
}

func getFieldOptions(field *protogen.Field) *gorm.GormFieldOptions {
	if field.Desc.Options() == nil {
		return nil
	}
	v := proto.GetExtension(field.Desc.Options(), gorm.E_Field)
	opts, ok := v.(*gorm.GormFieldOptions)
	if !ok {
		return nil
	}
	return opts
}

func getServiceOptions(service *protogen.Service) *gorm.AutoServerOptions {
	if service.Desc.Options() == nil {
		return nil
	}
	v := proto.GetExtension(service.Desc.Options(), gorm.E_Server)
	opts, ok := v.(*gorm.AutoServerOptions)
	if !ok {
		return nil
	}
	return opts
}

func getMethodOptions(method *protogen.Method) *gorm.MethodOptions {
	if method.Desc.Options() == nil {
		return nil
	}
	v := proto.GetExtension(method.Desc.Options(), gorm.E_Method)
	opts, ok := v.(*gorm.MethodOptions)
	if !ok {
		return nil
	}
	return opts
}

func (p *OrmPlugin) isSpecialType(field *protogen.Field) bool {
	var ident protogen.GoIdent
	if field.Message != nil {
		ident = field.Message.GoIdent
	} else {
		ident = field.GoIdent
	}
	if p.currentPackage == ident.GoImportPath {
		return false
	}

	_, specialPkg := specialImports[string(ident.GoImportPath)]
	if !specialPkg {
		return false
	}

	typeName := p.fieldType(field)
	switch typeName {
	case protoTypeJSON,
		protoTypeUUID,
		protoTypeUUIDValue,
		protoTypeResource,
		protoTypeInet,
		protoTimeOnly,
		protoTypeTimestamp:
		return true
	}
	return false
}

func (p *OrmPlugin) fieldType(field *protogen.Field) string {
	var tmp string
	switch {
	case field.Desc.Message() == nil:
		tmp = protoPrimitiveKinds[field.Desc.Kind()]
	case field.Message != nil:
		tmp = p.messageType(field.Message)
	default:
		tmp = string(field.Desc.Message().Name())
	}
	if _, ok := builtinTypes[tmp]; ok && field.Desc.IsList() {
		tmp = "[]" + tmp
	}
	return tmp
}

func fieldName(field *protogen.Field) string {
	return string(field.GoName)
}

func fieldIdent(field *protogen.Field) protogen.GoIdent {
	if field.Enum != nil {
		return field.Enum.GoIdent
	}
	if field.Message != nil {
		return messageIdent(field.Message)
	}
	return field.GoIdent
}

func (p *OrmPlugin) messageType(message *protogen.Message) string {
	return string(message.Desc.Name())
}

func messageName(message *protogen.Message) string {
	return message.GoIdent.GoName
}

func messageIdent(message *protogen.Message) protogen.GoIdent {
	return message.GoIdent
}
