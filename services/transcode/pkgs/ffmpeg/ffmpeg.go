package ffmpeg

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/google/uuid"
	"github.com/romashorodok/stream-source/services/transcode/pkgs/ffmpeg/fragments"
)

type FFMpeg struct {
	Input string
	Pipe  *string

	Codec     string
	Muxer     string
	Bitrate   string
	Fragments []fragments.FFMpegFragment
}

/**
 * Create filesystem pipe to pass bytes into that and consumer may read it from that
 * It helps pass data throough different process.
 * And in my case it prevent saving ffmpeg files into disk, because i need pack files for streaming
 */
func (s *FFMpeg) NewPipe(dirpath string) error {
	pipePath := filepath.Join(dirpath,
		fmt.Sprintf("pipe-%s.fifo", uuid.New()),
	)

	if err := syscall.Mkfifo(pipePath, 0666); err != nil {
		fmt.Println("Error creating data pipe:", err)
		return errors.New("cannot create pipe")
	}

	s.Pipe = &pipePath

	return nil
}

func (s *FFMpeg) DelPipe() error {
	return os.Remove(*s.Pipe)
}

func (s *FFMpeg) NewProcess(sourcefile *string) *exec.Cmd {
	cmd := exec.Command("ffmpeg", fragments.INPUT, *sourcefile)

	for _, fragment := range s.Fragments {
		command, value := fragment.GetFragment()

		cmd.Args = append(cmd.Args, command, value)
	}

	cmd.Args = append(cmd.Args,
		fragments.MUXER, s.Muxer,
		fragments.BITRATE, s.Bitrate,
		fragments.CODEC, s.Codec,
		fragments.PIPE,
	)

	log.Println("Command for ffmpeg", cmd.Args)

	return cmd
}

func (s *FFMpeg) Run() error {
	process := s.NewProcess(&s.Input)

	pipe, err := os.OpenFile(*s.Pipe, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		log.Println("Error opening named pipe:", err)
		return err
	}
	defer pipe.Close()

	process.Stdout = pipe

	if err = process.Start(); err != nil {
		log.Printf("Something went wrong on fmpeg process for %s", s.Input)
		return err
	}

	return nil
}
