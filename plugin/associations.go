package plugin

import (
	"fmt"
	jgorm "github.com/jinzhu/gorm"
	"github.com/jinzhu/inflection"
	"github.com/kirinse/atlas-app-toolkit/util/cases"
	gorm "github.com/kirinse/protoc-gen-gorm/options"
	"google.golang.org/protobuf/compiler/protogen"
	"strings"
)

func (p *OrmPlugin) parseAssociations(msg *protogen.Message) {
	typeName := p.messageType(msg)
	ormable := p.getOrmable(typeName)

	for _, field := range msg.Fields {
		fieldOpts := getFieldOptions(field)
		if fieldOpts.GetDrop() {
			continue
		}
		fieldName := fieldName(field)
		// tmp := protodesc.ToFieldDescriptorProto(field)
		var fieldType string
		if field.Desc.Message() != nil {
			tmp := string(field.Desc.Message().Name())
			parts := strings.Split(tmp, ".")
			fieldType = parts[len(parts)-1]
		} else {
			fieldType = field.Desc.Kind().String()
		}
		fieldType = strings.Trim(fieldType, "[]*")

		parts := strings.Split(fieldType, ".")
		fieldTypeShort := parts[len(parts)-1]
		if p.isOrmable(fieldType) {
			if fieldOpts == nil {
				fieldOpts = &gorm.GormFieldOptions{}
			}
			assocOrmable := p.getOrmable(fieldType)
			if field.Desc.IsList() {
				if fieldOpts.GetManyToMany() != nil {
					p.parseManyToMany(msg, ormable, fieldName, fieldTypeShort, field, assocOrmable, fieldOpts)

				} else {
					p.parseHasMany(msg, ormable, fieldName, fieldTypeShort, field, assocOrmable, fieldOpts)
				}
				fieldType = fmt.Sprintf("[]*%sORM", fieldType)
			} else if fieldOpts.GetTag().GetEmbedded() {
				fieldType = fmt.Sprintf("*%sORM", fieldType)
			} else {
				if fieldOpts.GetBelongsTo() != nil {
					p.parseBelongsTo(msg, ormable, fieldName, fieldTypeShort, field, assocOrmable, fieldOpts)
				} else {
					p.parseHasOne(msg, ormable, fieldName, fieldTypeShort, field, assocOrmable, fieldOpts)
				}
				fieldType = fmt.Sprintf("*%sORM", fieldType)
			}
			field.GoIdent.GoName = fieldType
			field.GoIdent.GoImportPath = assocOrmable.File.GoImportPath
			ormable.Fields[fieldName] = &Field{Type: fieldType, GormFieldOptions: fieldOpts, F: field}
		}
	}
}

func (p *OrmPlugin) parseHasMany(msg *protogen.Message, parent *OrmableType, fieldName string, fieldType string, field *protogen.Field, child *OrmableType, opts *gorm.GormFieldOptions) {
	typeName := msg.GoIdent.GoName
	hasMany := opts.GetHasMany()
	if hasMany == nil {
		hasMany = &gorm.HasManyOptions{}
		opts.Association = &gorm.GormFieldOptions_HasMany{HasMany: hasMany}
	}
	assocKeyName, assocKey := p.getAssocKeyName(hasMany, parent)
	hasMany.AssociationForeignkey = &assocKeyName
	foreignKey := p.getForeignKey(hasMany, assocKey, child, field)
	var foreignKeyName string
	if foreignKeyName = hasMany.GetForeignkey(); foreignKeyName == "" {
		if p.countHasAssociationDimension(msg, fieldType) == 1 {
			foreignKeyName = fmt.Sprintf(typeName + assocKeyName)
		} else {
			foreignKeyName = fmt.Sprintf(fieldName + typeName + assocKeyName)
		}
	}
	if _, ok := child.Fields[foreignKeyName]; child.Package != parent.Package && !ok {
		p.Fail(`Object`, child.Name, `from package`, child.Package, `cannot be used for has-many in`, parent.Name, `since it`,
			`does not have FK`, foreignKeyName, `defined. Manually define the key, or switch to many-to-many`)
	}
	hasMany.Foreignkey = &foreignKeyName
	p.setChildForeignKeyFieldExternal(child, parent, foreignKey, foreignKeyName)

	var posField string
	if posField = cases.GoCamelCase(hasMany.GetPositionField()); posField != "" {
		if exField, ok := child.Fields[posField]; !ok {
			child.Fields[posField] = &Field{Type: "int", GormFieldOptions: &gorm.GormFieldOptions{Tag: hasMany.GetPositionFieldTag()}}
		} else if !strings.Contains(exField.Type, "int") {
			p.Fail("Cannot include", posField, "field into", child.Name, "as it already exists there with a different type.")
		}
		hasMany.PositionField = &posField
	}
}

