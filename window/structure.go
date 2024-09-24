package window

import (
	"CertViewer/cert"
	"encoding/base64"
	"encoding/hex"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	. "github.com/zaneway/cain-go/x509"
)

// 构造解析证书核心图形模块
func Structure() *fyne.Container {
	structure := container.NewVBox()
	detail := container.NewVBox()
	inputCertEntry := buildInputCertEntry("please input base64/hex cert")

	inputCertEntry.Text = "MIICETCCAbWgAwIBAgINKl81oFaaablKOp0YTjAMBggqgRzPVQGDdQUAMGExCzAJBgNVBAYMAkNOMQ0wCwYDVQQKDARCSkNBMSUwIwYDVQQLDBxCSkNBIEFueXdyaXRlIFRydXN0IFNlcnZpY2VzMRwwGgYDVQQDDBNUcnVzdC1TaWduIFNNMiBDQS0xMB4XDTIwMDgxMzIwMTkzNFoXDTIwMTAyNDE1NTk1OVowHjELMAkGA1UEBgwCQ04xDzANBgNVBAMMBuWGr+i9rDBZMBMGByqGSM49AgEGCCqBHM9VAYItA0IABAIF97Sqq0Rv616L2PjFP3xt16QGJLmi+W8Ht+NLHiXntgUey0Nz+ZVnSUKUMzkKuGTikY3h2v7la20b6lpKo8WjgZIwgY8wCwYDVR0PBAQDAgbAMB0GA1UdDgQWBBSxiaS6z4Uguz3MepS2zblkuAF/LTAfBgNVHSMEGDAWgBTMZyRCGsP4rSes0vLlhIEf6cUvrjBABgNVHSAEOTA3MDUGCSqBHIbvMgICAjAoMCYGCCsGAQUFBwIBFhpodHRwOi8vd3d3LmJqY2Eub3JnLmNuL2NwczAMBggqgRzPVQGDdQUAA0gAMEUCIG6n6PG0BOK1EdFcvetQlC+9QhpsTuTui2wkeqWiPKYWAiEAvqR8Z+tSiYR5DIs7SyHJPWZ+sa8brtQL/1jURvHGxU8="
	//确认按钮
	confirm := buildButton("确认", func() {
		inputCert := inputCertEntry.Text
		detail.RemoveAll()
		decodeCert, err := base64.StdEncoding.DecodeString(inputCert)
		if err != nil {
			fyne.LogError("解析请求错误", err)
			return
		}
		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
		certificate, err := cert.ParseCertificate(decodeCert)
		if err != nil {
			fyne.LogError("解析证书错误", err)
			return
		}
		//构造证书解析详情
		keys, value := buildCertificateDetail(certificate)

		//展示证书详情
		showCertificateDetail(keys, value, detail)
	})
	//清除按钮
	clear := buildButton("清除", func() {
		inputCertEntry.Text = ""
		inputCertEntry.Refresh()
	})

	//对所有按钮进行表格化
	allButton := container.New(layout.NewGridLayout(2), confirm, clear)
	structure.Add(inputCertEntry)
	structure.Add(allButton)
	structure.Add(detail)
	return structure
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
	//PublicKey
	certDetail[keys[6]] = ParsePublicKeyAlg(certificate.PublicKeyAlgorithm)
	//SignatureAlgorithm
	certDetail[keys[7]] = certificate.SignatureAlgorithm.String()
	//KeyUsage
	certDetail[keys[8]] = cert.ParseKeyUsage(certificate.KeyUsage)

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
		value := widget.NewEntry()
		value.SetText(certDetail[orderKey])
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
	inputCert := widget.NewEntry()
	inputCert.SetPlaceHolder(data)
	return inputCert
}

func buildButton(data string, fun func()) *widget.Button {
	button := widget.NewButton(data, fun)
	return button
}
