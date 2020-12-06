package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

func main() {
	fmt.Printf("from: %s\nto: %s\n分隔符: %s\n并发数: %d\n超时时间: %d(毫秒)\n重试次数: %d\n错误日志: %s\n", *from, *to, *split, *num, *timeout, *retry, errorLog)
	var wg sync.WaitGroup
	downloadChan := make(chan []string, total)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(downloadChan)
		for _, arr := range content {
			downloadChan <- arr
		}
	}()
	for i := 0; i < int(*num); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for tmp := range downloadChan {
				run(strings.Trim(tmp[0], " "), filepath.Join(*to, strings.Trim(tmp[1], " ")))
				print()
			}
		}()
	}
	wg.Wait()
	fmt.Println("\nDownload completed")
}
