package security

const (
	RSA_1024 = "RSA_1024"
	RSA_2048 = "RSA_2048"
	RSA_4096 = "RSA_4096"

	SM2_256 = "SM2_256"

	AES_128 = "AES_128"
	AES_256 = "AES_256"
	AES_384 = "AES_384"
	AES_512 = "AES_512"

	SM4_128 = "SM4_128"
)

// 算法_长度
// 对称算法
var ALL_SYM_KEYS = []string{AES_128, AES_256, AES_384, AES_512, SM4_128}

// 非对称算法
var ALL_ASYM_KEYS = []string{SM2_256, RSA_1024, RSA_2048, RSA_4096}
