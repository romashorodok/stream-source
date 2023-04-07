package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	uploadpb "github.com/romashorodok/stream-source/pb/go/upload/v1"
)

const HOST = "localhost:9898"

func main() {
	lis, err := net.Listen("tcp", HOST)

	if err != nil {
		log.Printf("Failed to listen: %v\n", HOST)
		log.Printf("ERROR: %v\n", err)
	}

	minioPool, err := NewMinioPool(5, &MinioCredentials{
		user:     "minioadmin",
		password: "minioadmin",
		endpoint: "localhost:9000",
	})

	if err != nil {
		log.Panic("Cannot init minio pool clients. Error: ", err)
	}

	minioService := &MinioService{
		pool: minioPool,
	}

	uploadService := &UploadService{
		miniosvc: minioService,
	}

	server := grpc.NewServer()
	uploadpb.RegisterUploadServiceServer(server, uploadService)

	log.Printf("Listen on %v\n", HOST)
	server.Serve(lis)
}
