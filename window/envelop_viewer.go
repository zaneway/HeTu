package window

import (
	"HeTu/gm"
	"HeTu/helper"
	"HeTu/util"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/zaneway/cain-go/sm2"
	"github.com/zaneway/cain-go/x509"
)

var knownAlgOIDs = map[string]string{
	"1.2.156.10197.1.104.1":  "SM4-ECB",
	"1.2.156.10197.1.104.2":  "SM4-CBC",
	"1.2.156.10197.1.104.3":  "SM4-OFB",
	"1.2.156.10197.1.104.4":  "SM4-CFB",
	"1.2.156.10197.1.104":    "SM4",
	"1.2.156.10197.1.301":    "SM2",
	"1.2.156.10197.1.401":    "SM3",
	"2.16.840.113549.3.4":    "RC4",
	"2.16.840.1.101.3.4.1.1": "AES-128-ECB",
	"2.16.840.1.101.3.4.1.2": "AES-128-CBC",
}

var (
	currentEnvelopedKey *gm.SM2EnvelopedKey
	currentDecodeData   []byte
)

func SM2EnvelopedPfxStructure(input *widget.Entry) *fyne.Container {
	input.Wrapping = fyne.TextWrapWord
	structure := container.NewVBox()

	keyInput := buildInputCertEntry("请输入 Base64/Hex 格式的 SM2 私钥")
	keyInput.Wrapping = fyne.TextWrapWord

	detail := container.NewVBox()

	parseFunc := func() {
		inputEnveloped := strings.TrimSpace(input.Text)

		if inputEnveloped == "" {
			dialog.ShowError(fmt.Errorf("请输入信封数据"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		decodeEnveloped, err := decodeInput(inputEnveloped)
		if err != nil {
			dialog.ShowError(fmt.Errorf("信封数据解码失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		util.GetHistoryDB().AddHistory("📦 信封解析", inputEnveloped)
		if historyManager := GetGlobalHistoryManager(); historyManager != nil {
			historyManager.LoadHistoryForTab("📦 信封解析")
		}

		sm2EnvelopedKey, err := ParseSM2EnvelopedKey(decodeEnveloped)
		if err != nil {
			dialog.ShowError(fmt.Errorf("信封结构解析失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		currentEnvelopedKey = sm2EnvelopedKey
		currentDecodeData = decodeEnveloped

		detail.RemoveAll()
		detail.Add(buildEnvelopeStructureCard(sm2EnvelopedKey))
		detail.Refresh()
	}

	decryptFunc := func() {
		inputEnveloped := strings.TrimSpace(input.Text)
		inputKey := strings.TrimSpace(keyInput.Text)

		if inputKey == "" {
			dialog.ShowError(fmt.Errorf("请输入解密私钥"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		if currentEnvelopedKey == nil || currentDecodeData == nil {
			if inputEnveloped == "" {
				dialog.ShowError(fmt.Errorf("请输入信封数据"), fyne.CurrentApp().Driver().AllWindows()[0])
				return
			}

			decodeEnveloped, err := decodeInput(inputEnveloped)
			if err != nil {
				dialog.ShowError(fmt.Errorf("信封数据解码失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
				return
			}

			util.GetHistoryDB().AddHistory("📦 信封解析", inputEnveloped)
			if historyManager := GetGlobalHistoryManager(); historyManager != nil {
				historyManager.LoadHistoryForTab("📦 信封解析")
			}

			sm2EnvelopedKey, err := ParseSM2EnvelopedKey(decodeEnveloped)
			if err != nil {
				dialog.ShowError(fmt.Errorf("信封结构解析失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
				return
			}

			currentEnvelopedKey = sm2EnvelopedKey
			currentDecodeData = decodeEnveloped
		}

		decodeKey, err := decodeInput(inputKey)
		if err != nil {
			dialog.ShowError(fmt.Errorf("私钥解码失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		sm2SignPrivateKey, err := parsePrivateKey(decodeKey)
		if err != nil {
			dialog.ShowError(fmt.Errorf("私钥解析失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		sm2CipherBytes, _ := asn1.Marshal(currentEnvelopedKey.Sm2cipher)
		sm4Key, err := sm2SignPrivateKey.DecryptAsn1(sm2CipherBytes)
		if err != nil {
			dialog.ShowError(fmt.Errorf("对称密钥解密失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		encPrivateKey, err := gm.DecryptDataUseSm4Key(currentEnvelopedKey.Sm2EncryptedPrivateKey.Bytes, sm4Key)
		if err != nil {
			dialog.ShowError(fmt.Errorf("私钥解密失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		detail.RemoveAll()
		detail.Add(buildDecryptResultCard(sm4Key, encPrivateKey, currentEnvelopedKey.PublicKey.Bytes))
		detail.Refresh()
	}

	parseBtn := widget.NewButtonWithIcon("解析", theme.ConfirmIcon(), parseFunc)
	decryptBtn := widget.NewButtonWithIcon("解密", theme.VisibilityIcon(), decryptFunc)
	clearBtn := widget.NewButtonWithIcon("清除", theme.CancelIcon(), func() {
		input.Text = ""
		keyInput.Text = ""
		input.Refresh()
		keyInput.Refresh()
		detail.RemoveAll()
		currentEnvelopedKey = nil
		currentDecodeData = nil
		detail.Refresh()
	})

	buttonRow := container.New(layout.NewGridLayout(3), parseBtn, decryptBtn, clearBtn)

	structure.Add(keyInput)
	structure.Add(buttonRow)
	structure.Add(detail)

	scrollContainer := container.NewScroll(structure)
	return container.NewMax(scrollContainer)
}

func buildEnvelopeStructureCard(env *gm.SM2EnvelopedKey) *widget.Card {
	oidStr := env.SymAlgID.Algorithm.String()
	algName := oidStr
	if name, ok := knownAlgOIDs[oidStr]; ok {
		algName = fmt.Sprintf("%s (%s)", name, oidStr)
	}

	xHex := fmt.Sprintf("%064x", env.Sm2cipher.X)
	yHex := fmt.Sprintf("%064x", env.Sm2cipher.Y)
	hashHex := hex.EncodeToString(env.Sm2cipher.Hash)
	cipherHex := hex.EncodeToString(env.Sm2cipher.CipherText)
	pubKeyHex := hex.EncodeToString(env.PublicKey.Bytes)
	encPrivKeyHex := hex.EncodeToString(env.Sm2EncryptedPrivateKey.Bytes)

	form := widget.NewForm(
		widget.NewFormItem("对称算法 OID", newSelectableLabel(algName)),
		widget.NewFormItem("SM2Cipher.X", newSelectableLabel(xHex)),
		widget.NewFormItem("SM2Cipher.Y", newSelectableLabel(yHex)),
		widget.NewFormItem("SM2Cipher.Hash", newSelectableLabel(hashHex)),
		widget.NewFormItem("SM2Cipher.CipherText", newSelectableLabel(cipherHex)),
		widget.NewFormItem("SM2 公钥", newSelectableLabel(pubKeyHex)),
		widget.NewFormItem("加密的私钥", newSelectableLabel(encPrivKeyHex)),
	)

	return widget.NewCard("📋 信封结构", "SM2EnvelopedKey ASN.1 结构", form)
}

func buildDecryptResultCard(sm4Key, privateKey, publicKey []byte) *widget.Card {
	form := widget.NewForm(
		widget.NewFormItem("公钥 (Hex)", newCopyableEntry(hex.EncodeToString(publicKey))),
		widget.NewFormItem("公钥 (Base64)", newCopyableEntry(base64.StdEncoding.EncodeToString(publicKey))),
		widget.NewFormItem("私钥明文 (Hex)", newCopyableEntry(hex.EncodeToString(privateKey))),
		widget.NewFormItem("私钥明文 (Base64)", newCopyableEntry(base64.StdEncoding.EncodeToString(privateKey))),
		widget.NewFormItem("对称密钥 (Hex)", newCopyableEntry(hex.EncodeToString(sm4Key))),
		widget.NewFormItem("对称密钥 (Base64)", newCopyableEntry(base64.StdEncoding.EncodeToString(sm4Key))),
	)
	return widget.NewCard("🔓 解密结果", "信封解密成功", form)
}

func newSelectableLabel(text string) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetText(text)
	entry.Disable()
	return entry
}

func newCopyableEntry(text string) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetText(text)
	return entry
}

func decodeInput(input string) ([]byte, error) {
	if strings.Contains(input, "-----BEGIN") {
		var b64 strings.Builder
		for _, line := range strings.Split(input, "\n") {
			line = strings.TrimSpace(strings.TrimRight(line, "\r"))
			if line == "" || strings.HasPrefix(line, "-----") {
				continue
			}
			b64.WriteString(line)
		}
		input = b64.String()
	}
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, "\n", "")
	input = strings.ReplaceAll(input, "\r", "")
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err == nil {
		return decoded, nil
	}
	decoded, err = hex.DecodeString(input)
	if err == nil {
		return decoded, nil
	}
	return nil, fmt.Errorf("无法识别的编码格式，请使用 Base64 或 Hex")
}

func ParseSM2EnvelopedKey(data []byte) (*gm.SM2EnvelopedKey, error) {
	var sm2EnvelopedKey gm.SM2EnvelopedKey
	_, err := asn1.Unmarshal(data, &sm2EnvelopedKey)
	return &sm2EnvelopedKey, err
}

func DecryptSM2EnvelopedKey(data []byte, signPrivateKey []byte) ([]byte, []byte, error) {
	sm2SignPrivateKey, err := parsePrivateKey(signPrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("私钥解析失败: %v", err)
	}

	sm2EnvelopedKey, err := ParseSM2EnvelopedKey(data)
	if err != nil {
		return nil, nil, fmt.Errorf("信封解析失败: %v", err)
	}

	sm2Cipher, _ := asn1.Marshal(sm2EnvelopedKey.Sm2cipher)
	sm4Key, err := sm2SignPrivateKey.DecryptAsn1(sm2Cipher)
	if err != nil {
		return nil, nil, fmt.Errorf("对称密钥解密失败: %v", err)
	}

	encPrivateKey, err := gm.DecryptDataUseSm4Key(sm2EnvelopedKey.Sm2EncryptedPrivateKey.Bytes, sm4Key)
	if err != nil {
		return nil, nil, fmt.Errorf("私钥解密失败: %v", err)
	}

	return sm2EnvelopedKey.PublicKey.Bytes, encPrivateKey, nil
}

func parsePrivateKey(data []byte) (*sm2.PrivateKey, error) {
	if len(data) == 32 {
		return helper.BuildPrivateKeyUseRaw(data), nil
	}
	if len(data) == 30 {
		padded := make([]byte, 32)
		copy(padded[2:], data)
		return helper.BuildPrivateKeyUseRaw(padded), nil
	}

	privKey, err := parsePKCS8PrivateKey(data)
	if err == nil {
		return privKey, nil
	}

	return x509.ParseSm2PrivateKey(data)
}

var (
	oidPublicKeyEC       = asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}
	oidSM2Curve          = asn1.ObjectIdentifier{1, 2, 156, 10197, 1, 301}
	oidNamedCurveP256SM2 = asn1.ObjectIdentifier{1, 2, 156, 10197, 1, 301}
)

type pkcs8PrivateKey struct {
	Version    int
	Algo       pkix.AlgorithmIdentifier
	PrivateKey []byte
}

type pkixAlgorithmIdentifier struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.RawValue `optional:"true"`
}

type ecPrivateKey struct {
	Version    int
	PrivateKey []byte
	Parameters asn1.RawValue `optional:"true"`
	PublicKey  asn1.RawValue `optional:"true"`
}

type pkixPublicKeyAlgorithmIdentifier struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.RawValue `optional:"true"`
}

func parsePKCS8PrivateKey(data []byte) (*sm2.PrivateKey, error) {
	var privKeyInfo pkcs8PrivateKey
	_, err := asn1.Unmarshal(data, &privKeyInfo)
	if err != nil {
		return nil, fmt.Errorf("unmarshal PKCS8 failed: %v", err)
	}

	if !privKeyInfo.Algo.Algorithm.Equal(oidPublicKeyEC) {
		return nil, fmt.Errorf("not EC key")
	}

	var ecPriv ecPrivateKey
	_, err = asn1.Unmarshal(privKeyInfo.PrivateKey, &ecPriv)
	if err != nil {
		return nil, fmt.Errorf("unmarshal EC private key failed: %v", err)
	}

	privKey := &sm2.PrivateKey{}
	privKey.Curve = sm2.P256Sm2()
	privKey.D = new(big.Int).SetBytes(ecPriv.PrivateKey)
	privKey.PublicKey.X, privKey.PublicKey.Y = privKey.Curve.ScalarBaseMult(ecPriv.PrivateKey)

	return privKey, nil
}
