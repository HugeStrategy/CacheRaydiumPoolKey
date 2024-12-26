package downloader

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io"
	"net/http"
	"os"
)

// DownloadFile 从指定的 URL 下载文件并保存到 filepath，同时显示下载进度条
func DownloadFile(url, filepath string) error {
	// 创建HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// 检查HTTP响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// 获取内容长度
	contentLength := resp.ContentLength

	// 创建输出文件
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	var writer io.Writer

	if contentLength > 0 {
		// 初始化具有已知总大小的进度条
		bar := progressbar.NewOptions64(
			contentLength,
			progressbar.OptionSetDescription("Downloading"),
			progressbar.OptionSetWriter(os.Stdout),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(40),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "=",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}),
		)
		writer = io.MultiWriter(out, bar)
	} else {
		// 初始化无总大小的进度指示器（旋转器）
		fmt.Println("Downloading (unknown size)...")
		bar := progressbar.NewOptions(
			-1,
			progressbar.OptionSetDescription("Downloading"),
			progressbar.OptionSetWriter(os.Stdout),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionSetWidth(40),
			progressbar.OptionFullWidth(),
			progressbar.OptionClearOnFinish(),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "=",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}),
		)
		writer = io.MultiWriter(out, bar)
	}

	// 开始下载并写入数据
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	// 确保进度条完成
	if contentLength > 0 {
		fmt.Println("\nDownload completed successfully.")
	} else {
		fmt.Println("Download completed successfully.")
	}

	return nil
}
