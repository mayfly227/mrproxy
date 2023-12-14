package common

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Config struct {
	config map[string]any
}

var once sync.Once
var globalConfig *Config

func (c *Config) GetNode(name string) any {
	return c.config[name]
}

func (c *Config) readConfig(fileName string) ([]byte, error) {
	fileContent, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil, err
	}
	return fileContent, nil
}

func NewConfig(fileName string) (*Config, error) {
	c := &Config{config: nil}
	readConfig, err := c.readConfig(fileName)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(readConfig, &c.config)
	if err != nil {
		return nil, err
	}
	return c, err
}

// InitGlobalConfig GetInstance 返回 Singleton 的唯一实例
func InitGlobalConfig(fileName string) {
	once.Do(func() {
		globalConfig, _ = NewConfig(fileName)
		if globalConfig == nil {
			panic("load config error")
		}

	})
}

func GetConfig() *Config {
	return globalConfig
}
