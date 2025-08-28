package window

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/zaneway/cain-go/x509"
)

// 输入证书、私钥、密码，生成pfx文件
func CrlStructure(input *widget.Entry) *fyne.Container {
	input = widget.NewMultiLineEntry()
	input.SetPlaceHolder("Please input base64/hex crl")
	input.Refresh()
	structure := container.NewVBox()
	input.Wrapping = fyne.TextWrapWord
	certSNInput := buildInputCertEntry("Please input certSN")
	certSNInput.Wrapping = fyne.TextWrapWord

	// 创建输出框，供用户输入数据
	output := widget.NewMultiLineEntry()
	output.Hide()

	//确认按钮
	confirm := buildButton("确认", theme.ConfirmIcon(), func() {
		inputCrl := input.Text
		inputCertSN := certSNInput.Text
		decodeCrl, err := base64.StdEncoding.DecodeString(inputCrl)
		if err != nil {
			decodeCrl, err = hex.DecodeString(inputCrl)
			if err != nil {
				fyne.LogError("解析CRL请求错误", err)
				return
			}
		}
		crl, err := x509.ParseCRL(decodeCrl)
		certificates := crl.TBSCertList.RevokedCertificates
		output.Text = "false"
		for _, revokeCert := range certificates {
			hexStr := fmt.Sprintf("%x", revokeCert.SerialNumber)
			if hexStr == inputCertSN {
				output.Text = "true"
			}
		}
		if err != nil {
			return
		}
		output.Show()
	})
	//清除按钮
	clear := buildButton("清除", theme.CancelIcon(), func() {
		input.Text = ""
		output.Text = ""
		input.Refresh()
		output.Refresh()
	})

	//对所有按钮进行表格化
	allButton := container.New(layout.NewGridLayout(2), confirm, clear)
	structure.Add(input)

	structure.Add(certSNInput)
	structure.Add(allButton)
	structure.Add(output)
	return structure
}
