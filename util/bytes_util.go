package util

import (
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"
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
