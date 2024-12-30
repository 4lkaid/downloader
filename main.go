package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	from          *string     // Text data source
	split         *string     // Delimiter
	to            *string     // Download root directory
	retry         *uint       // Retry count on failure
	num           *uint       // Concurrency (default to CPU cores)
	timeout       *uint       // Timeout (ms)
	content       [][]string  // Data from text source
	startTime     time.Time   // Program start time
	errorLog      string      // Error log file path
	total         uint64      // Total download count
	successTotal  uint64      // Successful downloads
	failuresTotal uint64      // Failed downloads
	client        http.Client // HTTP client
	errorLogger   *log.Logger // Error logger
)

func init() {
	// Flag initialization
	from = flag.String("from", "", "Text data source")
	split = flag.String("split", "=>>", "Delimiter")
	to = flag.String("to", "downloads", "Download root directory")
	retry = flag.Uint("retry", 3, "Retry count on failure")
	num = flag.Uint("num", uint(runtime.NumCPU()), "Concurrency (default to CPU cores)")
	timeout = flag.Uint("timeout", 1000, "Timeout (ms)")

	flag.Parse()

	// Validate the flags
	validateFlags()

	// Load the content from the data source file
	loadContent()

	// Initialize error log and directories
	initErrorLog()

	// Create a global HTTP client
	client = http.Client{
		Timeout: time.Duration(*timeout) * time.Millisecond,
	}

	startTime = time.Now()
}

// validateFlags validates command-line flags.
func validateFlags() {
	if *from == "" {
		fmt.Println("Please set the -from flag")
		os.Exit(-1)
	}
	_, err := os.Stat(*from)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	if *split == "" {
		fmt.Println("The -split flag cannot be empty")
		os.Exit(-1)
	}
	if *num == 0 {
		fmt.Println("The -num flag must be greater than 0")
		os.Exit(-1)
	}
	if *timeout == 0 {
		fmt.Println("The -timeout flag must be greater than 0")
		os.Exit(-1)
	}
}

// loadContent loads content from the data source file.
func loadContent() {
	bytes, err := os.ReadFile(*from)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	lines := strings.Split(strings.Replace(string(bytes), "\r", "", -1), "\n")
	content = make([][]string, len(lines))
	for k, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		content[total] = strings.Split(line, *split)
		if len(content[total]) != 2 {
			fmt.Printf("Format error in line %d of %s", k+1, *from)
			os.Exit(-1)
		}
		atomic.AddUint64(&total, 1)
	}
	content = content[:total]
}

// ensureDir ensures that the given directory exists.
func ensureDir(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModePerm)
	}
	return err
}

// initErrorLog initializes the error log file.
func initErrorLog() {
	filename := filepath.Base(*from)
	errorLog = filepath.Join("logs", strings.Join([]string{
		strings.TrimSuffix(filename, filepath.Ext(filename)),
		time.Now().Format("20060102150405"),
		"error.log"}, "_"))
	path := filepath.Dir(errorLog)
	if err := ensureDir(path); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	if err := setupErrorLog(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	// Initialize error logger
	file, err := os.OpenFile(errorLog, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	errorLogger = log.New(file, "", log.LstdFlags)
}

// setupErrorLog creates a new error log file, backing up the existing one if needed.
func setupErrorLog() error {
	_, err := os.Stat(errorLog)
	if err == nil {
		backup := errorLog + "_backup"
		if err := os.Rename(errorLog, backup); err != nil {
			return err
		}
	}
	_, err = os.Create(errorLog)
	return err
}

// download performs the download and saves the file.
func download(url, name string) error {
	path := filepath.Dir(name)
	if err := ensureDir(path); err != nil {
		return err
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to get URL %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid status code %d for URL %s", resp.StatusCode, url)
	}

	dst, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", name, err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, resp.Body)
	return err
}

// logError logs errors into the error log file.
func logError(message string) {
	errorLogger.Println(message)
}

// run executes the download with retry logic.
func run(url, name string) {
	var retries uint
	for retries = 0; retries < *retry; retries++ {
		if err := download(url, name); err == nil {
			atomic.AddUint64(&successTotal, 1)
			return
		}
	}

	logError(fmt.Sprintf("%s %s %s\n", url, name, "Failed to download after retries"))
	atomic.AddUint64(&failuresTotal, 1)
}

// printStats prints the current download statistics.
func printStats() {
	fmt.Printf("\rTotal: %d, Success: %d, Failures: %d, Time: %s\r", total, successTotal, failuresTotal, time.Since(startTime))
}

// setupSignalHandler sets up the signal handler to gracefully exit on interruption.
func setupSignalHandler() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal, terminating...")
		os.Exit(0)
	}()
}

func main() {
	// Print configuration parameters
	fmt.Printf("From: %s\nTo: %s\nDelimiter: %s\nConcurrency: %d\nTimeout: %dms\nRetry Count: %d\nError Log: %s\n",
		*from, *to, *split, *num, *timeout, *retry, errorLog)

	// Setup signal handler
	setupSignalHandler()

	var wg sync.WaitGroup
	downloadChan := make(chan []string, total)
	wg.Add(1)

	// Send content to the download channel
	go func() {
		defer wg.Done()
		defer close(downloadChan)
		for _, arr := range content {
			downloadChan <- arr
		}
	}()

	// Start concurrent downloads
	for i := 0; i < int(*num); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for tmp := range downloadChan {
				run(strings.TrimSpace(tmp[0]), filepath.Join(*to, strings.TrimSpace(tmp[1])))
				printStats()
			}
		}()
	}

	wg.Wait()
	fmt.Println("\nDownload completed")
}
