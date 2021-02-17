package plugin

import "google.golang.org/protobuf/compiler/protogen"

var (
	// stdlib idents
	identCtx                = newKnownIdent("Context", "context")
	identTime               = newKnownIdent("Time", "time")
	identStringsHasPrefixFn = newKnownIdent("HasPrefix", "strings")
	identJsonMarshal        = newKnownIdent("Marshal", "encoding/json")
	// gorm idents
	identGormDB = newKnownIdent("DB", "github.com/jinzhu/gorm")
	// timestamp idents
	identTimestamp      = newKnownIdent("Timestamp", "github.com/golang/protobuf/ptypes")
	identTimestampProto = newKnownIdent("TimestampProto", "github.com/golang/protobuf/ptypes")
	// error idents
	identNilArgumentError             = newKnownIdent("NilArgumentError", "github.com/edhaight/protoc-gen-gorm/errors")
	identEmptyIDError                 = newKnownIdent("EmptyIdError", "github.com/edhaight/protoc-gen-gorm/errors")
	identBadRepeatedFieldMaskTplError = newKnownIdent("BadRepeatedFieldMaskTpl", "github.com/edhaight/protoc-gen-gorm/errors")
	identNoTransactionError           = newKnownIdent("NoTransactionError", "github.com/edhaight/protoc-gen-gorm/errors")
	// field selection idents
	identQueryFieldSelection = newKnownIdent("FieldSelection", "github.com/infobloxopen/atlas-app-toolkit/query")
	identQueryPagination     = newKnownIdent("Pagination", "github.com/infobloxopen/atlas-app-toolkit/query")
	identQuerySorting        = newKnownIdent("Sorting", "github.com/infobloxopen/atlas-app-toolkit/query")
	identQueryFiltering      = newKnownIdent("Filtering", "github.com/infobloxopen/atlas-app-toolkit/query")

	identApplyFieldSelectionFn      = newKnownIdent("ApplyFieldSelection", "github.com/infobloxopen/atlas-app-toolkit/gorm")
	identMergeWithMaskFn            = newKnownIdent("MergeWithMask", "github.com/infobloxopen/atlas-app-toolkit/gorm")
	identApplyCollectionOperatorsFn = newKnownIdent("ApplyCollectionOperators", "github.com/infobloxopen/atlas-app-toolkit/gorm")
	identTkFromContextFn            = newKnownIdent("FromContext", "github.com/infobloxopen/atlas-app-toolkit/gorm")

	// GetAccountID function ident
	identGetAccountIDFn = newKnownIdent("GetAccountID", "github.com/infobloxopen/atlas-app-toolkit/auth")
	// fieldMask ident
	identFieldMask = newKnownIdent("FieldMask", "google.golang.org/genproto/protobuf/field_mask")
	// uuid idents
	identNilUUID          = newKnownIdent("Nil", "github.com/satori/go.uuid")
	identUUID             = newKnownIdent("UUID", "github.com/satori/go.uuid")
	identUUIDFromStringFn = newKnownIdent("FromString", "github.com/satori/go.uuid")
)

func newKnownIdent(goName, goImportPath string) protogen.GoIdent {
	return protogen.GoIdent{
		GoName:       goName,
		GoImportPath: protogen.GoImportPath(goImportPath),
	}
}
