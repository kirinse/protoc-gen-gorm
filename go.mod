module github.com/edhaight/protoc-gen-gorm

require (
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/infobloxopen/atlas-app-toolkit v0.20.0
	github.com/jinzhu/gorm v1.9.2
	github.com/jinzhu/inflection v1.0.0
	github.com/lib/pq v1.9.0
	github.com/mattn/go-sqlite3 v2.0.1+incompatible // indirect
	github.com/satori/go.uuid v1.2.0
	go.opencensus.io v0.22.6
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/genproto v0.0.0-20200806141610-86f49bd18e98
	google.golang.org/grpc v1.36.0
	google.golang.org/grpc/examples v0.0.0-20210517220359-39015b9c5e19 // indirect
	google.golang.org/protobuf v1.25.0
	gorm.io/datatypes v1.0.1
	gorm.io/gorm v1.21.9
)

replace github.com/infobloxopen/atlas-app-toolkit => github.com/edhaight/atlas-app-toolkit v1.1.0-alpha

go 1.13
