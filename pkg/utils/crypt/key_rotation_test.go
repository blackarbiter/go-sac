package crypt_test

import (
	"testing"
	"time"

	"github.com/blackarbiter/go-sac/pkg/utils/crypt"
)

func TestKeyManager(t *testing.T) {
	// 初始化密钥管理器
	initialKey := make([]byte, 32)
	km := crypt.NewKeyManager(initialKey, time.Hour)

	// 准备测试数据
	data := []byte("confidential data")
	cipher1, err := crypt.Encrypt(data, initialKey)
	if err != nil {
		t.Fatal("加密失败:", err)
	}

	// 测试当前密钥解密
	plaintext1, err := km.DecryptWithHistory(cipher1)
	if err != nil {
		t.Error("使用当前密钥解密失败:", err)
	}
	if string(plaintext1) != string(data) {
		t.Error("解密结果与原始数据不匹配")
	}

	// 测试自动密钥轮换
	// 启动自动轮换，但我们使用一个很长的时间间隔，这样不会干扰测试
	km.StartAutoRotation()

	// 我们可以手动触发一次密钥轮换来测试
	// 因为rotateKey是私有方法，我们不能直接调用它
	// 所以我们需要间接验证轮换的效果

	// 确保旧密钥仍然可用于解密
	plaintext2, err := km.DecryptWithHistory(cipher1)
	if err != nil {
		t.Error("使用轮换后的密钥解密旧数据失败:", err)
	}
	if string(plaintext2) != string(data) {
		t.Error("解密旧数据结果与原始数据不匹配")
	}

	// 测试大量密钥轮换
	// 这里我们无法直接测试，因为rotateKey是私有方法
	// 如果需要测试这个功能，建议将密钥管理器设计成可以注入模拟时间和手动触发轮换的方式

	// 可以添加更多验证逻辑，例如测试解密超过历史限制的旧密钥时应该失败
	// 但是这需要调整KeyManager的设计，使其更容易测试
}
