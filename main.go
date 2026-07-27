package main

import (
	"bufio"
	"context"
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

// checkTarget attempts to establish a TCP connection with Context support
func checkTarget(ctx context.Context, target string, timeout time.Duration) CheckResult {
	startTime := time.Now()

	// 使用 dialer + context 可以更好地支持响应取消与精细控制
	var dialer net.Dialer
	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	conn, err := dialer.DialContext(dialCtx, "tcp", target)
	latency := time.Since(startTime)

	if err != nil {
		return CheckResult{
			Target:  target,
			Success: false,
			Latency: latency,
			Err:     err,
			ErrMsg:  formatNetError(err),
		}
	}
	conn.Close() // 成功连接后立即关闭

	return CheckResult{
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

	checkEnvironment()

	targets := []string{
		"114.114.114.114:53",
		"8.8.8.8:53",
		"1.1.1.1:53",
		"223.5.5.5:53",
		"180.76.76.76:53",
	}

	timeout := 2 * time.Second
	maxConcurrency := 10 // 控制最大并发数，防止 FD 耗尽或打满出口

	jobs := make(chan string, len(targets))
	results := make(chan CheckResult, len(targets))

	// 将所有待检测目标放入 jobs 通道
	for _, t := range targets {
		jobs <- t
	}
	close(jobs)

	fmt.Println("--------------------------------------------------")
	fmt.Println(" Starting concurrent connectivity check...")
	fmt.Println("--------------------------------------------------")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// 启动固定数量的 Worker Goroutines
	workerCount := maxConcurrency
	if len(targets) < workerCount {
		workerCount = len(targets)
	}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for target := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
					res := checkTarget(ctx, target, timeout)
					results <- res
				}
			}
		}()
	}

	// 异步等待所有工作完成并关闭 results 通道
	go func() {
		wg.Wait()
		close(results)
	}()

	hasFailure := false

	// 实时读取并打印结果
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

	// Windows 下等待 Enter 退出，避免闪退
	if runtime.GOOS == "windows" {
		fmt.Println("\nPress Enter to exit...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}

	// Linux 下检测失败返回 exit status 1，方便被自动化脚本/CI-CD捕获
	if hasFailure && runtime.GOOS != "windows" {
		os.Exit(1)
	}
}
