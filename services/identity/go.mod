module github.com/romashorodok/stream-source/services/identity

go 1.19

replace github.com/romashorodok/stream-source => ../../

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.3.0
	github.com/romashorodok/stream-source v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.8.0
	google.golang.org/grpc v1.54.0
	gorm.io/driver/postgres v1.5.0
	gorm.io/gorm v1.25.0
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.3.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20230331144136-dcfb400f0633 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
)
