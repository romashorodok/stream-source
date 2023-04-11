module github.com/romashorodok/stream-source/services/transcode

go 1.19

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.1.0
	github.com/minio/minio-go/v7 v7.0.51
	github.com/romashorodok/stream-source v0.0.0-00010101000000-000000000000
	github.com/romashorodok/stream-source/services/upload v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.30.0
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.16.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/rs/xid v1.4.0 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	golang.org/x/crypto v0.6.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
)

replace github.com/romashorodok/stream-source => ../../

replace github.com/romashorodok/stream-source/services/upload => ../upload/
