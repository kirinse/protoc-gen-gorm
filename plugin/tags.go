package plugin

import (
	"fmt"
	"strconv"
	"strings"

	gorm "github.com/edhaight/protoc-gen-gorm/options"
)

type tagString string

func (t *tagString) checkAndSetString(attr *string, key string, omitEmpty bool) {
	if attr != nil {
		if omitEmpty && *attr == "" {
			(*t) += tagString(fmt.Sprintf("%s;", key))
		} else {
			(*t) += tagString(fmt.Sprintf("%s:%s;", key, *attr))
		}
	}
}

func (t *tagString) checkAndSetInt32(attr *int32, key string) {
	if attr != nil {
		(*t) += tagString(fmt.Sprintf("%s:%d;", key, *attr))
	}
}

func (t *tagString) checkAndSetBool(attr *bool, key string, formatBool bool) {
	if attr != nil && *attr {
		if formatBool {
			(*t) += tagString(fmt.Sprintf("%s:%s;", key, strconv.FormatBool(*attr)))
		} else {
			(*t) += tagString(fmt.Sprintf("%s;", key))
		}
	}
}

func (t *tagString) format(key string) string {
	if string(*t) == "" {
		return ""
	}
	return fmt.Sprintf("%s:\"%s\"", key, strings.TrimRight(string(*t), ";"))
}

func (p *OrmPlugin) renderGormTag(field *Field) string {
	var gormRes, atlasRes tagString
	tag := field.GetTag()
	if tag == nil {
		tag = &gorm.GormTag{}
	}

	gormRes.checkAndSetString(tag.Column, "column", false)
	gormRes.checkAndSetString(tag.Type, "type", false)

	gormRes.checkAndSetInt32(tag.Size, "size")
	gormRes.checkAndSetInt32(tag.Precision, "precision")

	gormRes.checkAndSetBool(tag.PrimaryKey, "primary_key", false)
	gormRes.checkAndSetBool(tag.Unique, "unique", false)

	gormRes.checkAndSetString(tag.Default, "default", false)
	gormRes.checkAndSetBool(tag.NotNull, "not null", false)
	gormRes.checkAndSetBool(tag.AutoIncrement, "auto_increment", false)

	gormRes.checkAndSetString(tag.Index, "index", true)
	gormRes.checkAndSetString(tag.UniqueIndex, "unique_index", true)

	gormRes.checkAndSetBool(tag.Embedded, "embedded", false)
	gormRes.checkAndSetString(tag.EmbeddedPrefix, "embedded_prefix", false)
	gormRes.checkAndSetBool(tag.Ignore, "-", false)

	var foreignKey, associationForeignKey, joinTable, joinTableForeignKey, associationJoinTableForeignKey *string
	var associationAutoupdate, associationAutocreate, associationSaveReference, preload, replace, append, clear *bool
	if hasOne := field.GetHasOne(); hasOne != nil {
		foreignKey = hasOne.Foreignkey
		associationForeignKey = hasOne.AssociationForeignkey
		associationAutoupdate = hasOne.AssociationAutoupdate
		associationAutocreate = hasOne.AssociationAutocreate
		associationSaveReference = hasOne.AssociationSaveReference
		preload = hasOne.Preload
		clear = hasOne.Clear
		replace = hasOne.Replace
		append = hasOne.Append
	} else if belongsTo := field.GetBelongsTo(); belongsTo != nil {
		foreignKey = belongsTo.Foreignkey
		associationForeignKey = belongsTo.AssociationForeignkey
		associationAutoupdate = belongsTo.AssociationAutoupdate
		associationAutocreate = belongsTo.AssociationAutocreate
		associationSaveReference = belongsTo.AssociationSaveReference
		preload = belongsTo.Preload
	} else if hasMany := field.GetHasMany(); hasMany != nil {
		foreignKey = hasMany.Foreignkey
		associationForeignKey = hasMany.AssociationForeignkey
		associationAutoupdate = hasMany.AssociationAutoupdate
		associationAutocreate = hasMany.AssociationAutocreate
		associationSaveReference = hasMany.AssociationSaveReference
		clear = hasMany.Clear
		preload = hasMany.Preload
		replace = hasMany.Replace
		append = hasMany.Append
		atlasRes.checkAndSetString(hasMany.PositionField, "position", false)
	} else if mtm := field.GetManyToMany(); mtm != nil {
		foreignKey = mtm.Foreignkey
		associationForeignKey = mtm.AssociationForeignkey
		joinTable = mtm.Jointable
		joinTableForeignKey = mtm.JointableForeignkey
		associationJoinTableForeignKey = mtm.AssociationJointableForeignkey
		associationAutoupdate = mtm.AssociationAutoupdate
		associationAutocreate = mtm.AssociationAutocreate
		associationSaveReference = mtm.AssociationSaveReference
		preload = mtm.Preload
		clear = mtm.Clear
		replace = mtm.Replace
		append = mtm.Append
	} else {
		foreignKey = tag.Foreignkey
		associationForeignKey = tag.AssociationForeignkey
		joinTable = tag.ManyToMany
		joinTableForeignKey = tag.JointableForeignkey
		associationJoinTableForeignKey = tag.AssociationJointableForeignkey
		associationAutoupdate = tag.AssociationAutoupdate
		associationAutocreate = tag.AssociationAutocreate
		associationSaveReference = tag.AssociationSaveReference
		preload = tag.Preload
	}

	gormRes.checkAndSetString(foreignKey, "foreignkey", false)
	gormRes.checkAndSetString(associationForeignKey, "association_foreignkey", false)
	gormRes.checkAndSetString(joinTable, "many2many", false)
	gormRes.checkAndSetString(joinTableForeignKey, "jointable_foreignkey", false)
	gormRes.checkAndSetString(associationJoinTableForeignKey, "association_jointable_foreignkey", false)
	gormRes.checkAndSetBool(associationAutoupdate, "association_autoupdate", true)
	gormRes.checkAndSetBool(associationAutocreate, "association_autocreate", true)
	gormRes.checkAndSetBool(associationSaveReference, "association_save_reference", true)
	gormRes.checkAndSetBool(preload, "preload", true)
	gormRes.checkAndSetBool(clear, "clear", true)
	gormRes.checkAndSetBool(replace, "replace", true)
	gormRes.checkAndSetBool(append, "append", true)

	finalTag := strings.TrimSpace(strings.Join([]string{gormRes.format("gorm"), atlasRes.format("atlas")}, " "))
	if finalTag == "" {
		return ""
	}
	return fmt.Sprintf("`%s`", finalTag)
}
