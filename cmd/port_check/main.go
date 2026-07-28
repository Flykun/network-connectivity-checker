package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
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
	fmt.Printf("[INFO] Go Version : %s\n", runtime.Version())
	fmt.Printf("[INFO] OS/Arch    : %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func main() {
	// 1. 定义与解析命令行参数
	targetsFlag := flag.String("t", "114.114.114.114:53,8.8.8.8:53", "指定检测目标，多个目标用逗号分隔 (例: -t \"192.168.1.1:80, 10.0.0.1:443\")")
	timeoutFlag := flag.Duration("timeout", 2*time.Second, "单次 TCP 连接超时时间 (例: 2s, 500ms)")
	concurrencyFlag := flag.Int("c", 10, "最大并发 Worker 数量")

	flag.Parse()

	fmt.Println("==================================================")
	fmt.Println("  Network Connectivity Checker")
	fmt.Println("==================================================")

	checkEnvironment()

	// 2. 解析 targets 列表并清洗无效/包含空格的字符串
	rawTargets := strings.Split(*targetsFlag, ",")
	var targets []string
	for _, t := range rawTargets {
		cleaned := strings.TrimSpace(t)
		if cleaned != "" {
			targets = append(targets, cleaned)
		}
	}

	if len(targets) == 0 {
		fmt.Println("[ERROR] 没有有效的检测目标！请使用 -t 参数指定。")
		os.Exit(1)
	}

	timeout := *timeoutFlag
	maxConcurrency := *concurrencyFlag
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	jobs := make(chan string, len(targets))
	results := make(chan CheckResult, len(targets))

	// 将所有待检测目标放入 jobs 通道
	for _, t := range targets {
		jobs <- t
	}
	close(jobs)

	fmt.Println("--------------------------------------------------")
	fmt.Printf(" 开始并发检测 [目标数: %d | 超时: %v | 并发数: %d]...\n", len(targets), timeout, maxConcurrency)
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
			fmt.Printf("[SUCCESS] %-22s | Latency: %8v\n", res.Target, res.Latency.Round(time.Millisecond))
		} else {
			hasFailure = true
			fmt.Printf("[FAILED]  %-22s | Latency: %8v | Error: %s\n", res.Target, res.Latency.Round(time.Millisecond), res.ErrMsg)
		}
	}

	fmt.Println("==================================================")
	fmt.Println(" Connectivity check completed.")

	// 如果有任何一个端口探测失败，返回 exit status 1，便于 CI/CD 或 Shell 脚本捕捉
	if hasFailure {
		os.Exit(1)
	}
}
