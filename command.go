package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
)

var (
	from    *string    // 文本数据源
	split   *string    // 分割符
	to      *string    // 下载根目录
	retry   *uint      // 失败重试次数
	num     *uint      // 并发数
	timeout *uint      // 超时时间(毫秒)
	content [][]string // 文本数据源数据
)

func init() {
	from = flag.String("from", "", "文本数据源")
	split = flag.String("split", "=>>", "分割符")
	to = flag.String("to", "downloads", "下载根目录")
	retry = flag.Uint("retry", 3, "失败重试次数")
	num = flag.Uint("num", uint(runtime.NumCPU()), "并发数(默认为CPU核数)")
	timeout = flag.Uint("timeout", 1000, "超时时间(毫秒)")
	flag.Parse()
	checkFrom()
	checkSplit()
	checkNum()
	checkTimeout()
	checkContent()
}

func checkFrom() {
	if *from == "" {
		fmt.Println("请设置 -from 参数")
		os.Exit(-1)
	}
	_, err := os.Stat(*from)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func checkSplit() {
	if *split == "" {
		fmt.Println("-split 参数设置错误(PS: 不能为空字符)")
		os.Exit(-1)
	}
}

func checkNum() {
	if *num == 0 {
		fmt.Println("-num 参数设置错误(PS: 大于 0)")
		os.Exit(-1)
	}
}

func checkTimeout() {
	if *timeout == 0 {
		fmt.Println("-timeout 参数设置错误(PS: 大于 0)")
		os.Exit(-1)
	}
}

func checkContent() {
	bytes, err := ioutil.ReadFile(*from)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	str := strings.Split(strings.Replace(string(bytes), "\r", "", -1), "\n")
	content = make([][]string, len(str))
	for k, v := range str {
		tmp := strings.Trim(v, " ")
		if len(tmp) == 0 {
			continue
		}
		content[total] = strings.Split(tmp, *split)
		if len(content[total]) != 2 {
			fmt.Printf("%s 第 %d 行 格式错误", *from, k+1)
			os.Exit(-1)
		}
		atomic.AddUint64(&total, 1)
	}
	content = content[:total]
}
