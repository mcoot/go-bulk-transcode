package main

import (
	"flag"
	"fmt"

	"github.com/mcoot/go-bulk-transcode/internal/ffmpeg"
)

var (
	inputFile  = flag.String("i", "", "input file")
	outputFile = flag.String("o", "", "output file")
)

func main() {
	flag.Parse()

	t := ffmpeg.DefaultTranscoder()

	err := t.Transcode(*inputFile, *outputFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("done")
}