func (p *OrmPlugin) parseHasOne(msg *protogen.Message, parent *OrmableType, fieldName string, fieldType string, field *protogen.Field, child *OrmableType, opts *gorm.GormFieldOptions) {
	typeName := msg.GoIdent.GoName
	hasOne := opts.GetHasOne()
	if hasOne == nil {
		hasOne = &gorm.HasOneOptions{}
		opts.Association = &gorm.GormFieldOptions_HasOne{HasOne: hasOne}
	}
	assocKeyName, assocKey := p.getAssocKeyName(hasOne, parent)
	hasOne.AssociationForeignkey = &assocKeyName
	foreignKey := p.getForeignKey(hasOne, assocKey, child, field)
	foreignKeyName := cases.GoCamelCase(hasOne.GetForeignkey())
	if foreignKeyName == "" {
		if p.countHasAssociationDimension(msg, fieldType) == 1 {
			foreignKeyName = fmt.Sprintf(typeName + assocKeyName)
		} else {
			foreignKeyName = fmt.Sprintf(fieldName + typeName + assocKeyName)
		}
	}
	if _, ok := child.Fields[foreignKeyName]; child.Package != parent.Package && !ok {
		p.Fail(`Object`, child.Name, `from package`, child.Package, `cannot be used for has-one in`, parent.Name, `since it`,
			`does not have FK field`, foreignKeyName, `defined. Manually define the key, or switch to belongs-to`)
	}
	hasOne.Foreignkey = &foreignKeyName
	p.setChildForeignKeyFieldExternal(child, parent, foreignKey, foreignKeyName)
}

func (p *OrmPlugin) parseBelongsTo(msg *protogen.Message, child *OrmableType, fieldName string, fieldType string, field *protogen.Field, parent *OrmableType, opts *gorm.GormFieldOptions) {
	belongsTo := opts.GetBelongsTo()
	if belongsTo == nil {
		belongsTo = &gorm.BelongsToOptions{}
		opts.Association = &gorm.GormFieldOptions_BelongsTo{BelongsTo: belongsTo}
	}
	assocKeyName, assocKey := p.getAssocKeyName(belongsTo, parent)
	belongsTo.AssociationForeignkey = &assocKeyName
	foreignKey := p.getForeignKey(belongsTo, assocKey, child, field)
	foreignKeyName := cases.GoCamelCase(belongsTo.GetForeignkey())
	if foreignKeyName == "" {
		if p.countBelongsToAssociationDimension(msg, fieldType) == 1 {
			foreignKeyName = fmt.Sprintf(fieldType + assocKeyName)
		} else {
			foreignKeyName = fmt.Sprintf(fieldName + assocKeyName)
		}
	}
	belongsTo.Foreignkey = &foreignKeyName
	p.setChildForeignKeyFieldExternal(child, parent, foreignKey, foreignKeyName)
}

