package window

import (
	"HeTu/gm"
	"encoding/asn1"
	"github.com/zaneway/cain-go/x509"
)

// 解析信封
func ParseSM2EnvelopedKey(data []byte) (*gm.SM2EnvelopedKey, error) {
	var sm2EnvelopedKey gm.SM2EnvelopedKey
	_, err := asn1.Unmarshal(data, &sm2EnvelopedKey)
	return &sm2EnvelopedKey, err
}

// 使用私钥解密数字信封
func DecryptSM2EnvelopedKey(data []byte, privateKey []byte) ([]byte, error) {
	//将SM2私钥转换为制定对象
	sm2PrivateKey, err := x509.ParseSm2PrivateKey(privateKey)
	//解析SM2EnvelopedKey
	sm2EnvelopedKey, err := ParseSM2EnvelopedKey(data)
	if err != nil {
		return nil, err
	}
	sm2Cipher, _ := asn1.Marshal(sm2EnvelopedKey.Sm2cipher)
	//解密信封,得到对称密钥
	sm4Key, err := gm.DecryptDataUsePrivateKey(sm2Cipher, sm2PrivateKey)
	if err != nil {
		return nil, err
	}
	//对称密钥解密私钥
	return gm.DecryptDataUseSm4Key(sm2EnvelopedKey.Sm2EncryptedPrivateKey.Bytes, sm4Key)
}
