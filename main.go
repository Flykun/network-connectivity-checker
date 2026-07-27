package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"runtime"
	"sync"
	"time"
)

// CheckResult holds the outcome of testing a single IP/Port
type CheckResult struct {
	Target  string
	Success bool
	Latency time.Duration
	Err     error
	ErrMsg  string
}

// formatNetError parses net.Error to return human-readable failure reason
func formatNetError(err error) string {
	if err == nil {
		return ""
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "Connection timed out"
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return fmt.Sprintf("Network error (%s %s): %v", opErr.Op, opErr.Net, opErr.Err)
	}

	return err.Error()
}

// checkTarget attempts to establish a TCP connection to a given target
func checkTarget(target string, timeout time.Duration, wg *sync.WaitGroup, results chan<- CheckResult) {
	defer wg.Done()

	startTime := time.Now()
	// Attempt TCP connection
	conn, err := net.DialTimeout("tcp", target, timeout)
	latency := time.Since(startTime)

	if err != nil {
		results <- CheckResult{
			Target:  target,
			Success: false,
			Latency: latency,
			Err:     err,
			ErrMsg:  formatNetError(err),
		}
		return
	}
	conn.Close() // Close connection immediately after successful test

	results <- CheckResult{
		Target:  target,
		Success: true,
		Latency: latency,
		Err:     nil,
		ErrMsg:  "",
	}
}

// checkEnvironment performs runtime system architecture and OS checks
func checkEnvironment() {
	fmt.Printf("[INFO] Go Version  : %s\n", runtime.Version())
	fmt.Printf("[INFO] OS/Arch     : %s/%s\n", runtime.GOOS, runtime.GOARCH)

	if runtime.GOOS != "linux" {
		fmt.Printf("[WARN] Warning: Current OS is '%s', intended target is 'linux'.\n", runtime.GOOS)
	}
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("  Network Connectivity Checker")
	fmt.Println("==================================================")

	// Check system OS and Go environment
	checkEnvironment()

	// 1. List of IP and Port targets to test
	targets := []string{
		"114.114.114.114:53",
		"8.8.8.8:53",
		"1.1.1.1:53",
		"223.5.5.5:53",
		"180.76.76.76:53",
	}

	timeout := 2 * time.Second
	var wg sync.WaitGroup
	results := make(chan CheckResult, len(targets))

	fmt.Println("--------------------------------------------------")
	fmt.Println(" Starting concurrent connectivity check...")
	fmt.Println("--------------------------------------------------")

	// 2. Launch a goroutine for each target
	for _, target := range targets {
		wg.Add(1)
		go checkTarget(target, timeout, &wg, results)
	}

	// 3. Wait for all goroutines in background and close channel
	go func() {
		wg.Wait()
		close(results)
	}()

	hasFailure := false

	// 4. Read and print results in real-time
	for res := range results {
		if res.Success {
			fmt.Printf("[SUCCESS] %-20s | Latency: %8v\n", res.Target, res.Latency.Round(time.Millisecond))
		} else {
			hasFailure = true
			fmt.Printf("[FAILED]  %-20s | Latency: %8v | Error: %s\n", res.Target, res.Latency.Round(time.Millisecond), res.ErrMsg)
		}
	}

	fmt.Println("==================================================")
	fmt.Println(" Connectivity check completed.")

	// 如果是在 Windows 下双击运行，等待按 Enter 键，防止窗口直接闪退关闭
	if runtime.GOOS == "windows" {
		fmt.Println("\nPress Enter to exit...")
		var input string
		fmt.Scanln(&input)
	}

	// Linux 自动化测试时如果存在失败则返回 exit status 1
	if hasFailure && runtime.GOOS != "windows" {
		os.Exit(1)
	}
}
