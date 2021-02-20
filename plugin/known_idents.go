package plugin

import "google.golang.org/protobuf/compiler/protogen"

var (
	// stdlib idents
	identCtx                = newKnownIdent("Context", "context")
	identTime               = newKnownIdent("Time", "time")
	identStringsHasPrefixFn = newKnownIdent("HasPrefix", "strings")
	identJsonMarshal        = newKnownIdent("Marshal", "encoding/json")
	identFmtErrorf          = newKnownIdent("Errorf", "fmt")
	// proto custom types
	identTypesInet               = newKnownIdent("Inet", "github.com/edhaight/protoc-gen-gorm/types")
	identTypesInetValue          = newKnownIdent("InetValue", "github.com/edhaight/protoc-gen-gorm/types")
	identTypesUUIDValue          = newKnownIdent("UUIDValue", "github.com/edhaight/protoc-gen-gorm/types")
	identTypesUUID               = newKnownIdent("UUID", "github.com/edhaight/protoc-gen-gorm/types")
	identTypesJSONValue          = newKnownIdent("JSONValue", "github.com/edhaight/protoc-gen-gorm/types")
	identTypesParseInetFn        = newKnownIdent("ParseInet", "github.com/edhaight/protoc-gen-gorm/types")
	identTypesParseTimeFn        = newKnownIdent("ParseTime", "github.com/edhaight/protoc-gen-gorm/types")
	identTypesTimeOnlyByStringFn = newKnownIdent("TimeOnlyByString", "github.com/edhaight/protoc-gen-gorm/types")
	// gorm idents
	identGormDB         = newKnownIdent("DB", "github.com/jinzhu/gorm")
	identpqJsonb        = newKnownIdent("Jsonb", "github.com/jinzhu/gorm/dialects/postgres")
	identpqBoolArray    = newKnownIdent("Jsonb", "github.com/jinzhu/gorm/dialects/postgres")
	identpqFloat64Array = newKnownIdent("Float64Array", "github.com/lib/pq")
	identpqInt64Array   = newKnownIdent("Int64Array", "github.com/lib/pq")
	identpqStringArray  = newKnownIdent("StringArray", "github.com/lib/pq")
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
	identQueryPageInfo       = newKnownIdent("PageInfo", "github.com/infobloxopen/atlas-app-toolkit/query")

	identApplyFieldSelectionFn      = newKnownIdent("ApplyFieldSelection", "github.com/infobloxopen/atlas-app-toolkit/gorm")
	identMergeWithMaskFn            = newKnownIdent("MergeWithMask", "github.com/infobloxopen/atlas-app-toolkit/gorm")
	identApplyCollectionOperatorsFn = newKnownIdent("ApplyCollectionOperators", "github.com/infobloxopen/atlas-app-toolkit/gorm")
	identTkFromContextFn            = newKnownIdent("FromContext", "github.com/infobloxopen/atlas-app-toolkit/gorm")
	// Atlas resources idents
	identResourceEncodeFn      = newKnownIdent("Encode", "github.com/infobloxopen/atlas-app-toolkit/gorm/resource")
	identResourceDecodeFn      = newKnownIdent("Decode", "github.com/infobloxopen/atlas-app-toolkit/gorm/resource")
	identResourceDecodeBytesFn = newKnownIdent("DecodeBytes", "github.com/infobloxopen/atlas-app-toolkit/gorm/resource")
	identResourceDecodeInt64Fn = newKnownIdent("DecodeInt64", "github.com/infobloxopen/atlas-app-toolkit/gorm/resource")
	// trace idents
	identTraceSpan              = newKnownIdent("Span", "go.opencensus.io/trace")
	identTraceStartSpanFn       = newKnownIdent("StartSpan", "go.opencensus.io/trace")
	identTraceAttribute         = newKnownIdent("Attribute", "go.opencensus.io/trace")
	identTraceStringAttributeFn = newKnownIdent("StringAttribute", "go.opencensus.io/trace")
	identTraceStatus            = newKnownIdent("Status", "go.opencensus.io/trace")
	identTraceStatusCodeUnknown = newKnownIdent("StatusCodeUnknown", "go.opencensus.io/trace")
	// gateway idents
	identGatewaySetCreatedFn = newKnownIdent("SetCreated", "github.com/infobloxopen/atlas-app-toolkit/gateway")

	// GetAccountID function ident
	identGetAccountIDFn = newKnownIdent("GetAccountID", "github.com/infobloxopen/atlas-app-toolkit/auth")
	// fieldMask ident
	identFieldMask = newKnownIdent("FieldMask", "google.golang.org/genproto/protobuf/field_mask")
	// uuid idents
	identNilUUID          = newKnownIdent("Nil", "github.com/satori/go.uuid")
	identUUID             = newKnownIdent("UUID", "github.com/satori/go.uuid")
	identUUIDFromStringFn = newKnownIdent("FromString", "github.com/satori/go.uuid")
)

var specialImports = map[string]struct{}{
	"github.com/edhaight/protoc-gen-gorm/types":              {},
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource": {},
	"github.com/golang/protobuf/ptypes/timestamp":            {},
}

func newKnownIdent(goName, goImportPath string) protogen.GoIdent {
	return protogen.GoIdent{
		GoName:       goName,
		GoImportPath: protogen.GoImportPath(goImportPath),
	}
}

func ptrIdent(ident protogen.GoIdent) protogen.GoIdent {
	ident.GoName = "*" + ident.GoName
	return ident
}
