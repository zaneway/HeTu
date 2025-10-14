package window

import (
	"HeTu/gm"
	"HeTu/util"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/zaneway/cain-go/pkcs12"
	"github.com/zaneway/cain-go/sm2"
	"github.com/zaneway/cain-go/x509"
)

func SM2EnvelopedPfxStructure(input *widget.Entry) *fyne.Container {
	// 移除占位符设置，由主界面统一管理
	input.Wrapping = fyne.TextWrapWord
	//input.Text = "MIHxMAwGCCqBHM9VAWgBBQAwegIhAMn/+ClYld5HKOj5JFdYZz8J4INMb+xT64hE5vnn+uFNAiEA/x7Zs47KTpO3DVJBQF9ccegoIYLEbBsRdPV3vy+yqg8EIDQPDQXzf2I0GvERWZuYPxTl0635mJOesnFPD+Wj1AO2BBBfllNH03r8WZ2cvK3tlACxA0IABBRwHZgGMVEB2SnMRxGWmHnP0pwRLE8M1X4b9G47345dpVTkML5kbrde6OufsBIFLfLfGcrydVkeXRt3AY1uH40DIQCUyF3nhuu+9ibomzX4IcwcArNBOBiSoY9fe16RLZLJOg=="
	structure := container.NewVBox()
	KeyInput := buildInputCertEntry("Please input base64/hex private key")
	KeyInput.Wrapping = fyne.TextWrapWord
	//KeyInput.Text = "MHcCAQEEIP7J6j7OktAgLXGxKXNkD11Ua/Int8FyOpou21ClJ86JoAoGCCqBHM9VAYItoUQDQgAEXi1Fo4RreqNuDZlHmCKfII93S+YpKeN5fXgQt2aG/G66UKklbEweWvjRbbaXYA/zLYaEpOTisvjguwKUKOVhCQ=="
	// 创建输出框，供用户输入数据
	output := widget.NewMultiLineEntry()
	output.Hide()

	//确认按钮
	confirm := buildButton("确认", theme.ConfirmIcon(), func() {
		inputEnveloped := input.Text
		inputKey := KeyInput.Text

		// 保存到历史记录
		if inputEnveloped != "" {
			util.GetHistoryDB().AddHistory("📦 信封解析", inputEnveloped)

			// 刷新历史记录下拉框
			if historyManager := GetGlobalHistoryManager(); historyManager != nil {
				historyManager.LoadHistoryForTab("📦 信封解析")
			}
		}

		decodeEnveloped, err := base64.StdEncoding.DecodeString(inputEnveloped)
		if err != nil {
			decodeEnveloped, err = hex.DecodeString(inputEnveloped)
			if err != nil {
				fyne.LogError("解析信封请求错误", err)
				return
			}
		}

		decodeKey, err := base64.StdEncoding.DecodeString(inputKey)
		if err != nil {
			decodeKey, err = hex.DecodeString(inputKey)
			if err != nil {
				fyne.LogError("解析Key请求错误", err)
				return
			}
		}

		publicKey, privateKey, err := DecryptSM2EnvelopedKey(decodeEnveloped, decodeKey)
		if err != nil {
			fyne.LogError("解析信封失败", err)
			return
		}
		base64PublicKey := base64.StdEncoding.EncodeToString(publicKey)

		priKey := sm2.PrivateKey{
			D: new(big.Int).SetBytes(privateKey),
		}
		key := pkcs12.BuildPrivateKeyInfoNoPublicKey(&priKey, pkcs12.OidNamedCurveP256SM2)
		privateKey, _ = asn1.Marshal(key)
		base64PrivateKey := base64.StdEncoding.EncodeToString(privateKey)
		output.Text = fmt.Sprintf("publicKey:%s\nprivateKey:%s", base64PublicKey, base64PrivateKey)
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
	// 不添加全局输入框，它已经在主界面的固定位置
	// structure.Add(input)
	structure.Add(KeyInput)
	structure.Add(allButton)
	structure.Add(output)
	// 使用滚动容器支持长内容
	scrollContainer := container.NewScroll(structure)
	return container.NewMax(scrollContainer)
}

// 解析信封
func ParseSM2EnvelopedKey(data []byte) (*gm.SM2EnvelopedKey, error) {
	var sm2EnvelopedKey gm.SM2EnvelopedKey
	_, err := asn1.Unmarshal(data, &sm2EnvelopedKey)
	return &sm2EnvelopedKey, err
}

// 使用私钥解密数字信封，返回密钥对
func DecryptSM2EnvelopedKey(data []byte, signPrivateKey []byte) ([]byte, []byte, error) {
	//将SM2私钥转换为制定对象
	sm2SignPrivateKey, err := x509.ParseSm2PrivateKey(signPrivateKey)
	//解析SM2EnvelopedKey
	sm2EnvelopedKey, err := ParseSM2EnvelopedKey(data)
	if err != nil {
		return nil, nil, err
	}
	sm2Cipher, _ := asn1.Marshal(sm2EnvelopedKey.Sm2cipher)
	//解密信封,得到对称密钥
	sm4Key, err := sm2SignPrivateKey.DecryptAsn1(sm2Cipher)
	if err != nil {
		return nil, nil, err
	}
	//对称密钥 解密 加密密钥对私钥
	encPrivateKey, err := gm.DecryptDataUseSm4Key(sm2EnvelopedKey.Sm2EncryptedPrivateKey.Bytes, sm4Key)
	if err != nil {
		return nil, nil, err
	}
	return sm2EnvelopedKey.PublicKey.Bytes, encPrivateKey, nil

}
