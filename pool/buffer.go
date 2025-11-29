package pool

import (
	"bytes"
	"strings"
	"sync"
)

// 预定义的缓冲池
var (
	// BufferPool 字节缓冲池
	BufferPool = NewObject(
		func() *bytes.Buffer { return new(bytes.Buffer) },
		func(b *bytes.Buffer) { b.Reset() },
	)

	// ByteSlicePool 字节切片池（按大小分级）
	byteSlicePools = [...]sync.Pool{
		{New: func() any { return make([]byte, 0, 64) }},      // 64B
		{New: func() any { return make([]byte, 0, 256) }},     // 256B
		{New: func() any { return make([]byte, 0, 1024) }},    // 1KB
		{New: func() any { return make([]byte, 0, 4096) }},    // 4KB
		{New: func() any { return make([]byte, 0, 16384) }},   // 16KB
		{New: func() any { return make([]byte, 0, 65536) }},   // 64KB
	}
	byteSizes = [...]int{64, 256, 1024, 4096, 16384, 65536}
)

// GetBuffer 获取缓冲区
func GetBuffer() *bytes.Buffer {
	return BufferPool.Get()
}

// PutBuffer 归还缓冲区
func PutBuffer(buf *bytes.Buffer) {
	BufferPool.Put(buf)
}

// GetBytes 获取指定大小的字节切片
func GetBytes(size int) []byte {
	idx := selectPool(size)
	if idx < 0 {
		return make([]byte, 0, size)
	}
	return byteSlicePools[idx].Get().([]byte)[:0]
}

// PutBytes 归还字节切片
func PutBytes(b []byte) {
	cap := cap(b)
	idx := selectPool(cap)
	if idx >= 0 && cap == byteSizes[idx] {
		byteSlicePools[idx].Put(b[:0])
	}
}

func selectPool(size int) int {
	for i, s := range byteSizes {
		if size <= s {
			return i
		}
	}
	return -1
}

// StringBuilderPool 字符串构建器池
var StringBuilderPool = &sync.Pool{
	New: func() any { return new(strings.Builder) },
}

// GetStringBuilder 获取字符串构建器
func GetStringBuilder() *strings.Builder {
	return StringBuilderPool.Get().(*strings.Builder)
}

// PutStringBuilder 归还字符串构建器
func PutStringBuilder(sb *strings.Builder) {
	sb.Reset()
	StringBuilderPool.Put(sb)
}
