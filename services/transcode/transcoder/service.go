package transcoder

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/romashorodok/stream-source/services/transcode/pkgs/ffmpeg"
	"github.com/romashorodok/stream-source/services/transcode/pkgs/ffmpeg/codecs"
	"github.com/romashorodok/stream-source/services/transcode/pkgs/ffmpeg/fragments"
	"github.com/romashorodok/stream-source/services/upload/storage"
)

type TranscodeData struct {
	Bucket     string
	OriginFile string
}

type TranscoderService struct {
	ctx      *context.Context
	miniosvc *storage.MinioService
}

func NewTranscoderService(ctx *context.Context) *TranscoderService {
	minioPool, err := storage.NewMinioPool(4, &storage.MinioCredentials{
		User:     "minioadmin",
		Password: "minioadmin",
		Endpoint: "localhost:9000",
	})

	if err != nil {
		log.Panic("Cannot init minio pool clients. Error: ", err)
	}

	return &TranscoderService{
		ctx:      ctx,
		miniosvc: &storage.MinioService{Pool: minioPool},
	}
}

func (s *TranscoderService) TranscodeAudio(t *TranscodeData) error {

	dir, err := os.MkdirTemp("", fmt.Sprintf("%s-*", uuid.New()))
	if err != nil {
		log.Println("Cannot create temp dir")
	}
	defer os.RemoveAll(dir)

	url, err := s.miniosvc.GetObjectURL(*s.ctx, t.Bucket, t.OriginFile)

	if err != nil {
		log.Println("Cannot reach input file")
		return err
	}

	parts := strings.Split(url.String(), "?")
	urlWithoutQuery := parts[0]

	pipeline := &ffmpeg.FFMpegProcessingPipeline{
		Sourcefile: urlWithoutQuery,
		Manifest:   fmt.Sprintf("%s.mpd", uuid.New()),
		Workdir:    dir,

		Items: []*ffmpeg.FFMpeg{
			{
				Codec:     codecs.VORBIS,
				Muxer:     fragments.MUXER_WEBM,
				Bitrate:   fragments.BITRATE_LOW,
				Fragments: []fragments.FFMpegFragment{&fragments.NoMetadata{}},
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

	return err
}
