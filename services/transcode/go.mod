module github.com/romashorodok/stream-source/services/transcode

go 1.19

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.1.0
	github.com/romashorodok/stream-source v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.30.0
)

replace github.com/romashorodok/stream-source => ../../