func (p *OrmPlugin) parseManyToMany(msg *protogen.Message, ormable *OrmableType, fieldName string, fieldType string, field *protogen.Field, assoc *OrmableType, opts *gorm.GormFieldOptions) {
	typeName := p.messageType(msg)
	mtm := opts.GetManyToMany()
	if mtm == nil {
		mtm = &gorm.ManyToManyOptions{}
		opts.Association = &gorm.GormFieldOptions_ManyToMany{ManyToMany: mtm}
	}
	var foreignKeyName string
	if foreignKeyName = cases.GoCamelCase(mtm.GetForeignkey()); foreignKeyName == "" {
		foreignKeyName, _ = p.findPrimaryKey(ormable)
	} else {
		var ok bool
		_, ok = ormable.Fields[foreignKeyName]
		if !ok {
			p.Fail("Missing", foreignKeyName, "field in", ormable.Name, ".")
		}
	}
	mtm.Foreignkey = &foreignKeyName
	assocKeyName, _ := p.getAssocKeyName(mtm, assoc)
	mtm.AssociationForeignkey = &assocKeyName
	var jt string
	if jt = jgorm.ToDBName(mtm.GetJointable()); jt == "" {
		if p.countManyToManyAssociationDimension(msg, fieldType) == 1 && typeName != fieldType {
			jt = jgorm.ToDBName(typeName + inflection.Plural(fieldType))
		} else {
			jt = jgorm.ToDBName(typeName + inflection.Plural(fieldName))
		}
	}
	mtm.Jointable = &jt
	var jtForeignKey string
	if jtForeignKey = cases.GoCamelCase(mtm.GetJointableForeignkey()); jtForeignKey == "" {
		jtForeignKey = jgorm.ToDBName(typeName + foreignKeyName)
	}
	mtm.JointableForeignkey = &jtForeignKey
	var jtAssocForeignKey string
	if jtAssocForeignKey = cases.GoCamelCase(mtm.GetAssociationJointableForeignkey()); jtAssocForeignKey == "" {
		if typeName == fieldType {
			jtAssocForeignKey = jgorm.ToDBName(inflection.Singular(fieldName) + assocKeyName)
		} else {
			jtAssocForeignKey = jgorm.ToDBName(fieldType + assocKeyName)
		}
	}
	mtm.AssociationJointableForeignkey = &jtAssocForeignKey
}

type assocForeignKeyGetter interface {
	GetAssociationForeignkey() string
}

func (p *OrmPlugin) getAssocKeyName(i assocForeignKeyGetter, parent *OrmableType) (string, *Field) {
	assocKeyName := cases.GoCamelCase(i.GetAssociationForeignkey())
	if assocKeyName == "" {
		return p.findPrimaryKey(parent)
	}
	assocKey, ok := parent.Fields[assocKeyName]
	if !ok {
		p.Fail("Missing", assocKeyName, "field in", parent.Name, ".")
	}
	return assocKeyName, assocKey
}

type foreignKeyTagGetter interface {
	GetForeignkeyTag() *gorm.GormTag
}

func (p *OrmPlugin) getForeignKey(i foreignKeyTagGetter, assocKey *Field, child *OrmableType, field *protogen.Field) *Field {
	var foreignKeyType string
	if i.GetForeignkeyTag().GetNotNull() {
		foreignKeyType = strings.TrimPrefix(assocKey.F.GoIdent.GoName, "*")
	} else if strings.HasPrefix(assocKey.Type, "*") {
		foreignKeyType = assocKey.F.GoIdent.GoName
	} else if strings.Contains(assocKey.Type, "[]byte") {
		foreignKeyType = assocKey.F.GoIdent.GoName
	} else {
		foreignKeyType = "*" + assocKey.Type
	}
	copiedF := *field
	copiedF.GoIdent = protogen.GoIdent{
		GoName:       foreignKeyType,
		GoImportPath: assocKey.F.GoIdent.GoImportPath,
	}
	foreignKey := &Field{Type: foreignKeyType, F: &copiedF, Package: string(field.GoIdent.GoImportPath), GormFieldOptions: &gorm.GormFieldOptions{Tag: i.GetForeignkeyTag()}}
	return foreignKey
}

