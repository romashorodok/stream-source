package main

import (
	"fmt"
	"log"
	"net"

	audiopb "github.com/romashorodok/stream-source/pb/go/audio/v1"
	"github.com/romashorodok/stream-source/services/audio/types"
	"github.com/romashorodok/stream-source/services/upload/storage"

	"google.golang.org/grpc"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const HOST = "localhost:9292"

var creds *storage.MinioCredentials = &storage.MinioCredentials{
	User:     "minioadmin",
	Password: "minioadmin",
	Endpoint: "localhost:9000",
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

	lis, err := net.Listen("tcp", HOST)

	if err != nil {
		log.Printf("Failed to listen: %v\n", HOST)
		log.Printf("ERROR: %v\n", err)
	}

	minioService := &storage.MinioService{Pool: minioPool}

	audiosvc := &AudioService{db: db, miniosvc: minioService}

	server := grpc.NewServer()
	audiopb.RegisterAudioServiceServer(server, audiosvc)
	server.Serve(lis)
}
