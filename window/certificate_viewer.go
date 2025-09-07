package window

import (
	"HeTu/helper"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	. "github.com/zaneway/cain-go/x509"
)

// 构造解析证书核心图形模块
func CertificateStructure(input *widget.Entry) *fyne.Container {
	structure := container.NewVBox()
	detail := container.NewVBox()
	input.Wrapping = fyne.TextWrapWord
	// 为公共输入框设置适当的高度
	//input := buildInputCertEntry("Please input base64/hex cert")

	//inputCertEntry.Text = "MIICETCCAbWgAwIBAgINKl81oFaaablKOp0YTjAMBggqgRzPVQGDdQUAMGExCzAJBgNVBAYMAkNOMQ0wCwYDVQQKDARCSkNBMSUwIwYDVQQLDBxCSkNBIEFueXdyaXRlIFRydXN0IFNlcnZpY2VzMRwwGgYDVQQDDBNUcnVzdC1TaWduIFNNMiBDQS0xMB4XDTIwMDgxMzIwMTkzNFoXDTIwMTAyNDE1NTk1OVowHjELMAkGA1UEBgwCQ04xDzANBgNVBAMMBuWGr+i9rDBZMBMGByqGSM49AgEGCCqBHM9VAYItA0IABAIF97Sqq0Rv616L2PjFP3xt16QGJLmi+W8Ht+NLHiXntgUey0Nz+ZVnSUKUMzkKuGTikY3h2v7la20b6lpKo8WjgZIwgY8wCwYDVR0PBAQDAgbAMB0GA1UdDgQWBBSxiaS6z4Uguz3MepS2zblkuAF/LTAfBgNVHSMEGDAWgBTMZyRCGsP4rSes0vLlhIEf6cUvrjBABgNVHSAEOTA3MDUGCSqBHIbvMgICAjAoMCYGCCsGAQUFBwIBFhpodHRwOi8vd3d3LmJqY2Eub3JnLmNuL2NwczAMBggqgRzPVQGDdQUAA0gAMEUCIG6n6PG0BOK1EdFcvetQlC+9QhpsTuTui2wkeqWiPKYWAiEAvqR8Z+tSiYR5DIs7SyHJPWZ+sa8brtQL/1jURvHGxU8="
	//确认按钮
	confirm := buildButton("确认", theme.ConfirmIcon(), func() {
		inputCert := strings.TrimSpace(input.Text)
		if inputCert == "" {
			dialog.ShowError(fmt.Errorf("请输入证书数据"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		detail.RemoveAll()

		// 尝试Base64解码
		var decodeCert []byte
		var err error

		// 检查是否是PEM格式
		if strings.Contains(inputCert, "-----BEGIN CERTIFICATE-----") {
			// 处理PEM格式证书
			decodeCert, err = parsePEMCertificate(inputCert)
			if err != nil {
				dialog.ShowError(fmt.Errorf("PEM格式证书解析失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
				return
			}
		} else {
			// 清理输入数据，移除空格和换行符
			cleanedInput := cleanInputData(inputCert)

			// 尝试Base64解码
			decodeCert, err = base64.StdEncoding.DecodeString(cleanedInput)
			if err != nil {
				// 如果Base64失败，尝试Hex解码
				decodeCert, err = hex.DecodeString(cleanedInput)
				if err != nil {
					dialog.ShowError(fmt.Errorf("无法解码输入数据，请确保输入的是有效的Base64、Hex或PEM格式证书数据\n\n输入数据长度: %d\n清理后数据长度: %d\n\nBase64错误: %v\nHex错误: %v", len(inputCert), len(cleanedInput), err, err), fyne.CurrentApp().Driver().AllWindows()[0])
					return
				}
			}
		}

		// 验证解码后的数据长度
		if len(decodeCert) < 50 { // 证书通常至少有几百字节
			dialog.ShowError(fmt.Errorf("解码后的数据太短（%d 字节），不像是有效的证书数据", len(decodeCert)), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		// 解析证书
		certificate, err := helper.ParseCertificate(decodeCert)
		if err != nil {
			dialog.ShowError(fmt.Errorf("证书解析失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		//构造证书解析详情
		keys, value := buildCertificateDetail(certificate)

		//展示证书详情
		showCertificateDetail(keys, value, detail)

		// 解析并展示证书扩展项
		if len(certificate.Extensions) > 0 {
			extensionKeys, extensionValues := buildCertificateExtensions(certificate)
			showCertificateExtensions(extensionKeys, extensionValues, detail)
		}
	})
	//清除按钮
	clear := buildButton("清除", theme.CancelIcon(), func() {
		input.Text = ""
		input.Refresh()
	})

	//对所有按钮进行表格化
	allButton := container.New(layout.NewGridLayout(2), confirm, clear)
	// 不添加全局输入框，它已经在主界面的固定位置
	// structure.Add(input)
	structure.Add(allButton)
	structure.Add(detail)

	// 使用带滚动条的容器包装
	scrollContainer := container.NewScroll(structure)
	scrollContainer.SetMinSize(fyne.NewSize(600, 400))
	return container.NewMax(scrollContainer)
}

func buildCertificateDetail(certificate *Certificate) (keys []string, certDetail map[string]string) {
	certDetail = make(map[string]string)
	//有序的key放切片，值对应在map
	keys = []string{"SerialNumber", "SubjectName", "IssueName", "NotBefore", "NotAfter", "PublicKey", "PublicKeyAlgorithm", "SignatureAlgorithm"}
	//SerialNumber
	certDetail[keys[0]] = hex.EncodeToString(certificate.SerialNumber.Bytes())
	//SubjectName
	certDetail[keys[1]] = certificate.Subject.String()
	//IssueName
	certDetail[keys[2]] = certificate.Issuer.String()
	//NotBefore
	certDetail[keys[3]] = certificate.NotBefore.String()
	//NotAfter
	certDetail[keys[4]] = certificate.NotAfter.String()
	//PublicKeyAlgorithm
	certDetail[keys[5]] = base64.StdEncoding.EncodeToString(certificate.RawSubjectPublicKeyInfo)
	//PublicKey Alg
	certDetail[keys[6]] = ParsePublicKeyAlg(certificate.PublicKeyAlgorithm)
	//SignatureAlgorithm
	//.String()被重构
	certDetail[keys[7]] = certificate.SignatureAlgorithm.String()
	//KeyUsage
	//certDetail[keys[8]] = helper.ParseKeyUsage(certificate.KeyUsage)

	return keys, certDetail
}

// 解析证书扩展项
func buildCertificateExtensions(certificate *Certificate) (keys []string, certExtensions map[string]string) {
	certExtensions = make(map[string]string)
	keys = make([]string, 0)

	// OID到名称的映射
	oidToName := map[string]string{
		"2.5.29.14":         "Subject Key Identifier",
		"2.5.29.15":         "Key Usage",
		"2.5.29.17":         "Subject Alternative Name",
		"2.5.29.19":         "Basic Constraints",
		"2.5.29.31":         "CRL Distribution Points",
		"2.5.29.32":         "Certificate Policies",
		"2.5.29.35":         "Authority Key Identifier",
		"2.5.29.37":         "Extended Key Usage",
		"1.3.6.1.5.5.7.1.1": "Authority Information Access",
	}

	for i, ext := range certificate.Extensions {
		// 获取扩展项的OID
		oidStr := ext.Id.String()

		// 根据OID获取扩展项名称
		name, exists := oidToName[oidStr]
		if !exists {
			name = fmt.Sprintf("Extension %d (%s)", i+1, oidStr)
		} else {
			name = fmt.Sprintf("%s (%s)", name, oidStr)
		}

		// 添加到keys切片中
		keys = append(keys, name)

		// 根据OID类型解析扩展项值
		var value string
		switch oidStr {
		case "2.5.29.32": // Certificate Policies
			value = parseCertificatePolicies(ext.Value)
		case "2.5.29.14": // Subject Key Identifier
			value = parseSubjectKeyIdentifier(ext.Value)
		case "2.5.29.35": // Authority Key Identifier
			value = parseAuthorityKeyIdentifier(ext.Value)
		case "2.5.29.19": // Basic Constraints
			value = parseBasicConstraints(ext.Value)
		case "2.5.29.15": // Key Usage
			value = parseKeyUsage(ext.Value)
		case "2.5.29.37": // Extended Key Usage
			value = parseExtendedKeyUsage(ext.Value)
		default:
			// 默认情况下，将扩展项的值转换为十六进制字符串
			value = hex.EncodeToString(ext.Value)
			if len(value) > 1000 {
				value = value[:1000] + "...(已截断)"
			}
		}

		// 如果是关键扩展项，添加标记
		if ext.Critical {
			value = "[Critical] " + value
		}

		certExtensions[name] = value
	}

	return keys, certExtensions
}

// 解析证书策略扩展项
func parseCertificatePolicies(data []byte) string {
	// 尝试解析证书策略中的CPS URL
	result := fmt.Sprintf("Certificate Policies (Length: %d bytes)\n", len(data))

	// 显示原始十六进制值
	hexValue := hex.EncodeToString(data)
	result += fmt.Sprintf("Raw Hex: %s\n\n", formatHexDisplay(hexValue))

	// 尝试解析ASN.1结构以提取CPS URL
	cpsUrls := extractCPSUrls(data)
	if len(cpsUrls) > 0 {
		result += "CPS (Certificate Practice Statement) URLs:\n"
		for i, url := range cpsUrls {
			result += fmt.Sprintf("  %d. %s\n", i+1, url)
		}
		result += "\n"
	} else {
		result += "未找到CPS URLs或解析失败\n\n"
	}

	// 添加解析说明
	result += "解析说明:\n"
	result += "证书策略扩展包含以下信息:\n"
	result += "1. 策略标识符(Policy Identifiers) - OID格式\n"
	result += "2. 可选的策略限定符(Policy Qualifiers) - 包括CPS指针和用户声明\n"

	// 添加常见证书策略OID的说明
	result += "\n常见证书策略OID:\n"
	result += "- 1.3.6.1.5.5.7.2.1: CPS (Certificate Practice Statement)\n"
	result += "- 1.3.6.1.5.5.7.2.2: User Notice\n"

	// 如果数据较大，提醒用户使用专业工具进行完整解析
	if len(data) > 200 {
		result += "\n注意: 对于复杂的证书策略结构，建议使用专业ASN.1解析工具进行完整解析"
	}

	return result
}

// 格式化十六进制显示
func formatHexDisplay(hexStr string) string {
	var result string
	for i := 0; i < len(hexStr); i += 32 {
		end := i + 32
		if end > len(hexStr) {
			end = len(hexStr)
		}
		result += hexStr[i:end] + "\n"
	}
	return result
}

// 从证书策略数据中提取CPS URLs
func extractCPSUrls(data []byte) []string {
	var urls []string

	// 将字节数据转换为十六进制字符串
	hexData := hex.EncodeToString(data)

	// 查找http://或https://模式
	httpPattern := "687474703a2f2f"    // "http://"
	httpsPattern := "68747470733a2f2f" // "https://"

	// 查找所有可能的URL
	urls = append(urls, findUrlsInHex(hexData, httpPattern)...)
	urls = append(urls, findUrlsInHex(hexData, httpsPattern)...)

	// 去重
	uniqueUrls := []string{}
	seen := make(map[string]bool)
	for _, url := range urls {
		if !seen[url] {
			seen[url] = true
			uniqueUrls = append(uniqueUrls, url)
		}
	}

	return uniqueUrls
}

// 在十六进制数据中查找URL
func findUrlsInHex(hexData, pattern string) []string {
	var urls []string

	// 查找所有匹配的模式
	start := 0
	for {
		idx := strings.Index(hexData[start:], pattern)
		if idx == -1 {
			break
		}

		// 计算实际索引位置
		actualIdx := start + idx

		// 提取URL
		url := extractURLFromHex(hexData, actualIdx)
		if url != "" {
			urls = append(urls, url)
		}

		// 移动起始位置
		start = actualIdx + len(pattern)
	}

	return urls
}

// 从十六进制数据中提取URL
func extractURLFromHex(hexData string, startIndex int) string {
	// 从指定位置开始，尝试提取URL直到遇到结束字符
	// 这是一个简化的实现，实际应用中可能需要更复杂的ASN.1解析

	// 查找URL的结束位置
	endIndex := len(hexData)

	// 查找可能的结束标记
	// 在ASN.1中，URL通常以0x00或下一个结构标记结束
	for i := startIndex; i < len(hexData)-2; i += 2 {
		// 检查是否是结束标记
		if i+2 <= len(hexData) {
			nextBytes := hexData[i : i+2]
			// 00可能是字符串结束标记
			// 13是IA5String的标记
			// 0c是UTF8String的标记
			if nextBytes == "00" || nextBytes == "13" || nextBytes == "0c" {
				endIndex = i
				break
			}
		}
	}

	// 提取URL部分的十六进制数据
	urlHex := hexData[startIndex:endIndex]

	// 将十六进制转换为字节，再转换为字符串
	bytes, err := hex.DecodeString(urlHex)
	if err != nil {
		return ""
	}

	// 检查是否是有效的URL字符
	urlStr := string(bytes)
	if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") {
		// 验证URL格式，确保包含域名
		if strings.Contains(urlStr, ".") && len(urlStr) > 10 {
			// 简单清理URL，移除可能的尾随字符
			// 查找URL的结束位置
			endChars := []string{"\x00", "\x13", "\x0c", "\n", "\r"}
			for _, endChar := range endChars {
				if idx := strings.Index(urlStr, endChar); idx != -1 {
					urlStr = urlStr[:idx]
				}
			}
			return urlStr
		}
	}

	return ""
}

// 解析主题密钥标识符扩展项
func parseSubjectKeyIdentifier(data []byte) string {
	return "Key ID: " + hex.EncodeToString(data)
}

// 解析颁发机构密钥标识符扩展项
func parseAuthorityKeyIdentifier(data []byte) string {
	// 简化的解析，实际结构可能更复杂
	return "Authority Key ID: " + hex.EncodeToString(data)
}

// 解析基本约束扩展项
func parseBasicConstraints(data []byte) string {
	// 简化的解析
	hexValue := hex.EncodeToString(data)
	if len(hexValue) > 1000 {
		hexValue = hexValue[:1000] + "...(已截断)"
	}
	return "Basic Constraints (Hex): " + hexValue
}

// 解析密钥用法扩展项
func parseKeyUsage(data []byte) string {
	// 尝试解析密钥用法的ASN.1结构
	result := fmt.Sprintf("Key Usage (Length: %d bytes)\n", len(data))

	// 显示原始十六进制值
	hexValue := hex.EncodeToString(data)
	result += fmt.Sprintf("Raw Hex: %s\n\n", formatHexDisplay(hexValue))

	// 尝试解析密钥用法的位图
	keyUsage := parseKeyUsageBitmap(data)
	if keyUsage != "" {
		result += "解析结果:\n"
		result += keyUsage
	} else {
		result += "解析失败，显示原始十六进制值\n"
	}

	// 添加常见密钥用法说明
	result += "\n常见密钥用法位定义:\n"
	result += "  bit 0: Digital Signature (数字签名)\n"
	result += "  bit 1: Non Repudiation (不可否认)\n"
	result += "  bit 2: Key Encipherment (密钥加密)\n"
	result += "  bit 3: Data Encipherment (数据加密)\n"
	result += "  bit 4: Key Agreement (密钥协商)\n"
	result += "  bit 5: Cert Sign (证书签发)\n"
	result += "  bit 6: CRL Sign (CRL签发)\n"
	result += "  bit 7: Encipher Only (仅加密)\n"
	result += "  bit 8: Decipher Only (仅解密)\n"

	return result
}

// 解析密钥用法位图
func parseKeyUsageBitmap(data []byte) string {
	// 提取位图数据
	var keyUsageBytes []byte

	// 如果数据以BIT STRING标记开始 (0x03)
	if len(data) >= 3 && data[0] == 0x03 {
		length := int(data[1])
		if length > 0 && 2+length <= len(data) {
			// unusedBits := data[2]  // 未使用位数
			if length > 1 {
				keyUsageBytes = data[3 : 2+length]
			}
		}
	} else {
		// 假设数据本身就是位图
		keyUsageBytes = data
	}

	if len(keyUsageBytes) == 0 {
		return ""
	}

	// 解析位图
	var usageStrings []string

	// 定义密钥用法位的含义
	usageMap := map[int]string{
		0: "Digital Signature (数字签名)",
		1: "Non Repudiation (不可否认)",
		2: "Key Encipherment (密钥加密)",
		3: "Data Encipherment (数据加密)",
		4: "Key Agreement (密钥协商)",
		5: "Cert Sign (证书签发)",
		6: "CRL Sign (CRL签发)",
		7: "Encipher Only (仅加密)",
		8: "Decipher Only (仅解密)",
	}

	// 解析每个字节的位 (从最高位到最低位)
	for i, b := range keyUsageBytes {
		for j := 7; j >= 0; j-- { // 从bit 7到bit 0
			// 检查位是否设置
			if b&(1<<uint(j)) != 0 {
				bitPos := i*8 + (7 - j) // 转换为从0开始的位置
				if usage, exists := usageMap[bitPos]; exists {
					usageStrings = append(usageStrings, usage)
				} else {
					usageStrings = append(usageStrings, fmt.Sprintf("Unknown Usage (Bit %d)", bitPos))
				}
			}
		}
	}

	if len(usageStrings) > 0 {
		result := ""
		for _, usage := range usageStrings {
			result += "  - " + usage + "\n"
		}
		return result
	}

	return ""
}

// 解析扩展密钥用法扩展项
func parseExtendedKeyUsage(data []byte) string {
	// 简化的解析
	hexValue := hex.EncodeToString(data)
	if len(hexValue) > 1000 {
		hexValue = hexValue[:1000] + "...(已截断)"
	}
	return "Extended Key Usage (Hex): " + hexValue
}

// 展示证书扩展项
func showCertificateExtensions(orderKeys []string, certExtensions map[string]string, box *fyne.Container) {
	// 添加一个分隔线和标题
	box.Add(widget.NewSeparator())
	extensionTitle := widget.NewLabel("Certificate Extensions:")
	extensionTitle.TextStyle = fyne.TextStyle{Bold: true}
	box.Add(extensionTitle)

	for _, orderKey := range orderKeys {
		key := widget.NewLabel(orderKey)
		data := certExtensions[orderKey]
		value := widget.NewEntry()
		if len(data) > 100 {
			value = widget.NewMultiLineEntry()
			// 为多行输入框设置最小高度
			value.Resize(fyne.NewSize(400, 100))
		}
		value.Wrapping = fyne.TextWrapWord
		value.SetText(data)
		//防止值被修改
		value.OnChanged = func(s string) {
			text := certExtensions[key.Text]
			value.SetText(text)
		}
		realKey := container.New(layout.NewGridWrapLayout(fyne.Size{150, 30}), key)
		realValue := container.NewStack(value)
		line := container.New(layout.NewFormLayout(), realKey, realValue)
		box.Add(line)
	}
	box.Refresh()
}

func ParsePublicKeyAlg(alg PublicKeyAlgorithm) string {
	switch alg {
	case RSA:
		return "RSA"
	case SM2:
		return "SM2"
	case ECDSA:
		return "ECDSA"
	default:
		return ""
	}

}

// 将证书详情以表格的形式添加在最后
func showCertificateDetail(orderKeys []string, certDetail map[string]string, box *fyne.Container) {
	for _, orderKey := range orderKeys {
		key := widget.NewLabel(orderKey)
		data := certDetail[orderKey]
		value := widget.NewEntry()
		if len(data) > 100 {
			value = widget.NewMultiLineEntry()
			// 为多行输入框设置最小高度
			value.Resize(fyne.NewSize(400, 100))
		}
		value.Wrapping = fyne.TextWrapWord
		value.SetText(data)
		//防止值被修改
		value.OnChanged = func(s string) {
			text := certDetail[key.Text]
			value.SetText(text)
		}
		realKey := container.New(layout.NewGridWrapLayout(fyne.Size{150, 30}), key)
		realValue := container.NewStack(value)
		line := container.New(layout.NewFormLayout(), realKey, realValue)
		box.Add(line)
	}
	box.Refresh()
}

func buildInputCertEntry(data string) *widget.Entry {
	inputCert := widget.NewMultiLineEntry()
	inputCert.Wrapping = fyne.TextWrapWord
	inputCert.SetPlaceHolder(data)
	return inputCert
}

func buildButton(data string, icon fyne.Resource, fun func()) *widget.Button {
	if icon == nil {
		icon = theme.ConfirmIcon()
	}
	button := widget.NewButtonWithIcon(data, icon, fun)
	return button
}

// parsePEMCertificate 解析PEM格式证书
func parsePEMCertificate(pemData string) ([]byte, error) {
	// 清理输入数据，移除多余的空格和换行
	pemData = strings.TrimSpace(pemData)

	// 解析PEM块
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("无法解析PEM数据，请检查格式是否正确")
	}

	// 检查PEM块类型
	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("PEM块类型不正确，期望为CERTIFICATE，实际为: %s", block.Type)
	}

	return block.Bytes, nil
}

// cleanInputData 清理输入数据，移除可能影响解析的字符
func cleanInputData(input string) string {
	// 移除所有空格、换行符、制表符等空白字符
	cleaned := strings.ReplaceAll(input, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "\n", "")
	cleaned = strings.ReplaceAll(cleaned, "\r", "")
	cleaned = strings.ReplaceAll(cleaned, "\t", "")
	return strings.TrimSpace(cleaned)
}
