package plugin

import (
	"fmt"
	"strings"

	gorm "github.com/edhaight/protoc-gen-gorm/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

func ormIdent(ident protogen.GoIdent) protogen.GoIdent {
	ident.GoName += "ORM"
	return ident
}

func (p *OrmPlugin) qualifiedGoIdent(ident protogen.GoIdent) string {
	return p.currentFile.QualifiedGoIdent(ident)
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

func (p *OrmPlugin) isSpecialType(typeName string, ident protogen.GoIdent) bool {
	if p.currentPackage == string(ident.GoImportPath) {
		return false
	}

	// anything that looks like a google_protobufX should be considered special
	if strings.HasPrefix(strings.TrimLeft(typeName, "[]*"), "google_protobuf") {
		return true
	}
	switch typeName {
	case protoTypeJSON,
		protoTypeUUID,
		protoTypeUUIDValue,
		protoTypeResource,
		protoTypeInet,
		protoTimeOnly:
		return true
	}
	return false
}

func fieldType(field *protogen.Field) string {
	if field.Desc.Message() == nil {
		return field.Desc.Kind().String()
	}
	return string(field.Desc.Message().Name())
}

func fieldName(field *protogen.Field) string {
	return string(field.GoName)
}

func fieldIdent(field *protogen.Field) protogen.GoIdent {
	return field.GoIdent
}

func messageType(message *protogen.Message) string {
	return string(message.Desc.Name())
}

func messageName(message *protogen.Message) string {
	return message.GoIdent.GoName
}

func messageIdent(message *protogen.Message) protogen.GoIdent {
	return message.GoIdent
}
