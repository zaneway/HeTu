package gm

import (
	"encoding/asn1"
	"github.com/zaneway/cain-go/sm2"
	"github.com/zaneway/cain-go/sm4"
	"github.com/zaneway/cain-go/x509"
	"math/big"
)

type SM2Cipher struct {
	X          *big.Int `asn1:"integer"`
	Y          *big.Int `asn1:"integer"`
	Hash       []byte
	CipherText []byte
}

type KeyPair struct {
	PublicKey  *sm2.PublicKey
	PrivateKey *sm2.PrivateKey
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

func BuildKeyPair(publicKey []byte, privateKey []byte) (*KeyPair, error) {
	sm2PrivateKey, err := x509.ParseSm2PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	sm2PublicKey, err := x509.ParseSm2PublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	return &KeyPair{
		PrivateKey: sm2PrivateKey,
		PublicKey:  sm2PublicKey,
	}, nil
}
