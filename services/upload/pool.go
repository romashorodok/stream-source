package main

import (
	"log"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioCredentials struct {
	user     string
	password string
	endpoint string
}

type MinioPool struct {
	clients []*minio.Client
	mutex   sync.Mutex
}

func NewMinioPool(size int, creds *MinioCredentials) (*MinioPool, error) {
	pool := &MinioPool{
		clients: make([]*minio.Client, size),
	}

	for i := 0; i < size; i++ {
		client, err := minio.New(creds.endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(creds.user, creds.password, ""),
			Secure: false,
		})

		if err != nil {
			log.Println("Cannot open minio client in pool: ", err)
			return nil, err
		}

		pool.clients[i] = client
	}

	return pool, nil
}

func (p *MinioPool) Client() *minio.Client {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	client := p.clients[0]
	p.clients = p.clients[1:]

	return client
}

func (p *MinioPool) Put(client *minio.Client) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.clients = append(p.clients, client)
}
