package main

import (
	"fmt"
	"log"
	"net"

	identitypb "github.com/romashorodok/stream-source/pb/go/identity/v1"
	"github.com/romashorodok/stream-source/services/identity/jwt"
	"github.com/romashorodok/stream-source/services/identity/services"
	"github.com/romashorodok/stream-source/services/identity/stores"
	"github.com/romashorodok/stream-source/services/identity/types"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"google.golang.org/grpc"
)

const (
	HOST = "localhost:9393"
)

func main() {
	lis, err := net.Listen("tcp", HOST)

	if err != nil {
		log.Printf("Failed to listen: %v\n", HOST)
		log.Printf("ERROR: %v\n", err)
	}

	user := "user"
	password := "user"
	dbname := "postgresdb"
	port := "5432"

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Kiev", user, password, dbname, port),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})

	if err != nil {
		log.Panic("Unable connect to pgdb. Error:", err)
	}

	db.AutoMigrate(&types.User{})
	db.AutoMigrate(&types.UserAuthentication{})

	userstore := &stores.UserStoreGORM{DB: db}
	jwt := &jwt.JWT{
		SecretKey:                 "MySecretKey",
		AccessTokenDurationInSec:  120,
		RefreshTokenDurationInSec: 604800,
	}

	identitysvc := &services.IdentityService{UserStore: userstore, JWT: jwt, DB: db}

	server := grpc.NewServer()
	identitypb.RegisterIdentityServiceServer(server, identitysvc)
	log.Printf("Listen on %v\n", HOST)
	server.Serve(lis)
}
