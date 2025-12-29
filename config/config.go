package config

import (
	"encoding/json"
	"io/ioutil"
)

// DMConfig 达梦数据库配置
type DMConfig struct {
	DSN string
}

// MySQLConfig MySQL数据库配置
type MySQLConfig struct {
	DSN string
}

// TablesConfig 表配置
type TablesConfig struct {
	Tables []string `json:"tables"`
}

// Config 迁移配置
type Config struct {
	DM         DMConfig
	MySQL      MySQLConfig
	SchemaOnly bool
}

// LoadTablesConfig 从JSON文件加载表配置
func LoadTablesConfig(filename string) (*TablesConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var tablesConfig TablesConfig
	err = json.Unmarshal(data, &tablesConfig)
	if err != nil {
		return nil, err
	}

	return &tablesConfig, nil
}