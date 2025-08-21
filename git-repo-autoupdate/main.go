package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Config 用于存储从配置文件中读取的内容
type Config struct {
	Repositories []string `json:"repositories"`
	Proxy        string   `json:"proxy"`
}

// 颜色定义
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
)

func main() {
	configFile := "config.json" // 默认配置文件路径

	// 如果用户提供了参数，则使用该路径
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	// 读取配置文件
	config, err := readConfig(configFile)
	if err != nil {
		fmt.Printf("%s错误: %s%s\n", Red, err, Reset)
		return
	}

	// 创建日志文件
	logFile, err := os.OpenFile("update_log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("%s无法创建日志文件: %s%s\n", Red, err, Reset)
		return
	}
	defer logFile.Close()

	// 遍历所有 Git 仓库路径
	for _, repoPath := range config.Repositories {
		updateRepository(repoPath, config.Proxy, logFile)
	}
}

// readConfig 读取和解析配置文件
func readConfig(filePath string) (Config, error) {
	var config Config
	file, err := os.Open(filePath)
	if err != nil {
		return config, fmt.Errorf("无法打开配置文件: %s", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return config, fmt.Errorf("解析配置文件时出错: %s", err)
	}
	return config, nil
}

// updateRepository 更新指定的 Git 仓库
func updateRepository(repoPath, proxy string, logFile *os.File) {
	// 确保路径存在
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		logAndPrint(fmt.Sprintf("%s仓库路径不存在: %s%s\n", Red, repoPath, Reset), logFile)
		return
	}

	repoName := filepath.Base(repoPath)
	logAndPrint(fmt.Sprintf("\n%s🚀 正在更新仓库: %s%s\n", Cyan, repoName, Reset), logFile)

	// 设置代理环境变量
	os.Setenv("http_proxy", proxy)
	os.Setenv("https_proxy", proxy)

	// 进入仓库目录并执行 git pull
	cmd := exec.Command("git", "pull")
	cmd.Dir = repoPath

	// 获取命令的输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		logAndPrint(fmt.Sprintf("%s❌ 更新失败: %s\n输出: %s%s\n", Red, err, string(output), Reset), logFile)
		return
	}

	// 成功更新的输出
	logAndPrint(fmt.Sprintf("%s✔️ 更新成功:\n%s%s", Green, string(output), Reset), logFile)
}

// logAndPrint 输出日志并打印到终端
func logAndPrint(message string, logFile *os.File) {
	// 打印到终端
	fmt.Print(message)

	// 追加到日志文件
	logFile.WriteString(fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), message))
}
