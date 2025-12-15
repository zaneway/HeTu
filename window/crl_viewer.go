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
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// CRL证书撤销列表解析和验证功能
func CrlStructure(input *widget.Entry) *fyne.Container {
	// 移除占位符设置，由主界面统一管理
	structure := container.NewVBox()
	input.Wrapping = fyne.TextWrapWord
	certSNInput := buildInputCertEntry("请输入要验证的证书序列号")
	certSNInput.Wrapping = fyne.TextWrapWord

	// 创建CRL详情显示区域
	crlDetails := widget.NewMultiLineEntry()
	crlDetails.SetPlaceHolder("CRL详细信息将在这里显示")
	crlDetails.Hide()

	// 创建验证结果输出框
	output := widget.NewMultiLineEntry()
	output.SetPlaceHolder("证书验证结果将在这里显示")
	output.Hide()

	// 当前加载的CRL信息
	var currentCRLInfo *helper.CRLInfo

	// 解析CRL按钮 - 从输入框解析Base64/Hex/PEM格式的CRL数据
	parseBtn := buildButton("解析CRL", theme.ConfirmIcon(), func() {
		inputData := strings.TrimSpace(input.Text)
		if inputData == "" {
			dialog.ShowError(fmt.Errorf("请输入CRL数据"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		// 尝试解析CRL数据
		var decodeData []byte
		var err error
		var isPEMFormat bool

		// 检查是否是PEM格式
		trimmedInput := strings.TrimSpace(inputData)
		if strings.HasPrefix(trimmedInput, "-----BEGIN X509 CRL-----") ||
			strings.HasPrefix(trimmedInput, "-----BEGIN CRL-----") ||
			strings.Contains(trimmedInput, "-----BEGIN") {
			isPEMFormat = true
			// 尝试处理PEM格式CRL
			decodeData, err = parsePEMCRL(inputData)
			if err != nil {
				// PEM解析失败，回退到Base64/Hex解码
				isPEMFormat = false
			}
		}

		// 如果不是PEM格式，或者PEM解析失败，尝试Base64/Hex解码
		if !isPEMFormat {
			// 清理输入数据，移除空格和换行符
			cleanedInput := cleanInputData(inputData)

			// 尝试Base64解码
			decodeData, err = base64.StdEncoding.DecodeString(cleanedInput)
			if err != nil {
				// 尝试URL-safe Base64
				decodeData, err = base64.URLEncoding.DecodeString(cleanedInput)
				if err != nil {
					// 尝试添加填充后解码
					cleanedWithPadding := addBase64Padding(cleanedInput)
					decodeData, err = base64.StdEncoding.DecodeString(cleanedWithPadding)
					if err != nil {
						// 如果Base64失败，尝试Hex解码
						decodeData, err = hex.DecodeString(cleanedInput)
						if err != nil {
							dialog.ShowError(fmt.Errorf("无法解码输入数据，请确保输入的是有效的Base64、Hex或PEM格式CRL数据\n\nBase64错误: %v\nHex错误: %v", err, err), fyne.CurrentApp().Driver().AllWindows()[0])
							return
						}
					}
				}
			}
		}

		// 验证解码后的数据长度
		if len(decodeData) < 50 {
			dialog.ShowError(fmt.Errorf("解码后的数据太短（%d 字节），不像是有效的CRL数据", len(decodeData)), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		// 解析CRL
		crlInfo, err := helper.ParseCRL(decodeData)
		if err != nil {
			dialog.ShowError(fmt.Errorf("解析CRL失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		currentCRLInfo = crlInfo
		// 显示CRL详情
		displayCRLDetails(crlDetails, crlInfo)
		crlDetails.Show()
	})

	// 文件选择按钮
	selectFileBtn := buildButton("选择CRL文件", theme.FolderOpenIcon(), func() {
		// 创建文件选择对话框
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(fmt.Errorf("打开文件失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			// 获取文件路径并显示
			filePath := reader.URI().Path()
			input.SetText(fmt.Sprintf("已选择文件: %s", filePath))

			// 读取文件内容
			data := make([]byte, 0)
			buffer := make([]byte, 1024)
			for {
				n, err := reader.Read(buffer)
				if n > 0 {
					data = append(data, buffer[:n]...)
				}
				if err != nil {
					break
				}
			}

			// 解析CRL
			crlInfo, err := helper.ParseCRL(data)
			if err != nil {
				dialog.ShowError(fmt.Errorf("解析CRL文件失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
				return
			}

			currentCRLInfo = crlInfo
			// 显示CRL详情
			displayCRLDetails(crlDetails, crlInfo)
			crlDetails.Show()
		}, fyne.CurrentApp().Driver().AllWindows()[0])

		// 设置文件过滤器
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".crl", ".der", ".pem", ".cer", ".crt"}))
		fileDialog.Show()
	})

	//验证证书按钮
	verifyBtn := buildButton("验证证书", theme.ConfirmIcon(), func() {
		if currentCRLInfo == nil {
			dialog.ShowInformation("提示", "请先解析CRL或选择CRL文件", fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		inputCertSN := strings.TrimSpace(certSNInput.Text)
		if inputCertSN == "" {
			dialog.ShowInformation("提示", "请输入要验证的证书序列号", fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		isRevoked, revokedCert := helper.CheckCertificateRevocation(currentCRLInfo, inputCertSN)
		displayVerificationResult(output, inputCertSN, isRevoked, revokedCert)
		output.Show()
	})

	//清除按钮
	clear := buildButton("清除", theme.CancelIcon(), func() {
		input.SetText("")
		certSNInput.SetText("")
		crlDetails.SetText("")
		output.SetText("")
		crlDetails.Hide()
		output.Hide()
		currentCRLInfo = nil
		input.Refresh()
		certSNInput.Refresh()
	})

	// 按钮布局 - 添加解析按钮和文件选择按钮
	buttonRow1 := container.New(layout.NewGridLayout(2), parseBtn, selectFileBtn)
	buttonRow2 := container.New(layout.NewGridLayout(2), verifyBtn, clear)

	// 组装界面 - 不添加全局输入框，它已经在主界面的固定位置
	// structure.Add(widget.NewLabel("CRL数据输入:"))
	// structure.Add(input)
	structure.Add(buttonRow1)
	structure.Add(widget.NewSeparator())
	structure.Add(widget.NewLabel("证书序列号:"))
	structure.Add(certSNInput)
	structure.Add(buttonRow2)
	structure.Add(widget.NewSeparator())
	structure.Add(widget.NewLabel("CRL详情:"))
	structure.Add(crlDetails)
	structure.Add(widget.NewLabel("验证结果:"))
	structure.Add(output)

	// 使用滚动容器支持长内容
	scrollContainer := container.NewScroll(structure)
	return container.NewMax(scrollContainer)
}

// displayCRLDetails 显示CRL详细信息
func displayCRLDetails(detailsWidget *widget.Entry, crlInfo *helper.CRLInfo) {
	details := fmt.Sprintf(`CRL详细信息:
`+
		`颁发者: %s
`+
		`本次更新时间: %s
`+
		`下次更新时间: %s
`+
		`签名算法: %s
`+
		`被吊销证书总数: %d
`+
		`\n被吊销证书列表:\n`,
		crlInfo.Issuer,
		crlInfo.ThisUpdate.Format("2006-01-02 15:04:05"),
		crlInfo.NextUpdate.Format("2006-01-02 15:04:05"),
		crlInfo.SignatureAlgorithm,
		crlInfo.TotalRevoked)

	for i, cert := range crlInfo.RevokedCerts {
		if i >= 20 { // 限制显示前20个，避免界面过长
			details += fmt.Sprintf("... 还有 %d 个被吊销的证书\n", len(crlInfo.RevokedCerts)-20)
			break
		}
		details += fmt.Sprintf("%d. 序列号: %s, 吊销时间: %s, 原因: %s\n",
			i+1, cert.SerialNumber,
			cert.RevocationTime.Format("2006-01-02 15:04:05"),
			cert.Reason)
	}

	detailsWidget.SetText(details)
}

// displayVerificationResult 显示验证结果
func displayVerificationResult(outputWidget *widget.Entry, serialNumber string, isRevoked bool, revokedCert *helper.RevokedCertificate) {
	var result string
	if isRevoked {
		result = fmt.Sprintf(`🔴 证书已被吊销
`+
			`查询序列号: %s
`+
			`吊销时间: %s
`+
			`吊销原因: %s
`+
			`
⚠️  警告: 该证书不应被信任！`,
			serialNumber,
			revokedCert.RevocationTime.Format("2006-01-02 15:04:05"),
			revokedCert.Reason)
	} else {
		result = fmt.Sprintf(`🟢 证书未被吊销
`+
			`查询序列号: %s
`+
			`
✅ 该证书在当前CRL中未被列为已吊销状态`,
			serialNumber)
	}

	outputWidget.SetText(result)
}

// parsePEMCRL 解析PEM格式CRL
func parsePEMCRL(pemData string) ([]byte, error) {
	// 清理输入数据，移除多余的空格和换行
	pemData = strings.TrimSpace(pemData)

	// 解析PEM块
	rest := []byte(pemData)
	for len(rest) > 0 {
		block, remaining := pem.Decode(rest)
		if block == nil {
			break
		}

		// 检查PEM块类型，支持CRL格式
		if block.Type == "X509 CRL" || block.Type == "CRL" {
			return block.Bytes, nil
		}

		rest = remaining
	}

	return nil, fmt.Errorf("未找到有效的CRL块")
}
