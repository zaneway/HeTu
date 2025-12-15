package helper

import (
	"HeTu/util"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/zaneway/cain-go/sm2"
)

// 定义tag和名称的映射关系
var TagToName = map[int]string{
	0:  "",
	1:  "BOOLEAN",
	2:  "INTEGER",
	3:  "BIT STRING",
	4:  "OCTET STRING",
	5:  "NULL",
	6:  "OBJECT IDENTIFIER",
	8:  "EXTERNAL",
	10: "ENUMERATED",
	12: "UTF8String",
	16: "SEQUENCE",
	17: "SET",
	19: "PrintableString",
	20: "T61String",
	22: "IA5String",
	23: "UTCTime",
	24: "GeneralizedTime",
	27: "TagGeneralString",
	30: "TagBMPString",
}

type ASN1Content struct {
	TypeName string
	RealType interface{}
}

var ClassToNum = map[int]int{
	0: 0,
	1: 64,  //0x40,第7位为1
	2: 128, //0x80,第8位为1
	3: 192, //0xc0,第8位和第7位为1
}

type ASN1Node struct {
	//this Tag is real Number in asn1
	Tag, Class, Length int
	Children           []*ASN1Node
	Value              string
	Content, FullBytes []byte
	Error              string // 添加错误信息字段
	Depth              int    // 添加深度字段防止无限递归
}

// ParseAsn1WithMaxDepth 带最大深度限制的ASN1解析
func ParseAsn1(data []byte) ASN1Node {
	return parseAsn1WithDepth(data, 0, 20) // 最大深度20层
}