func (p *OrmPlugin) setChildForeignKeyFieldExternal(child *OrmableType, parent *OrmableType, foreignKey *Field, foreignKeyName string) {
	if exField, ok := child.Fields[foreignKeyName]; !ok {
		child.Fields[foreignKeyName] = foreignKey
	} else if exField.Type == "interface{}" {
		exField.Type = foreignKey.Type
		exField.F.GoIdent.GoName = foreignKey.F.GoIdent.GoName
	} else if !p.sameType(exField, foreignKey) {
		p.Fail("Cannot include", foreignKeyName, "field into", child.Name, "as it already exists there with a different type:", exField.Type, foreignKey.Type)
	}
	child.Fields[foreignKeyName].ParentOriginName = parent.OriginName
}

func (p *OrmPlugin) findPrimaryKeyHelper(ormable *OrmableType) (bool, string, *Field) {
	for fieldName, field := range ormable.Fields {
		if field.GetTag().GetPrimaryKey() {
			return true, fieldName, field
		}
	}

	for fieldName, field := range ormable.Fields {
		if strings.ToLower(fieldName) == "id" {
			return true, fieldName, field
		}
	}

	// get PrimaryKey from embedded
	//for _, field := range ormable.Fields {
	//	fmt.Println("##### try to get PrimaryKey from embedded:", field.GetTag().GetEmbedded())
	//	if field.GetTag().GetEmbedded() {
	//		typeName := field.Type
	//		embedOrmable := p.getOrmable(typeName)
	//		return p.findPrimaryKeyHelper(embedOrmable)
	//	}
	//}
	return false, "", nil
}

func (p *OrmPlugin) findPrimaryKey(ormable *OrmableType) (string, *Field) {
	found, a, b := p.findPrimaryKeyHelper(ormable)
	if !found {
		p.Fail("Primary key cannot be found in", ormable.Name, ".")
	}
	return a, b
}

func (p *OrmPlugin) hasPrimaryKey(ormable *OrmableType) bool {
	found, _, _ := p.findPrimaryKeyHelper(ormable)
	return found
}

func (p *OrmPlugin) countDimensionGeneric(msg *protogen.Message, typeName string, conditional func(fieldOpts *gorm.GormFieldOptions) bool) int {
	dim := 0
	for _, field := range msg.Fields {
		fieldOpts := getFieldOptions(field)
		if fieldOpts.GetDrop() {
			continue
		}
		fieldType := p.fieldType(field)
		if conditional(fieldOpts) == true {
			if strings.Trim(typeName, "[]*") == strings.Trim(fieldType, "[]*") {
				dim++
			}
		}
	}
	return dim
}

func (p *OrmPlugin) countHasAssociationDimension(msg *protogen.Message, typeName string) int {
	return p.countDimensionGeneric(msg, typeName, func(opts *gorm.GormFieldOptions) bool {
		return opts.GetManyToMany() == nil && opts.GetBelongsTo() == nil
	})
}

func (p *OrmPlugin) countBelongsToAssociationDimension(msg *protogen.Message, typeName string) int {
	return p.countDimensionGeneric(msg, typeName, func(opts *gorm.GormFieldOptions) bool {
		return opts.GetBelongsTo() != nil
	})
}

func (p *OrmPlugin) countManyToManyAssociationDimension(msg *protogen.Message, typeName string) int {
	return p.countDimensionGeneric(msg, typeName, func(opts *gorm.GormFieldOptions) bool {
		return opts.GetManyToMany() != nil
	})
}

func (p *OrmPlugin) sameType(field1 *Field, field2 *Field) bool {
	isPointer1 := strings.HasPrefix(field1.Type, "*")
	typeParts1 := strings.Split(field1.Type, ".")
	if len(typeParts1) == 2 {
		isPointer2 := strings.HasPrefix(field2.Type, "*")
		typeParts2 := strings.Split(field2.Type, ".")
		if len(typeParts2) == 2 && isPointer1 == isPointer2 && typeParts1[1] == typeParts2[1] && field1.Package == field2.Package {
			return true
		}
		return false
	}
	return field1.Type == field2.Type
}
