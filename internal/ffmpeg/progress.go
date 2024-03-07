package ffmpeg

import (
	"fmt"
	"strings"
	"time"
)

type progressState struct {
	latestFrame string
	outTime     string
	fps         string
	status      string
}

func updateProgressState(line string, state *progressState) {
	rawKey, rawVal, found := strings.Cut(line, "=")
	if !found {
		return
	}
	key := strings.TrimSpace(rawKey)
	val := strings.TrimSpace(rawVal)

	switch key {
	case "frame":
		state.latestFrame = val
	case "out_time":
		state.outTime = val
	case "fps":
		state.fps = val
	case "progress":
		state.status = val
	}
}

func setupProgressOutputHandler(stdout chan string, done chan bool) {
	state := progressState{
		latestFrame: "0",
		outTime:     "00:00:00.000000",
		fps:         "0",
		status:      "",
	}

	ticker := time.NewTicker(5 * time.Second)

	// Read process output to update the state and output at intervals
	go func() {
		for {
			select {
			case line := <-stdout:
				updateProgressState(line, &state)
			case <-ticker.C:
				fmt.Printf("frame: %s, time: %s, fps: %s, status: %s\n", state.latestFrame, state.outTime, state.fps, state.status)
			case <-done:
				ticker.Stop()
				close(stdout)
				return
			}
		}
	}()

}
