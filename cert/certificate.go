package cert

import (
	"encoding/asn1"
	"fmt"
	gm "github.com/tjfoc/gmsm/x509"
	"time"
)

type Cert struct {
	//有效期开始/结束日期
	NotBefore, NotAfter time.Time
	//使用者 ， 颁发者
	Subject, Issue string
	//公钥算法
	Alg string
}

func ParseCertificate(cert []byte) (result Cert, error error) {
	var certificate gm.Certificate
	_, err := asn1.Unmarshal(cert, &certificate)
	if err != nil {
		fmt.Println(err)
		return result, err
	}
	result.Subject = certificate.Subject.String()
	result.Issue = certificate.Issuer.String()
	result.NotBefore = certificate.NotBefore
	result.NotAfter = certificate.NotAfter
	result.Alg = certificate.SignatureAlgorithm.String()
	return result, nil

}
