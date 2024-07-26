package window

import (
	"CertViewer/cert"
	"encoding/base64"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// 构造解析证书核心图形模块
func Structure() *fyne.Container {
	structure := container.NewVBox()
	inputCertEntry := buildInputCertEntry("please input base64/hex cert")
	confirm := buildButton("确认", func() {
		inputCert := inputCertEntry.Text
		decodeCert, err := base64.StdEncoding.DecodeString(inputCert)
		if err != nil {
			fyne.LogError("Error decoding base64 cert", err)
		}
		certificate := cert.ParseCertificate(decodeCert)
		inputCertEntry.Text = certificate.Subject
	})
	clear := buildButton("清除", func() {
		inputCertEntry.Text = ""
		inputCertEntry.Refresh()
	})
	//对所有按钮进行表格化
	allButton := container.New(layout.NewGridLayout(2), confirm, clear)
	structure.Add(inputCertEntry)
	structure.Add(allButton)
	return structure
}

func buildInputCertEntry(data string) *widget.Entry {
	inputCert := widget.NewEntry()
	inputCert.SetPlaceHolder(data)
	return inputCert
}

func buildButton(data string, fun func()) *widget.Button {
	button := widget.NewButton(data, fun)
	return button
}
