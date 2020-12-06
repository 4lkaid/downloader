downloader
==========
整理个文本 [demo.txt](./demo.txt) 让 downloader 来帮你自动完成批量下载

### Contents
- 导入文本批量下载
- 支持自定义文本数据源
- 支持自定义下载根目录
- 支持自定义分割符
- 支持自定义并发数
- 支持自定义失败重试次数
- 支持自定义超时时间
- 自动校验文本内容格式
- 兼容windows/linux换行
- 自动跳过空白行

### Quick start

```
$ go get github.com/4lkaid/downloader
```

```
$ downloader -h
Usage of ./downloader:
  -from string
    	文本数据源
  -num uint
    	并发数(默认为CPU核数) (default 4)
  -retry uint
    	失败重试次数 (default 3)
  -split string
    	分割符 (default "=>>")
  -timeout uint
    	超时时间(毫秒) (default 1000)
  -to string
    	下载根目录 (default "downloads")
```

```
$ downloader -from demo.txt
from: demo.txt
to: downloads
分隔符: =>>
并发数: 4
超时时间: 1000(毫秒)
重试次数: 3
错误日志: log/demo_20201206174512_error.log
total: 15, success: 15, failures: 0, time: 554.568918ms
Download completed
```