// parseAsn1WithDepth 内部解析函数，带深度控制
func parseAsn1WithDepth(data []byte, currentDepth, maxDepth int) ASN1Node {
	var thisNode ASN1Node
	thisNode.Depth = currentDepth

	// 检查深度限制
	if currentDepth >= maxDepth {
		thisNode.Error = fmt.Sprintf("超过最大解析深度 %d", maxDepth)
		thisNode.Value = "解析深度过深"
		return thisNode
	}

	// 检查数据长度
	if len(data) == 0 {
		thisNode.Error = "数据为空"
		thisNode.Value = "空数据"
		return thisNode
	}

	// 检查数据长度是否过大（防止内存溢出）
	if len(data) > 10*1024*1024 { // 10MB限制
		thisNode.Error = fmt.Sprintf("数据过大: %d bytes", len(data))
		thisNode.Value = "数据过大，无法解析"
		return thisNode
	}

	var node asn1.RawValue

	// 安全的ASN1解析，捕获panic
	defer func() {
		if r := recover(); r != nil {
			thisNode.Error = fmt.Sprintf("解析panic: %v", r)
			thisNode.Value = "解析失败"
		}
	}()

	_, err := asn1.Unmarshal(data, &node)
	if err != nil {
		// 检查是否是IA5String相关的错误，如果是则尝试手动解析
		errStr := err.Error()
		// 更宽松的错误检测：只要包含 "IA5String" 和 "invalid" 或 "syntax error" 就尝试手动解析
		if strings.Contains(errStr, "IA5String") && (strings.Contains(errStr, "invalid") || strings.Contains(errStr, "syntax error")) {
			// 尝试手动解析IA5String
			manualNode, manualErr := parseIA5StringManually(data)
			if manualErr == nil {
				// 手动解析成功，使用手动解析的结果
				thisNode.Tag = manualNode.Tag
				thisNode.Class = manualNode.Class
				thisNode.Length = manualNode.Length
				thisNode.FullBytes = manualNode.FullBytes
				thisNode.Content = manualNode.Content
				// 添加警告信息，但不阻止解析
				thisNode.Error = fmt.Sprintf("警告: IA5String包含非标准字符，已尝试解析")
				// 继续处理，不返回
			} else {
				// 手动解析也失败，尝试更通用的手动解析
				genericNode, genericErr := parseASN1Manually(data)
				if genericErr == nil {
					thisNode.Tag = genericNode.Tag
					thisNode.Class = genericNode.Class
					thisNode.Length = genericNode.Length
					thisNode.FullBytes = genericNode.FullBytes
					thisNode.Content = genericNode.Content
					thisNode.Error = fmt.Sprintf("警告: ASN1解析遇到问题，已使用手动解析 (原始错误: %v)", err)
				} else {
					// 所有手动解析都失败，返回错误
					thisNode.Error = fmt.Sprintf("ASN1解析错误: %v (手动解析也失败: %v)", err, manualErr)
					thisNode.Value = "解析错误"
					return thisNode
				}
			}
		} else {
			// 其他类型的错误，也尝试通用手动解析
			genericNode, genericErr := parseASN1Manually(data)
			if genericErr == nil {
				thisNode.Tag = genericNode.Tag
				thisNode.Class = genericNode.Class
				thisNode.Length = genericNode.Length
				thisNode.FullBytes = genericNode.FullBytes
				thisNode.Content = genericNode.Content
				thisNode.Error = fmt.Sprintf("警告: ASN1解析遇到问题，已使用手动解析 (原始错误: %v)", err)
			} else {
				// 手动解析也失败，返回错误
				thisNode.Error = fmt.Sprintf("ASN1解析错误: %v", err)
				thisNode.Value = "解析错误"
				return thisNode
			}
		}
	} else {
		// 正常解析成功
		// 设置基本属性
		thisNode.Tag = node.Tag
		if node.IsCompound {
			thisNode.Tag = node.Tag + 32
		}
		thisNode.Tag += ClassToNum[node.Class]
		thisNode.Class = node.Class
		thisNode.Length = len(node.FullBytes)
		thisNode.FullBytes = node.FullBytes

		// 处理复合类型
		if node.IsCompound || isCompoundSafe(node) {
			thisNodeValue := node.Bytes
			childCount := 0
			maxChildren := 1000 // 限制子节点数量

			for len(thisNodeValue) > 0 && childCount < maxChildren {
				// 安全的子节点解析
				childNode := parseAsn1WithDepth(thisNodeValue, currentDepth+1, maxDepth)

				// 如果子节点解析失败，停止继续解析
				if childNode.Error != "" {
					break
				}

				thisNode.Children = append(thisNode.Children, &childNode)
				childCount++

				// 防止无限循环
				if childNode.Length <= 0 || childNode.Length > len(thisNodeValue) {
					break
				}

				thisNodeValue = thisNodeValue[childNode.Length:]

			}

			if childCount >= maxChildren {
				thisNode.Error = fmt.Sprintf("子节点过多，已截断（显示前%d个）", maxChildren)
			}

		} else {
			thisNode.Content = node.Bytes
		}

		// 构建显示值
		thisNode.Value = buildAsn1ValueSafe(thisNode)
		return thisNode
	}

	// 手动解析路径：处理复合类型
	if thisNode.Tag >= 32 && thisNode.Tag < 64 || thisNode.Tag >= 96 && thisNode.Tag < 128 ||
		thisNode.Tag >= 160 && thisNode.Tag < 192 || thisNode.Tag >= 224 && thisNode.Tag < 256 {
		// 可能是复合类型，尝试解析子节点
		thisNodeValue := thisNode.Content
		childCount := 0
		maxChildren := 1000

		for len(thisNodeValue) > 0 && childCount < maxChildren {
			childNode := parseAsn1WithDepth(thisNodeValue, currentDepth+1, maxDepth)
			if childNode.Error != "" {
				break
			}
			thisNode.Children = append(thisNode.Children, &childNode)
			childCount++
			if childNode.Length <= 0 || childNode.Length > len(thisNodeValue) {
				break
			}
			thisNodeValue = thisNodeValue[childNode.Length:]
		}
	}

	// 构建显示值
	thisNode.Value = buildAsn1ValueSafe(thisNode)
	return thisNode
}

// isCompoundSafe 安全的复合类型检查
func isCompoundSafe(node asn1.RawValue) bool {
	if len(node.Bytes) == 0 {
		return false
	}

	defer func() {
		// 捕获可能的panic
		recover()
	}()

	var nodeNext asn1.RawValue
	_, err := asn1.Unmarshal(node.Bytes, &nodeNext)
	return err == nil
}

// getOriginalTag 从修改后的tag中提取原始tag值
func getOriginalTag(tag int) int {
	// 根据tag的范围提取原始tag值
	if tag < 32 {
		// Universal Simple (0-31)
		return tag
	} else if 32 <= tag && tag < 64 {
		// Universal Structure (32-63)
		return tag - 32
	} else if 64 <= tag && tag < 96 {
		// Application Simple (64-95)
		return tag - 64
	} else if 96 <= tag && tag < 128 {
		// Application Structure (96-127)
		return tag - 96
	} else if 128 <= tag && tag < 160 {
		// Context Specific Simple (128-159)
		return tag - 128
	} else if 160 <= tag && tag < 192 {
		// Context Specific Structure (160-191)
		return tag - 160
	} else if 192 <= tag && tag < 224 {
		// Private Simple (192-223)
		return tag - 192
	} else if 224 <= tag && tag < 256 {
		// Private Structure (224-255)
		return tag - 224
	}
	// 如果超出范围，返回原值
	return tag
}

