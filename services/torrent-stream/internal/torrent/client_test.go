package torrent

import (
	"errors"
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
