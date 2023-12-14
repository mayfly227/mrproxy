package common

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func SetWorkDir() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return
	}

	exeDir := filepath.Dir(exePath)

	err = os.Chdir(exeDir)
	if err != nil {
		fmt.Println("Error changing working directory:", err)
		return
	}

	// 打印当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current working directory:", err)
		return
	}
	slog.Info(fmt.Sprintf("current work dir=%s", currentDir))
}
