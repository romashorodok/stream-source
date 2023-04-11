package transcoder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
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
	Pipe      *string
}

const (
	INPUT_SLOT   = "-i"
	MUXER_SLOT   = "-f"
	CODEC_SLOT   = "-c:a"
	BITRATE_SLOT = "-b:a"
)

func (s *FFmpeg) NewProcess(sourcefile *string) *exec.Cmd {

	cmd := exec.Command(
		"ffmpeg", "-y",
		INPUT_SLOT, *sourcefile,
	)

	for _, fragment := range s.Fragments {
		command, value := fragment.GetFragment()
		cmd.Args = append(cmd.Args, command, value)
	}

	cmd.Args = append(
		cmd.Args,
		"-map", "0",
		"-ldash", "1",
		MUXER_SLOT, s.Muxer,
		BITRATE_SLOT, s.Bitrate,
		CODEC_SLOT, s.Codec,
		"pipe:",
	)

	log.Println("Command for ffmpeg", cmd.Args)

	return cmd
}

type FFmpegPipeline struct {
	Sourcefile string
	Workdir    string
	Items      []*FFmpeg
	Packager   string
}

func prepPipe(dir string, ffmpeg *FFmpeg) error {
	pipe, err := NewDataPipe(dir)
	if err != nil {
		return errors.New("cannot prepare")
	}
	ffmpeg.Pipe = &pipe
	return nil
}

func readOutput(output io.ReadCloser, out chan string) {
	defer close(out)

	buf := make([]byte, 1024)

	for {
		n, err := output.Read(buf)

		if err != nil {
			if err != io.EOF {
				fmt.Println("Error:", err)
			}
			break
		}

		out <- string(buf[:n])
	}
}

func (s *FFmpegPipeline) Start() {
	// var wg sync.WaitGroup
	var pwg sync.WaitGroup
	pwg.Add(1)

	for _, ffmpeg := range s.Items {
		if err := prepPipe(s.Workdir, ffmpeg); err != nil {
			log.Printf("cannot prepare pipe for %s on %s", s.Sourcefile, ffmpeg.Name)
			return
		}
	}

	go func() {
		defer func() {
			pwg.Done()
		}()

		packager := exec.Command("packager")

		for _, ffmpeg := range s.Items {
			init := fmt.Sprintf("init_segment=%s/%s", s.Workdir, ffmpeg.Name)
			segment := fmt.Sprintf("segment_template=%s/$Number$-%s", s.Workdir, ffmpeg.Name)
			in := fmt.Sprintf("in=%s,stream=audio,%s,%s", *ffmpeg.Pipe, init, segment)
			packager.Args = append(packager.Args, in)
		}

		packager.Args = append(packager.Args, "--mpd_output", fmt.Sprintf("%s/%s", s.Workdir, "manifest.mpd"))
		packager.Args = append(packager.Args, "--segment_duration", "4")
		packager.Args = append(packager.Args, "--min_buffer_time", "4")

		log.Println(packager.Args)

		stdout, _ := packager.StdoutPipe()
		stderr, _ := packager.StderrPipe()

		out := make(chan string, 1024)
		errCh := make(chan string, 1024)

		go readOutput(stdout, out)
		go readOutput(stderr, errCh)

		go func() {
			for line := range errCh {
				fmt.Println(line)
			}
		}()

		go func() {
			for line := range out {
				fmt.Println(line)
			}
		}()

		if err := packager.Start(); err != nil {
			log.Println("Start ", err)
		}
		if err := packager.Wait(); err != nil {
			// log.Printf("Error on ffmpeg processing for %s on %s", s.Sourcefile, ffmpeg.Name)
			log.Println("Wait ", err)
		}
	}()

	for _, ffmpeg := range s.Items {
		// wg.Add(1)

		go func(ffmpeg *FFmpeg) {
			defer func() {
				// wg.Done()
			}()

			process := ffmpeg.NewProcess(&s.Sourcefile)

			stdout, _ := process.StdoutPipe()
			stderr, _ := process.StderrPipe()

			pipe, err := os.OpenFile(*ffmpeg.Pipe, os.O_WRONLY, os.ModeNamedPipe)
			if err != nil {
				fmt.Println("Error opening named pipe:", err)
				return
			}
			defer pipe.Close()
			process.Stdout = pipe

			out := make(chan string, 1024)
			errCh := make(chan string, 1024)

			go readOutput(stdout, out)
			go readOutput(stderr, errCh)

			go func() {
				for line := range errCh {
					fmt.Println(line)
				}
			}()

			go func() {
				for line := range out {
					fmt.Println(line)
				}
			}()

			if err := process.Start(); err != nil {
				log.Printf("Cannot start ffmpeg process for %s on %s", s.Sourcefile, ffmpeg.Name)
			}
			if err := process.Wait(); err != nil {
				// log.Printf("Error on ffmpeg processing for %s on %s", s.Sourcefile, ffmpeg.Name)
				log.Println(err)
			}
		}(ffmpeg)
	}

	pwg.Wait()

	log.Println("Success")
}

func NewDataPipe(dir string) (string, error) {
	pipePath := filepath.Join(dir, fmt.Sprintf("pipe-%s.fifo", uuid.New()))
	if err := syscall.Mkfifo(pipePath, 0666); err != nil {
		fmt.Println("Error creating data pipe:", err)
		return "", errors.New("cannot create pipe")
	}
	return pipePath, nil
}

var noMetadata = &FFmpegFragment_NoMetadata{}
var signals = make(chan os.Signal, 1)

func (s *TranscoderService) TranscodeAudio(t *TranscodeData) error {
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	title := "testsong"

	dir, err := os.MkdirTemp("", fmt.Sprintf("%s-*", title))
	if err != nil {
		log.Println("Cannot create temp dir")
	}
	defer os.RemoveAll(dir)

	go func() {
		<-signals
		os.RemoveAll(dir)
		log.Printf("Termination signal received. Cleen up transcode resources. For:\n %s", dir)
	}()

	url, err := s.miniosvc.GetObjectURL(*s.ctx, t.Bucket, t.OriginFile)

	if err != nil {
		log.Println("Cannot reach input file")
		return err
	}

	names := []string{
		fmt.Sprintf("%s-%s.%s", title, LOW, WEBM),
		fmt.Sprintf("%s-%s.%s", title, NORMAL, WEBM),
		fmt.Sprintf("%s-%s.%s", title, HIGHT, WEBM),
	}

	parts := strings.Split(url.String(), "?")
	urlWithoutQuery := parts[0]

	ffmpeg := &FFmpegPipeline{
		Workdir:    dir,
		Sourcefile: urlWithoutQuery,

		Items: []*FFmpeg{
			{
				Name:      names[0],
				Codec:     VORBIS,
				Muxer:     WEBM,
				Bitrate:   LOW,
				Fragments: []FFmpegFragment{noMetadata},
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

	time.Sleep(10 * time.Minute)

	return nil
}
