package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

var (
	startTime     time.Time // 程序开始执行时间
	errorLog      string    // 错误日志
	total         uint64    // 下载总数
	successTotal  uint64    // 下载成功总数
	failuresTotal uint64    // 下载失败总数
)

// init 初始化
// 设置开始时间并检查错误日志文件
// 日志文件存在时备份并生成新的日志文件
func init() {
	startTime = time.Now()
	filename := filepath.Base(*from)
	errorLog = filepath.Join("log", strings.Join([]string{strings.TrimSuffix(filename, filepath.Ext(filename)), time.Now().Format("20060102150405"), "error.log"}, "_"))
	path := filepath.Dir(errorLog)
	err := checkPath(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	_, err1 := os.Stat(errorLog)
	if err1 == nil {
		err2 := os.Rename(errorLog, strings.Join([]string{errorLog, "backup"}, "_"))
		if err2 != nil {
			fmt.Println(err2)
			os.Exit(-1)
		}
	}
	_, err3 := os.Create(errorLog)
	if err3 != nil {
		fmt.Println(err3)
		os.Exit(-1)
	}
}

// checkPath 检测路径
// 示例: checkPath("/a/.../b")
// path 不存在时自动创建
// path 存在时返回 nil, 创建失败返回 error
func checkPath(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err1 := os.MkdirAll(path, os.ModePerm)
		if err1 != nil {
			return err1
		}
		return nil
	}
	return err
}

// download 下载
// 示例: download("https://a.com/b.png", "/c/.../d.png")
// 下载成功返回 nil, 失败返回 error
func download(url, name string) error {
	path := filepath.Dir(name)
	err := checkPath(path)
	if err != nil {
		return err
	}
	client := http.Client{
		Timeout: time.Duration(*timeout) * time.Millisecond,
	}
	resp, err1 := client.Get(url)
	if err1 != nil {
		return err1
	}
	statusCode := resp.StatusCode
	if statusCode != 200 {
		return errors.New("invalid url")
	}
	dst, err2 := os.Create(name)
	if err2 != nil {
		return err2
	}
	defer dst.Close()
	src := resp.Body
	_, err3 := io.Copy(dst, src)
	if err3 != nil {
		return err3
	}
	return nil
}

// log 写入错误日志
// 当日志文件打开失败时会终止程序
func log(s string) {
	f, err := os.OpenFile(errorLog, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	_, _ = f.WriteString(s)
	f.Close()
}

// run 执行下载
// 默认下载失败重试 3 次, 重试后仍失败时记录错误日志
// 原子计数器统计成功/失败总数
func run(url, name string) {
	var tmp uint = 0
DOWNLOAD:
	err := download(url, name)
	if err != nil {
		if tmp < *retry {
			tmp++
			goto DOWNLOAD
		}
		log(strings.Join([]string{strings.Join([]string{url, name, err.Error()}, *split), "\n"}, ""))
		atomic.AddUint64(&failuresTotal, 1)
		return
	}
	atomic.AddUint64(&successTotal, 1)
}

// print 输出当前下载统计
// 打印输出: total: 1, success: 1, failures: 0, time: 523.0075ms
func print() {
	fmt.Printf("\rtotal: %d, success: %d, failures: %d, time: %s\r", total, successTotal, failuresTotal, time.Since(startTime))
}
