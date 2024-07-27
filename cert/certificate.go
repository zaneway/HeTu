package cert

import (
	"fmt"
	gm "github.com/tjfoc/gmsm/x509"
)

func ParseCertificate(cert []byte) (*gm.Certificate, error) {
	certificate, err := gm.ParseCertificate(cert)
	if err != nil {
		fmt.Println(err)
	}
	return certificate, nil
}
