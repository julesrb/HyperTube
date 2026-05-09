package transcode

import (
	"os/exec"
	"bytes"
	"fmt"
)

// Flags: -movflags frag_keyframe+empty_moov are required so the browser can
// begin playback before the full file is received.
func TranscodeHLS(inputPath string, outputPath string) error {
	var stderr bytes.Buffer
    cmd := exec.Command(
        "ffmpeg",
        "-i", inputPath,
        "-c", "copy",
        "-avoid_negative_ts", "make_zero",
        "-hls_time", "5",
        "-hls_list_size", "0", // keep all segments in the playlist, using a sliding windows reduce the size of the buffer but comes with problems of loosing the ability to go backward.
		"-hls_flags", "append_list",  // append new segments instead of rewriting
        "-f", "hls",
        outputPath,
    )
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg error: %v, details: %s", err, stderr.String())
	}
	return nil
}


