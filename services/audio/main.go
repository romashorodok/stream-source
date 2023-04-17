package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	audiopb "github.com/romashorodok/stream-source/pb/go/audio/v1"
	audiotopicpb "github.com/romashorodok/stream-source/pb/go/topic/audio/v1"
	"github.com/romashorodok/stream-source/pkgs/consumer"
	"github.com/romashorodok/stream-source/services/audio/services"
	"github.com/romashorodok/stream-source/services/audio/types"
	"github.com/romashorodok/stream-source/services/audio/workers"
	"github.com/romashorodok/stream-source/services/upload/storage"

	"google.golang.org/grpc"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	HOST  = "localhost:9292"
	KAFKA = "localhost:9092"
)

var creds *storage.MinioCredentials = &storage.MinioCredentials{
	User:     "minioadmin",
	Password: "minioadmin",
	Endpoint: "localhost:9000",
}

var kafkaConfig = &kafka.ConfigMap{
	"bootstrap.servers": KAFKA,
	"group.id":          "upload-consumers",
	"auto.offset.reset": "earliest",
}

func main() {
	user := "user"
	password := "user"
	dbname := "postgresdb"
	port := "5432"

	minioPool, err := storage.NewMinioPool(5, creds)

	if err != nil {
		log.Panic("Cannot init minio pool clients. Error: ", err)
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Kiev", user, password, dbname, port),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})

	if err != nil {
		log.Panic("Unable connect to pgdb. Error: ", err)
	}

	db.AutoMigrate(&types.Audio{})
	db.AutoMigrate(&types.AudioBucket{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lis, err := net.Listen("tcp", HOST)

	if err != nil {
		log.Printf("Failed to listen: %v\n", HOST)
		log.Printf("ERROR: %v\n", err)
	}

	transcodedAudioChan := make(chan *consumer.ConsumerBox[*audiotopicpb.AudioTranscoded])
	go consumer.ConsumeProtobufTopic(kafkaConfig, audiotopicpb.Default_AudioTranscoded_Topic, transcodedAudioChan)

	transcodedAudioWorker := &workers.TranscodedAudioWorker{DB: db}
	go transcodedAudioWorker.ProcessTranscodedAudio(ctx, transcodedAudioChan, 4)

	minioService := &storage.MinioService{Pool: minioPool}

	audiosvc := &services.AudioService{DB: db, Miniosvc: minioService}

	server := grpc.NewServer()
	audiopb.RegisterAudioServiceServer(server, audiosvc)
	server.Serve(lis)
}
