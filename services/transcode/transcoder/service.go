package transcoder

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	audiotopicpb "github.com/romashorodok/stream-source/pb/go/topic/audio/v1"
	"github.com/romashorodok/stream-source/services/transcode/pkgs/ffmpeg"
	"github.com/romashorodok/stream-source/services/transcode/pkgs/ffmpeg/codecs"
	"github.com/romashorodok/stream-source/services/transcode/pkgs/ffmpeg/fragments"
	"github.com/romashorodok/stream-source/services/upload/storage"
)

type TranscodeData struct {
	Bucket     string
	BucketId   string
	OriginFile string
}

type TranscoderService struct {
	ctx      *context.Context
	miniosvc *storage.MinioService
	producer *kafka.Producer
}

var creds *storage.MinioCredentials = &storage.MinioCredentials{
	User:     "minioadmin",
	Password: "minioadmin",
	Endpoint: "localhost:9000",
}

func NewTranscoderService(ctx *context.Context, producer *kafka.Producer) *TranscoderService {
	minioPool, err := storage.NewMinioPool(4, creds)

	if err != nil {
		log.Panic("Cannot init minio pool clients. Error: ", err)
	}

	return &TranscoderService{
		ctx:      ctx,
		miniosvc: &storage.MinioService{Pool: minioPool},
		producer: producer,
	}
}

func (s *TranscoderService) TranscodeAudio(t *TranscodeData) error {
	url, err := s.miniosvc.GetObjectURL(*s.ctx, t.Bucket, t.OriginFile)
	if err != nil {
		log.Println("Cannot reach input file")
		return err
	}
	dir, err := os.MkdirTemp("", fmt.Sprintf("%s-*", uuid.New()))
	if err != nil {
		log.Println("Cannot create temp dir")
	}
	defer os.RemoveAll(dir)

	parts := strings.Split(url.String(), "?")
	urlWithoutQuery := parts[0]
	manifest := fmt.Sprintf("%s.mpd", uuid.New())

	pipeline := &ffmpeg.FFMpegProcessingPipeline{
		Sourcefile: urlWithoutQuery,
		Manifest:   manifest,
		Workdir:    dir,

		Items: []*ffmpeg.FFMpeg{
			{
				Codec:     codecs.VORBIS,
				Muxer:     fragments.MUXER_WEBM,
				Bitrate:   fragments.BITRATE_LOW,
				Fragments: []fragments.FFMpegFragment{&fragments.NoMetadata{}, &fragments.EchoEffect{}},
			},
			{
				Codec:     codecs.VORBIS,
				Muxer:     fragments.MUXER_WEBM,
				Bitrate:   fragments.BITRATE_NORMAL,
				Fragments: []fragments.FFMpegFragment{&fragments.NoMetadata{}},
			},
			{
				Codec:     codecs.VORBIS,
				Muxer:     fragments.MUXER_WEBM,
				Bitrate:   fragments.BITRATE_HIGHT,
				Fragments: []fragments.FFMpegFragment{&fragments.NoMetadata{}},
			},
		},
	}

	if err = pipeline.Run(); err != nil {
		log.Println("Something goes wrong on pipeline", err)
	}

	for _, ffmpeg := range pipeline.Items {
		_ = ffmpeg.DelPipe()
	}

	s.DeliverFiles(dir, t.Bucket)

	topic := audiotopicpb.Default_AudioTranscoded_Topic

	transcoded := &audiotopicpb.AudioTranscoded{
		ManifestFile: &manifest,
		Bucket:       &t.Bucket,
		BucketId:     &t.BucketId,
	}

	transcodedBytes, err := proto.Marshal(transcoded)

	if err != nil {
		log.Println("Failed to serialize transcoded topic", err)
	}

	s.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          transcodedBytes,
	}, nil)

	return err
}

type FileBox struct {
	path string
	file string
}

func (s *TranscoderService) DeliverFiles(workdir, bucket string) {
	clientPool := storage.AsyncMinioPool(10, creds)
	files := make(chan *FileBox, 20)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		filepath.Walk(workdir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				folder := fmt.Sprintf("%s/", workdir)
				file := strings.SplitAfter(path, folder)

				wg.Add(1)
				files <- &FileBox{file: file[1], path: path}
			}

			return nil
		})

		close(files)
	}()

	go func() {
		defer wg.Done()
		for file := range files {
			go func(file *FileBox) {
				client := <-clientPool
				defer func() {
					clientPool <- client
					wg.Done()
				}()

				client.FPutObject(*s.ctx, bucket, file.file, file.path, minio.PutObjectOptions{})

			}(file)
		}
	}()

	wg.Wait()
}
