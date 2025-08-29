package helper

import (
	"crypto/x509/pkix"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"time"

	"github.com/zaneway/cain-go/x509"
)

// CRLInfo CRL信息结构体
type CRLInfo struct {
	Issuer             string
	ThisUpdate         time.Time
	NextUpdate         time.Time
	RevokedCerts       []RevokedCertificate
	TotalRevoked       int
	SignatureAlgorithm string
}

// RevokedCertificate 被吊销的证书信息
type RevokedCertificate struct {
	SerialNumber   string
	RevocationTime time.Time
	Reason         string
}

// ParseCRLFromFile 从文件路径解析CRL
func ParseCRLFromFile(filePath string) (*CRLInfo, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取CRL文件失败: %v", err)
	}

	return ParseCRL(data)
}

// ParseCRL 解析CRL数据
func ParseCRL(data []byte) (*CRLInfo, error) {
	crl, err := x509.ParseCRL(data)
	if err != nil {
		return nil, fmt.Errorf("解析CRL失败: %v", err)
	}

	crlInfo := &CRLInfo{
		Issuer:             crl.TBSCertList.Issuer.String(),
		ThisUpdate:         crl.TBSCertList.ThisUpdate,
		NextUpdate:         crl.TBSCertList.NextUpdate,
		TotalRevoked:       len(crl.TBSCertList.RevokedCertificates),
		SignatureAlgorithm: formatSignatureAlgorithm(crl.SignatureAlgorithm),
	}

	// 解析被吊销的证书列表
	for _, revoked := range crl.TBSCertList.RevokedCertificates {
		revokedCert := RevokedCertificate{
			SerialNumber:   hex.EncodeToString(revoked.SerialNumber.Bytes()),
			RevocationTime: revoked.RevocationTime,
			Reason:         parseRevocationReason(revoked.Extensions),
		}
		crlInfo.RevokedCerts = append(crlInfo.RevokedCerts, revokedCert)
	}

	return crlInfo, nil
}

// CheckCertificateRevocation 检查证书序列号是否在CRL中被吊销
func CheckCertificateRevocation(crlInfo *CRLInfo, serialNumber string) (bool, *RevokedCertificate) {
	// 标准化序列号格式（移除空格、冒号等分隔符，转换为小写）
	normalizedInput := normalizeSerialNumber(serialNumber)

	for _, revoked := range crlInfo.RevokedCerts {
		normalizedRevoked := normalizeSerialNumber(revoked.SerialNumber)
		if normalizedRevoked == normalizedInput {
			return true, &revoked
		}
	}

	return false, nil
}

// CheckCertificateRevocationFromFile 从文件检查证书是否被吊销
func CheckCertificateRevocationFromFile(filePath, serialNumber string) (bool, *RevokedCertificate, error) {
	crlInfo, err := ParseCRLFromFile(filePath)
	if err != nil {
		return false, nil, err
	}

	isRevoked, revokedCert := CheckCertificateRevocation(crlInfo, serialNumber)
	return isRevoked, revokedCert, nil
}

// normalizeSerialNumber 标准化序列号格式
func normalizeSerialNumber(serialNumber string) string {
	normalized := ""
	for _, char := range serialNumber {
		if (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F') {
			if char >= 'A' && char <= 'F' {
				normalized += string(char - 'A' + 'a') // 转为小写
			} else {
				normalized += string(char)
			}
		}
	}
	return normalized
}

// parseRevocationReason 解析吊销原因
func parseRevocationReason(extensions []pkix.Extension) string {
	for _, ext := range extensions {
		// CRL Reason Code OID: 2.5.29.21
		if ext.Id.Equal([]int{2, 5, 29, 21}) {
			if len(ext.Value) > 0 {
				reason := int(ext.Value[0])
				return getRevocationReasonText(reason)
			}
		}
	}
	return "未指定"
}

// getRevocationReasonText 获取吊销原因文本
func getRevocationReasonText(reason int) string {
	reasons := map[int]string{
		0:  "未指定",
		1:  "密钥泄露",
		2:  "CA泄露",
		3:  "从属关系变更",
		4:  "被取代",
		5:  "停止运营",
		6:  "证书暂停",
		8:  "移除从属关系",
		9:  "特权撤销",
		10: "AA泄露",
	}

	if text, exists := reasons[reason]; exists {
		return text
	}
	return fmt.Sprintf("未知原因(%d)", reason)
}

// formatSignatureAlgorithm 格式化签名算法
func formatSignatureAlgorithm(alg pkix.AlgorithmIdentifier) string {
	// 根据OID返回算法名称
	oidToName := map[string]string{
		"1.2.840.113549.1.1.1":  "RSA",
		"1.2.840.113549.1.1.5":  "SHA1-RSA",
		"1.2.840.113549.1.1.11": "SHA256-RSA",
		"1.2.840.113549.1.1.12": "SHA384-RSA",
		"1.2.840.113549.1.1.13": "SHA512-RSA",
		"1.2.840.10045.4.1":     "SHA1-ECDSA",
		"1.2.840.10045.4.3.2":   "SHA256-ECDSA",
		"1.2.840.10045.4.3.3":   "SHA384-ECDSA",
		"1.2.840.10045.4.3.4":   "SHA512-ECDSA",
		"1.2.156.10197.1.501":   "SM3-SM2",
	}

	oidStr := alg.Algorithm.String()
	if name, exists := oidToName[oidStr]; exists {
		return name
	}
	return oidStr
}

// ConvertSerialNumberToBigInt 将序列号字符串转换为大整数（用于比较）
func ConvertSerialNumberToBigInt(serialNumber string) (*big.Int, error) {
	normalized := normalizeSerialNumber(serialNumber)
	bigInt := new(big.Int)
	_, ok := bigInt.SetString(normalized, 16)
	if !ok {
		return nil, fmt.Errorf("无效的序列号格式: %s", serialNumber)
	}
	return bigInt, nil
}
