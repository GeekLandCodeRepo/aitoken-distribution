package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

const (
	keyPrefix = "sk-"
	keyLength = 48
)

// GenerateAPIKey 生成 API Key
// 返回: 明文key, hash, prefix, suffix, error
func GenerateAPIKey() (plainKey, hash, prefix, suffix string, err error) {
	// 生成随机字节
	randomBytes := make([]byte, keyLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// 转换为base62字符
	keyBody := toBase62(randomBytes)
	plainKey = keyPrefix + keyBody

	// 计算 SHA-256 hash
	hashBytes := sha256.Sum256([]byte(plainKey))
	hash = hex.EncodeToString(hashBytes[:])

	// 取展示用前六位和后六位，不保存完整密钥
	prefix = plainKey[:6]
	suffix = plainKey[len(plainKey)-6:]

	return plainKey, hash, prefix, suffix, nil
}

// HashAPIKey 计算 API Key 的 hash
func HashAPIKey(key string) string {
	hashBytes := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hashBytes[:])
}

func toBase62(bytes []byte) string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	result := strings.Builder{}

	for _, b := range bytes {
		num := big.NewInt(int64(b))
		for num.Int64() > 0 {
			mod := new(big.Int)
			num.DivMod(num, big.NewInt(62), mod)
			result.WriteByte(charset[mod.Int64()])
		}
	}

	// 确保长度一致
	resultStr := result.String()
	if len(resultStr) < keyLength {
		// 补充随机字符
		for i := len(resultStr); i < keyLength; i++ {
			idx, _ := rand.Int(rand.Reader, big.NewInt(62))
			result.WriteByte(charset[idx.Int64()])
		}
	}

	return result.String()[:keyLength]
}
