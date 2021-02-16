package plugin

import (
	"fmt"
	"strings"

	jgorm "github.com/jinzhu/gorm"
	"google.golang.org/protobuf/compiler/protogen"
)

func (p *OrmPlugin) generateDefaultHandlers(file *protogen.File) {
	for _, message := range file.Messages {
		if getMessageOptions(message).GetOrmable() {
			//context package is a global import because it used in function parameters
			p.UsingGoImports(stdCtxImport)

			p.generateCreateHandler(message)
			// FIXME: Temporary fix for Ormable objects that have no ID field but
			// have pk.

			if p.hasPrimaryKey(p.getOrmable(message.GoIdent.GoName)) && p.hasIDField(message) {
				p.generateReadHandler(message)
				p.generateDeleteHandler(message)
				p.generateDeleteSetHandler(message)
				p.generateStrictUpdateHandler(message)
				p.generatePatchHandler(message)
				p.generatePatchSetHandler(message)
			}

			p.generateApplyFieldMask(message)
			p.generateListHandler(message)
		}
	}
}

func (p *OrmPlugin) generateAccountIdWhereClause() {
	p.P(`accountID, err := `, identGetAccountIDFn, `(ctx, nil)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`db = db.Where(map[string]interface{}{"account_id": accountID})`)
}

func (p *OrmPlugin) generateBeforeHookDef(orm *OrmableType, method string) {
	p.P(`type `, orm.Name, `WithBefore`, method, ` interface {`)
	p.P(`Before`, method, `(`, identCtx, `, *`, identGormDB, `) (*`, identGormDB, `, error)`)
	p.P(`}`)
}

func (p *OrmPlugin) generateAfterHookDef(orm *OrmableType, method string) {
	p.P(`type `, orm.Name, `WithAfter`, method, ` interface {`)
	p.P(`After`, method, `(`, identCtx, `, *`, identGormDB, `) error`)
	p.P(`}`)
}

func (p *OrmPlugin) generateBeforeHookCall(orm *OrmableType, method string) {
	p.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `WithBefore`, method, `); ok {`)
	p.P(`if db, err = hook.Before`, method, `(ctx, db); err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generateAfterHookCall(orm *OrmableType, method string) {
	p.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `WithAfter`, method, `); ok {`)
	p.P(`if err = hook.After`, method, `(ctx, db); err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generateCreateHandler(message *protogen.Message) {
	typeName := message.GoIdent.GoName
	orm := p.getOrmable(typeName)
	p.P(`// DefaultCreate`, typeName, ` executes a basic gorm create call`)
	p.P(`func DefaultCreate`, typeName, `(ctx `, identCtx, `, in *`,
		typeName, `, db *`, identGormDB, `) (*`, typeName, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, `, identNilArgumentError)
	p.P(`}`)
	p.P(`ormObj, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	create := "Create_"
	p.generateBeforeHookCall(orm, create)
	p.P(`if err = db.Create(&ormObj).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.generateAfterHookCall(orm, create)
	p.P(`pbResponse, err := ormObj.ToPB(ctx)`)
	p.P(`return &pbResponse, err`)
	p.P(`}`)
	p.generateBeforeHookDef(orm, create)
	p.generateAfterHookDef(orm, create)
}

func (p *OrmPlugin) generateReadHandler(message *protogen.Message) {
	typeName := message.GoIdent.GoName
	ident := message.GoIdent
	ormable := p.getOrmable(typeName)
	p.P(`// DefaultRead`, ident, ` executes a basic gorm read call`)
	// Different behavior if there is a
	if p.readHasFieldSelection(ormable) {
		p.P(`func DefaultRead`, ident, `(ctx `, identCtx, `, in `,
			p.qualifiedGoIdentPtr(ident), `, db `, p.qualifiedGoIdentPtr(identGormDB), ` fs `, p.qualifiedGoIdentPtr(identQueryFieldSelection), `) (`, p.qualifiedGoIdentPtr(ident), `, error) {`)
	} else {
		p.P(`func DefaultRead`, ident, `(ctx `, identCtx, `, in `,
			p.qualifiedGoIdentPtr(ident), `, db `, p.qualifiedGoIdentPtr(identGormDB), `) (`, p.qualifiedGoIdentPtr(ident), `, error) {`)
	}
	p.P(`if in == nil {`)
	p.P(`return nil, `, identNilArgumentError)
	p.P(`}`)

	p.P(`ormObj, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	k, f := p.findPrimaryKey(ormable)
	if strings.Contains(f.Type, "*") {
		p.P(`if ormObj.`, k, ` == nil || *ormObj.`, k, ` == `, p.guessZeroValue(f.Type), ` {`)
	} else {
		p.P(`if ormObj.`, k, ` == `, p.guessZeroValue(f.Type), ` {`)
	}
	p.P(`return nil, `, identEmptyIDError)
	p.P(`}`)

	var fs string
	if p.readHasFieldSelection(ormable) {
		fs = "fs"
	} else {
		fs = "nil"
	}

	p.generateBeforeReadHookCall(ormable, "ApplyQuery")
	p.P(`if db, err = `, identApplyFieldSelectionFn, `(ctx, db, `, fs, `, &`, ormable.Name, `{}); err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	p.generateBeforeReadHookCall(ormable, "Find")
	p.P(`ormResponse := `, ormable.Name, `{}`)
	p.P(`if err = db.Where(&ormObj).First(&ormResponse).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.generateAfterReadHookCall(ormable)
	p.P(`pbResponse, err := ormResponse.ToPB(ctx)`)
	p.P(`return &pbResponse, err`)
	p.P(`}`)
	p.generateBeforeReadHookDef(ormable, "ApplyQuery")
	p.generateBeforeReadHookDef(ormable, "Find")
	p.generateAfterReadHookDef(ormable)
}

func (p *OrmPlugin) generateBeforeReadHookDef(orm *OrmableType, suffix string) {
	p.P(`type `, orm.Name, `WithBeforeRead`, suffix, ` interface {`)
	hookSign := fmt.Sprint(`BeforeRead`, suffix, `(`, p.qualifiedGoIdent(identCtx), `, `, p.qualifiedGoIdentPtr(identGormDB))
	if p.readHasFieldSelection(orm) {
		hookSign += fmt.Sprint(`, `, p.qualifiedGoIdentPtr(identQueryFieldSelection))
	}
	hookSign += fmt.Sprint(`) (`, p.qualifiedGoIdentPtr(identGormDB), `, error)`)
	p.P(hookSign)
	p.P(`}`)
}

func (p *OrmPlugin) generateAfterReadHookDef(orm *OrmableType) {
	p.P(`type `, orm.Name, `WithAfterReadFind interface {`)
	hookSign := fmt.Sprint(`AfterReadFind`, `(`, p.qualifiedGoIdent(identCtx), `, `, p.qualifiedGoIdentPtr(identGormDB))
	if p.readHasFieldSelection(orm) {
		hookSign += fmt.Sprint(`, `, p.qualifiedGoIdentPtr(identQueryFieldSelection))
	}
	hookSign += `) error`
	p.P(hookSign)
	p.P(`}`)
}

func (p *OrmPlugin) generateBeforeReadHookCall(orm *OrmableType, suffix string) {
	p.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `WithBeforeRead`, suffix, `); ok {`)
	hookCall := fmt.Sprint(`if db, err = hook.BeforeRead`, suffix, `(ctx, db`)
	if p.readHasFieldSelection(orm) {
		hookCall += `, fs`
	}
	hookCall += `); err != nil{`
	p.P(hookCall)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generateAfterReadHookCall(orm *OrmableType) {
	p.P(`if hook, ok := interface{}(&ormResponse).(`, orm.Name, `WithAfterReadFind`, `); ok {`)
	hookCall := fmt.Sprint(`if err = hook.AfterReadFind(ctx, db`)
	if p.readHasFieldSelection(orm) {
		hookCall += `, fs`
	}
	hookCall += `); err != nil {`
	p.P(hookCall)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generateApplyFieldMask(message *protogen.Message) {
	// return
	typeName := messageType(message)
	p.P(`// DefaultApplyFieldMask`, typeName, ` patches an pbObject with patcher according to a field mask.`)
	p.P(`func DefaultApplyFieldMask`, typeName, `(ctx `, identCtx, `, patchee *`,
		typeName, `, patcher *`, typeName, `, updateMask `, p.qualifiedGoIdentPtr(identFieldMask),
		`, prefix string, db `, p.qualifiedGoIdentPtr(identGormDB), `) (*`, typeName, `, error) {`)

	p.P(`if patcher == nil {`)
	p.P(`return nil, nil`)
	p.P(`} else if patchee == nil {`)
	p.P(`return nil, `, identNilArgumentError)
	p.P(`}`)
	p.P(`var err error`)
	hasNested := false
	for _, field := range message.Fields {
		desc := field.Desc
		fieldType := fieldType(field)
		fieldName := fieldName(field)
		notSpecialType := !p.isSpecialType(fieldType, field.GoIdent)

		if desc.Message() != nil && notSpecialType && !desc.IsList() {
			p.P(`var updated`, fieldName, ` bool`)
			hasNested = true
		} else if strings.HasSuffix(fieldType, protoTypeJSON) {
			p.P(`var updated`, fieldName, ` bool`)
		}
	}
	// Patch pbObj with input according to a field mask.
	if hasNested {
		p.P(`for i, f := range updateMask.Paths {`)
	} else {
		p.P(`for _, f := range updateMask.Paths {`)
	}
	for _, field := range message.Fields {
		desc := field.Desc
		ccName := fieldName(field)
		fieldType := fieldType(field)
		//  for ormable message, do recursive patching
		if desc.Message() != nil && p.isOrmable(fieldType) && !desc.IsList() {
			p.UsingGoImports(stdStringsImport)
			p.P(`if !updated`, ccName, ` && strings.HasPrefix(f, prefix+"`, ccName, `.") {`)
			p.P(`updated`, ccName, ` = true`)
			p.P(`if patcher.`, ccName, ` == nil {`)
			p.P(`patchee.`, ccName, ` = nil`)
			p.P(`continue`)
			p.P(`}`)
			p.P(`if patchee.`, ccName, ` == nil {`)
			p.P(`patchee.`, ccName, ` = &`, strings.TrimPrefix(fieldType, "*"), `{}`)
			p.P(`}`)
			if s := strings.Split(fieldType, "."); len(s) == 2 {
				p.P(`if o, err := `, strings.TrimLeft(s[0], "*"), `.DefaultApplyFieldMask`, s[1], `(ctx, patchee.`, ccName,
					`, patcher.`, ccName, `, &`, identFieldMask,
					`{Paths:updateMask.Paths[i:]}, prefix+"`, ccName, `.", db); err != nil {`)
			} else {
				p.P(`if o, err := DefaultApplyFieldMask`, strings.TrimPrefix(fieldType, "*"), `(ctx, patchee.`, ccName,
					`, patcher.`, ccName, `, &`, identFieldMask,
					`{Paths:updateMask.Paths[i:]}, prefix+"`, ccName, `.", db); err != nil {`)
			}
			p.P(`return nil, err`)
			p.P(`} else {`)
			p.P(`patchee.`, ccName, ` = o`)
			p.P(`}`)
			p.P(`continue`)
			p.P(`}`)
			p.P(`if f == prefix+"`, ccName, `" {`)
			p.P(`updated`, ccName, ` = true`)
			p.P(`patchee.`, ccName, ` = patcher.`, ccName)
			p.P(`continue`)
			p.P(`}`)
		} else if desc.Message() != nil && !p.isSpecialType(fieldType, field.GoIdent) && !desc.IsList() {
			p.UsingGoImports(stdStringsImport)
			p.P(`if !updated`, ccName, ` && strings.HasPrefix(f, prefix+"`, ccName, `.") {`)
			p.P(`if patcher.`, ccName, ` == nil {`)
			p.P(`patchee.`, ccName, ` = nil`)
			p.P(`continue`)
			p.P(`}`)
			p.P(`if patchee.`, ccName, ` == nil {`)
			p.P(`patchee.`, ccName, ` = &`, strings.TrimPrefix(fieldType, "*"), `{}`)
			p.P(`}`)
			p.P(`childMask := &`, identFieldMask, `{}`)
			p.P(`for j := i; j < len(updateMask.Paths); j++ {`)
			p.P(`if trimPath := strings.TrimPrefix(updateMask.Paths[j], prefix+"`, ccName, `."); trimPath != updateMask.Paths[j] {`)
			p.P(`childMask.Paths = append(childMask.Paths, trimPath)`)
			p.P(`}`)
			p.P(`}`)
			p.P(`if err := `, p.identFnCall(identMergeWithMaskFn, "patcher."+ccName, "patchee."+ccName, "childMask"), `; err != nil {`)
			p.P(`return nil, nil`)
			p.P(`}`)
			p.P(`}`)
			p.P(`if f == prefix+"`, ccName, `" {`)
			p.P(`updated`, ccName, ` = true`)
			p.P(`patchee.`, ccName, ` = patcher.`, ccName)
			p.P(`continue`)
			p.P(`}`)
		} else if strings.HasSuffix(fieldType, protoTypeJSON) && !desc.IsList() {
			p.UsingGoImports(stdStringsImport)
			p.P(`if !updated`, ccName, ` && strings.HasPrefix(f, prefix+"`, ccName, `") {`)
			p.P(`patchee.`, ccName, ` = patcher.`, ccName)
			p.P(`updated`, ccName, ` = true`)
			p.P(`continue`)
			p.P(`}`)
		} else {
			p.P(`if f == prefix+"`, ccName, `" {`)
			p.P(`patchee.`, ccName, ` = patcher.`, ccName)
			p.P(`continue`)
			p.P(`}`)
		}
	}
	p.P(`}`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`return patchee, nil`)
	p.P(`}`)
	p.P()
}

func (p *OrmPlugin) hasIDField(message *protogen.Message) bool {
	for _, field := range message.Fields {
		if strings.ToLower(fieldName(field)) == "id" {
			return true
		}
	}
	return false
}

func (p *OrmPlugin) generatePatchHandler(message *protogen.Message) {
	var isMultiAccount bool

	typeName := messageType(message)
	ormable := p.getOrmable(typeName)

	if getMessageOptions(message).GetMultiAccount() {
		isMultiAccount = true
	}

	if isMultiAccount && !p.hasIDField(message) {
		p.P(fmt.Sprintf("// Cannot autogen DefaultPatch%s: this is a multi-account table without an \"id\" field in the message.\n", typeName))
		return
	}

	p.P(`// DefaultPatch`, typeName, ` executes a basic gorm update call with patch behavior`)
	p.P(`func DefaultPatch`, typeName, `(ctx `, identCtx, `, in *`,
		typeName, `, updateMask `, p.qualifiedGoIdentPtr(identFieldMask), `, db `, p.qualifiedGoIdentPtr(identGormDB), `) (*`, typeName, `, error) {`)

	p.P(`if in == nil {`)
	p.P(`return nil, `, identNilArgumentError)
	p.P(`}`)
	p.P(`var pbObj `, typeName)
	p.P(`var err error`)
	p.generateBeforePatchHookCall(ormable, "Read")
	if p.readHasFieldSelection(ormable) {
		p.P(`pbReadRes, err := DefaultRead`, typeName, `(ctx, &`, typeName, `{Id: in.GetId()}, db, nil)`)
	} else {
		p.P(`pbReadRes, err := DefaultRead`, typeName, `(ctx, &`, typeName, `{Id: in.GetId()}, db)`)
	}

	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	p.P(`pbObj = *pbReadRes`)

	p.generateBeforePatchHookCall(ormable, "ApplyFieldMask")
	p.P(`if _, err := DefaultApplyFieldMask`, typeName, `(ctx, &pbObj, in, updateMask, "", db); err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	p.generateBeforePatchHookCall(ormable, "Save")
	p.P(`pbResponse, err := DefaultStrictUpdate`, typeName, `(ctx, &pbObj, db)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.generateAfterPatchHookCall(ormable, "Save")

	p.P(`return pbResponse, nil`)
	p.P(`}`)

	p.generateBeforePatchHookDef(ormable, "Read")
	p.generateBeforePatchHookDef(ormable, "ApplyFieldMask")
	p.generateBeforePatchHookDef(ormable, "Save")
	p.generateAfterPatchHookDef(ormable, "Save")
}

func (p *OrmPlugin) generateBeforePatchHookDef(orm *OrmableType, suffix string) {
	p.P(`type `, orm.OriginName, `WithBeforePatch`, suffix, ` interface {`)
	p.P(`BeforePatch`, suffix, `(`, identCtx, `, *`, orm.OriginName, `, `, p.qualifiedGoIdentPtr(identFieldMask), `, *`, identGormDB,
		`) (*`, identGormDB, `, error)`)
	p.P(`}`)
}

func (p *OrmPlugin) generateBeforePatchHookCall(orm *OrmableType, suffix string) {
	p.P(`if hook, ok := interface{}(&pbObj).(`, orm.OriginName, `WithBeforePatch`, suffix, `); ok {`)
	p.P(`if db, err = hook.BeforePatch`, suffix, `(ctx, in, updateMask, db); err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generateAfterPatchHookDef(orm *OrmableType, suffix string) {
	p.P(`type `, orm.OriginName, `WithAfterPatch`, suffix, ` interface {`)
	p.P(`AfterPatch`, suffix, `(`, identCtx, `, *`, orm.OriginName, `, `, p.qualifiedGoIdentPtr(identFieldMask), `, `, p.qualifiedGoIdentPtr(identGormDB),
		`) error`)
	p.P(`}`)
}

func (p *OrmPlugin) generateAfterPatchHookCall(orm *OrmableType, suffix string) {
	p.P(`if hook, ok := interface{}(pbResponse).(`, orm.OriginName, `WithAfterPatch`, suffix, `); ok {`)
	p.P(`if err = hook.AfterPatch`, suffix, `(ctx, in, updateMask, db); err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generatePatchSetHandler(message *protogen.Message) {
	var isMultiAccount bool

	typeName := messageType(message)
	if getMessageOptions(message).GetMultiAccount() {
		isMultiAccount = true
	}

	if isMultiAccount && !p.hasIDField(message) {
		p.P(fmt.Sprintf("// Cannot autogen DefaultPatchSet%s: this is a multi-account table without an \"id\" field in the message.\n", typeName))
		return
	}

	p.UsingGoImports(stdFmtImport)
	p.P(`// DefaultPatchSet`, typeName, ` executes a bulk gorm update call with patch behavior`)
	p.P(`func DefaultPatchSet`, typeName, `(ctx `, identCtx, `, objects []*`,
		typeName, `, updateMasks []`, p.qualifiedGoIdentPtr(identFieldMask), `, db `, p.qualifiedGoIdentPtr(identGormDB), `) ([]*`, typeName, `, error) {`)
	p.P(`if len(objects) != len(updateMasks) {`)
	p.P(`return nil, fmt.Errorf(`, identBadRepeatedFieldMaskTplError, `, len(updateMasks), len(objects))`)
	p.P(`}`)
	p.P(``)
	p.P(`results := make([]*`, typeName, `, 0, len(objects))`)
	p.P(`for i, patcher := range objects {`)
	p.P(`pbResponse, err := DefaultPatch`, typeName, `(ctx, patcher, updateMasks[i], db)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(``)
	p.P(`results = append(results, pbResponse)`)
	p.P(`}`)
	p.P(``)
	p.P(`return results, nil`)
	p.P(`}`)
}

func (p *OrmPlugin) generateDeleteHandler(message *protogen.Message) {
	typeName := messageType(message)
	p.P(`func DefaultDelete`, typeName, `(ctx `, identCtx, `, in *`,
		typeName, `, db `, p.qualifiedGoIdentPtr(identGormDB), `) error {`)
	p.P(`if in == nil {`)
	p.P(`return `, identNilArgumentError)
	p.P(`}`)
	p.P(`ormObj, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return err`)
	p.P(`}`)
	ormable := p.getOrmable(typeName)
	pkName, pk := p.findPrimaryKey(ormable)
	if strings.Contains(pk.Type, "*") {
		p.P(`if ormObj.`, pkName, ` == nil || *ormObj.`, pkName, ` == `, p.guessZeroValue(pk.Type), ` {`)
	} else {
		p.P(`if ormObj.`, pkName, ` == `, p.guessZeroValue(pk.Type), `{`)
	}
	p.P(`return `, identEmptyIDError)
	p.P(`}`)
	p.generateBeforeDeleteHookCall(ormable)
	p.P(`err = db.Where(&ormObj).Delete(&`, ormable.Name, `{}).Error`)
	p.P(`if err != nil {`)
	p.P(`return err`)
	p.P(`}`)
	p.generateAfterDeleteHookCall(ormable)
	p.P(`return err`)
	p.P(`}`)
	delete := "Delete_"
	p.generateBeforeHookDef(ormable, delete)
	p.generateAfterHookDef(ormable, delete)
}

func (p *OrmPlugin) generateBeforeDeleteHookCall(orm *OrmableType) {
	p.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `WithBeforeDelete_); ok {`)
	p.P(`if db, err = hook.BeforeDelete_(ctx, db); err != nil {`)
	p.P(`return err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generateAfterDeleteHookCall(orm *OrmableType) {
	p.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `WithAfterDelete_); ok {`)
	p.P(`err = hook.AfterDelete_(ctx, db)`)
	p.P(`}`)
}

func (p *OrmPlugin) generateDeleteSetHandler(message *protogen.Message) {
	typeName := messageType(message)
	p.P(`func DefaultDelete`, typeName, `Set(ctx `, identCtx, `, in []*`,
		typeName, `, db *`, identGormDB, `) error {`)
	p.P(`if in == nil {`)
	p.P(`return `, identNilArgumentError)
	p.P(`}`)
	p.P(`var err error`)
	ormable := p.getOrmable(typeName)
	pkName, pk := p.findPrimaryKey(ormable)
	column := pk.GetTag().GetColumn()
	if len(column) != 0 {
		pkName = column
	}
	p.P(`keys := []`, pk.Type, `{}`)
	p.P(`for _, obj := range in {`)
	p.P(`ormObj, err := obj.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return err`)
	p.P(`}`)
	if strings.Contains(pk.Type, "*") {
		p.P(`if ormObj.`, pkName, ` == nil || *ormObj.`, pkName, ` == `, p.guessZeroValue(pk.Type), ` {`)
	} else {
		p.P(`if ormObj.`, pkName, ` == `, p.guessZeroValue(pk.Type), `{`)
	}
	p.P(`return `, identEmptyIDError)
	p.P(`}`)
	p.P(`keys = append(keys, ormObj.`, pkName, `)`)
	p.P(`}`)
	p.generateBeforeDeleteSetHookCall(ormable)
	if getMessageOptions(message).GetMultiAccount() {
		p.P(`acctId, err := `, identGetAccountIDFn, `(ctx, nil)`)
		p.P(`if err != nil {`)
		p.P(`return err`)
		p.P(`}`)
		p.P(`err = db.Where("account_id = ? AND `, jgorm.ToDBName(pkName), ` in (?)", acctId, keys).Delete(&`, ormable.Name, `{}).Error`)
	} else {
		p.P(`err = db.Where("`, jgorm.ToDBName(pkName), ` in (?)", keys).Delete(&`, ormable.Name, `{}).Error`)
	}
	p.P(`if err != nil {`)
	p.P(`return err`)
	p.P(`}`)
	p.generateAfterDeleteSetHookCall(ormable)
	p.P(`return err`)
	p.P(`}`)
	p.P(`type `, ormable.Name, `WithBeforeDeleteSet interface {`)
	p.P(`BeforeDeleteSet(`, identCtx, `, []*`, ormable.OriginName, `, `, p.qualifiedGoIdentPtr(identGormDB), `) (`, p.qualifiedGoIdentPtr(identGormDB), `, error)`)
	p.P(`}`)
	p.P(`type `, ormable.Name, `WithAfterDeleteSet interface {`)
	p.P(`AfterDeleteSet(`, identCtx, `, []*`, ormable.OriginName, `, `, p.qualifiedGoIdentPtr(identGormDB), `) error`)
	p.P(`}`)
}

func (p *OrmPlugin) generateBeforeDeleteSetHookCall(orm *OrmableType) {
	p.P(`if hook, ok := (interface{}(&`, orm.Name, `{})).(`, orm.Name, `WithBeforeDeleteSet); ok {`)
	p.P(`if db, err = hook.BeforeDeleteSet(ctx, in, db); err != nil {`)
	p.P(`return err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generateAfterDeleteSetHookCall(orm *OrmableType) {
	p.P(`if hook, ok := (interface{}(&`, orm.Name, `{})).(`, orm.Name, `WithAfterDeleteSet); ok {`)
	p.P(`err = hook.AfterDeleteSet(ctx, in, db)`)
	p.P(`}`)
}

func (p *OrmPlugin) generateListHandler(message *protogen.Message) {
	typeName := messageType(message)
	ormable := p.getOrmable(typeName)

	p.P(`// DefaultList`, typeName, ` executes a gorm list call`)
	listSign := fmt.Sprint(`func DefaultList`, typeName, `(ctx `, p.qualifiedGoIdent(identCtx), `, db `, p.qualifiedGoIdentPtr(identGormDB))
	var f, s, pg, fs string
	if p.listHasFiltering(ormable) {
		listSign += fmt.Sprint(`, f `, p.qualifiedGoIdentPtr(identQueryFiltering))
		f = "f"
	} else {
		f = "nil"
	}
	if p.listHasSorting(ormable) {
		listSign += fmt.Sprint(`, s `, p.qualifiedGoIdentPtr(identQuerySorting))
		s = "s"
	} else {
		s = "nil"
	}
	if p.listHasPagination(ormable) {
		listSign += fmt.Sprint(`, p `, p.qualifiedGoIdentPtr(identQueryPagination))
		pg = "p"
	} else {
		pg = "nil"
	}
	if p.listHasFieldSelection(ormable) {
		listSign += fmt.Sprint(`, fs `, p.qualifiedGoIdentPtr(identQueryFieldSelection))
		fs = "fs"
	} else {
		fs = "nil"
	}
	listSign += fmt.Sprint(`) ([]*`, typeName, `, error) {`)
	p.P(listSign)
	p.P(`in := `, typeName, `{}`)
	p.P(`ormObj, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.generateBeforeListHookCall(ormable, "ApplyQuery")
	p.P(`db, err = `, p.Import(tkgormImport), `.ApplyCollectionOperators(ctx, db, &`, ormable.Name, `{}, &`, typeName, `{}, `, f, `,`, s, `,`, pg, `,`, fs, `)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.generateBeforeListHookCall(ormable, "Find")
	p.P(`db = db.Where(&ormObj)`)

	// add default ordering by primary key
	if p.hasPrimaryKey(ormable) {
		pkName, pk := p.findPrimaryKey(ormable)
		column := pk.GetTag().GetColumn()
		if len(column) == 0 {
			column = jgorm.ToDBName(pkName)
		}
		p.P(`db = db.Order("`, column, `")`)
	}

	p.P(`ormResponse := []`, ormable.Name, `{}`)
	p.P(`if err := db.Find(&ormResponse).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.generateAfterListHookCall(ormable, "Find")
	p.P(`pbResponse := []*`, typeName, `{}`)
	p.P(`for _, responseEntry := range ormResponse {`)
	p.P(`temp, err := responseEntry.ToPB(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse = append(pbResponse, &temp)`)
	p.P(`}`)
	p.P(`return pbResponse, nil`)
	p.P(`}`)
	p.generateBeforeListHookDef(ormable, "ApplyQuery")
	p.generateBeforeListHookDef(ormable, "Find")
	p.generateAfterListHookDef(ormable, "Find")
}

func (p *OrmPlugin) generateListHookDefHelper(orm *OrmableType, suffix string, returnDB bool) {
	p.P(`type `, orm.Name, `With`, suffix, ` interface {`)
	hookSign := fmt.Sprint(suffix, `(`, p.qualifiedGoIdent(identCtx), `, `, p.qualifiedGoIdentPtr(identGormDB))
	if returnDB {
		hookSign += fmt.Sprint(`, *[]`, orm.Name)
	}
	if p.listHasFiltering(orm) {
		hookSign += fmt.Sprint(`, *`, p.Import(queryImport), `.Filtering`)
	}
	if p.listHasSorting(orm) {
		hookSign += fmt.Sprint(`, *`, p.Import(queryImport), `.Sorting`)
	}
	if p.listHasPagination(orm) {
		hookSign += fmt.Sprint(`, *`, p.Import(queryImport), `.Pagination`)
	}
	if p.listHasFieldSelection(orm) {
		hookSign += fmt.Sprint(`, `, p.qualifiedGoIdentPtr(identQueryFieldSelection))
	}
	hookSign += fmt.Sprint(`) `)
	if returnDB {
		hookSign += fmt.Sprint(`error`)
	} else {
		hookSign += fmt.Sprint(`(`, p.qualifiedGoIdentPtr(identGormDB), `, error)`)
	}
	p.P(hookSign)
	p.P(`}`)
}

func (p *OrmPlugin) generateBeforeListHookDef(orm *OrmableType, suffix string) {
	p.generateListHookDefHelper(orm, "BeforeList"+suffix, false)
}

func (p *OrmPlugin) generateAfterListHookDef(orm *OrmableType, suffix string) {
	p.generateListHookDefHelper(orm, "AfterList"+suffix, true)
}

func (p *OrmPlugin) generateListHookCallHelper(orm *OrmableType, suffix string, passORMResponse bool) {
	p.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `With`, suffix, `); ok {`)
	hookCall := fmt.Sprint(`if db, err = hook.`, suffix, `(ctx, db`)
	if passORMResponse {
		hookCall += fmt.Sprint(` &ormResponse`)
	}
	if p.listHasFiltering(orm) {
		hookCall += `,f`
	}
	if p.listHasSorting(orm) {
		hookCall += `,s`
	}
	if p.listHasPagination(orm) {
		hookCall += `,p`
	}
	if p.listHasFieldSelection(orm) {
		hookCall += `,fs`
	}
	hookCall += `); err != nil {`
	p.P(hookCall)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generateBeforeListHookCall(orm *OrmableType, suffix string) {
	p.generateListHookCallHelper(orm, "BeforeList"+suffix, false)
}

func (p *OrmPlugin) generateAfterListHookCall(orm *OrmableType, suffix string) {
	p.generateListHookCallHelper(orm, "AfterList"+suffix, true)
}

func (p *OrmPlugin) generateStrictUpdateHandler(message *protogen.Message) {
	p.UsingGoImports(stdFmtImport)
	typeName := messageType(message)
	p.P(`// DefaultStrictUpdate`, typeName, ` clears / replaces / appends first level 1:many children and then executes a gorm update call`)
	p.P(`func DefaultStrictUpdate`, typeName, `(ctx `, identCtx, `, in *`,
		typeName, `, db *`, identGormDB, `) (*`, typeName, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, fmt.Errorf("Nil argument to DefaultStrictUpdate`, typeName, `")`)
	p.P(`}`)
	p.P(`ormObj, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	if getMessageOptions(message).GetMultiAccount() {
		p.generateAccountIdWhereClause()
	}
	ormable := p.getOrmable(typeName)
	if p.Gateway {
		p.P(`var count int64`)
	}
	if p.hasPrimaryKey(ormable) {
		pkName, pk := p.findPrimaryKey(ormable)
		column := pk.GetTag().GetColumn()
		if len(column) == 0 {
			column = jgorm.ToDBName(pkName)
		}
		p.P(`lockedRow := &`, typeName, `ORM{}`)
		var count string
		var rowsAffected string
		if p.Gateway {
			count = `count = `
			rowsAffected = `.RowsAffected`
		}
		p.P(count+`db.Model(&ormObj).Set("gorm:query_option", "FOR UPDATE").Where("`, column, `=?", ormObj.`, pkName, `).First(lockedRow)`+rowsAffected)
	}
	p.generateBeforeHookCall(ormable, "StrictUpdateCleanup")
	p.handleChildAssociations(message)
	p.generateBeforeHookCall(ormable, "StrictUpdateSave")
	p.P(`if err = db.Save(&ormObj).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.generateAfterHookCall(ormable, "StrictUpdateSave")
	p.P(`pbResponse, err := ormObj.ToPB(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	if p.Gateway {
		p.P(`if count == 0 {`)
		p.P(`err = `, p.Import(gatewayImport), `.SetCreated(ctx, "")`)
		p.P(`}`)
	}

	p.P(`return &pbResponse, err`)
	p.P(`}`)
	p.generateBeforeHookDef(ormable, "StrictUpdateCleanup")
	p.generateBeforeHookDef(ormable, "StrictUpdateSave")
	p.generateAfterHookDef(ormable, "StrictUpdateSave")
}

func (p *OrmPlugin) isFieldOrmable(message *protogen.Message, fieldName string) bool {
	_, ok := p.getOrmableMessage(message).Fields[fieldName]
	return ok
}

func (p *OrmPlugin) handleChildAssociations(message *protogen.Message) {
	ormable := p.getOrmableMessage(message)
	for _, fieldName := range p.getSortedFieldNames(ormable.Fields) {
		p.handleChildAssociationsByName(message, fieldName)
	}
}

func (p *OrmPlugin) handleChildAssociationsByName(message *protogen.Message, fieldName string) {
	ormable := p.getOrmableMessage(message)
	field := ormable.Fields[fieldName]
	if field == nil {
		return
	}

	if field.GetHasMany() != nil || field.GetHasOne() != nil || field.GetManyToMany() != nil {
		var assocHandler string
		switch {
		case field.GetHasMany() != nil:
			switch {
			case field.GetHasMany().GetClear():
				assocHandler = "Clear"
			case field.GetHasMany().GetAppend():
				assocHandler = "Append"
			case field.GetHasMany().GetReplace():
				assocHandler = "Replace"
			default:
				assocHandler = "Remove"
			}
		case field.GetHasOne() != nil:
			switch {
			case field.GetHasOne().GetClear():
				assocHandler = "Clear"
			case field.GetHasOne().GetAppend():
				assocHandler = "Append"
			case field.GetHasOne().GetReplace():
				assocHandler = "Replace"
			default:
				assocHandler = "Remove"
			}
		case field.GetManyToMany() != nil:
			switch {
			case field.GetManyToMany().GetClear():
				assocHandler = "Clear"
			case field.GetManyToMany().GetAppend():
				assocHandler = "Append"
			case field.GetManyToMany().GetReplace():
				assocHandler = "Replace"
			default:
				assocHandler = "Replace"
			}
		}

		if assocHandler == "Remove" {
			p.removeChildAssociationsByName(message, fieldName)
			return
		}

		action := fmt.Sprintf("%s(ormObj.%s)", assocHandler, fieldName)
		if assocHandler == "Clear" {
			action = fmt.Sprintf("%s()", assocHandler)
		}

		p.P(`if err = db.Model(&ormObj).Association("`, fieldName, `").`, action, `.Error; err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
		p.P(`ormObj.`, fieldName, ` = nil`)
	}
}

func (p *OrmPlugin) removeChildAssociationsByName(message *protogen.Message, fieldName string) {
	ormable := p.getOrmableMessage(message)
	field := ormable.Fields[fieldName]

	if field == nil {
		return
	}

	if field.GetHasMany() != nil || field.GetHasOne() != nil {
		var assocKeyName, foreignKeyName string
		switch {
		case field.GetHasMany() != nil:
			assocKeyName = field.GetHasMany().GetAssociationForeignkey()
			foreignKeyName = field.GetHasMany().GetForeignkey()
		case field.GetHasOne() != nil:
			assocKeyName = field.GetHasOne().GetAssociationForeignkey()
			foreignKeyName = field.GetHasOne().GetForeignkey()
		}
		assocKeyType := ormable.Fields[assocKeyName].Type
		assocOrmable := p.getOrmable(field.Type)
		foreignKeyType := assocOrmable.Fields[foreignKeyName].Type
		p.P(`filter`, fieldName, ` := `, strings.Trim(field.Type, "[]*"), `{}`)
		zeroValue := p.guessZeroValue(assocKeyType)
		if strings.Contains(assocKeyType, "*") {
			p.P(`if ormObj.`, assocKeyName, ` == nil || *ormObj.`, assocKeyName, ` == `, zeroValue, `{`)
		} else {
			p.P(`if ormObj.`, assocKeyName, ` == `, zeroValue, `{`)
		}
		p.P(`return nil, `, identEmptyIDError)
		p.P(`}`)
		filterDesc := "filter" + fieldName + "." + foreignKeyName
		ormDesc := "ormObj." + assocKeyName
		if strings.HasPrefix(foreignKeyType, "*") {
			p.P(filterDesc, ` = new(`, strings.TrimPrefix(foreignKeyType, "*"), `)`)
			filterDesc = "*" + filterDesc
		}
		if strings.HasPrefix(assocKeyType, "*") {
			ormDesc = "*" + ormDesc
		}
		p.P(filterDesc, " = ", ormDesc)
		p.P(`if err = db.Where(filter`, fieldName, `).Delete(`, strings.Trim(field.Type, "[]*"), `{}).Error; err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
	}
}

// guessZeroValue of the input type, so that we can check if a (key) value is set or not
func (p *OrmPlugin) guessZeroValue(typeName string) string {
	typeName = strings.ToLower(typeName)

	type tmp struct {
		cmp string
		ret interface{}
	}
	for _, current := range []tmp{
		{cmp: "string", ret: `""`},
		{cmp: "int", ret: `0`},
		{cmp: "uuid", ret: identNilUUID},
		{cmp: "[]byte", ret: `nil`},
		{cmp: "bool", ret: `false`},
	} {
		if strings.Contains(typeName, current.cmp) {
			switch v := current.ret.(type) {
			case protogen.GoIdent:
				return p.qualifiedGoIdent(v)
			case string:
				return v
			default:
				panic("invalid guessZeroValue tmp structs")
			}
		}
	}
	return ``
}

func (p *OrmPlugin) hasMethodGenericHelper(ormable *OrmableType, idx string, callback func(*protogen.Message) string) bool {
	if read, ok := ormable.Methods[idx]; ok {
		if s := callback(read.inType); s != "" {
			return true
		}
	}
	return false
}

func (p *OrmPlugin) readHasFieldSelection(ormable *OrmableType) bool {
	return p.hasMethodGenericHelper(ormable, readService, p.getFieldSelection)
}

func (p *OrmPlugin) listHasFiltering(ormable *OrmableType) bool {
	return p.hasMethodGenericHelper(ormable, listService, p.getFiltering)
}

func (p *OrmPlugin) listHasSorting(ormable *OrmableType) bool {
	return p.hasMethodGenericHelper(ormable, listService, p.getSorting)
}

func (p *OrmPlugin) listHasPagination(ormable *OrmableType) bool {
	return p.hasMethodGenericHelper(ormable, listService, p.getPagination)
}

func (p *OrmPlugin) listHasFieldSelection(ormable *OrmableType) bool {
	return p.hasMethodGenericHelper(ormable, listService, p.getFieldSelection)
}
