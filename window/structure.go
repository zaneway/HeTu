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
			fyne.LogError("解析请求错误", err)
			return
		}
		//MIICETCCAbWgAwIBAgINKl81oFaaablKOp0YTjAMBggqgRzPVQGDdQUAMGExCzAJBgNVBAYMAkNOMQ0wCwYDVQQKDARCSkNBMSUwIwYDVQQLDBxCSkNBIEFueXdyaXRlIFRydXN0IFNlcnZpY2VzMRwwGgYDVQQDDBNUcnVzdC1TaWduIFNNMiBDQS0xMB4XDTIwMDgxMzIwMTkzNFoXDTIwMTAyNDE1NTk1OVowHjELMAkGA1UEBgwCQ04xDzANBgNVBAMMBuWGr+i9rDBZMBMGByqGSM49AgEGCCqBHM9VAYItA0IABAIF97Sqq0Rv616L2PjFP3xt16QGJLmi+W8Ht+NLHiXntgUey0Nz+ZVnSUKUMzkKuGTikY3h2v7la20b6lpKo8WjgZIwgY8wCwYDVR0PBAQDAgbAMB0GA1UdDgQWBBSxiaS6z4Uguz3MepS2zblkuAF/LTAfBgNVHSMEGDAWgBTMZyRCGsP4rSes0vLlhIEf6cUvrjBABgNVHSAEOTA3MDUGCSqBHIbvMgICAjAoMCYGCCsGAQUFBwIBFhpodHRwOi8vd3d3LmJqY2Eub3JnLmNuL2NwczAMBggqgRzPVQGDdQUAA0gAMEUCIG6n6PG0BOK1EdFcvetQlC+9QhpsTuTui2wkeqWiPKYWAiEAvqR8Z+tSiYR5DIs7SyHJPWZ+sa8brtQL/1jURvHGxU8=
		certificate, err := cert.ParseCertificate(decodeCert)
		if err != nil {
			fyne.LogError("解析证书错误", err)
			return
		}
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
