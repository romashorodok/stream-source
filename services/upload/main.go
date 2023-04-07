package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

    "github.com/romashorodok/stream-source/services/upload/storage"

	uploadpb "github.com/romashorodok/stream-source/pb/go/upload/v1"
)

const HOST = "localhost:9898"

func main() {
	lis, err := net.Listen("tcp", HOST)

	if err != nil {
		log.Printf("Failed to listen: %v\n", HOST)
		log.Printf("ERROR: %v\n", err)
	}

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

	uploadService := &UploadService{
		miniosvc: minioService,
	}

	server := grpc.NewServer()
	uploadpb.RegisterUploadServiceServer(server, uploadService)

	log.Printf("Listen on %v\n", HOST)
	server.Serve(lis)
}