// buildAsn1ValueSafe 安全的ASN1值构建函数
func buildAsn1ValueSafe(node ASN1Node) (data string) {
	// 如果有错误，返回错误信息
	if node.Error != "" {
		return fmt.Sprintf("错误: %s", node.Error)
	}

	// 防止内容过大
	if len(node.Content) > 1024*1024 { // 1MB限制
		return fmt.Sprintf("内容过大 (%d bytes)", len(node.Content))
	}

	defer func() {
		if r := recover(); r != nil {
			data = fmt.Sprintf("值解析panic: %v", r)
		}
	}()

	// 默认显示Hex编码
	data = hex.EncodeToString(node.Content)

	// 提取原始tag值用于匹配
	originalTag := getOriginalTag(node.Tag)

	// 根据标签类型特殊处理
	switch originalTag {
	case 2: // INTEGER
		if bigInt, err := parseBigIntSafe(node.Content); err == nil {
			data = bigInt.String()
			if len(data) > 1000 {
				data = data[:1000] + "...已截断"
			}
		}
	case 3: // BIT STRING
		if ret, err := parseBitStringSafe(node.Content); err == nil {
			data = hex.EncodeToString(ret.Bytes)
			if len(data) > 200 {
				data = data[:200] + "...已截断"
			}
			// 尝试SM2签名解析（安全版本）
			if r, s, err := sm2SignDataSafe(ret.Bytes); err == nil {
				data = fmt.Sprintf("r: %s\ns: %s", r, s)
			}
		}
	case 6: // OBJECT IDENTIFIER
		if oid, err := ParseObjectIdentifierSafe(node.FullBytes); err == nil {
			data = oid
		}
	case 12: // UTF8String
		if isValidUTF8(node.Content) {
			data = string(node.Content)
			if len(data) > 500 {
				data = data[:500] + "...已截断"
			}
		}
	case 19: // PrintableString
		if isPrintableString(node.Content) {
			data = string(node.Content)
			if len(data) > 500 {
				data = data[:500] + "...已截断"
			}
		}
	case 22: // IA5String
		// 即使包含非标准字符，也尝试显示为字符串
		// 这样可以显示那些不符合IA5String标准但实际存在的内容
		if isIA5String(node.Content) {
			// 标准IA5String，直接显示
			data = string(node.Content)
			if len(data) > 500 {
				data = data[:500] + "...已截断"
			}
		} else {
			// 包含非标准字符，尝试显示但添加警告
			// 过滤掉控制字符，保留可打印字符
			var sb strings.Builder
			nonPrintableCount := 0
			for _, b := range node.Content {
				if b >= 32 && b < 127 {
					sb.WriteByte(b)
				} else if b == 9 || b == 10 || b == 13 {
					// 允许制表符、换行符、回车符
					sb.WriteByte(b)
				} else {
					nonPrintableCount++
					// 对于非打印字符，显示为转义序列
					sb.WriteString(fmt.Sprintf("\\x%02x", b))
				}
			}
			result := sb.String()
			if len(result) > 500 {
				result = result[:500] + "...已截断"
			}
			if nonPrintableCount > 0 {
				data = fmt.Sprintf("[警告: 包含%d个非标准字符] %s", nonPrintableCount, result)
			} else {
				data = result
			}
		}
	case 23: // UTCTime
		if timeStr, err := parseUTCTimeSafe(node.Content); err == nil {
			data = timeStr
		}
	case 24: // GeneralizedTime
		if timeStr, err := parseGeneralizedTimeSafe(node.Content); err == nil {
			data = timeStr
		}
	}

	return
}

// 安全的辅助函数

// parseBigIntSafe 安全的大整数解析
func parseBigIntSafe(bytes []byte) (*big.Int, error) {
	if len(bytes) > 1024 { // 限制大整数大小
		return nil, fmt.Errorf("数据过大")
	}
	return parseBigInt(bytes)
}

// parseBitStringSafe 安全的位字符串解析
func parseBitStringSafe(bytes []byte) (asn1.BitString, error) {
	if len(bytes) > 10240 { // 限制位字符串大小
		return asn1.BitString{}, fmt.Errorf("数据过大")
	}
	return parseBitString(bytes)
}

// sm2SignDataSafe 安全的SM2签名解析
func sm2SignDataSafe(data []byte) (string, string, error) {
	if len(data) > 1024 {
		return "", "", fmt.Errorf("数据过大")
	}

	defer func() {
		recover() // 防止panic
	}()

	r, s, err := sm2.SignDataToSignDigit(data)
	if err != nil {
		return "", "", err
	}

	return r.String(), s.String(), nil
}

