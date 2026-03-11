package window

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
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
	"github.com/zaneway/cain-go/sm4"
)

func KeyStructure(input *widget.Entry) *fyne.Container {
	structure := container.NewVBox()

	algoSelect := widget.NewSelect([]string{
		"SM2", "RSA-2048",
		"SM4", "AES-128", "AES-256",
	}, nil)
	algoSelect.PlaceHolder = "选择加密算法"

	keyInput := widget.NewMultiLineEntry()
	keyInput.SetPlaceHolder("输入密钥 (Base64/Hex格式)\n- 对称算法：输入对称密钥\n- 非对称算法：输入公钥(加密)或私钥(解密)")
	keyInput.Wrapping = fyne.TextWrapWord
	keyInput.Resize(fyne.NewSize(600, 80))

	dataInput := widget.NewMultiLineEntry()
	dataInput.SetPlaceHolder("输入待处理数据 (明文或Base64/Hex密文)")
	dataInput.Wrapping = fyne.TextWrapWord
	dataInput.Resize(fyne.NewSize(600, 80))

	resultArea := widget.NewMultiLineEntry()
	resultArea.SetPlaceHolder("处理结果")
	resultArea.Wrapping = fyne.TextWrapWord
	resultArea.Resize(fyne.NewSize(600, 100))

	statusLabel := widget.NewLabel("💡 选择算法，输入密钥和数据")
	statusLabel.TextStyle = fyne.TextStyle{Italic: true}

	processBtn := widget.NewButtonWithIcon("加密", theme.ConfirmIcon(), func() {
		algo := algoSelect.Selected

		if algo == "" {
			dialog.ShowError(fmt.Errorf("请选择加密算法"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		keyStr := strings.TrimSpace(keyInput.Text)
		if keyStr == "" {
			dialog.ShowError(fmt.Errorf("请输入密钥"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		dataStr := strings.TrimSpace(dataInput.Text)
		if dataStr == "" {
			dialog.ShowError(fmt.Errorf("请输入数据"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		statusLabel.SetText("🔄 加密中...")

		go func() {
			result, err := processData(algo, "加密", keyStr, dataStr)
			fyne.Do(func() {
				if err != nil {
					statusLabel.SetText(fmt.Sprintf("❌ 加密失败: %v", err))
				} else {
					resultArea.SetText(result)
					resultArea.Refresh()
					statusLabel.SetText("✅ 加密完成")
				}
			})
		}()
	})

	decryptBtn := widget.NewButtonWithIcon("解密", theme.VisibilityIcon(), func() {
		algo := algoSelect.Selected

		if algo == "" {
			dialog.ShowError(fmt.Errorf("请选择加密算法"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		keyStr := strings.TrimSpace(keyInput.Text)
		if keyStr == "" {
			dialog.ShowError(fmt.Errorf("请输入密钥"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		dataStr := strings.TrimSpace(dataInput.Text)
		if dataStr == "" {
			dialog.ShowError(fmt.Errorf("请输入数据"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		statusLabel.SetText("🔄 解密中...")

		go func() {
			result, err := processData(algo, "解密", keyStr, dataStr)
			fyne.Do(func() {
				if err != nil {
					statusLabel.SetText(fmt.Sprintf("❌ 解密失败: %v", err))
				} else {
					resultArea.SetText(result)
					resultArea.Refresh()
					statusLabel.SetText("✅ 解密完成")
				}
			})
		}()
	})

	clearBtn := widget.NewButtonWithIcon("清除", theme.CancelIcon(), func() {
		algoSelect.ClearSelected()
		keyInput.SetText("")
		dataInput.SetText("")
		resultArea.SetText("")
		statusLabel.SetText("💡 选择算法，输入密钥和数据")
	})

	copyResultBtn := widget.NewButtonWithIcon("复制结果", theme.ContentCopyIcon(), func() {
		if strings.TrimSpace(resultArea.Text) == "" {
			dialog.ShowError(fmt.Errorf("没有可复制的结果"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}
		clipboard := fyne.CurrentApp().Driver().AllWindows()[0].Clipboard()
		clipboard.SetContent(resultArea.Text)
		statusLabel.SetText("📋 结果已复制到剪贴板")
	})

	buttonRow := container.New(layout.NewGridLayout(4), processBtn, decryptBtn, clearBtn, copyResultBtn)

	structure.Add(statusLabel)
	structure.Add(widget.NewSeparator())
	structure.Add(algoSelect)
	structure.Add(keyInput)
	structure.Add(dataInput)
	structure.Add(buttonRow)
	structure.Add(resultArea)

	scrollContainer := container.NewScroll(structure)
	return container.NewMax(scrollContainer)
}

func processData(algo, mode, keyStr, dataStr string) (string, error) {
	keyData, err := decodeKey(keyStr, algo)
	if err != nil {
		return "", fmt.Errorf("密钥解析失败: %v", err)
	}

	var inputData []byte
	if mode == "加密" {
		inputData, err = decodeData(dataStr, false)
	} else {
		inputData, err = decodeData(dataStr, true)
	}
	if err != nil {
		return "", fmt.Errorf("数据解析失败: %v", err)
	}

	var result []byte
	switch algo {
	case "SM2":
		if mode == "加密" {
			result, err = sm2Encrypt(keyData, inputData)
		} else {
			result, err = sm2Decrypt(keyData, inputData)
		}
	case "RSA-2048":
		if mode == "加密" {
			result, err = rsaEncrypt(keyData, inputData)
		} else {
			result, err = rsaDecrypt(keyData, inputData)
		}
	case "SM4":
		if mode == "加密" {
			result, err = sm4Encrypt(keyData, inputData)
		} else {
			result, err = sm4Decrypt(keyData, inputData)
		}
	case "AES-128", "AES-256":
		if mode == "加密" {
			result, err = aesEncrypt(keyData, inputData)
		} else {
			result, err = aesDecrypt(keyData, inputData)
		}
	default:
		return "", fmt.Errorf("不支持的算法: %s", algo)
	}

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Hex: %s\nBase64: %s", hex.EncodeToString(result), base64.StdEncoding.EncodeToString(result)), nil
}

func decodeKey(keyStr, algo string) ([]byte, error) {
	cleaned := strings.TrimSpace(keyStr)

	if data, err := base64.StdEncoding.DecodeString(cleaned); err == nil {
		return data, nil
	}
	if data, err := hex.DecodeString(cleaned); err == nil {
		return data, nil
	}

	if strings.Contains(cleaned, "-----BEGIN") {
		block, _ := pem.Decode([]byte(cleaned))
		if block != nil {
			return block.Bytes, nil
		}
	}

	cleaned = strings.TrimSpace(keyStr)
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "\n", "")
	cleaned = strings.ReplaceAll(cleaned, "\r", "")
	cleaned = strings.ReplaceAll(cleaned, "\t", "")

	if data, err := base64.StdEncoding.DecodeString(cleaned); err == nil {
		return data, nil
	}
	if data, err := hex.DecodeString(cleaned); err == nil {
		return data, nil
	}

	return nil, fmt.Errorf("无法解析密钥格式")
}

func decodeData(dataStr string, isCipher bool) ([]byte, error) {
	cleaned := strings.TrimSpace(dataStr)

	if data, err := base64.StdEncoding.DecodeString(cleaned); err == nil {
		return data, nil
	}
	if data, err := hex.DecodeString(cleaned); err == nil {
		return data, nil
	}

	if !isCipher {
		return []byte(cleaned), nil
	}

	return nil, fmt.Errorf("无法解析数据格式，请使用Base64或Hex编码")
}

func sm2Encrypt(keyData []byte, data []byte) ([]byte, error) {
	pubKey, err := parseSM2PublicKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("解析SM2公钥失败: %v", err)
	}
	return sm2.Encrypt(pubKey, data, rand.Reader, sm2.C1C3C2)
}

func sm2Decrypt(keyData []byte, data []byte) ([]byte, error) {
	privKey, err := parseSM2PrivateKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("解析SM2私钥失败: %v", err)
	}
	return sm2.Decrypt(privKey, data, sm2.C1C3C2)
}

func rsaEncrypt(keyData []byte, data []byte) ([]byte, error) {
	pubKey, err := parseRSAPublicKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("解析RSA公钥失败: %v", err)
	}
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, data, nil)
}

func rsaDecrypt(keyData []byte, data []byte) ([]byte, error) {
	privKey, err := parseRSAPrivateKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("解析RSA私钥失败: %v", err)
	}
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, data, nil)
}

func sm4Encrypt(keyData []byte, data []byte) ([]byte, error) {
	return sm4.Sm4EcbNoPaddingCipher(keyData, data, true)
}

func sm4Decrypt(keyData []byte, data []byte) ([]byte, error) {
	return sm4.Sm4EcbNoPaddingCipher(keyData, data, false)
}

func aesEncrypt(keyData []byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(keyData)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, data, nil), nil
}

func aesDecrypt(keyData []byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(keyData)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("密文数据太短")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func parseSM2PublicKey(data []byte) (*sm2.PublicKey, error) {
	if len(data) >= 64 {
		pubKey := &sm2.PublicKey{}
		pubKey.Curve = sm2.P256Sm2()
		pubKey.X = new(big.Int).SetBytes(data[:32])
		pubKey.Y = new(big.Int).SetBytes(data[32:64])
		return pubKey, nil
	}

	pubKeyInterface, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return nil, err
	}
	pubKey, ok := pubKeyInterface.(*sm2.PublicKey)
	if !ok {
		return nil, errors.New("不是有效的SM2公钥")
	}
	return pubKey, nil
}

func parseSM2PrivateKey(data []byte) (*sm2.PrivateKey, error) {
	if len(data) >= 30 && len(data) <= 32 {
		keyBytes := data
		if len(data) == 30 {
			padded := make([]byte, 32)
			copy(padded[2:], data)
			keyBytes = padded
		}
		privKey := &sm2.PrivateKey{}
		privKey.Curve = sm2.P256Sm2()
		privKey.D = new(big.Int).SetBytes(keyBytes)
		privKey.PublicKey.X, privKey.PublicKey.Y = privKey.Curve.ScalarBaseMult(keyBytes)
		return privKey, nil
	}

	privKeyInterface, err := x509.ParsePKCS8PrivateKey(data)
	if err == nil {
		if privKey, ok := privKeyInterface.(*sm2.PrivateKey); ok {
			return privKey, nil
		}
	}

	privKeyInterface, err = x509.ParsePKIXPublicKey(data)
	if err == nil {
		return nil, errors.New("输入的是公钥，请输入私钥")
	}

	return nil, fmt.Errorf("不是有效的SM2私钥: %v", err)
}

func parseRSAPublicKey(data []byte) (*rsa.PublicKey, error) {
	pubKeyInterface, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return nil, err
	}
	pubKey, ok := pubKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("不是有效的RSA公钥")
	}
	return pubKey, nil
}

func parseRSAPrivateKey(data []byte) (*rsa.PrivateKey, error) {
	privKeyInterface, err := x509.ParsePKCS8PrivateKey(data)
	if err != nil {
		return nil, err
	}
	privKey, ok := privKeyInterface.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("不是有效的RSA私钥")
	}
	return privKey, nil
}
