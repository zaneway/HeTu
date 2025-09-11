package util

import (
	"encoding/base64"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"unicode"
)

func Base64EncodeToString(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func Base64DecodeFromString(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

func HexEncodeBytesToString(data []byte) string {
	return strings.ToUpper(hex.EncodeToString(data))
}

func HexEncodeIntToString(data int) string {
	result := strings.ToUpper(strconv.FormatInt(int64(data), 16))
	if len(result)%2 != 0 {
		return "0" + result
	}
	return result
}

func HexDecodeStringToBytes(data string) ([]byte, error) {
	return hex.DecodeString(data)
}

func HexDecodeStringToInt(data string) (int, error) {
	i, err := strconv.ParseInt(data, 16, 64)
	return int(i), err
}

// IsASCIIOrChinese 判断内容是否为ASCII字符或汉字
func IsASCIIOrChinese(content []byte) bool {
	s := string(content)
	for _, r := range s {
		// ASCII字符范围是0-127
		if r <= 127 || unicode.Is(unicode.Han, r) {
			continue
		}
		return false
	}
	return true
}

// ReadFileContent 读取文件内容
func ReadFileContent(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}
