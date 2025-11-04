package util

import (
	"encoding/json"
	"encoding/xml"
	"strings"
)

// IsJSON 检查字符串是否为有效的JSON格式
func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// FormatJSON 格式化JSON字符串
func FormatJSON(jsonStr string) (string, error) {
	var jsonObj interface{}
	err := json.Unmarshal([]byte(jsonStr), &jsonObj)
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