// ParseObjectIdentifierSafe 安全的OID解析（导出版本）
func ParseObjectIdentifierSafe(fullBytes []byte) (string, error) {
	if len(fullBytes) > 1024 {
		return "", fmt.Errorf("数据过大")
	}

	defer func() {
		recover()
	}()

	var identifier asn1.ObjectIdentifier
	_, err := asn1.Unmarshal(fullBytes, &identifier)
	if err != nil {
		return "", err
	}

	oidStr := identifier.String()

	// 常见OID的友好名称
	knownOIDs := map[string]string{
		"1.2.840.113549.1.1.1":  "RSA",
		"1.2.840.10045.2.1":     "ECDSA",
		"1.2.156.10197.1.301":   "SM2",
		"1.2.840.113549.1.1.11": "SHA256WithRSA",
		"1.2.840.113549.1.1.5":  "SHA1WithRSA",
		"1.2.840.10045.4.3.2":   "SHA256WithECDSA",
		"2.5.4.3":               "commonName",
		"2.5.4.6":               "countryName",
		"2.5.4.7":               "localityName",
		"2.5.4.8":               "stateOrProvinceName",
		"2.5.4.10":              "organizationName",
		"2.5.4.11":              "organizationalUnitName",
	}

	if name, exists := knownOIDs[oidStr]; exists {
		return fmt.Sprintf("%s (%s)", name, oidStr), nil
	}

	return oidStr, nil
}

// parseObjectIdentifierSafe 安全的OID解析
func parseObjectIdentifierSafe(fullBytes []byte) (string, error) {
	return ParseObjectIdentifierSafe(fullBytes)
}

// parseUTCTimeSafe 安全的UTC时间解析
func parseUTCTimeSafe(content []byte) (string, error) {
	if len(content) > 100 {
		return "", fmt.Errorf("时间数据过长")
	}

	s := string(content)
	if !isValidTimeString(s) {
		return hex.EncodeToString(content), nil
	}

	defer func() {
		recover()
	}()

	parse, err := time.Parse(util.FormatStr, s)
	if err != nil {
		return s, nil // 返回原始字符串
	}

	return parse.Format(util.DateTime), nil
}

// parseGeneralizedTimeSafe 安全的通用时间解析
func parseGeneralizedTimeSafe(content []byte) (string, error) {
	if len(content) > 100 {
		return "", fmt.Errorf("时间数据过长")
	}

	s := string(content)
	if !isValidTimeString(s) {
		return hex.EncodeToString(content), nil
	}

	defer func() {
		recover()
	}()

	// 尝试多种时间格式
	timeFormats := []string{
		"20060102150405Z",
		"20060102150405-0700",
		"20060102150405+0700",
		util.FormatStr,
	}

	for _, format := range timeFormats {
		if parse, err := time.Parse(format, s); err == nil {
			return parse.Format(util.DateTime), nil
		}
	}

	return s, nil // 返回原始字符串
}

// 字符串验证函数

// isValidUTF8 检查是否为有效UTF8
func isValidUTF8(data []byte) bool {
	if len(data) > 10240 {
		return false
	}
	for _, b := range data {
		if b < 32 && b != 9 && b != 10 && b != 13 { // 控制字符检查
			return false
		}
	}
	return true
}

// isPrintableString 检查是否为可打印字符串
func isPrintableString(data []byte) bool {
	if len(data) > 10240 {
		return false
	}
	for _, b := range data {
		if !((b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') ||
			b == ' ' || b == '\'' || b == '(' || b == ')' || b == '+' || b == ',' ||
			b == '-' || b == '.' || b == '/' || b == ':' || b == '=' || b == '?') {
			return false
		}
	}
	return true
}

// isIA5String 检查是否为IA5字符串
func isIA5String(data []byte) bool {
	if len(data) > 10240 {
		return false
	}
	for _, b := range data {
		if b > 127 { // IA5只允许7位ASCII
			return false
		}
	}
	return true
}

// isValidTimeString 检查时间字符串格式
func isValidTimeString(s string) bool {
	if len(s) < 6 || len(s) > 30 {
		return false
	}
	// 简单检查是否包含数字
	digitCount := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			digitCount++
		}
	}
	return digitCount >= 6 // 至少月6个数字
}

// 原始解析函数（保留兼容性）

