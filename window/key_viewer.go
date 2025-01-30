package window

import (
	"HeTu/security"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/zaneway/cain-go/sm2"
)

func KeyStructure() *fyne.Container {
	//算法\长度
	newSelect := widget.NewSelect(append(security.ALL_ASYM_KEYS, security.ALL_SYM_KEYS...), func(alg string) {
		switch alg {
		case security.SM2_256:
			priKey, _ := sm2.GenerateKey(nil)
			println("Pub:", hex.EncodeToString(append(priKey.PublicKey.X.Bytes(), priKey.PublicKey.Y.Bytes()...)))
			println("Pri:", hex.EncodeToString(priKey.D.Bytes()))
			break
		case security.RSA_1024:
			key, _ := rsa.GenerateKey(nil, 1024)
			println("Pub:", hex.EncodeToString(key.PublicKey.N.Bytes()))
			println("Pri:", hex.EncodeToString(key.D.Bytes()))
			break
		case security.AES_128:
			aesKey := make([]byte, 16)
			rand.Read(aesKey)
			print(" AES:", hex.EncodeToString(aesKey))
			break
		}
	})
	structure := container.NewVBox()
	structure.Add(newSelect)
	return structure
}
