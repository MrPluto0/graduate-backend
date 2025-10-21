package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateTaskID() string {
	// 生成 8 字节随机数,转换为 16 位十六进制字符串
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
