package torrent

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestPickMainVideoFile(t *testing.T) {
	t.Run("returns an error for empty input", func(t *testing.T) {
		_, err := pickMainVideoFile(nil)
		if !errors.Is(err, errNoMainVideoFile) {
			t.Fatalf("expected errNoMainVideoFile, got %v", err)
		}
	})

	t.Run("returns an error when there are no video files", func(t *testing.T) {
		files := []fileCandidate{
			{Path: "poster.jpg", Size: 4_096},
			{Path: "notes.nfo", Size: 1_024},
			{Path: "subtitles.srt", Size: 2_048},
		}

		_, err := pickMainVideoFile(files)
		if !errors.Is(err, errNoMainVideoFile) {
			t.Fatalf("expected errNoMainVideoFile, got %v", err)
		}
	})

	t.Run("returns the only valid video file", func(t *testing.T) {
		files := []fileCandidate{
			{Path: "movie/feature.mkv", Size: 1_500_000_000},
		}

		got, err := pickMainVideoFile(files)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		want := files[0]
		if got != want {
			t.Fatalf("expected %+v, got %+v", want, got)
		}
	})

	t.Run("returns the largest valid video file", func(t *testing.T) {
		files := []fileCandidate{
			{Path: "movie/trailer.mp4", Size: 200_000_000},
			{Path: "movie/feature.mkv", Size: 1_500_000_000},
			{Path: "movie/bonus.webm", Size: 700_000_000},
		}

		got, err := pickMainVideoFile(files)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		want := files[1]
		if got != want {
			t.Fatalf("expected %+v, got %+v", want, got)
		}
	})

	t.Run("ignores sample files even when they are larger", func(t *testing.T) {
		files := []fileCandidate{
			{Path: "movie/Movie.Sample.mkv", Size: 2_000_000_000},
			{Path: "movie/Movie.1080p.mkv", Size: 1_400_000_000},
		}

		got, err := pickMainVideoFile(files)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		want := files[1]
		if got != want {
			t.Fatalf("expected %+v, got %+v", want, got)
		}
	})

	t.Run("matches video extensions case-insensitively", func(t *testing.T) {
		files := []fileCandidate{
			{Path: "movie/feature.MKV", Size: 1_500_000_000},
		}

		got, err := pickMainVideoFile(files)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		want := files[0]
		if got != want {
			t.Fatalf("expected %+v, got %+v", want, got)
		}
	})
}

// TestClientAdd covers only the orchestration added in iteration two.
// The opener, torrent and files are stubs on purpose so the tests stay fast,
// deterministic and independent from a real torrent backend.
func TestClientAdd(t *testing.T) {
	t.Run("returns an error for an empty magnet URI", func(t *testing.T) {
		opener := &stubOpener{}
		client := Client{opener: opener}

		_, err := client.Add("   ")
		if !errors.Is(err, errEmptyMagnetURI) {
			t.Fatalf("expected errEmptyMagnetURI, got %v", err)
		}

		if opener.calls != 0 {
			t.Fatalf("expected opener to not be called, got %d calls", opener.calls)
		}
	})

	t.Run("returns an error when no opener is configured", func(t *testing.T) {
		client := Client{}

		_, err := client.Add("magnet:?xt=urn:btih:test")
		if !errors.Is(err, errNoMagnetOpener) {
			t.Fatalf("expected errNoMagnetOpener, got %v", err)
		}
	})

	t.Run("propagates opener errors", func(t *testing.T) {
		wantErr := errors.New("open magnet")
		client := Client{
			opener: &stubOpener{err: wantErr},
		}

		_, err := client.Add("magnet:?xt=urn:btih:test")
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected %v, got %v", wantErr, err)
		}
	})

	t.Run("returns an error when no main video file exists", func(t *testing.T) {
		client := Client{
			opener: &stubOpener{
				torrent: &stubTorrent{
					files: []torrentFile{
						&stubFile{path: "poster.jpg", size: 4_096},
						&stubFile{path: "notes.nfo", size: 1_024},
					},
				},
			},
		}

		_, err := client.Add("magnet:?xt=urn:btih:test")
		if !errors.Is(err, errNoMainVideoFile) {
			t.Fatalf("expected errNoMainVideoFile, got %v", err)
		}
	})

	t.Run("returns the selected main video reader", func(t *testing.T) {
		sample := &stubFile{
			path:   "movie/Movie.Sample.mkv",
			size:   2_000_000_000,
			reader: strings.NewReader("sample"),
		}
		feature := &stubFile{
			path:   "movie/Movie.1080p.mkv",
			size:   1_400_000_000,
			reader: strings.NewReader("feature"),
		}

		opener := &stubOpener{
			torrent: &stubTorrent{
				files: []torrentFile{sample, feature},
			},
		}
		client := Client{opener: opener}

		got, err := client.Add("magnet:?xt=urn:btih:test")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		body, err := io.ReadAll(got)
		if err != nil {
			t.Fatalf("expected to read the selected reader, got %v", err)
		}

		if string(body) != "feature" {
			t.Fatalf("expected feature reader contents, got %q", string(body))
		}

		if sample.openCalls != 0 {
			t.Fatalf("expected sample file to stay closed, got %d opens", sample.openCalls)
		}

		if feature.openCalls != 1 {
			t.Fatalf("expected feature file to be opened once, got %d opens", feature.openCalls)
		}

		if opener.lastMagnetURI != "magnet:?xt=urn:btih:test" {
			t.Fatalf("expected opener to receive the magnet URI, got %q", opener.lastMagnetURI)
		}
	})

	t.Run("propagates reader errors", func(t *testing.T) {
		wantErr := errors.New("open reader")
		feature := &stubFile{
			path: "movie/Movie.1080p.mkv",
			size: 1_400_000_000,
			err:  wantErr,
		}
		client := Client{
			opener: &stubOpener{
				torrent: &stubTorrent{
					files: []torrentFile{feature},
				},
			},
		}

		_, err := client.Add("magnet:?xt=urn:btih:test")
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected %v, got %v", wantErr, err)
		}
	})
}

// stubOpener records how Add interacts with the dependency that will later
// become the real torrent integration.
type stubOpener struct {
	torrent       torrentHandle
	err           error
	calls         int
	lastMagnetURI string
}

func (o *stubOpener) Open(magnetURI string) (torrentHandle, error) {
	o.calls++
	o.lastMagnetURI = magnetURI

	if o.err != nil {
		return nil, o.err
	}

	return o.torrent, nil
}

type stubTorrent struct {
	files []torrentFile
}

func (t *stubTorrent) Files() []torrentFile {
	return t.files
}

type stubFile struct {
	path      string
	size      int64
	reader    io.Reader
	err       error
	openCalls int
}

func (f *stubFile) Path() string {
	return f.path
}

func (f *stubFile) Size() int64 {
	return f.size
}

func (f *stubFile) NewReader() (io.Reader, error) {
	f.openCalls++

	if f.err != nil {
		return nil, f.err
	}

	return f.reader, nil
}
