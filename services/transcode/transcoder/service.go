package transcoder

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"sync"

	"github.com/romashorodok/stream-source/services/upload/storage"
)

const (
	WEBM = "webm"
)

const (
	VORBIS = "libvorbis"
	OPUS   = "libopus"
)

const (
	LOW    = "64K"
	NORMAL = "128K"
	HIGHT  = "320K"
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
	minioPool, err := storage.NewMinioPool(1, &storage.MinioCredentials{
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

type FFmpegFragment interface {
	GetFragment() (command string, value string)
}

type FFmpegFragment_NoMetadata struct {
	FFmpegFragment
}

func (*FFmpegFragment_NoMetadata) GetFragment() (string, string) {
	return "-map_metadata", "0:s:0"
}

type FFmpeg struct {
	Name      string
	Codec     string
	Muxer     string
	Bitrate   string
	Fragments []FFmpegFragment
}

const (
	INPUT_SLOT = "-i"
	MUXER_SLOT = "-f"
	CODEC_SLOT = "-c:a"
)

func (s *FFmpeg) NewProcess(sourcefile *string) *exec.Cmd {
	commands := map[string]*string{
		INPUT_SLOT: sourcefile,
		MUXER_SLOT: &s.Muxer,
		CODEC_SLOT: &s.Codec,
	}

	for _, fragment := range s.Fragments {
		command, value := fragment.GetFragment()
		commands[command] = &value
	}

	cmd := exec.Command("ffmpeg")

	for command, value := range commands {
		cmd.Args = append(cmd.Args, command, *value)
	}

	cmd.Args = append(cmd.Args, s.Name)

	log.Println("Command for ffmpeg", cmd.Args)

	return cmd
}

type FFmpegPipeline struct {
	Sourcefile string
	Items      []*FFmpeg
	Packager   string
}

func (s *FFmpegPipeline) Start() {
	var wg sync.WaitGroup

	for _, ffmpeg := range s.Items {
		wg.Add(1)

		go func(ffmpeg *FFmpeg) {
			defer func() {
				wg.Done()
			}()

			process := ffmpeg.NewProcess(&s.Sourcefile)
			if err := process.Start(); err != nil {
				log.Printf("Cannot start ffmpeg process for %s on %s", s.Sourcefile, ffmpeg.Name)
			}
			if err := process.Wait(); err != nil {
				log.Printf("Error on ffmpeg processing for %s on %s", s.Sourcefile, ffmpeg.Name)
			}
		}(ffmpeg)
	}

	wg.Wait()
}

var noMetadata = &FFmpegFragment_NoMetadata{}

func (s *TranscoderService) TranscodeAudio(t *TranscodeData) error {
	url, err := s.miniosvc.GetObjectURL(*s.ctx, t.Bucket, t.OriginFile)

	if err != nil {
		log.Println("Cannot reach input file")
		return err
	}

	title := "testsong"

	names := []string{
		fmt.Sprintf("%s-%s.%s", title, LOW, WEBM),
		fmt.Sprintf("%s-%s.%s", title, NORMAL, WEBM),
		fmt.Sprintf("%s-%s.%s", title, HIGHT, WEBM),
	}

	ffmpeg := &FFmpegPipeline{
		Sourcefile: url.String(),

		Items: []*FFmpeg{
			{
				Name:      names[0],
				Codec:     VORBIS,
				Muxer:     WEBM,
				Bitrate:   LOW,
				Fragments: []FFmpegFragment{},
			},

			{
				Name:      names[1],
				Codec:     VORBIS,
				Muxer:     WEBM,
				Bitrate:   NORMAL,
				Fragments: []FFmpegFragment{},
			},

			{
				Name:      names[2],
				Codec:     VORBIS,
				Muxer:     WEBM,
				Bitrate:   HIGHT,
				Fragments: []FFmpegFragment{},
			},
		},
	}

	ffmpeg.Start()

	return nil
}
