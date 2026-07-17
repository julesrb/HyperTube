package transcode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
)

type ffprobeStream struct {
	Index     int    `json:"index"`
	CodecType string `json:"codec_type"`
	CodecName string `json:"codec_name"`
	Tags      struct {
		Language string `json:"language"`
		Title    string `json:"title"`
	} `json:"tags"`
}

type ffprobeOutput struct {
	Streams []ffprobeStream `json:"streams"`
}

const defaultAudioMap = "0:a:0"

func probeStreamsFromReader(r io.Reader) ([]ffprobeStream, error) {
	var stdout bytes.Buffer
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		"pipe:0",
	)
	cmd.Stdin = r
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffprobe: %v", err)
	}
	var out ffprobeOutput
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return nil, fmt.Errorf("ffprobe parse: %v", err)
	}
	return out.Streams, nil
}

func videoCodecName(streams []ffprobeStream) string {
	for _, s := range streams {
		if s.CodecType == "video" {
			return s.CodecName
		}
	}
	return ""
}

// iso639_1to2 maps ISO 639-1
var iso639_1to2 = map[string][]string{
	"en": {"eng"},
	"fr": {"fre", "fra"},
	"de": {"ger", "deu"},
	"es": {"spa"},
	"it": {"ita"},
	"pt": {"por"},
	"ru": {"rus"},
	"ja": {"jpn"},
	"ko": {"kor"},
	"zh": {"chi", "zho"},
	"hi": {"hin"},
	"ar": {"ara"},
	"nl": {"dut", "nld"},
	"sv": {"swe"},
	"no": {"nor"},
	"da": {"dan"},
	"fi": {"fin"},
	"pl": {"pol"},
	"tr": {"tur"},
	"cs": {"cze", "ces"},
	"el": {"gre", "ell"},
	"he": {"heb"},
	"th": {"tha"},
	"uk": {"ukr"},
	"hu": {"hun"},
	"ro": {"rum", "ron"},
}

func hasAudioStream(streams []ffprobeStream) bool {
	for _, s := range streams {
		if s.CodecType == "audio" {
			return true
		}
	}
	return false
}

func audioMapForLanguage(streams []ffprobeStream, originalLang string) string {
	if !hasAudioStream(streams) {
		return ""
	}
	if originalLang == "" {
		return defaultAudioMap
	}
	lang := strings.ToLower(strings.TrimSpace(originalLang))

	want := map[string]bool{lang: true}
	for _, code := range iso639_1to2[lang] {
		want[code] = true
	}

	for _, s := range streams {
		if s.CodecType != "audio" {
			continue
		}
		if want[strings.ToLower(s.Tags.Language)] {
			return fmt.Sprintf("0:%d", s.Index)
		}
	}
	return defaultAudioMap
}

func audioArgs(audioMap string) []string {
	if audioMap == "" {
		return nil
	}
	return []string{
		"-map", audioMap,
		"-c:a", "aac",
		"-b:a", "256k",
		"-ac", "2",
	}
}

// buildH264Args repackages an h264 source into HLS
func buildH264Args(input, outputDir, audioMap string) []string {
	args := []string{
		"-i", input,
		"-map", "0:v:0",
		"-c:v", "copy",
	}
	args = append(args, audioArgs(audioMap)...)
	return append(args,
		"-bsf:v", "h264_mp4toannexb",
		"-avoid_negative_ts", "make_zero",
		"-hls_time", "5",
		"-hls_list_size", "0",
		"-hls_flags", "append_list",
		"-f", "hls",
		outputDir+"/stream.m3u8",
	)
}

// buildHEVCArgs repackages an hevc source into HLS
func buildHEVCArgs(input, outputDir, audioMap string) []string {
	args := []string{
		"-i", input,
		"-map", "0:v:0",
		"-c:v", "copy",
	}
	args = append(args, audioArgs(audioMap)...)
	return append(args,
		"-bsf:v", "hevc_mp4toannexb",
		"-avoid_negative_ts", "make_zero",
		"-hls_time", "5",
		"-hls_list_size", "0",
		"-hls_flags", "append_list",
		"-f", "hls",
		outputDir+"/stream.m3u8",
	)
}

// transcode other sources source into h264 and packages it into HLS
func defaultTranscodeArgs(input, outputDir, audioMap string) []string {
	args := []string{
		"-i", input,
		"-map", "0:v:0",
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-crf", "23",
	}
	args = append(args, audioArgs(audioMap)...)
	return append(args,
		"-bsf:v", "h264_mp4toannexb",
		"-avoid_negative_ts", "make_zero",
		"-hls_time", "5",
		"-hls_list_size", "0",
		"-hls_flags", "append_list",
		"-f", "hls",
		outputDir+"/stream.m3u8",
	)
}

// ConvertPipeHLS probes the first 10 MB of reader to detect the video codec and
// pick the audio track matching originalLang
func ConvertPipeHLS(reader io.ReadSeeker, outputDir string, originalLang string) error {
	videoCodec := ""
	audioMap := defaultAudioMap
	if streams, err := probeStreamsFromReader(io.LimitReader(reader, 10*1024*1024)); err != nil {
		log.Printf("probe streams failed: %v", err)
	} else {
		videoCodec = videoCodecName(streams)
		audioMap = audioMapForLanguage(streams, originalLang)
		log.Printf("source video codec: %s, original language: %q, selected audio map: %s", videoCodec, originalLang, audioMap)
	}
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek: %w", err)
	}

	var args []string
	switch videoCodec {
	case "h264":
		args = buildH264Args("pipe:0", outputDir, audioMap)
	// case "hevc":
	// 	args = buildHEVCArgs("pipe:0", outputDir, audioMap)
	default:
		args = defaultTranscodeArgs("pipe:0", outputDir, audioMap)

	}

	var stderr bytes.Buffer
	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdin = reader
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg: %v\n%s", err, stderr.String())
	}
	return nil
}
