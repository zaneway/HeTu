package window

import (
	"HeTu/security"
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
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/zaneway/cain-go/sm2"
	"github.com/zaneway/cain-go/sm4"
)

// KeyData 存储密钥数据
type KeyData struct {
	Algorithm    string
	PublicKey    []byte
	PrivateKey   []byte
	SymmetricKey []byte
	KeySize      int
}

// 当前生成的密钥
var currentKeyData *KeyData

func KeyStructure(input *widget.Entry) *fyne.Container {
	// 密钥生成模块使用独立的输入区域，不隐藏全局输入框以保持其他标签页正常工作

	// 创建算法选择下拉框
	algorithmSelect := widget.NewSelect(append(security.ALL_ASYM_KEYS, security.ALL_SYM_KEYS...), nil)
	// algorithmSelect.SetPlaceHolder("🔐 请选择密钥算法") // 注释掉不支持的方法

	// 创建密钥显示区域
	keyDisplayArea := widget.NewMultiLineEntry()
	keyDisplayArea.SetPlaceHolder("生成的密钥将在这里显示...")
	keyDisplayArea.Wrapping = fyne.TextWrapWord
	keyDisplayArea.Resize(fyne.NewSize(600, 150))

	// 创建待加密数据输入区域
	dataInput := widget.NewMultiLineEntry()
	dataInput.SetPlaceHolder("📝 请输入要加密的数据（明文或Base64/Hex编码数据）")
	dataInput.Wrapping = fyne.TextWrapWord
	dataInput.Resize(fyne.NewSize(600, 100))

	// 创建加解密结果显示区域
	resultArea := widget.NewMultiLineEntry()
	resultArea.SetPlaceHolder("加解密结果将在这里显示...")
	resultArea.Wrapping = fyne.TextWrapWord
	resultArea.Resize(fyne.NewSize(600, 150))

	// 创建状态标签
	statusLabel := widget.NewLabel("💡 请选择算法并生成密钥")
	statusLabel.TextStyle = fyne.TextStyle{Italic: true}

	// 生成密钥按钮
	generateBtn := widget.NewButtonWithIcon("🔑 生成密钥", theme.ConfirmIcon(), func() {
		selectedAlg := algorithmSelect.Selected
		if selectedAlg == "" {
			dialog.ShowError(fmt.Errorf("请先选择密钥算法"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		statusLabel.SetText("🔄 正在生成密钥...")

		// 异步生成密钥
		go func() {
			time.Sleep(time.Millisecond * 200) // 显示生成过程

			keyData, err := generateKey(selectedAlg)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("密钥生成失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
					statusLabel.SetText("❌ 密钥生成失败")
				})
				return
			}

			currentKeyData = keyData
			fyne.Do(func() {
				displayKeyInfo(keyDisplayArea, keyData)
				statusLabel.SetText("✅ 密钥生成成功")
			})
		}()
	})

	// 加密按钮
	encryptBtn := widget.NewButtonWithIcon("🔒 加密", theme.ContentAddIcon(), func() {
		if currentKeyData == nil {
			dialog.ShowError(fmt.Errorf("请先生成密钥"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		inputData := strings.TrimSpace(dataInput.Text)
		if inputData == "" {
			dialog.ShowError(fmt.Errorf("请输入要加密的数据"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		statusLabel.SetText("🔄 正在加密...")

		go func() {
			encryptedData, err := encryptData(currentKeyData, inputData)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("加密失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
					statusLabel.SetText("❌ 加密失败")
				})
				return
			}

			fyne.Do(func() {
				resultArea.SetText(fmt.Sprintf("🔒 加密结果 (%s):\n%s", currentKeyData.Algorithm, encryptedData))
				statusLabel.SetText("✅ 加密完成")
			})
		}()
	})

	// 解密按钮
	decryptBtn := widget.NewButtonWithIcon("🔓 解密", theme.ContentRemoveIcon(), func() {
		if currentKeyData == nil {
			dialog.ShowError(fmt.Errorf("请先生成密钥"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		inputData := strings.TrimSpace(dataInput.Text)
		if inputData == "" {
			dialog.ShowError(fmt.Errorf("请输入要解密的数据"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		statusLabel.SetText("🔄 正在解密...")

		go func() {
			decryptedData, err := decryptData(currentKeyData, inputData)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("解密失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
					statusLabel.SetText("❌ 解密失败")
				})
				return
			}

			fyne.Do(func() {
				resultArea.SetText(fmt.Sprintf("🔓 解密结果 (%s):\n%s", currentKeyData.Algorithm, decryptedData))
				statusLabel.SetText("✅ 解密完成")
			})
		}()
	})

	// 清除按钮
	clearBtn := widget.NewButtonWithIcon("🗑️ 清除", theme.CancelIcon(), func() {
		algorithmSelect.ClearSelected()
		keyDisplayArea.SetText("")
		dataInput.SetText("")
		resultArea.SetText("")
		currentKeyData = nil
		statusLabel.SetText("💡 请选择算法并生成密钥")
	})

	// 复制密钥按钮
	copyKeyBtn := widget.NewButtonWithIcon("📋 复制密钥", theme.ContentCopyIcon(), func() {
		if currentKeyData == nil {
			dialog.ShowError(fmt.Errorf("没有可复制的密钥"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		clipboard := fyne.CurrentApp().Driver().AllWindows()[0].Clipboard()
		clipboard.SetContent(keyDisplayArea.Text)
		statusLabel.SetText("📋 密钥已复制到剪贴板")
	})

	// 复制结果按钮
	copyResultBtn := widget.NewButtonWithIcon("📋 复制结果", theme.ContentCopyIcon(), func() {
		if strings.TrimSpace(resultArea.Text) == "" {
			dialog.ShowError(fmt.Errorf("没有可复制的结果"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		clipboard := fyne.CurrentApp().Driver().AllWindows()[0].Clipboard()
		clipboard.SetContent(resultArea.Text)
		statusLabel.SetText("📋 结果已复制到剪贴板")
	})

	// 按钮布局
	actionButtons := container.New(layout.NewGridLayout(2), encryptBtn, decryptBtn)
	// controlButtons := container.New(layout.NewGridLayout(2), copyKeyBtn, copyResultBtn) // 已集成到各自区域
	mainButtons := container.New(layout.NewGridLayout(2), generateBtn, clearBtn)

	// 创建分组容器
	keyGenGroup := widget.NewCard("🔑 密钥生成", "", container.NewVBox(
		algorithmSelect,
		mainButtons,
		keyDisplayArea,
		container.NewHBox(copyKeyBtn, layout.NewSpacer()),
	))

	cryptoGroup := widget.NewCard("🔐 加解密操作", "", container.NewVBox(
		dataInput,
		actionButtons,
		resultArea,
		container.NewHBox(copyResultBtn, layout.NewSpacer()),
	))

	// 主容器
	mainContainer := container.NewVBox(
		statusLabel,
		widget.NewSeparator(),
		keyGenGroup,
		widget.NewSeparator(),
		cryptoGroup,
	)

	// 使用滚动容器
	scrollContainer := container.NewScroll(mainContainer)
	return container.NewMax(scrollContainer)
}

// generateKey 根据算法生成密钥
func generateKey(algorithm string) (*KeyData, error) {
	keyData := &KeyData{
		Algorithm: algorithm,
	}

	switch algorithm {
	case security.SM2_256:
		return generateSM2Key(keyData)
	case security.RSA_1024:
		return generateRSAKey(keyData, 1024)
	case security.RSA_2048:
		return generateRSAKey(keyData, 2048)
	case security.RSA_4096:
		return generateRSAKey(keyData, 4096)
	case security.AES_128:
		return generateAESKey(keyData, 16)
	case security.AES_256:
		return generateAESKey(keyData, 32)
	case security.AES_384:
		return generateAESKey(keyData, 48)
	case security.AES_512:
		return generateAESKey(keyData, 64)
	case security.SM4_128:
		return generateSM4Key(keyData)
	default:
		return nil, fmt.Errorf("不支持的算法: %s", algorithm)
	}
}

// generateSM2Key 生成SM2密钥对
func generateSM2Key(keyData *KeyData) (*KeyData, error) {
	privateKey, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("SM2密钥生成失败: %v", err)
	}

	// 提取公钥坐标
	pubKeyBytes := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)
	privKeyBytes := privateKey.D.Bytes()

	keyData.PublicKey = pubKeyBytes
	keyData.PrivateKey = privKeyBytes
	keyData.KeySize = 256

	return keyData, nil
}

// generateRSAKey 生成RSA密钥对
func generateRSAKey(keyData *KeyData, keySize int) (*KeyData, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, fmt.Errorf("RSA密钥生成失败: %v", err)
	}

	// 序列化私钥
	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("RSA私钥序列化失败: %v", err)
	}

	// 序列化公钥
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("RSA公钥序列化失败: %v", err)
	}

	keyData.PublicKey = pubKeyBytes
	keyData.PrivateKey = privKeyBytes
	keyData.KeySize = keySize

	return keyData, nil
}

// generateAESKey 生成AES对称密钥
func generateAESKey(keyData *KeyData, keySize int) (*KeyData, error) {
	key := make([]byte, keySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("AES密钥生成失败: %v", err)
	}

	keyData.SymmetricKey = key
	keyData.KeySize = keySize * 8 // 转换为位数

	return keyData, nil
}

// generateSM4Key 生成SM4对称密钥
func generateSM4Key(keyData *KeyData) (*KeyData, error) {
	key := make([]byte, 16) // SM4固定128位密钥
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("SM4密钥生成失败: %v", err)
	}

	keyData.SymmetricKey = key
	keyData.KeySize = 128

	return keyData, nil
}

// displayKeyInfo 显示密钥信息
func displayKeyInfo(display *widget.Entry, keyData *KeyData) {
	var info strings.Builder

	info.WriteString(fmt.Sprintf("🔐 算法: %s\n", keyData.Algorithm))
	info.WriteString(fmt.Sprintf("🔢 密钥长度: %d 位\n", keyData.KeySize))
	info.WriteString("\n")

	if keyData.SymmetricKey != nil {
		// 对称密钥
		info.WriteString("🔑 对称密钥:\n")
		info.WriteString(fmt.Sprintf("Hex: %s\n", hex.EncodeToString(keyData.SymmetricKey)))
		info.WriteString(fmt.Sprintf("Base64: %s\n", base64.StdEncoding.EncodeToString(keyData.SymmetricKey)))
	} else {
		// 非对称密钥对
		info.WriteString("🔓 公钥:\n")
		if strings.Contains(keyData.Algorithm, "RSA") {
			// RSA公钥PEM格式
			pubKeyPEM := &pem.Block{
				Type:  "PUBLIC KEY",
				Bytes: keyData.PublicKey,
			}
			info.WriteString(string(pem.EncodeToMemory(pubKeyPEM)))
		} else {
			// SM2公钥Hex格式
			info.WriteString(fmt.Sprintf("Hex: %s\n", hex.EncodeToString(keyData.PublicKey)))
			info.WriteString(fmt.Sprintf("Base64: %s\n", base64.StdEncoding.EncodeToString(keyData.PublicKey)))
		}

		info.WriteString("\n🔐 私钥:\n")
		if strings.Contains(keyData.Algorithm, "RSA") {
			// RSA私钥PEM格式
			privKeyPEM := &pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: keyData.PrivateKey,
			}
			info.WriteString(string(pem.EncodeToMemory(privKeyPEM)))
		} else {
			// SM2私钥Hex格式
			info.WriteString(fmt.Sprintf("Hex: %s\n", hex.EncodeToString(keyData.PrivateKey)))
			info.WriteString(fmt.Sprintf("Base64: %s\n", base64.StdEncoding.EncodeToString(keyData.PrivateKey)))
		}
	}

	info.WriteString("\n💡 提示: 可以点击复制按钮将密钥复制到剪贴板")

	display.SetText(info.String())
}

// encryptData 加密数据
func encryptData(keyData *KeyData, input string) (string, error) {
	// 预处理输入数据
	data, err := preprocessInputData(input)
	if err != nil {
		return "", fmt.Errorf("数据预处理失败: %v", err)
	}

	switch keyData.Algorithm {
	case security.SM2_256:
		return encryptWithSM2(keyData, data)
	case security.RSA_1024, security.RSA_2048, security.RSA_4096:
		return encryptWithRSA(keyData, data)
	case security.AES_128, security.AES_256, security.AES_384, security.AES_512:
		return encryptWithAES(keyData, data)
	case security.SM4_128:
		return encryptWithSM4(keyData, data)
	default:
		return "", fmt.Errorf("不支持的加密算法: %s", keyData.Algorithm)
	}
}

// decryptData 解密数据
func decryptData(keyData *KeyData, input string) (string, error) {
	// 预处理输入数据（Base64或Hex解码）
	data, err := preprocessCipherData(input)
	if err != nil {
		return "", fmt.Errorf("密文数据预处理失败: %v", err)
	}

	switch keyData.Algorithm {
	case security.SM2_256:
		return decryptWithSM2(keyData, data)
	case security.RSA_1024, security.RSA_2048, security.RSA_4096:
		return decryptWithRSA(keyData, data)
	case security.AES_128, security.AES_256, security.AES_384, security.AES_512:
		return decryptWithAES(keyData, data)
	case security.SM4_128:
		return decryptWithSM4(keyData, data)
	default:
		return "", fmt.Errorf("不支持的解密算法: %s", keyData.Algorithm)
	}
}

// preprocessInputData 预处理输入数据
func preprocessInputData(input string) ([]byte, error) {
	cleaned := strings.TrimSpace(input)
	if cleaned == "" {
		return nil, errors.New("输入数据为空")
	}

	// 尝试Base64解码
	if data, err := base64.StdEncoding.DecodeString(cleaned); err == nil {
		return data, nil
	}

	// 尝试Hex解码
	if data, err := hex.DecodeString(cleaned); err == nil {
		return data, nil
	}

	// 作为普通字符串处理
	return []byte(cleaned), nil
}

// preprocessCipherData 预处理密文数据
func preprocessCipherData(input string) ([]byte, error) {
	cleaned := strings.TrimSpace(input)
	if cleaned == "" {
		return nil, errors.New("密文数据为空")
	}

	// 优先尝试Base64解码
	if data, err := base64.StdEncoding.DecodeString(cleaned); err == nil {
		return data, nil
	}

	// 然后尝试Hex解码
	if data, err := hex.DecodeString(cleaned); err == nil {
		return data, nil
	}

	return nil, errors.New("无法解码密文数据，请确保输入正确的Base64或Hex格式")
}

// SM2加密
func encryptWithSM2(keyData *KeyData, data []byte) (string, error) {
	// 构建SM2公钥
	pubKey, err := buildSM2PublicKey(keyData.PublicKey)
	if err != nil {
		return "", fmt.Errorf("构建SM2公钥失败: %v", err)
	}

	// SM2加密
	ciphertext, err := sm2.Encrypt(pubKey, data, rand.Reader, sm2.C1C3C2)
	if err != nil {
		return "", fmt.Errorf("SM2加密失败: %v", err)
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// SM2解密
func decryptWithSM2(keyData *KeyData, data []byte) (string, error) {
	// 构建SM2私钥
	privKey, err := buildSM2PrivateKey(keyData.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("构建SM2私钥失败: %v", err)
	}

	// SM2解密
	plaintext, err := sm2.Decrypt(privKey, data, sm2.C1C3C2)
	if err != nil {
		return "", fmt.Errorf("SM2解密失败: %v", err)
	}

	return string(plaintext), nil
}

// RSA加密
func encryptWithRSA(keyData *KeyData, data []byte) (string, error) {
	// 解析RSA公钥
	pubKeyInterface, err := x509.ParsePKIXPublicKey(keyData.PublicKey)
	if err != nil {
		return "", fmt.Errorf("解析RSA公钥失败: %v", err)
	}

	pubKey, ok := pubKeyInterface.(*rsa.PublicKey)
	if !ok {
		return "", errors.New("不是有效的RSA公钥")
	}

	// RSA加密（使用OAEP填充）
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, data, nil)
	if err != nil {
		return "", fmt.Errorf("RSA加密失败: %v", err)
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// RSA解密
func decryptWithRSA(keyData *KeyData, data []byte) (string, error) {
	// 解析RSA私钥
	privKeyInterface, err := x509.ParsePKCS8PrivateKey(keyData.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("解析RSA私钥失败: %v", err)
	}

	privKey, ok := privKeyInterface.(*rsa.PrivateKey)
	if !ok {
		return "", errors.New("不是有效的RSA私钥")
	}

	// RSA解密（使用OAEP填充）
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, data, nil)
	if err != nil {
		return "", fmt.Errorf("RSA解密失败: %v", err)
	}

	return string(plaintext), nil
}

// AES加密
func encryptWithAES(keyData *KeyData, data []byte) (string, error) {
	// 创建AES加密器
	block, err := aes.NewCipher(keyData.SymmetricKey)
	if err != nil {
		return "", fmt.Errorf("创建AES加密器失败: %v", err)
	}

	// 使用GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM模式失败: %v", err)
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("生成nonce失败: %v", err)
	}

	// 加密
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AES解密
func decryptWithAES(keyData *KeyData, data []byte) (string, error) {
	// 创建AES解密器
	block, err := aes.NewCipher(keyData.SymmetricKey)
	if err != nil {
		return "", fmt.Errorf("创建AES解密器失败: %v", err)
	}

	// 使用GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM模式失败: %v", err)
	}

	// 检查数据长度
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("密文数据太短")
	}

	// 提取nonce和密文
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("AES解密失败: %v", err)
	}

	return string(plaintext), nil
}

// SM4加密
func encryptWithSM4(keyData *KeyData, data []byte) (string, error) {
	// SM4加密（ECB模式，无填充）
	ciphertext, err := sm4.Sm4EcbNoPaddingCipher(keyData.SymmetricKey, data, true)
	if err != nil {
		return "", fmt.Errorf("SM4加密失败: %v", err)
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// SM4解密
func decryptWithSM4(keyData *KeyData, data []byte) (string, error) {
	// SM4解密（ECB模式，无填充）
	plaintext, err := sm4.Sm4EcbNoPaddingCipher(keyData.SymmetricKey, data, false)
	if err != nil {
		return "", fmt.Errorf("SM4解密失败: %v", err)
	}

	return string(plaintext), nil
}

// 辅助函数：构建SM2公钥
func buildSM2PublicKey(pubKeyBytes []byte) (*sm2.PublicKey, error) {
	if len(pubKeyBytes) < 64 {
		return nil, errors.New("SM2公钥数据长度不足")
	}

	pubKey := &sm2.PublicKey{}
	pubKey.Curve = sm2.P256Sm2()
	pubKey.X = new(big.Int).SetBytes(pubKeyBytes[:32])
	pubKey.Y = new(big.Int).SetBytes(pubKeyBytes[32:64])

	return pubKey, nil
}

// 辅助函数：构建SM2私钥
func buildSM2PrivateKey(privKeyBytes []byte) (*sm2.PrivateKey, error) {
	if len(privKeyBytes) != 32 {
		return nil, errors.New("SM2私钥数据长度不正确")
	}

	privKey := &sm2.PrivateKey{}
	privKey.Curve = sm2.P256Sm2()
	privKey.D = new(big.Int).SetBytes(privKeyBytes)

	// 计算对应的公钥
	privKey.PublicKey.X, privKey.PublicKey.Y = privKey.Curve.ScalarBaseMult(privKeyBytes)

	return privKey, nil
}
