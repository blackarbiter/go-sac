package compression

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipCompression(t *testing.T) {
	t.Run("基本压缩和解压缩", func(t *testing.T) {
		// 测试数据
		originalData := []byte("Hello, this is a test message for gzip compression")

		// 压缩数据
		compressed, err := GzipCompress(originalData)
		require.NoError(t, err)
		require.NotNil(t, compressed)

		// 确认压缩是有效的（压缩后的数据应该小于或等于原始数据，但要考虑特殊情况）
		// 对于非常小的数据，压缩后可能会更大，所以不做严格的大小比较

		// 确认压缩数据与原始数据不同
		assert.False(t, bytes.Equal(originalData, compressed), "压缩数据应该与原始数据不同")

		// 解压缩数据
		decompressed, err := GzipDecompress(compressed)
		require.NoError(t, err)
		require.NotNil(t, decompressed)

		// 验证解压后的数据与原始数据相同
		assert.Equal(t, originalData, decompressed, "解压缩后的数据应该与原始数据相同")
	})

	t.Run("压缩空数据", func(t *testing.T) {
		// 测试空数据
		emptyData := []byte{}

		// 压缩空数据
		compressed, err := GzipCompress(emptyData)
		require.NoError(t, err)
		require.NotNil(t, compressed)

		// 解压缩
		decompressed, err := GzipDecompress(compressed)
		require.NoError(t, err)

		// 验证结果
		assert.Equal(t, emptyData, decompressed, "解压缩空数据应该返回空数据")
	})

	t.Run("压缩大数据", func(t *testing.T) {
		// 创建一个大数据块
		largeData := make([]byte, 1024*1024) // 1MB
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}

		// 压缩大数据
		compressed, err := GzipCompress(largeData)
		require.NoError(t, err)
		require.NotNil(t, compressed)

		// 对于重复性高的数据，压缩应该有效
		assert.Less(t, len(compressed), len(largeData), "大型重复数据应该可以有效压缩")

		// 解压缩
		decompressed, err := GzipDecompress(compressed)
		require.NoError(t, err)
		require.NotNil(t, decompressed)

		// 验证解压后的数据与原始数据相同
		assert.Equal(t, largeData, decompressed, "解压缩后的大数据应该与原始数据相同")
	})

	t.Run("解压缩无效数据", func(t *testing.T) {
		// 尝试解压缩无效数据
		invalidData := []byte("this is not compressed data")

		// 解压应该失败
		_, err := GzipDecompress(invalidData)
		assert.Error(t, err, "解压缩无效数据应该返回错误")
	})

	t.Run("压缩-解压缩循环", func(t *testing.T) {
		// 测试多次压缩和解压缩
		originalData := []byte("test data for multiple compression cycles")

		// 第一次压缩
		compressed1, err := GzipCompress(originalData)
		require.NoError(t, err)

		// 第二次压缩（压缩已压缩的数据）
		compressed2, err := GzipCompress(compressed1)
		require.NoError(t, err)

		// 第一次解压缩
		decompressed1, err := GzipDecompress(compressed2)
		require.NoError(t, err)

		// 第二次解压缩
		decompressed2, err := GzipDecompress(decompressed1)
		require.NoError(t, err)

		// 验证最终解压的数据与原始数据相同
		assert.Equal(t, originalData, decompressed2, "多次压缩解压缩后的数据应该与原始数据相同")
	})
}
