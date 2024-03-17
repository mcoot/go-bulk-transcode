package ffmpeg

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
)

var (
	ValidExtensions = []string{".mp4", ".mkv"}
)

type videoConfig struct {
	resX int
	resY int
	crf  int
}

type noiseReductionConfig struct {
	enabled      bool
	reductionDb  int
	noiseFloorDb int
}

type audioConfig struct {
	bitrate           string
	micVolume         float64
	micNoiseReduction noiseReductionConfig
}

type Transcoder struct {
	video videoConfig
	audio audioConfig
}

func DefaultTranscoder() *Transcoder {
	return &Transcoder{
		video: videoConfig{
			resX: 2560,
			resY: 1440,
			crf:  24,
		},
		audio: audioConfig{
			bitrate:   "192k",
			micVolume: 0.85,
			micNoiseReduction: noiseReductionConfig{
				enabled:      true,
				reductionDb:  20,
				noiseFloorDb: -40,
			},
		},
	}
}

func (t *Transcoder) buildFilterComplex() string {
	noiseReductionFilter := ""
	if t.audio.micNoiseReduction.enabled {
		noiseReductionFilter = fmt.Sprintf("afftdn=nr=%d:nf=%d:tn=1[n];[n]", t.audio.micNoiseReduction.reductionDb, t.audio.micNoiseReduction.noiseFloorDb)
	}
	return fmt.Sprintf("[0:a:1]%svolume=%.2f[l];[0:a:0][l]amerge=inputs=2[a]", noiseReductionFilter, t.audio.micVolume)
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
		"-vf", fmt.Sprintf("scale=%d:%d", t.video.resX, t.video.resY),
		// Quality factor
		"-crf", fmt.Sprintf("%d", t.video.crf),
		// Slow preset for better filesize
		"-preset", "slow",
		// Pixel format
		"-pix_fmt", "yuv420p",

		// Audio encode in aac
		"-c:a", "aac",
		// 192k bitrate
		"-b:a", t.audio.bitrate,
		// Two audio channels
		"-ac", "2",

		// Complex filter to merge mic into main audio at lower volume
		"-filter_complex", "[0:a:1]afftdn=nr=20:nf=-40:tn=1[n];[n]volume=0.85[l];[0:a:0][l]amerge=inputs=2[a]",
		// Map the complex filter to the output
		"-map", "0:v:0",
		"-map", "[a]",

		// Overwrite
		"-y",
		// Log only errors
		"-loglevel", "error",
		// Progress...
		"-progress", "pipe:1",

		// Output file
		outputFile,
	}
}

func (t *Transcoder) Transcode(inputFile, outputFile string) error {
	// Create the command to execute and hook up stderr normally
	cmd := exec.Command("ffmpeg", t.buildArguments(inputFile, outputFile)...)
	cmd.Stderr = os.Stderr

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
	setupProgressOutputHandler(stdoutChan, done)

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
