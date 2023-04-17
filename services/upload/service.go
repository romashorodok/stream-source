package main

import (
	"context"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/golang/protobuf/proto"
	audiopb "github.com/romashorodok/stream-source/pb/go/audio/v1"
	transcodetopicpb "github.com/romashorodok/stream-source/pb/go/kafka/topic/v1"
	uploadpb "github.com/romashorodok/stream-source/pb/go/upload/v1"
	"github.com/romashorodok/stream-source/services/upload/storage"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UploadService struct {
	uploadpb.UnimplementedUploadServiceServer

	producer *kafka.Producer
	miniosvc *storage.MinioService
	audiosvc audiopb.AudioServiceClient
}

func (s *UploadService) GetPresignURL(ctx context.Context, in *uploadpb.GetPresignURLRequest) (*uploadpb.GetPresignURLResponse, error) {
	url, err := s.miniosvc.GetPresignURL(ctx, in.GetBucket(), in.GetFilename())
	if err != nil {
		log.Println(err.Error())
		return nil, status.Error(codes.Unavailable, "Cannot get upload url")
	}

	return &uploadpb.GetPresignURLResponse{
		Url: url.String(),
	}, nil
}

func (s *UploadService) SuccessAudioUpload(ctx context.Context, in *uploadpb.SuccessAudioUploadRequest) (*uploadpb.SuccessAudioUploadResponse, error) {

	resp, err := s.audiosvc.BindAudioToBucket(ctx, &audiopb.BindAudioToBucketRequest{
		Audio:  in.GetAudio(),
		Bucket: in.GetBucket(),
	})

	if err != nil {
		code, ok := status.FromError(err)

		if !ok {
			err = fmt.Errorf("cannot bind audio to bucket. Error %s", err)
			log.Println(err)
			return nil, err
		}

		switch code.Code() {
		case codes.AlreadyExists:
			return nil, status.Error(codes.AlreadyExists, "Cannot upload twice. Audio and bucket already exists")
		}
	}

	bucket := resp.GetBucket()

	transcode := &transcodetopicpb.TranscodeAudio{
		OriginFile: &bucket.OriginFile,
		Bucket:     &bucket.Bucket,
		BucketId:   &bucket.AudioBucketId,
	}

	transcodeBytes, err := proto.Marshal(transcode)

	if err != nil {
		log.Println("Failed to serialize transcode topic", err)
		return nil, err
	}

	topic := transcodetopicpb.Default_TranscodeAudio_Topic

	s.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          transcodeBytes,
	}, nil)

	return &uploadpb.SuccessAudioUploadResponse{}, nil
}

func (s *UploadService) FailAudioUpload(ctx context.Context, in *uploadpb.FailAudioUploadRequest) (*uploadpb.FailAudioUploadResponse, error) {
	return nil, nil
}
