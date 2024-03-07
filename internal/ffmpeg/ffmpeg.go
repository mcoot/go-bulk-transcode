package ffmpeg

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
)

type Transcoder struct {
	resolution string
	crf        string
}

func DefaultTranscoder() *Transcoder {
	return &Transcoder{
		resolution: "2560:1440",
		crf:        "24",
	}
}

func (t *Transcoder) buildArguments(inputFile, outputFile string) []string {
	return []string{
		// Input file
		"-i", inputFile,

		// Faststart for quicker upload
		"-movflags", "+faststart",

		// Video encode to h.264
		"-c:v", "libx264",
		// 1440p resolution
		"-vf", fmt.Sprintf("scale=%s", t.resolution),
		// Quality factor
		"-crf", t.crf,
		// Slow preset for better filesize
		"-preset", "slow",
		// Pixel format
		"-pix_fmt", "yuv420p",

		// Audio encode in aac
		"-c:a", "aac",
		// 192k bitrate
		"-b:a", "192k",
		// Two audio channels
		"-ac", "2",

		// Complex filter to merge mic into main audio at lower volume
		"-filter_complex", "[0:a:1]volume=0.8[l];[0:a:0][l]amerge=inputs=2[a]",
		// Map the complex filter to the output
		"-map", "0:v:0",
		"-map", "[a]",

		// Overwrite
		"-y",

		// Progress...
		"-progress", "pipe:1",

		// Output file
		outputFile,
	}
}

func (t *Transcoder) Transcode(inputFile, outputFile string) error {
	// Create the command to execute and hook up stderr normally
	cmd := exec.Command("ffmpeg", t.buildArguments(inputFile, outputFile)...)

	// Set up its stdout to route to a channel so we can parse progress
	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stdoutChan := setupPipeToChannel(stdoutReader)

	done := make(chan bool)
	defer func() {
		close(done)
	}()
	go handleProgressOutput(stdoutChan, done)

	return cmd.Run()
}

func setupPipeToChannel(r io.ReadCloser) chan string {
	scanner := bufio.NewScanner(r)
	c := make(chan string)
	go func() {
		for scanner.Scan() {
			c <- scanner.Text()
		}
	}()
	return c
}
