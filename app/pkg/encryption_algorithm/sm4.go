package encryption_algorithm

import (
	"bytes"
	"crypto/cipher"
	"fmt"

	"github.com/tjfoc/gmsm/sm4"
)

// SM4 分组大小（BlockSize）为 16 字节
const blockSize = sm4.BlockSize

// --- 填充/去填充函数 ---

// PKCS7Padding 对数据进行 PKCS#7 填充
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS7UnPadding 去除数据中的 PKCS#7 填充
func PKCS7UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	if length == 0 {
		return nil, fmt.Errorf("decryption failed: data is empty")
	}
	// 最后一个字节的值就是填充的长度
	unpadding := int(origData[length-1])
	if unpadding > length || unpadding == 0 {
		return nil, fmt.Errorf("decryption failed: invalid padding size %d", unpadding)
	}
	return origData[:(length - unpadding)], nil
}

// --- SM4 CBC 加密函数 ---

// Sm4CbcEncrypt 使用 SM4 CBC 模式加密数据
func Sm4CbcEncrypt(plainText, key, iv []byte) ([]byte, error) {
	// 检查 Key 和 IV 长度是否符合 SM4 标准
	if len(key) != blockSize {
		return nil, fmt.Errorf("invalid key length: got %d, want %d", len(key), blockSize)
	}
	if len(iv) != blockSize {
		return nil, fmt.Errorf("invalid iv length: got %d, want %d", len(iv), blockSize)
	}

	// 1. 创建 SM4 分组密码接口
	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 2. 填充明文
	plainText = PKCS7Padding(plainText, block.BlockSize())

	// 3. 创建 CBC 加密器
	blockMode := cipher.NewCBCEncrypter(block, iv)

	// 4. 加密
	cipherText := make([]byte, len(plainText))
	blockMode.CryptBlocks(cipherText, plainText)

	return cipherText, nil
}

// --- SM4 CBC 解密函数 ---

// Sm4CbcDecrypt 使用 SM4 CBC 模式解密数据
func Sm4CbcDecrypt(cipherText, key, iv []byte) ([]byte, error) {
	// 检查 Key 和 IV 长度是否符合 SM4 标准
	if len(key) != blockSize {
		return nil, fmt.Errorf("invalid key length: got %d, want %d", len(key), blockSize)
	}
	if len(iv) != blockSize {
		return nil, fmt.Errorf("invalid iv length: got %d, want %d", len(iv), blockSize)
	}

	// 1. 创建 SM4 分组密码接口
	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 2. 创建 CBC 解密器
	blockMode := cipher.NewCBCDecrypter(block, iv)

	// 3. 解密
	plainText := make([]byte, len(cipherText))
	blockMode.CryptBlocks(plainText, cipherText)

	// 4. 去除填充
	plainText, err = PKCS7UnPadding(plainText)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}

// --- 测试示例 ---
