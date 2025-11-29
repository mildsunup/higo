package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// --- UUID ---

// UUID 生成 UUID v4
func UUID() string {
	return uuid.New().String()
}

// UUIDShort 生成短 UUID（去掉横线）
func UUIDShort() string {
	id := uuid.New()
	return hex.EncodeToString(id[:])
}

// --- 随机 ID ---

// RandomID 生成随机 ID（指定长度的十六进制字符串）
func RandomID(length int) string {
	bytes := make([]byte, (length+1)/2)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

// RandomBase64 生成 Base64 编码的随机字符串
func RandomBase64(length int) string {
	bytes := make([]byte, length)
	_, _ = rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

// --- 雪花 ID ---

var (
	snowflakeEpoch    int64 = 1704067200000 // 2024-01-01 00:00:00 UTC
	snowflakeNodeID   int64 = 1
	snowflakeSequence atomic.Int64
	snowflakeLastTime atomic.Int64
)

// SetSnowflakeNode 设置雪花 ID 节点
func SetSnowflakeNode(nodeID int64) {
	snowflakeNodeID = nodeID & 0x3FF // 10 bits
}

// Snowflake 生成雪花 ID
// 格式: 41位时间戳 + 10位节点ID + 12位序列号
func Snowflake() int64 {
	for {
		now := time.Now().UnixMilli() - snowflakeEpoch
		last := snowflakeLastTime.Load()

		if now == last {
			seq := snowflakeSequence.Add(1) & 0xFFF // 12 bits
			if seq == 0 {
				// 序列号用尽，等待下一毫秒
				for now <= last {
					now = time.Now().UnixMilli() - snowflakeEpoch
				}
			}
			if snowflakeLastTime.CompareAndSwap(last, now) {
				return (now << 22) | (snowflakeNodeID << 12) | seq
			}
		} else if now > last {
			if snowflakeLastTime.CompareAndSwap(last, now) {
				snowflakeSequence.Store(0)
				return (now << 22) | (snowflakeNodeID << 12)
			}
		}
		// CAS 失败，重试
	}
}

// SnowflakeString 生成字符串格式的雪花 ID
func SnowflakeString() string {
	return fmt.Sprintf("%d", Snowflake())
}

// --- 有序 ID ---

// OrderedID 生成有序 ID（时间戳 + 随机数）
func OrderedID() string {
	timestamp := time.Now().UnixNano()
	random := RandomID(8)
	return fmt.Sprintf("%d%s", timestamp, random)
}

// --- 短 ID ---

const shortIDChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// ShortID 生成短 ID（指定长度，使用 62 进制字符）
func ShortID(length int) string {
	bytes := make([]byte, length)
	_, _ = rand.Read(bytes)
	for i := range bytes {
		bytes[i] = shortIDChars[bytes[i]%62]
	}
	return string(bytes)
}

// NanoID 生成 NanoID（默认 21 位）
func NanoID() string {
	return ShortID(21)
}
