package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	uploadpb "github.com/romashorodok/stream-source/pb/go/upload/v1"
)

const HOST = "localhost:9898"

type UploadService struct {
	uploadpb.UnimplementedUploadServiceServer
}

const (
	X_GRPC_WEB = "x-grpc-web"
	X_TOKEN    = "x-token"
)

// var header = metadata.New(map[string]string{"x-grpc-web": "", "x-token": ""})

func (*UploadService) UploadAudioStub(ctx context.Context, req *uploadpb.UploadAudioStubRequest) (*uploadpb.UploadAudioStubResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		log.Println("Cannot parse request meta-data")
	}

	isGrpcWeb := md.Get(X_GRPC_WEB)[0]
	token := md.Get(X_TOKEN)[0]

	log.Println("token is ", token)
	log.Println("is grpc-web ?", isGrpcWeb)

	log.Println(req.StubField)

	return &uploadpb.UploadAudioStubResponse{
		StubField: "Test response",
	}, nil
}

func (s *UploadService) UploadAudioProcess(stream uploadpb.UploadService_UploadAudioProcessServer) error {
	req, _ := stream.Recv()

	fileName := req.GetGetUploadUrl()

	uploadUrl := &uploadpb.UploadURL{
		Url: "my presign upload url will be here",
	}

	resp := &uploadpb.UploadAudioProcessResponse{
		Action: &uploadpb.UploadAudioProcessResponse_UploadUrl{
			UploadUrl: uploadUrl,
		},
	}

	stream.Send(resp)

	log.Println(fileName.FileName)

	return nil
}

func main() {
	lis, err := net.Listen("tcp", HOST)

	if err != nil {
		log.Printf("Failed to listen: %v\n", HOST)
		log.Printf("ERROR: %v\n", err)
	}

	uploadService := &UploadService{}

	server := grpc.NewServer()
	uploadpb.RegisterUploadServiceServer(server, uploadService)

	log.Printf("Listen on %v\n", HOST)
	server.Serve(lis)
}
