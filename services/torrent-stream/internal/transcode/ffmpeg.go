package transcode

import (
	"io"
	"os/exec"
)

// Pipe launches ffmpeg reading from src and writing fragmented MP4 to dst.
// Flags: -movflags frag_keyframe+empty_moov are required so the browser can
// begin playback before the full file is received.
func Pipe(src io.Reader, dst io.Writer) error {
	cmd := exec.Command(
		"ffmpeg",
		"-i", "pipe:0",
		"-movflags", "frag_keyframe+empty_moov",
		"-f", "mp4",
		"pipe:1",
	)
	cmd.Stdin = src
	cmd.Stdout = dst
	return cmd.Run()
}
