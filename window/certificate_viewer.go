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
	keys = []string{"SerialNumber", "SubjectName", "IssueName", "NotBefore", "NotAfter", "PublicKey", "PublicKeyAlgorithm", "SignatureAlgorithm", "KeyUsage"}
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
	certDetail[keys[8]] = helper.ParseKeyUsage(certificate.KeyUsage)

	return keys, certDetail
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
