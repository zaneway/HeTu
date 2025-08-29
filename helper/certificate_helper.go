package helper

import (
	"fmt"

	gm "github.com/zaneway/cain-go/x509"
)

var KeyUsages = make(map[gm.KeyUsage]string)

func init() {
	KeyUsages[gm.KeyUsageDigitalSignature] = "DigitalSignature(数字签名)"
	KeyUsages[gm.KeyUsageKeyEncipherment] = "KeyEncipherment(密钥加密)"
	KeyUsages[gm.KeyUsageDataEncipherment] = "DataEncipherment(数据加密)"
	KeyUsages[gm.KeyUsageKeyAgreement] = "KeyAgreement(密钥协商)"
	KeyUsages[gm.KeyUsageCertSign] = "CertSign(证书签发)"
	KeyUsages[gm.KeyUsageCRLSign] = "CRLSign(CRL签发)"
	KeyUsages[gm.KeyUsageEncipherOnly] = "EncipherOnly(仅加密)"
	KeyUsages[gm.KeyUsageDecipherOnly] = "DecipherOnly(仅解密)"
}

func ParseCertificate(cert []byte) (*gm.Certificate, error) {
	if len(cert) == 0 {
		return nil, fmt.Errorf("证书数据为空")
	}

	certificate, err := gm.ParseCertificate(cert)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %v", err)
	}
	return certificate, nil
}

// 根据密钥用途解析具体值
func ParseKeyUsage(usage gm.KeyUsage) string {
	var result string
	for key, value := range KeyUsages {
		if key&usage != 0 {
			result += value
			result += " "
		}
	}
	return result
}
