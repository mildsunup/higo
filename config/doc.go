// Package config 提供配置加载和管理。
//
// 核心功能：
//   - 多源配置（文件、环境变量、远程配置中心）
//   - 配置热更新和监听
//   - 结构化配置映射
//
// 使用示例：
//
//	cfg, err := config.Load("config.yaml")
//	if err != nil {
//	    log.Fatal(err)
//	}
package config