func parseBigInt(bytes []byte) (*big.Int, error) {
	if err := checkInteger(bytes); err != nil {
		return nil, err
	}
	ret := new(big.Int)
	if len(bytes) > 0 && bytes[0]&0x80 == 0x80 {
		// This is a negative number.
		notBytes := make([]byte, len(bytes))
		for i := range notBytes {
			notBytes[i] = ^bytes[i]
		}
		ret.SetBytes(notBytes)
		ret.Add(ret, bigOne)
		ret.Neg(ret)
		return ret, nil
	}
	ret.SetBytes(bytes)
	return ret, nil
}

var bigOne = big.NewInt(1)

func checkInteger(bytes []byte) error {
	if len(bytes) == 0 {
		return asn1.StructuralError{"empty integer"}
	}
	if len(bytes) == 1 {
		return nil
	}
	if (bytes[0] == 0 && bytes[1]&0x80 == 0) || (bytes[0] == 0xff && bytes[1]&0x80 == 0x80) {
		return asn1.StructuralError{"integer not minimally-encoded"}
	}
	return nil
}

// parseBitString parses an ASN.1 bit string from the given byte slice and returns it.
func parseBitString(bytes []byte) (ret asn1.BitString, err error) {
	if len(bytes) == 0 {
		err = asn1.SyntaxError{"zero length BIT STRING"}
		return
	}
	paddingBits := int(bytes[0])
	if paddingBits > 7 ||
		len(bytes) == 1 && paddingBits > 0 ||
		bytes[len(bytes)-1]&((1<<bytes[0])-1) != 0 {
		err = asn1.SyntaxError{"invalid padding bits in BIT STRING"}
		return
	}
	ret.BitLength = (len(bytes)-1)*8 - paddingBits
	ret.Bytes = bytes[1:]
	return
}

// ManualASN1Node 手动解析的ASN1节点结构
type ManualASN1Node struct {
	Tag       int
	Class     int
	Length    int
	FullBytes []byte
	Content   []byte
}

// parseASN1Manually 通用的手动解析ASN1结构，绕过标准库的严格验证
func parseASN1Manually(data []byte) (*ManualASN1Node, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("数据太短")
	}

	var node ManualASN1Node
	pos := 0

	// 解析tag和class
	tagByte := data[pos]
	pos++

	// 提取class (前2位)
	node.Class = int((tagByte >> 6) & 0x03)

	// 检查是否是复合类型 (第6位)
	isCompound := (tagByte & 0x20) != 0

	// 提取tag值
	if (tagByte & 0x1F) == 0x1F {
		// 长格式tag（多字节），这里简化处理
		return nil, fmt.Errorf("不支持长格式tag")
	}
	node.Tag = int(tagByte & 0x1F)

	// 如果是复合类型，添加32
	if isCompound {
		node.Tag += 32
	}

	// 添加class偏移
	node.Tag += ClassToNum[node.Class]

	// 解析长度
	if pos >= len(data) {
		return nil, fmt.Errorf("数据不足，无法解析长度")
	}

	lengthByte := data[pos]
	pos++

	var contentLength int
	var lengthBytes int

	if (lengthByte & 0x80) == 0 {
		// 短格式长度
		contentLength = int(lengthByte)
		lengthBytes = 1
	} else {
		// 长格式长度
		lengthBytesCount := int(lengthByte & 0x7F)
		if lengthBytesCount == 0 || lengthBytesCount > 4 {
			return nil, fmt.Errorf("长度字段格式错误")
		}

		if pos+lengthBytesCount > len(data) {
			return nil, fmt.Errorf("数据不足，无法读取长度字段")
		}

		contentLength = 0
		for i := 0; i < lengthBytesCount; i++ {
			contentLength = (contentLength << 8) | int(data[pos])
			pos++
		}
		lengthBytes = 1 + lengthBytesCount
	}

	// 计算总长度
	totalLength := 1 + lengthBytes + contentLength

	// 检查数据是否足够
	if pos+contentLength > len(data) {
		// 如果数据不足，使用实际可用数据
		contentLength = len(data) - pos
		totalLength = pos + contentLength
	}

	// 提取完整字节
	node.FullBytes = data[:totalLength]

	// 提取内容
	if pos+contentLength <= len(data) {
		node.Content = data[pos : pos+contentLength]
	} else {
		node.Content = data[pos:]
	}

	node.Length = totalLength

	return &node, nil
}

// parseIA5StringManually 手动解析IA5String，绕过标准库的严格验证
// 实际上调用通用的parseASN1Manually函数
func parseIA5StringManually(data []byte) (*ManualASN1Node, error) {
	return parseASN1Manually(data)
}
