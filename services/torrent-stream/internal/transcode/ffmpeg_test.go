package transcode

import (
    "os"
    "testing"
)

const CLEANUP_OUTPUT = true
// const CLEANUP_OUTPUT = false

func cleanupOutput(outputPath string) {
    if CLEANUP_OUTPUT {
        os.RemoveAll(outputPath)
    }
}

func TestTranscodeHLS_standardFile(t *testing.T) {
    file := "rubber"

    inputFile := "./test/" + file + ".mp4"
    outputPath := "./test/output/" + file + "/"
    outputFile := outputPath + "index" + ".m3u8"

    if err := os.MkdirAll(outputPath, 0755); err != nil {
        t.Fatalf("failed to create output dir: %v", err)
    }

    t.Cleanup(func() { 
        cleanupOutput(outputPath)
    })

    if err := TranscodeHLS(inputFile, outputFile); err != nil {
        t.Errorf("TranscodeHLS failed: %v", err)
    }
}


