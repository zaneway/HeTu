package cert

import "time"

type Cert struct {
	//有效期开始/结束日期
	NotBefore, NotAfter time.Time
	//使用者 ， 颁发者
	Subject, Issue string
	//公钥算法
	Alg string
}

func ParseCertificate(cert []byte) Cert {
	var certificate Cert

	return certificate

}
