package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
	Blue   = "\033[34m"
)

type Config struct {
	Repositories []string `json:"repositories"`
	Proxy        string   `json:"proxy"`
}

func main() {
	rootPath := getRootPath()
	if rootPath == "" {
		fmt.Println(Red + "未指定有效的路径" + Reset)
		return
	}

	repos := findGitRepositories(rootPath)
	if len(repos) == 0 {
		fmt.Println(Yellow + "没有找到任何 Git 仓库" + Reset)
		return
	}

	fmt.Println(Cyan + "找到以下 Git 仓库:" + Reset)
	for i, repo := range repos {
		fmt.Printf("%d. %s%s\n", i+1, Blue, repo, Reset) // 编号和蓝色显示仓库
	}

	configFile := "config.json"
	config, err := readConfig(configFile)
	if err != nil {
		fmt.Printf("%s读取配置文件时出错: %s%s\n", Red, err, Reset)
		return
	}

	for {
		var input string
		fmt.Print("请输入要添加的仓库编号（用逗号分隔，举例：1,2），或输入 y 继续，q 退出: ")
		fmt.Scanln(&input)

		if input == "q" {
			break
		} else if input == "y" {
			continue
		}

		for _, s := range splitCommaSeparated(input) {
			var index int
			if _, err := fmt.Sscanf(s, "%d", &index); err == nil && index > 0 && index <= len(repos) {
				repo := repos[index-1]
				if !contains(config.Repositories, repo) {
					config.Repositories = append(config.Repositories, repo)
					fmt.Printf("✅ 更新成功：%s\n", repo)
				} else {
					fmt.Printf("✅ %s 已经在配置中，不会重复添加。\n", repo)
				}
			} else {
				fmt.Printf("%s无效的输入: %s%s\n", Red, s, Reset)
			}
		}

		if err := saveConfig(configFile, config); err != nil {
			fmt.Printf("%s保存配置文件时出错: %s%s\n", Red, err, Reset)
			return
		}

		fmt.Printf("%s成功将选定的仓库添加到 config.json%s\n", Green, Reset)
	}

	fmt.Println("程序结束。按 Enter 键退出...")
	fmt.Scanln()
}

func getRootPath() string {
	cmd := exec.Command("cmd.exe", "/C", "echo | set /p=请输入要查看的Git仓库路径: ")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	var path string
	fmt.Scanln(&path)
	return path
}

func findGitRepositories(rootPath string) []string {
	var repos []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				fmt.Printf("%s无权限访问: %s%s\n", Yellow, path, Reset)
				return nil
			}
			return err
		}

		if info.IsDir() && path != rootPath {
			wg.Add(1)
			go func(p string) {
				defer wg.Done()
				if isGitRepo(p) {
					mu.Lock()
					repos = append(repos, p)
					fmt.Printf("\r正在遍历：%s%s\n", Blue, p) // 只有在找到 Git 仓库时才使用蓝色显示
					mu.Unlock()
				} else {
					fmt.Printf("\r正在遍历：%s", p) // 其他路径保持默认样式
				}
			}(path)
		}
		return nil
	})

	wg.Wait()
	fmt.Print("\r") // 打印一个回车以清除当前行
	fmt.Println()   // 换行
	return repos
}

func isGitRepo(path string) bool {
	gitPath := filepath.Join(path, ".git")
	if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
		return true
	}
	return false
}

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

func saveConfig(filePath string, config Config) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("无法打开配置文件进行写入: %s", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(config); err != nil {
		return fmt.Errorf("写入配置文件时出错: %s", err)
	}
	return nil
}

func splitCommaSeparated(input string) []string {
	return strings.Split(input, ",")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
