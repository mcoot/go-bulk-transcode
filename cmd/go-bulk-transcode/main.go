package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/mcoot/go-bulk-transcode/internal/ffmpeg"
)

var (
	inputDir  = flag.String("i", "", "input directory")
	outputDir = flag.String("o", "", "output directory")
)

func findFilenames(dir string) ([]string, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if !slices.Contains(ffmpeg.ValidExtensions, ext) {
			continue
		}
		files = append(files, entry.Name())
	}
	return files, nil
}

func main() {
	flag.Parse()

	inputFiles, err := findFilenames(*inputDir)
	if err != nil {
		panic(err)
	}

	t := ffmpeg.DefaultTranscoder()

	for i, inputFilename := range inputFiles {
		inputFile := filepath.Join(*inputDir, inputFilename)
		outputFile := filepath.Join(*outputDir, inputFilename)
		fmt.Printf("[%d/%d] Transcoding %s to %s\n", i+1, len(inputFiles), inputFile, outputFile)
		err = t.Transcode(inputFile, outputFile)
		if err != nil {
			panic(err)
		}
	}
}
