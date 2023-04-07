package main

import (
	"context"
	"log"

	uploadpb "github.com/romashorodok/stream-source/pb/go/upload/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UploadService struct {
	uploadpb.UnimplementedUploadServiceServer

	miniosvc *MinioService
}

func (s *UploadService) GetPresignURL(ctx context.Context, in *uploadpb.GetPresignURLRequest) (*uploadpb.GetPresignURLResponse, error) {

	url, err := s.miniosvc.GetPresignURL(ctx, "test", "test.png")

	if err != nil {
		log.Println(err.Error())
		return nil, status.Error(codes.Unavailable, "Cannot get upload url")
	}

	return &uploadpb.GetPresignURLResponse{
		Url: url.String(),
	}, nil
}

