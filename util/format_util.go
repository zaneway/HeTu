package util

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
)

// IsJSON 检查字符串是否为有效的JSON格式（支持对象和数组）
func IsJSON(str string) bool {
	trimmed := strings.TrimSpace(str)
	if trimmed == "" {
		return false
	}

	// 检查是否以 { 或 [ 开头，这是 JSON 对象或数组的标志
	firstChar := trimmed[0]
	if firstChar != '{' && firstChar != '[' {
		return false
	}

	// 尝试解析为 JSON（支持对象和数组）
	var js interface{}
	err := json.Unmarshal([]byte(trimmed), &js)
	return err == nil
}

// FormatJSON 格式化JSON字符串（支持对象和数组）
func FormatJSON(jsonStr string) (string, error) {
	// 先清理输入数据
	trimmed := strings.TrimSpace(jsonStr)
	if trimmed == "" {
		return "", fmt.Errorf("输入数据为空")
	}

	var jsonObj interface{}
	err := json.Unmarshal([]byte(trimmed), &jsonObj)
	if err != nil {
		return "", err
	}

	formattedBytes, err := json.MarshalIndent(jsonObj, "", "  ")
	if err != nil {
		return "", err
	}

	return string(formattedBytes), nil
}

// IsXML 检查字符串是否为有效的XML格式
func IsXML(str string) bool {
	trimmed := strings.TrimSpace(str)
	if !strings.HasPrefix(trimmed, "<") {
		return false
	}

	var x interface{}
	err := xml.Unmarshal([]byte(trimmed), &x)
	return err == nil
}

// FormatXML 格式化XML字符串
func FormatXML(xmlStr string) (string, error) {
	var x interface{}
	err := xml.Unmarshal([]byte(xmlStr), &x)
	if err != nil {
		return "", err
	}

	formattedBytes, err := xml.MarshalIndent(x, "", "  ")
	if err != nil {
		return "", err
	}

	return string(formattedBytes), nil
}
