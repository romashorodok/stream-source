package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	audiopb "github.com/romashorodok/stream-source/pb/go/audio/v1"
	uploadpb "github.com/romashorodok/stream-source/pb/go/upload/v1"
	"github.com/romashorodok/stream-source/services/upload/storage"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const HOST = "localhost:9898"
const KAFKA = "localhost:9092"
const REGISTRY = "localhost:8081"
const AUDIO_HOST = "localhost:9292"

func main() {
	lis, err := net.Listen("tcp", HOST)

	if err != nil {
		log.Printf("Failed to listen: %v\n", HOST)
		log.Printf("ERROR: %v\n", err)
	}

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": KAFKA})

	if err != nil {
		log.Panic(err)
	}

	defer p.Close()

	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					log.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	minioPool, err := storage.NewMinioPool(5, &storage.MinioCredentials{
		User:     "minioadmin",
		Password: "minioadmin",
		Endpoint: "localhost:9000",
	})

	if err != nil {
		log.Panic("Cannot init minio pool clients. Error: ", err)
	}

	minioService := &storage.MinioService{
		Pool: minioPool,
	}

	audioOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	audioConn, err := grpc.Dial(AUDIO_HOST, audioOpts...)
	if err != nil {
		log.Panicln("Failed to connect to audio service. Error:", err)
	}
	defer audioConn.Close()

	audioClient := audiopb.NewAudioServiceClient(audioConn)

	uploadService := &UploadService{
		miniosvc: minioService,
		audiosvc: audioClient,
		producer: p,
	}

	server := grpc.NewServer()
	uploadpb.RegisterUploadServiceServer(server, uploadService)

	log.Printf("Listen on %v\n", HOST)
	server.Serve(lis)
}
