package module

import (
	"fmt"
	"io"
	"strings"
	"time"
)

var (
	printRateInterval = time.Second / 2
)

type ProgressReader struct {
	io.Reader
	Total          int64
	Current        int64
	StartTime      time.Time
	LastPrinted    time.Time
	LastPrintedStr string
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Current += int64(n)
	pr.printProgress()
	return n, err
}

func (pr *ProgressReader) printProgress() {
	progress := float64(pr.Current) / float64(pr.Total) * 100
	now := time.Now()
	elapsed := now.Sub(pr.StartTime)
	uploadSpeed := float64(pr.Current) / elapsed.Seconds()

	if now.Sub(pr.LastPrinted) >= printRateInterval { // print rate every half second
		progressStr := fmt.Sprintf("uploading progress: %.2f%% [ %s / %s ], rate: %s    ",
			progress, getConvertSize(pr.Current), getConvertSize(pr.Total), getConvertRate(uploadSpeed))
		// Clear current line
		fmt.Print("\r", strings.Repeat(" ", len(pr.LastPrintedStr)), "\r")
		// Print new progress
		fmt.Print(progressStr)

		pr.LastPrinted = now
	}
}

func getConvertRate(rate float64) string {
	const (
		KB = 1024
		MB = 1024 * KB
	)

	switch {
	case rate >= MB:
		return fmt.Sprintf("%.2f MB/s", rate/MB)
	case rate >= KB:
		return fmt.Sprintf("%.2f KB/s", rate/KB)
	default:
		return fmt.Sprintf("%.2f Byte/s", rate)
	}
}

func getConvertSize(fileSize int64) string {
	var convertedSize string
	if fileSize > 1<<30 {
		convertedSize = fmt.Sprintf("%.2fG", float64(fileSize)/(1<<30))
	} else if fileSize > 1<<20 {
		convertedSize = fmt.Sprintf("%.2fM", float64(fileSize)/(1<<20))
	} else if fileSize > 1<<10 {
		convertedSize = fmt.Sprintf("%.2fK", float64(fileSize)/(1<<10))
	} else {
		convertedSize = fmt.Sprintf("%dB", fileSize)
	}
	return convertedSize
}
