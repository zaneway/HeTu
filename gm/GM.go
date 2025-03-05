package gm

import (
	"encoding/asn1"
	"github.com/zaneway/cain-go/sm2"
	"github.com/zaneway/cain-go/sm4"
	"math/big"
)

type SM2Cipher struct {
	X          *big.Int `asn1:"integer"`
	Y          *big.Int `asn1:"integer"`
	Hash       []byte
	CipherText []byte
}

type SM2EnvelopedKey struct {
	SymAlgID               AlgorithmIdentifier //SGD_SM4_ECB
	Sm2cipher              SM2Cipher
	PublicKey              asn1.BitString
	Sm2EncryptedPrivateKey asn1.BitString
}

type AlgorithmIdentifier struct {
	Algorithm  asn1.ObjectIdentifier // OID，标识算法
	Parameters asn1.RawValue         `asn1:"optional"` // 可选参数
}

func DecryptDataUsePrivateKey(data []byte, privateKey *sm2.PrivateKey) ([]byte, error) {
	return sm2.Decrypt(privateKey, data, sm2.C1C3C2)
}

func DecryptDataUseSm4Key(data []byte, key []byte) (out []byte, err error) {
	return sm4.Sm4Ecb(key, data, false)
}
