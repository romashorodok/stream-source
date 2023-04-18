package services

import (
	"context"
	"errors"
	"log"
	"strings"

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

	DB       *gorm.DB
	Miniosvc *storage.MinioService
}

func (s *AudioService) ListAudios(ctx context.Context, in *audiopb.ListAudiosRequest) (*audiopb.ListAudiosResponse, error) {
	var audios []types.Audio

	s.DB.Joins("JOIN audio_buckets ON audio_buckets.audio_bucket_id = audios.audio_bucket_id").
		Where("audio_buckets.manifest IS NOT NULL").
		Distinct().
		Preload("AudioBucket").
		Find(&audios)

	var resp []*audiopb.ListAudiosResponse_AudioWithManifest

	for _, audio := range audios {
		bucket := audio.AudioBucket

		resp = append(resp, &audiopb.ListAudiosResponse_AudioWithManifest{
			Audio:    &audiopb.Audio{AudioId: audio.AudioId.String(), Title: audio.Title},
			Manifest: "/" + strings.Join([]string{bucket.Bucket, bucket.Manifest}, "/"),
		})
	}

	return &audiopb.ListAudiosResponse{Audios: resp}, nil
}

func (s *AudioService) CreateAudioBucket(ctx context.Context, in *audiopb.CreateAudioBucketRequest) (*audiopb.CreateAudioBucketResponse, error) {
	client := s.Miniosvc.Pool.Client()
	defer s.Miniosvc.Pool.Put(client)

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

	s.Miniosvc.CreateBucketIfNotExist(ctx, client, bucket)

	if err != nil {
		return nil, status.Error(codes.Unavailable, "Cannot create bucket")
	}

	audioBucket := &types.AudioBucket{Bucket: bucket, AudioId: nil}
	s.DB.Create(audioBucket)

	return &audiopb.CreateAudioBucketResponse{AudioBucket: audioBucket.Proto()}, nil
}

func (s *AudioService) BindAudioToBucket(ctx context.Context, in *audiopb.BindAudioToBucketRequest) (*audiopb.BindAudioToBucketResponse, error) {
	reqBucket := &types.AudioBucket{}
	reqBucket = reqBucket.FromProto(in.GetBucket())
	bucket := &types.AudioBucket{}

	err := s.DB.Transaction(func(q *gorm.DB) error {

		if err := q.Where("audio_bucket_id = ?", reqBucket.AudioBucketId).First(bucket).Error; err != nil {
			return err
		}

		if containAudio := bucket.AudioId; containAudio != nil {
			return status.Error(codes.AlreadyExists, "audio and bucket alredy binded")
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

	return &audiopb.BindAudioToBucketResponse{
		Bucket: bucket.Proto(),
	}, nil
}
