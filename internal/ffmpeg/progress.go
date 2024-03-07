package ffmpeg

import "fmt"

func handleProgressOutput(stdout chan string, done chan bool) {
	for {
		select {
		case line := <-stdout:
			fmt.Printf("output: %s\n", line)
		case <-done:
			close(stdout)
			return
		}
	}

}
