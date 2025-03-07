package helper

import (
	"github.com/zaneway/cain-go/sm2"
	"github.com/zaneway/cain-go/x509"
	"math/big"
)

type KeyPair struct {
	PublicKey  *sm2.PublicKey
	PrivateKey *sm2.PrivateKey
}

// 裸公钥转换
func BuildPublicKeyUseRaw(realPublicKey []byte) *sm2.PublicKey {
	inputLen := len(realPublicKey)
	//04 || x || y ,长度小于64后无法区分
	if inputLen < 64 {
		return nil
	}
	publicKey := new(sm2.PublicKey)
	publicKey.Curve = sm2.P256Sm2()
	publicKey.X = new(big.Int).SetBytes(realPublicKey[len(realPublicKey)-64 : 32])
	publicKey.Y = new(big.Int).SetBytes(realPublicKey[len(realPublicKey)-32:])
	return publicKey
}

// 裸私钥转换
func BuildPrivateKeyUseRaw(realPrivateKey []byte) *sm2.PrivateKey {
	privateKey := new(sm2.PrivateKey)
	privateKey.D = new(big.Int).SetBytes(realPrivateKey)
	p256Sm2 := sm2.P256Sm2()
	privateKey.Curve = p256Sm2
	privateKey.PublicKey.X, privateKey.PublicKey.Y = p256Sm2.ScalarBaseMult(realPrivateKey)
	return privateKey
}

func BuildPublicKey(publicKey []byte) (*sm2.PublicKey, error) {
	return x509.ParseSm2PublicKey(publicKey)
}

func BuildPrivateKey(privateKey []byte) (*sm2.PrivateKey, error) {
	return x509.ParseSm2PrivateKey(privateKey)
}

// 密钥对构造
func BuildKeyPair(publicKey []byte, privateKey []byte) (*KeyPair, error) {
	return &KeyPair{
		PrivateKey: BuildPrivateKeyUseRaw(privateKey),
		PublicKey:  BuildPublicKeyUseRaw(publicKey),
	}, nil
}
