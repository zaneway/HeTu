package window

import (
	"HeTu/helper"
	"HeTu/util"
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
	// 为公共输入框设置适当的高度
	//input := buildInputCertEntry("Please input base64/hex cert")

	//inputCertEntry.Text = "MIICETCCAbWgAwIBAgINKl81oFaaablKOp0YTjAMBggqgRzPVQGDdQUAMGExCzAJBgNVBAYMAkNOMQ0wCwYDVQQKDARCSkNBMSUwIwYDVQQLDBxCSkNBIEFueXdyaXRlIFRydXN0IFNlcnZpY2VzMRwwGgYDVQQDDBNUcnVzdC1TaWduIFNNMiBDQS0xMB4XDTIwMDgxMzIwMTkzNFoXDTIwMTAyNDE1NTk1OVowHjELMAkGA1UEBgwCQ04xDzANBgNVBAMMBuWGr+i9rDBZMBMGByqGSM49AgEGCCqBHM9VAYItA0IABAIF97Sqq0Rv616L2PjFP3xt16QGJLmi+W8Ht+NLHiXntgUey0Nz+ZVnSUKUMzkKuGTikY3h2v7la20b6lpKo8WjgZIwgY8wCwYDVR0PBAQDAgbAMB0GA1UdDgQWBBSxiaS6z4Uguz3MepS2zblkuAF/LTAfBgNVHSMEGDAWgBTMZyRCGsP4rSes0vLlhIEf6cUvrjBABgNVHSAEOTA3MDUGCSqBHIbvMgICAjAoMCYGCCsGAQUFBwIBFhpodHRwOi8vd3d3LmJqY2Eub3JnLmNuL2NwczAMBggqgRzPVQGDdQUAA0gAMEUCIG6n6PG0BOK1EdFcvetQlC+9QhpsTuTui2wkeqWiPKYWAiEAvqR8Z+tSiYR5DIs7SyHJPWZ+sa8brtQL/1jURvHGxU8="
	// 创建状态标签和进度条
	statusLabel := widget.NewLabel("准备解析证书...")
	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	//确认按钮
	confirm := widget.NewButtonWithIcon("确认", theme.ConfirmIcon(), func() {
		inputCert := strings.TrimSpace(input.Text)
		if inputCert == "" {
			dialog.ShowError(fmt.Errorf("请输入证书数据"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		// 保存到历史记录
		if inputCert != "" {
			util.GetHistoryDB().AddHistory("🏆 证书解析", inputCert)

			// 刷新历史记录下拉框
			if historyManager := GetGlobalHistoryManager(); historyManager != nil {
				historyManager.LoadHistoryForTab("🏆 证书解析")
			}
		}

		// 清除旧内容并显示进度
		detail.RemoveAll()
		statusLabel.SetText("正在解析证书...")
		progressBar.Show()
		progressBar.SetValue(0.1)
		detail.Add(statusLabel)
		detail.Add(progressBar)
		detail.Refresh()

		// 在后台 goroutine 中执行解析操作
		go func() {
			fyne.Do(func() {
				statusLabel.SetText("正在解码证书数据...")
				progressBar.SetValue(0.3)
			})

			// 尝试Base64解码
			var decodeCert []byte
			var err error
			var isPEMFormat bool

			// 检查是否是PEM格式（更严格的检查）
			trimmedInput := strings.TrimSpace(inputCert)
			if strings.HasPrefix(trimmedInput, "-----BEGIN CERTIFICATE-----") ||
				strings.HasPrefix(trimmedInput, "-----BEGIN X509 CERTIFICATE-----") ||
				strings.Contains(trimmedInput, "-----BEGIN CERTIFICATE-----") {
				isPEMFormat = true
				// 尝试处理PEM格式证书
				decodeCert, err = parsePEMCertificate(inputCert)
				if err != nil {
					// PEM解析失败，回退到Base64/Hex解码
					fyne.Do(func() {
						statusLabel.SetText("PEM解析失败，尝试Base64/Hex解码...")
					})
					isPEMFormat = false
				}
			}

			// 如果不是PEM格式，或者PEM解析失败，尝试Base64/Hex解码
			if !isPEMFormat {
				// 清理输入数据，移除空格和换行符
				cleanedInput := cleanInputData(inputCert)

				// 尝试Base64解码（支持多种Base64格式）
				decodeCert, err = base64.StdEncoding.DecodeString(cleanedInput)
				if err != nil {
					// 尝试URL-safe Base64
					decodeCert, err = base64.URLEncoding.DecodeString(cleanedInput)
					if err != nil {
						// 尝试添加填充后解码
						cleanedWithPadding := addBase64Padding(cleanedInput)
						decodeCert, err = base64.StdEncoding.DecodeString(cleanedWithPadding)
						if err != nil {
							// 如果Base64失败，尝试Hex解码
							decodeCert, err = hex.DecodeString(cleanedInput)
							if err != nil {
								fyne.Do(func() {
									progressBar.Hide()
									dialog.ShowError(fmt.Errorf("无法解码输入数据，请确保输入的是有效的Base64、Hex或PEM格式证书数据\n\n输入数据长度: %d\n清理后数据长度: %d\n\nBase64错误: %v\nHex错误: %v\n\n提示：如果数据是Base64编码，请确保数据完整且格式正确", len(inputCert), len(cleanedInput), err, err), fyne.CurrentApp().Driver().AllWindows()[0])
									statusLabel.SetText("数据解码失败")
								})
								return
							}
						}
					}
				}
			}

			// 验证解码后的数据长度
			if len(decodeCert) < 50 { // 证书通常至少有几百字节
				fyne.Do(func() {
					progressBar.Hide()
					dialog.ShowError(fmt.Errorf("解码后的数据太短（%d 字节），不像是有效的证书数据", len(decodeCert)), fyne.CurrentApp().Driver().AllWindows()[0])
					statusLabel.SetText("数据长度不足")
				})
				return
			}

			fyne.Do(func() {
				statusLabel.SetText("正在解析证书结构...")
				progressBar.SetValue(0.6)
			})

			// 解析证书
			certificate, err := helper.ParseCertificate(decodeCert)
			if err != nil {
				fyne.Do(func() {
					progressBar.Hide()
					dialog.ShowError(fmt.Errorf("证书解析失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
					statusLabel.SetText("证书解析失败")
				})
				return
			}

			fyne.Do(func() {
				statusLabel.SetText("正在构建证书详情...")
				progressBar.SetValue(0.8)
			})

			//构造证书解析详情
			keys, value := buildCertificateDetail(certificate)

			// 更新UI显示结果
			fyne.Do(func() {
				statusLabel.SetText("正在显示结果...")
				progressBar.SetValue(0.9)

				//展示证书详情
				detail.RemoveAll()
				showCertificateDetail(keys, value, detail)

				// 解析并展示证书扩展项
				if len(certificate.Extensions) > 0 {
					extensionKeys, extensionValues := buildCertificateExtensions(certificate)
					showCertificateExtensions(extensionKeys, extensionValues, detail)
				}

				progressBar.Hide()
				detail.Refresh()
			})
		}()
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
		case "2.5.29.31": // CRL Distribution Points
			value = parseCRLDistributionPoints(ext.Value)
		case "2.5.29.37": // Extended Key Usage
			value = parseExtendedKeyUsage(ext.Value)
		case "1.3.6.1.5.5.7.1.1": // Authority Information Access
			value = parseAuthorityInformationAccess(ext.Value)
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
	var result string
	// 尝试解析证书策略中的CPS URL
	//result := fmt.Sprintf("Certificate Policies (Length: %d bytes)\n", len(data))
	//
	//// 显示原始十六进制值
	//hexValue := hex.EncodeToString(data)
	//result += fmt.Sprintf("Raw Hex: %s\n\n", formatHexDisplay(hexValue))

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
	return hex.EncodeToString(data)
}

// 解析颁发机构密钥标识符扩展项
func parseAuthorityKeyIdentifier(data []byte) string {
	// 简化的解析，实际结构可能更复杂
	return hex.EncodeToString(data)
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
	//result := fmt.Sprintf("Key Usage (Length: %d bytes)\n", len(data))
	//
	//// 显示原始十六进制值
	//hexValue := hex.EncodeToString(data)
	//result += fmt.Sprintf("Raw Hex: %s\n\n", formatHexDisplay(hexValue))
	var result string
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

// 解析CRL分发点扩展项
func parseCRLDistributionPoints(data []byte) string {
	//result := fmt.Sprintf("CRL Distribution Points (Length: %d bytes)\n", len(data))
	//
	//// 显示原始十六进制值
	//hexValue := hex.EncodeToString(data)
	//result += fmt.Sprintf("Raw Hex: %s\n\n", formatHexDisplay(hexValue))
	var result string
	// 尝试解析ASN.1结构以提取CRL URL
	crlUrls := extractCRLUrls(data)
	if len(crlUrls) > 0 {
		result += "CRL Distribution URLs:\n"
		for i, url := range crlUrls {
			result += fmt.Sprintf("  %d. %s\n", i+1, url)
		}
		result += "\n"
	} else {
		result += "未找到CRL URLs或解析失败\n\n"
	}

	// 添加解析说明
	result += "解析说明:\n"
	result += "CRL分发点扩展包含以下信息:\n"
	result += "1. CRL分发点URL - 用于下载证书吊销列表\n"
	result += "2. 可能包含多个分发点以提供冗余\n"

	return result
}

// 从CRL分发点数据中提取URLs
func extractCRLUrls(data []byte) []string {
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

// 解析扩展密钥用法扩展项
func parseExtendedKeyUsage(data []byte) string {
	// 简化的解析
	hexValue := hex.EncodeToString(data)
	if len(hexValue) > 1000 {
		hexValue = hexValue[:1000] + "...(已截断)"
	}
	return "Extended Key Usage (Hex): " + hexValue
}

// 解析Authority Information Access扩展项
func parseAuthorityInformationAccess(data []byte) string {
	result := fmt.Sprintf("Authority Information Access (Length: %d bytes)\n", len(data))

	// 显示原始十六进制值
	hexValue := hex.EncodeToString(data)
	result += fmt.Sprintf("Raw Hex: %s\n\n", formatHexDisplay(hexValue))

	// 使用重新设计的解析器
	parser := NewAuthorityInfoAccessParser(data)
	accessInfos := parser.Parse()

	if len(accessInfos) > 0 {
		result += "Authority Access Information:\n"
		for i, info := range accessInfos {
			result += fmt.Sprintf("  %d. Method: %s\n", i+1, info.Method)
			result += fmt.Sprintf("     Location: %s\n", info.Location)
		}
		result += "\n"
	} else {
		result += "未找到访问信息或解析失败\n\n"
	}

	// 添加解析说明
	result += "解析说明:\n"
	result += "Authority Information Access扩展包含以下信息:\n"
	result += "1. 访问方法 - 如OCSP、CA Issuers等\n"
	result += "2. 访问位置 - 对应的URL地址\n"
	result += "3. 可能包含多个访问信息条目\n\n"

	// 添加常见访问方法OID的说明
	result += "常见访问方法OID:\n"
	result += "- 1.3.6.1.5.5.7.48.1: OCSP (在线证书状态协议)\n"
	result += "- 1.3.6.1.5.5.7.48.2: CA Issuers (CA证书颁发者)\n"

	return result
}

// ParseAuthorityInformationAccessForTest 是parseAuthorityInformationAccess的导出版本，用于测试
func ParseAuthorityInformationAccessForTest(data []byte) string {
	return parseAuthorityInformationAccess(data)
}

// NewAuthorityInfoAccessParserForTest 是NewAuthorityInfoAccessParser的导出版本，用于测试
func NewAuthorityInfoAccessParserForTest(data []byte) *AuthorityInfoAccessParser {
	return NewAuthorityInfoAccessParser(data)
}

// AuthorityInfoAccessParser 用于解析Authority Information Access扩展
type AuthorityInfoAccessParser struct {
	data    []byte
	hexData string
}

// AuthorityAccessInfo 存储访问信息
type AuthorityAccessInfo struct {
	Method   string
	Location string
}

// NewAuthorityInfoAccessParser 创建新的解析器实例
func NewAuthorityInfoAccessParser(data []byte) *AuthorityInfoAccessParser {
	return &AuthorityInfoAccessParser{
		data:    data,
		hexData: hex.EncodeToString(data),
	}
}

// Parse 解析Authority Information Access扩展
func (p *AuthorityInfoAccessParser) Parse() []AuthorityAccessInfo {
	var accessInfos []AuthorityAccessInfo

	// 定义已知的访问方法OID
	accessMethods := map[string]string{
		"2b06010505073001": "OCSP",
		"2b06010505073002": "CA Issuers",
		"2b06010505073003": "Time Stamping",
		"2b06010505073004": "CA Repository",
	}

	// 首先尝试使用ASN.1解析
	accessInfos = p.parseWithASN1()

	// 添加标志变量，记录ASN.1是否解析成功
	asn1Parsed := len(accessInfos) > 0

	// 只有当ASN.1解析失败或没有找到信息时，才使用回退方法
	// 这样可以避免解析结果被覆盖的问题
	if !asn1Parsed {
		accessInfos = p.fallbackParsing(accessMethods)
	}

	// 去重
	return p.deduplicate(accessInfos)
}

// parseWithASN1 使用ASN.1结构解析Authority Information Access扩展
func (p *AuthorityInfoAccessParser) parseWithASN1() []AuthorityAccessInfo {
	var accessInfos []AuthorityAccessInfo

	// 确保数据不为空
	if len(p.data) == 0 {
		return accessInfos
	}

	// 查找SEQUENCE标记 (0x30) - AuthorityInfoAccessSyntax是SEQUENCE OF AccessDescription
	if len(p.data) < 2 || p.data[0] != 0x30 {
		return accessInfos
	}

	// 解析外层SEQUENCE长度
	seqLen, lenBytes := p.parseLength(1)
	// 改进长度检查逻辑：如果声明的长度超过实际数据长度，则使用实际数据长度
	actualDataLen := len(p.data) - 1 - lenBytes
	if seqLen <= 0 {
		return accessInfos
	} else if seqLen > actualDataLen {
		// 如果声明的长度大于实际可用数据，使用实际数据长度
		seqLen = actualDataLen
	}

	// 解析SEQUENCE中的AccessDescription条目
	pos := 1 + lenBytes
	endPos := 1 + lenBytes + seqLen

	// 添加安全检查，防止无限循环
	maxIterations := 100
	iterations := 0

	for pos < endPos && pos < len(p.data) && iterations < maxIterations {
		// 解析单个AccessDescription
		accessInfo := p.parseSingleAccessDescription(pos, endPos)
		if accessInfo.Method != "" && accessInfo.Location != "" {
			accessInfos = append(accessInfos, accessInfo)
		}

		// 移动到下一个AccessDescription
		// 需要解析当前AccessDescription的长度
		_, descTotalLen := p.parseAccessDescriptionLength(pos)
		if descTotalLen <= 0 {
			// 如果当前AccessDescription解析失败，尝试手动查找下一个
			for i := pos + 1; i < len(p.data) && i < endPos; i++ {
				if p.data[i] == 0x30 { // 找到下一个SEQUENCE标记
					pos = i
					break
				}
			}
			// 如果没找到下一个SEQUENCE标记，跳出循环
			if pos <= 1+lenBytes+seqLen {
				break
			}
		} else {
			// 确保不会超出边界
			if pos+descTotalLen > len(p.data) {
				break
			}
			pos += descTotalLen
		}
		iterations++
	}

	return accessInfos
}

// parseAccessDescriptionLength 解析单个AccessDescription的总长度
func (p *AuthorityInfoAccessParser) parseAccessDescriptionLength(startPos int) (contentLen int, totalLen int) {
	// 边界检查
	if startPos >= len(p.data) || startPos < 0 {
		return -1, -1
	}

	// AccessDescription是一个SEQUENCE
	if p.data[startPos] != 0x30 {
		return -1, -1
	}

	// 解析SEQUENCE长度
	seqLen, lenBytes := p.parseLength(startPos + 1)
	if seqLen <= 0 {
		return -1, -1
	}

	// 防止整数溢出
	if seqLen > 10000 || lenBytes > 10 {
		return -1, -1
	}

	totalLen = 1 + lenBytes + seqLen
	return seqLen, totalLen
}

// parseSingleAccessDescription 解析单个AccessDescription
func (p *AuthorityInfoAccessParser) parseSingleAccessDescription(startPos int, endPos int) AuthorityAccessInfo {
	// 确保有足够的数据
	// 修正边界检查条件，只需要确保有足够的数据来解析基本结构
	if startPos >= len(p.data) || startPos >= endPos || startPos < 0 || endPos > len(p.data) {
		return AuthorityAccessInfo{}
	}

	// AccessDescription是一个SEQUENCE
	if p.data[startPos] != 0x30 {
		return AuthorityAccessInfo{}
	}

	// 解析SEQUENCE长度
	seqLen, lenBytes := p.parseLength(startPos + 1)
	if seqLen <= 0 || startPos+1+lenBytes+seqLen > len(p.data) || startPos+1+lenBytes+seqLen > endPos {
		return AuthorityAccessInfo{}
	}

	// 解析AccessDescription内容
	contentStart := startPos + 1 + lenBytes
	contentEnd := contentStart + seqLen

	if contentStart >= contentEnd || contentEnd > len(p.data) || contentStart < 0 {
		return AuthorityAccessInfo{}
	}

	// 解析accessMethod (OID)
	oidPos := contentStart
	if oidPos >= len(p.data) || p.data[oidPos] != 0x06 { // OID标记
		return AuthorityAccessInfo{}
	}

	oidLen, oidLenBytes := p.parseLength(oidPos + 1)
	if oidLen <= 0 || oidPos+1+oidLenBytes+oidLen > len(p.data) || oidPos+1+oidLenBytes+oidLen > contentEnd {
		return AuthorityAccessInfo{}
	}

	// 确保OID长度合理
	if oidLen > 100 || oidLen <= 0 {
		return AuthorityAccessInfo{}
	}

	// 边界检查
	if oidPos+1+oidLenBytes+oidLen > len(p.data) {
		return AuthorityAccessInfo{}
	}

	oidBytes := p.data[oidPos+1+oidLenBytes : oidPos+1+oidLenBytes+oidLen]
	oidHex := hex.EncodeToString(oidBytes)

	// 将OID转换为方法名称
	method := p.oidToMethodName(oidHex)

	// 解析accessLocation (GeneralName)
	locationPos := oidPos + 1 + oidLenBytes + oidLen
	if locationPos >= len(p.data) || locationPos >= contentEnd {
		return AuthorityAccessInfo{}
	}

	// 查找URI标记 (context-specific tag 6 - 0x86)
	if p.data[locationPos] == 0x86 {
		// 解析URI长度
		uriLen, uriLenBytes := p.parseLength(locationPos + 1)
		if uriLen > 0 && uriLen < 1000 && locationPos+1+uriLenBytes+uriLen <= len(p.data) && locationPos+1+uriLenBytes+uriLen <= contentEnd {
			// 提取URI数据
			uriStart := locationPos + 1 + uriLenBytes
			uriEnd := locationPos + 1 + uriLenBytes + uriLen

			// 边界检查
			if uriStart >= len(p.data) || uriEnd > len(p.data) || uriStart >= uriEnd {
				return AuthorityAccessInfo{}
			}

			uriBytes := p.data[uriStart:uriEnd]
			uriStr := string(uriBytes)

			// 验证URI格式
			if (strings.HasPrefix(uriStr, "http://") || strings.HasPrefix(uriStr, "https://")) &&
				len(uriStr) > 10 && len(uriStr) < 500 {
				// 清理URL末尾可能的控制字符
				cleanURL := p.cleanURL(uriStr)
				// 即使清理后的URL为空，也返回原始URL
				if cleanURL == "" {
					cleanURL = uriStr
				}
				return AuthorityAccessInfo{
					Method:   method,
					Location: cleanURL,
				}
			}
		}
	}

	return AuthorityAccessInfo{}
}

// cleanURL 清理URL末尾可能的无效字符
func (p *AuthorityInfoAccessParser) cleanURL(url string) string {
	// 移除URL末尾的控制字符和无效字符
	for len(url) > 0 {
		lastChar := url[len(url)-1]
		// 如果是控制字符或非打印字符，则移除
		if lastChar < 32 || (lastChar >= 127 && lastChar <= 159) || lastChar == 0x00 {
			url = url[:len(url)-1]
		} else {
			break
		}
	}

	// 确保URL以有效的字符结尾
	for len(url) > 0 {
		lastChar := url[len(url)-1]
		if lastChar == '.' || lastChar == '/' {
			url = url[:len(url)-1]
		} else {
			break
		}
	}

	// 查找URL中第一个有效的结束位置（http://或https://之后的第一个控制字符或结构标记）
	if strings.Contains(url, "http://") {
		httpIdx := strings.Index(url, "http://")
		if httpIdx >= 0 {
			// 从http://之后开始查找结束位置
			startSearch := httpIdx + 7 // "http://".length
			for i := startSearch; i < len(url); i++ {
				char := url[i]
				// 如果遇到控制字符或特殊标记，截断URL
				if char < 32 || (char >= 127 && char <= 159) || char == 0x00 {
					url = url[:i]
					break
				}
			}
		}
	} else if strings.Contains(url, "https://") {
		httpsIdx := strings.Index(url, "https://")
		if httpsIdx >= 0 {
			// 从https://之后开始查找结束位置
			startSearch := httpsIdx + 8 // "https://".length
			for i := startSearch; i < len(url); i++ {
				char := url[i]
				// 如果遇到控制字符或特殊标记，截断URL
				if char < 32 || (char >= 127 && char <= 159) || char == 0x00 {
					url = url[:i]
					break
				}
			}
		}
	}

	return url
}

// parseLength 解析ASN.1长度字段
func (p *AuthorityInfoAccessParser) parseLength(startPos int) (length int, lenBytes int) {
	// 边界检查
	if startPos >= len(p.data) {
		return -1, 0
	}

	firstByte := p.data[startPos]
	if firstByte&0x80 == 0 { // 短格式
		return int(firstByte), 1
	} else { // 长格式
		lenBytesCount := int(firstByte & 0x7F)
		// 验证长度字节数是否合理
		if lenBytesCount > 4 || lenBytesCount <= 0 {
			return -1, 0
		}

		// 边界检查
		if startPos+1+lenBytesCount > len(p.data) {
			return -1, 0
		}

		length = 0
		for i := 0; i < lenBytesCount; i++ {
			// 防止整数溢出
			if length > 1000000 {
				return -1, 0
			}
			length = (length << 8) | int(p.data[startPos+1+i])
		}
		return length, 1 + lenBytesCount
	}
}

// oidToMethodName 将OID转换为方法名称
func (p *AuthorityInfoAccessParser) oidToMethodName(oidHex string) string {
	methodNames := map[string]string{
		"2b06010505073001": "OCSP",
		"2b06010505073002": "CA Issuers",
		"2b06010505073003": "Time Stamping",
		"2b06010505073004": "CA Repository",
	}

	if name, exists := methodNames[oidHex]; exists {
		return name
	}

	return "Unknown (" + oidHex + ")"
}

// OidToMethodNameForTest 是oidToMethodName的导出版本，用于测试
func (p *AuthorityInfoAccessParser) OidToMethodNameForTest(oidHex string) string {
	return p.oidToMethodName(oidHex)
}

// fallbackParsing 回退解析方法
func (p *AuthorityInfoAccessParser) fallbackParsing(accessMethods map[string]string) []AuthorityAccessInfo {
	var accessInfos []AuthorityAccessInfo

	// HTTP和HTTPS URL模式
	httpPattern := "687474703a2f2f"    // "http://"
	httpsPattern := "68747470733a2f2f" // "https://"

	// 查找所有OID位置
	for oid, methodName := range accessMethods {
		positions := p.findOIDPositions(oid)
		for _, pos := range positions {
			url := p.findURLForOID(pos, len(oid), httpPattern, httpsPattern)
			if url != "" {
				accessInfo := AuthorityAccessInfo{
					Method:   methodName,
					Location: url,
				}
				accessInfos = append(accessInfos, accessInfo)
			}
		}
	}

	return accessInfos
}

// findOIDPositions 查找OID在十六进制数据中的位置
func (p *AuthorityInfoAccessParser) findOIDPositions(oidHex string) []int {
	var positions []int
	start := 0
	for {
		idx := strings.Index(p.hexData[start:], oidHex)
		if idx == -1 {
			break
		}
		actualIdx := start + idx
		positions = append(positions, actualIdx)
		start = actualIdx + len(oidHex)
	}
	return positions
}

// findURLForOID 在指定OID附近查找URL
func (p *AuthorityInfoAccessParser) findURLForOID(oidPosition int, oidLength int, httpPattern, httpsPattern string) string {
	// 在OID之后查找URL
	searchStart := oidPosition + oidLength
	searchEnd := searchStart + 300 // 增加搜索范围到300字符

	if searchEnd > len(p.hexData) {
		searchEnd = len(p.hexData)
	}

	// 提取搜索范围内的数据
	searchData := p.hexData[searchStart:searchEnd]

	// 查找HTTP URL
	httpIdx := strings.Index(searchData, httpPattern)
	if httpIdx != -1 {
		// 计算URL在原始数据中的实际位置
		actualURLPos := searchStart + httpIdx
		url := p.extractURLFromHexPrecise(actualURLPos)
		if url != "" {
			return url
		}
	}

	// 查找HTTPS URL
	httpsIdx := strings.Index(searchData, httpsPattern)
	if httpsIdx != -1 {
		// 计算URL在原始数据中的实际位置
		actualURLPos := searchStart + httpsIdx
		url := p.extractURLFromHexPrecise(actualURLPos)
		if url != "" {
			return url
		}
	}

	return ""
}

// extractURLFromHexPrecise 精确地从十六进制数据中提取URL
func (p *AuthorityInfoAccessParser) extractURLFromHexPrecise(startIndex int) string {
	// 查找URL的结束位置
	endIndex := len(p.hexData)

	// 查找URL结束的明确标记
	for i := startIndex + 2; i < endIndex-4; i += 2 {
		// 检查两个字节的模式
		if i+4 <= endIndex {
			fourHex := p.hexData[i : i+4]
			// 常见的结构开始标记，表示URL可能在这里结束
			if fourHex == "0000" || // NULL标记
				fourHex == "0608" || // OID开始标记
				fourHex == "3081" || // SEQUENCE开始标记
				fourHex == "3082" || // SEQUENCE开始标记
				fourHex == "0c08" || // UTF8String开始标记
				fourHex == "1308" || // PrintableString开始标记
				fourHex == "863e" || // Context-specific tag 6
				fourHex == "863f" { // Context-specific tag 6
				endIndex = i
				break
			}
		}

		// 检查单个字节的模式
		if i+2 <= endIndex {
			twoHex := p.hexData[i : i+2]
			// 单字节结束标记
			if twoHex == "00" || // NULL标记
				twoHex == "30" || // SEQUENCE开始标记
				twoHex == "06" || // OID开始标记
				twoHex == "86" { // Context-specific tag 6
				endIndex = i
				break
			}
		}
	}

	// 确保不会越界
	if endIndex > len(p.hexData) {
		endIndex = len(p.hexData)
	}

	// 提取URL部分的十六进制数据
	urlHex := p.hexData[startIndex:endIndex]

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
			// 清理URL，移除可能的尾随字符
			// 查找URL中可能出现的控制字符或结构标记
			for i, char := range urlStr {
				// 如果遇到控制字符或特殊标记，截断URL
				if char < 32 || (char >= 127 && char <= 159) {
					urlStr = urlStr[:i]
					break
				}
			}
			return urlStr
		}
	}

	return ""
}

// deduplicate 对访问信息进行去重
func (p *AuthorityInfoAccessParser) deduplicate(accessInfos []AuthorityAccessInfo) []AuthorityAccessInfo {
	seen := make(map[string]bool)
	var uniqueInfos []AuthorityAccessInfo

	for _, info := range accessInfos {
		key := info.Method + ":" + info.Location
		if !seen[key] {
			seen[key] = true
			uniqueInfos = append(uniqueInfos, info)
		}
	}

	return uniqueInfos
}

// 将证书详情以表格的形式添加在最后
func showCertificateDetail(orderKeys []string, certDetail map[string]string, box *fyne.Container) {
	for _, orderKey := range orderKeys {
		key := widget.NewLabel(orderKey)
		data := certDetail[orderKey]
		var value *widget.Entry
		if len(data) > 100 {
			value = widget.NewMultiLineEntry()
			value.Wrapping = fyne.TextWrapWord
			// 为多行输入框设置最小高度
			value.Resize(fyne.NewSize(400, 100))
		} else {
			value = widget.NewEntry()
			// 不对单行Entry设置Wrapping属性，避免Fyne错误
		}
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
		var value *widget.Entry
		if len(data) > 100 {
			value = widget.NewMultiLineEntry()
			value.Wrapping = fyne.TextWrapWord
			// 为多行输入框设置最小高度
			value.Resize(fyne.NewSize(400, 100))
		} else {
			value = widget.NewEntry()
			// 不对单行Entry设置Wrapping属性，避免Fyne错误
		}
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

	// 解析PEM块（可能包含多个块）
	var certificateData []byte
	var hasCertificate bool

	rest := []byte(pemData)
	for len(rest) > 0 {
		block, remaining := pem.Decode(rest)
		if block == nil {
			if !hasCertificate {
				return nil, fmt.Errorf("无法解析PEM数据，请检查格式是否正确")
			}
			break
		}

		// 检查PEM块类型，支持多种证书格式
		if block.Type == "CERTIFICATE" || block.Type == "X509 CERTIFICATE" || block.Type == "TRUSTED CERTIFICATE" {
			if !hasCertificate {
				// 使用第一个证书块
				certificateData = block.Bytes
				hasCertificate = true
			}
			// 如果有多个证书块，可以在这里处理证书链
			// 目前只返回第一个证书
		}

		rest = remaining
	}

	if !hasCertificate {
		return nil, fmt.Errorf("未找到有效的CERTIFICATE块，请检查PEM格式是否正确")
	}

	return certificateData, nil
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

// addBase64Padding 为Base64字符串添加填充
func addBase64Padding(s string) string {
	// Base64编码要求长度是4的倍数，添加必要的填充
	switch len(s) % 4 {
	case 2:
		return s + "=="
	case 3:
		return s + "="
	default:
		return s
	}
}

// FallbackParsingForTest 是fallbackParsing的导出版本，用于测试
func (p *AuthorityInfoAccessParser) FallbackParsingForTest(accessMethods map[string]string) []AuthorityAccessInfo {
	return p.fallbackParsing(accessMethods)
}

// ParseWithASN1ForTest 是parseWithASN1的导出版本，用于测试
func (p *AuthorityInfoAccessParser) ParseWithASN1ForTest() []AuthorityAccessInfo {
	return p.parseWithASN1()
}

// ParseLengthForTest 是parseLength的导出版本，用于测试
func (p *AuthorityInfoAccessParser) ParseLengthForTest(startPos int) (length int, lenBytes int) {
	return p.parseLength(startPos)
}

// ParseSingleAccessDescriptionForTest 是parseSingleAccessDescription的导出版本，用于测试
func (p *AuthorityInfoAccessParser) ParseSingleAccessDescriptionForTest(startPos int, endPos int) AuthorityAccessInfo {
	return p.parseSingleAccessDescription(startPos, endPos)
}

// ParseAccessDescriptionLengthForTest 是parseAccessDescriptionLength的导出版本，用于测试
func (p *AuthorityInfoAccessParser) ParseAccessDescriptionLengthForTest(startPos int) (contentLen int, totalLen int) {
	return p.parseAccessDescriptionLength(startPos)
}
