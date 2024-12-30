# downloader

This program reads URLs and filenames from a [text](./demo.txt) file and concurrently downloads the files. It supports custom retry counts, concurrency, timeout settings, and logs errors during the download process.

## Features

- Reads URLs and filenames from a text file for downloading.
- Supports retrying downloads in case of temporary network failures.
- Downloads files concurrently using multiple threads to optimize speed.
- Generates error logs for failed download attempts.
- Allows custom timeout settings to prevent indefinite blocking of requests.

## Usage

```bash
git clone https://github.com/4lkaid/downloader.git
cd downloader
go build
./downloader -from demo.txt
```

## License

This project is licensed under the [MIT license](./LICENSE).