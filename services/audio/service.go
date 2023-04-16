package main

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	audiopb "github.com/romashorodok/stream-source/pb/go/audio/v1"
	"github.com/romashorodok/stream-source/services/audio/types"
	"github.com/romashorodok/stream-source/services/upload/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type AudioService struct {
	audiopb.UnimplementedAudioServiceServer

	db       *gorm.DB
	miniosvc *storage.MinioService
}

func (s *AudioService) CreateAudioBucket(ctx context.Context, in *audiopb.CreateAudioBucketRequest) (*audiopb.CreateAudioBucketResponse, error) {
	client := s.miniosvc.Pool.Client()
	defer s.miniosvc.Pool.Put(client)

	bucket := uuid.New().String()

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, status.Error(codes.Unavailable, "Cannot check if bucket exists")
	}

	for exists {
		bucket = uuid.New().String()
		exists, err = client.BucketExists(ctx, bucket)
		if err != nil {
			return nil, status.Error(codes.Unavailable, "Cannot check if bucket exists")
		}
	}

	s.miniosvc.CreateBucketIfNotExist(ctx, client, bucket)

	if err != nil {
		return nil, status.Error(codes.Unavailable, "Cannot create bucket")
	}

	audioBucket := &types.AudioBucket{Bucket: bucket, AudioId: nil}
	s.db.Create(audioBucket)

	return &audiopb.CreateAudioBucketResponse{AudioBucket: audioBucket.Proto()}, nil
}

func (s *AudioService) BindAudioToBucket(ctx context.Context, in *audiopb.BindAudioToBucketRequest) (*audiopb.BindAudioToBucketResponse, error) {

	err := s.db.Transaction(func(q *gorm.DB) error {

		reqBucket := &types.AudioBucket{}
		reqBucket = reqBucket.FromProto(in.GetBucket())
		bucket := &types.AudioBucket{}

		if err := q.Where("audio_bucket_id = ?", reqBucket.AudioBucketId).First(bucket).Error; err != nil {
			return err
		}

		if containAudio := bucket.AudioId; containAudio != nil {
			return errors.New("audio and bucket alredy binded")
		}

		if bucket.AudioBucketId != reqBucket.AudioBucketId || bucket.Bucket != reqBucket.Bucket {
			return errors.New("data was modified by user")
		}

		audio := &types.Audio{}
		audio = audio.FromProto(in.GetAudio())

		if err := q.Create(audio).Error; err != nil {
			return err
		}

		bucket.AudioId = &audio.AudioId
		bucket.OriginFile = reqBucket.OriginFile

		audio.AudioBucketId = &bucket.AudioBucketId

		q.Updates(bucket)
		q.Updates(audio)

		return nil
	})

	if err != nil {
		log.Println("Transaction was failed. Error:", err)
		return nil, err
	}

	return &audiopb.BindAudioToBucketResponse{}, nil
}
