package window

import (
	"HeTu/helper"
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
	// 使用共享的输入框，不重新创建
	input.SetPlaceHolder("请输入base64/hex格式的CRL数据，或点击'选择CRL文件'按钮")
	input.Refresh()
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

	// 按钮布局 - 只保留文件选择按钮
	buttonRow1 := container.New(layout.NewGridLayout(1), selectFileBtn)
	buttonRow2 := container.New(layout.NewGridLayout(2), verifyBtn, clear)

	// 组装界面
	structure.Add(widget.NewLabel("CRL数据输入:"))
	structure.Add(input)
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

	return structure
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
