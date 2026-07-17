package transcode

import (
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func mustContainSequence(t *testing.T, haystack []string, a, b string) {
	t.Helper()
	for i := 0; i < len(haystack)-1; i++ {
		if haystack[i] == a && haystack[i+1] == b {
			return
		}
	}
	t.Errorf("expected %q %q in args %v", a, b, haystack)
}

// buildAV1Args must TRANSCODE the video to h264, not copy it.
func TestBuildAV1ArgsTranscodesToH264(t *testing.T) {
	args := defaultTranscodeArgs("pipe:0", "/out", "0:a:0")

	mustContainSequence(t, args, "-c:v", "libx264")
	mustContainSequence(t, args, "-c:a", "aac")
	mustContainSequence(t, args, "-map", "0:v:0")
	mustContainSequence(t, args, "-map", "0:a:0")
	mustContainSequence(t, args, "-bsf:v", "h264_mp4toannexb")
	mustContainSequence(t, args, "-f", "hls")

	if slices.Contains(args, "copy") {
		t.Errorf("av1 path must not copy any stream, got %v", args)
	}
	if args[len(args)-1] != "/out/stream.m3u8" {
		t.Errorf("expected output playlist last, got %q", args[len(args)-1])
	}
}

// TestConvertPipeHLS_AV1ToH264 runs the real ffmpeg pipeline on an AV1 source
// and asserts the produced HLS video stream is h264. Requires ffmpeg/ffprobe.
func TestConvertPipeHLS_AV1ToH264(t *testing.T) {
	const fixture = "./test/sample_av1.mkv"
	if _, err := os.Stat(fixture); err != nil {
		t.Skipf("missing AV1 fixture %s: %v", fixture, err)
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not in PATH")
	}

	f, err := os.Open(fixture)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	// Write into ./test/output/ so the produced files can be inspected after the run.
	out := "./test/output/av1_to_h264"
	if err := os.RemoveAll(out); err != nil {
		t.Fatalf("clean output dir: %v", err)
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatalf("create output dir: %v", err)
	}
	if err := ConvertPipeHLS(f, out, "en"); err != nil {
		t.Fatalf("ConvertPipeHLS failed: %v", err)
	}

	if _, err := os.Stat(out + "/stream.m3u8"); err != nil {
		t.Fatalf("expected playlist: %v", err)
	}
	segs, _ := os.ReadDir(out)
	var tsCount int
	for _, e := range segs {
		if strings.HasSuffix(e.Name(), ".ts") {
			tsCount++
		}
	}
	if tsCount == 0 {
		t.Fatalf("expected at least one .ts segment in %v", segs)
	}

	// Probe a segment: the output video codec must now be h264.
	probe := exec.Command("ffprobe", "-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=codec_name",
		"-of", "default=nw=1:nk=1",
		out+"/stream.m3u8")
	codec, err := probe.Output()
	if err != nil {
		t.Fatalf("ffprobe output: %v", err)
	}
	// ffprobe reports the codec once per segment; all lines should be h264.
	for _, line := range strings.Fields(string(codec)) {
		if line != "h264" {
			t.Fatalf("output video codec = %q, want h264", line)
		}
	}
	t.Logf("OK: AV1 source -> %d HLS segments, video codec=h264", tsCount)
}
