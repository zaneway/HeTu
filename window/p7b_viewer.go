package window

import (
	"HeTu/util"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	. "github.com/zaneway/cain-go/x509"
)

// P7bStructure 构造解析P7B证书链核心图形模块
func P7bStructure(input *widget.Entry) *fyne.Container {
	structure := container.NewVBox()
	detail := container.NewVBox()

	// 创建状态标签和进度条
	statusLabel := widget.NewLabel("准备解析P7B证书链...")
	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	//确认按钮
	confirm := widget.NewButtonWithIcon("确认", theme.ConfirmIcon(), func() {
		inputData := strings.TrimSpace(input.Text)
		if inputData == "" {
			dialog.ShowError(fmt.Errorf("请输入P7B证书链数据"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		// 保存到历史记录
		if inputData != "" {
			util.GetHistoryDB().AddHistory("🔗 P7B证书链", inputData)

			// 刷新历史记录下拉框
			if historyManager := GetGlobalHistoryManager(); historyManager != nil {
				historyManager.LoadHistoryForTab("🔗 P7B证书链")
			}
		}

		// 清除旧内容并显示进度
		detail.RemoveAll()
		statusLabel.SetText("正在解析P7B证书链...")
		progressBar.Show()
		progressBar.SetValue(0.1)
		detail.Add(statusLabel)
		detail.Add(progressBar)
		detail.Refresh()

		// 在后台 goroutine 中执行解析操作
		go func() {
			fyne.Do(func() {
				statusLabel.SetText("正在解码数据...")
				progressBar.SetValue(0.3)
			})

			// 尝试Base64解码
			var decodeData []byte
			var err error

			// 清理输入数据，移除空格和换行符
			cleanedInput := strings.ReplaceAll(inputData, " ", "")
			cleanedInput = strings.ReplaceAll(cleanedInput, "\n", "")
			cleanedInput = strings.ReplaceAll(cleanedInput, "\r", "")
			cleanedInput = strings.ReplaceAll(cleanedInput, "\t", "")
			cleanedInput = strings.TrimSpace(cleanedInput)

			// 尝试Base64解码
			decodeData, err = base64.StdEncoding.DecodeString(cleanedInput)
			if err != nil {
				// 如果Base64失败，尝试Hex解码
				decodeData, err = hex.DecodeString(cleanedInput)
				if err != nil {
					fyne.Do(func() {
						progressBar.Hide()
						dialog.ShowError(fmt.Errorf("无法解码输入数据，请确保输入的是有效的Base64或Hex格式P7B数据\n\n输入数据长度: %d\n清理后数据长度: %d\n\nBase64错误: %v\nHex错误: %v", len(inputData), len(cleanedInput), err, err), fyne.CurrentApp().Driver().AllWindows()[0])
						statusLabel.SetText("数据解码失败")
					})
					return
				}
			}

			// 验证解码后的数据长度
			if len(decodeData) < 50 { // P7B通常至少有几百字节
				fyne.Do(func() {
					progressBar.Hide()
					dialog.ShowError(fmt.Errorf("解码后的数据太短（%d 字节），不像是有效的P7B数据", len(decodeData)), fyne.CurrentApp().Driver().AllWindows()[0])
					statusLabel.SetText("数据长度不足")
				})
				return
			}

			fyne.Do(func() {
				statusLabel.SetText("正在解析P7B结构...")
				progressBar.SetValue(0.6)
			})

			// 解析P7B证书链
			p7b, err := ParsePKCS7(decodeData)
			if err != nil {
				fyne.Do(func() {
					progressBar.Hide()
					dialog.ShowError(fmt.Errorf("P7B证书链解析失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
					statusLabel.SetText("P7B解析失败")
				})
				return
			}

			fyne.Do(func() {
				statusLabel.SetText("正在构建证书链信息...")
				progressBar.SetValue(0.8)
			})

			// 更新UI显示结果
			fyne.Do(func() {
				statusLabel.SetText("正在显示结果...")
				progressBar.SetValue(0.9)

				// 显示P7B信息
				detail.RemoveAll()
				showP7bInfo(p7b, detail)

				progressBar.Hide()
				detail.Refresh()
			})
		}()
	})

	//清除按钮
	clear := widget.NewButtonWithIcon("清除", theme.CancelIcon(), func() {
		input.Text = ""
		input.Refresh()
	})

	//对所有按钮进行表格化
	allButton := container.New(layout.NewGridLayout(2), confirm, clear)
	structure.Add(allButton)
	structure.Add(detail)

	// 使用带滚动条的容器包装
	scrollContainer := container.NewScroll(structure)
	scrollContainer.SetMinSize(fyne.NewSize(600, 400))
	return container.NewMax(scrollContainer)
}

// showP7bInfo 显示P7B证书链信息
func showP7bInfo(p7b *PKCS7, box *fyne.Container) {
	// 添加P7B基本信息
	infoTitle := widget.NewLabel("P7B证书链信息:")
	infoTitle.TextStyle = fyne.TextStyle{Bold: true}
	box.Add(infoTitle)

	// 显示证书数量
	certCount := widget.NewLabel(fmt.Sprintf("证书数量: %d", len(p7b.Certificates)))
	box.Add(certCount)

	// 显示CRL数量
	crlCount := widget.NewLabel(fmt.Sprintf("CRL数量: %d", len(p7b.CRLs)))
	box.Add(crlCount)

	//// 显示签名者数量
	//signerCount := widget.NewLabel(fmt.Sprintf("签名者数量: %d", len(p7b.Signers)))
	//box.Add(signerCount)

	// 显示内容长度（如果有）
	if len(p7b.Content) > 0 {
		contentInfo := widget.NewLabel(fmt.Sprintf("内容长度: %d 字节", len(p7b.Content)))
		box.Add(contentInfo)
	}

	// 显示每个证书的详细信息
	for i, certificate := range p7b.Certificates {
		// 添加分隔线和证书标题
		box.Add(widget.NewSeparator())
		certTitle := widget.NewLabel(fmt.Sprintf("证书 #%d", i+1))
		certTitle.TextStyle = fyne.TextStyle{Bold: true}
		box.Add(certTitle)

		// 构造证书解析详情
		keys, value := buildCertificateDetail(certificate)

		// 展示证书详情
		showCertificateDetail(keys, value, box)

		// 解析并展示证书扩展项
		if len(certificate.Extensions) > 0 {
			extensionKeys, extensionValues := buildCertificateExtensions(certificate)
			showCertificateExtensions(extensionKeys, extensionValues, box)
		}
	}

	// 验证证书链
	showCertificateChainValidation(p7b.Certificates, box)

	box.Refresh()
}

// showCertificateChainValidation 展示证书链验证结果
func showCertificateChainValidation(certificates []*Certificate, box *fyne.Container) {
	validationTitle := widget.NewLabel("证书链验证:")
	validationTitle.TextStyle = fyne.TextStyle{Bold: true}
	box.Add(validationTitle)

	validationResult := validateCertificateChain(certificates)

	validationEntry := widget.NewMultiLineEntry()
	validationEntry.SetText(validationResult)
	validationEntry.Wrapping = fyne.TextWrapWord
	validationEntry.Resize(fyne.NewSize(400, 200))

	box.Add(validationEntry)
	box.Refresh()
}

// validateCertificateChain 验证证书链的有效性
func validateCertificateChain(certificates []*Certificate) string {
	if len(certificates) == 0 {
		return "证书链为空"
	}

	if len(certificates) == 1 {
		return "单个证书，无需验证证书链"
	}

	result := fmt.Sprintf("证书链验证结果（共 %d 个证书）:\n\n", len(certificates))

	// 从根证书开始验证（假设证书按从叶到根的顺序排列）
	for i := 0; i < len(certificates)-1; i++ {
		childCert := certificates[i]
		parentCert := certificates[i+1]

		result += fmt.Sprintf("验证证书 #%d -> #%d:\n", i+1, i+2)

		// 验证签名
		err := childCert.CheckSignatureFrom(parentCert)
		if err != nil {
			result += fmt.Sprintf("  ❌ 签名验证失败: %v\n", err)
		} else {
			result += fmt.Sprintf("  ✅ 签名验证通过\n")
		}

		// 检查有效期
		now := time.Now()
		if now.Before(childCert.NotBefore) || now.After(childCert.NotAfter) {
			result += fmt.Sprintf("  ❌ 证书不在有效期内\n")
		} else {
			result += fmt.Sprintf("  ✅ 证书在有效期内\n")
		}

		result += "\n"
	}

	return result
}
