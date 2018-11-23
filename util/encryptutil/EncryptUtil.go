package encryptutil

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

func AESCBCPKCS5EncryptBase64(str32 string, data string) (string, error) {
	bs, err := AESCBCPKCS5Encrypt([]byte(str32[0:16]), []byte(str32[16:32]), []byte(data))
	return base64.StdEncoding.EncodeToString(bs), err
}
func AESCBCPKCS5Encrypt(key []byte, vector []byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ecb := cipher.NewCBCEncrypter(block, vector)
	content := PKCS5Padding(data, block.BlockSize())
	crypted := make([]byte, len(content))
	ecb.CryptBlocks(crypted, content)
	return crypted, nil
}
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
func AESCBCPKCS5DecryptBase64(ciphertext, key32 string) (string, error) {
	c, err := base64.StdEncoding.DecodeString(ciphertext)
	if nil != err {
		return "", err
	}
	bs, err := AESCBCPKCS5Decrypt(c, []byte(key32))
	return string(bs), err
}

func AESCBCPKCS5Decrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[0:16]) //选择加密算法
	if err != nil {
		return nil, err
	}
	blockModel := cipher.NewCBCDecrypter(block, key[16:32])
	plantText := make([]byte, len(ciphertext))
	blockModel.CryptBlocks(plantText, ciphertext)
	plantText = PKCS5Unpadding(plantText)
	return plantText, nil
}

func PKCS5Unpadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
