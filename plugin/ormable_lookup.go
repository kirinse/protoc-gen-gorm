package plugin

import (
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

// OrmableLookup is a helper map for tracking associations and relations between models.
type OrmableLookup map[string]*OrmableType

// TypeOk returns true if the parameter type is registered as ormable.
func (o *OrmableLookup) TypeOk(t string) bool {
	_, ok := (*o)[strings.Trim(t, "[]*")]
	return ok
}

// MessageOk returns true if the message's type is registered as ormable.
func (o *OrmableLookup) MessageOk(message *protogen.Message) bool {
	return o.TypeOk(messageName(message))
}

// GetOrmable returns the registered ormable object given the typename param exists.
func (o *OrmableLookup) GetOrmable(typeName string) *OrmableType {
	ormable, ok := (*o)[strings.TrimSuffix(strings.Trim(typeName, "[]*"), "ORM")]
	if !ok {
		return nil
	}
	return ormable
}

// GetOrmableMessage returns the registered ormable object given the message's typename param exists.
func (o *OrmableLookup) GetOrmableMessage(message *protogen.Message) *OrmableType {
	return o.GetOrmable(messageName(message))
}

// plugin adaptors
func (p *OrmPlugin) isOrmable(typeName string) bool {
	return p.ormableTypes.TypeOk(typeName)
}

func (p *OrmPlugin) isOrmableMessage(message *protogen.Message) bool {
	return p.ormableTypes.MessageOk(message)
}

func (p *OrmPlugin) getOrmable(typeName string) *OrmableType {
	ormable := p.ormableTypes.GetOrmable(typeName)
	if ormable == nil {
		p.Fail("%s is not ormable.", typeName)
	}
	return ormable
}

func (p *OrmPlugin) getOrmableMessage(message *protogen.Message) *OrmableType {
	return p.ormableTypes.GetOrmable(messageName(message))
}
