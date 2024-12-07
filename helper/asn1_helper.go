package helper

import (
	"HeTu/util"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"github.com/zaneway/cain-go/sm2"
	"log"
	"math/big"
	"time"
)

// 定义tag和名称的映射关系
var TagToName = map[int]string{
	0:  "",
	1:  "BOOLEAN",
	2:  "INTEGER",
	3:  "BIT STRING",
	4:  "OCTET STRING",
	5:  "NULL",
	6:  "OBJECT IDENTIFIER",
	8:  "EXTERNAL",
	10: "ENUMERATED",
	12: "UTF8String",
	16: "SEQUENCE",
	17: "SET",
	19: "PrintableString",
	20: "T61String",
	22: "IA5String",
	23: "UTCTime",
	24: "GeneralizedTime",
	27: "TagGeneralString",
	30: "TagBMPString",
}

type ASN1Content struct {
	TypeName string
	RealType interface{}
}

var ClassToNum = map[int]int{
	0: 0,
	1: 64,  //0x40,第7位为1
	2: 128, //0x80,第8位为1
	3: 192, //0xc0,第8位和第7位为1
}

type ASN1Node struct {
	//this Tag is real Number in asn1
	Tag, Class, Length int
	Children           []*ASN1Node
	Value              string
	Content, FullBytes []byte
}

func ParseAsn1(data []byte) ASN1Node {
	var thisNode ASN1Node
	var node asn1.RawValue

	//将输入转为base64
	_, err := asn1.Unmarshal(data, &node)
	if err != nil {
		log.Fatalln(err)
	}
	//tag to hex
	thisNode.Tag = node.Tag

	if node.IsCompound {
		//如果是结构类型,则第6位为1
		thisNode.Tag = node.Tag + 32
	}
	thisNode.Tag += ClassToNum[node.Class]
	thisNode.Class = node.Class
	//len
	thisNode.Length = len(node.FullBytes)
	//
	if node.IsCompound {
		thisNodeValue := node.Bytes
		for len(thisNodeValue) > 0 {
			childrenNode := ParseAsn1(thisNodeValue)
			thisNode.Children = append(thisNode.Children, &childrenNode)
			//上面可能只截取了第一段结构
			thisNodeValue = thisNodeValue[childrenNode.Length:]
		}

	} else {
		thisNode.Content = node.Bytes
	}
	thisNode.FullBytes = node.FullBytes
	thisNode.Value = buildAsn1Value(thisNode)
	return thisNode
}

func buildAsn1Value(node ASN1Node) (data string) {
	data = hex.EncodeToString(node.Content)
	switch node.Tag {
	//big int
	case 2:
		bigInt, _ := parseBigInt(node.Content)
		data = bigInt.String()
		break
	//bit string
	case 3:
		ret, _ := parseBitString(node.Content)
		data = hex.EncodeToString(ret.Bytes)
		r, s, err := sm2.SignDataToSignDigit(ret.Bytes)
		if err == nil {
			data = fmt.Sprintf("%s \n%s", r, s)
		}
		break
	//OID
	case 6:
		identifier := asn1.ObjectIdentifier{}
		asn1.Unmarshal(node.FullBytes, &identifier)
		//todo 根据OID解析出对应算法
		data = identifier.String()
		switch identifier {
		//RSA
		case asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}:
			data = "RSA"
			break
		case asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}:
			data = "ECDSA"
			break
		// SM2
		case asn1.ObjectIdentifier{1, 2, 156, 10197, 1, 301}:
			data = "SM2"
			break

		}
		break
	//UTF8String
	case 12:
		data = string(node.Content)
		break
	//UTC time
	case 23:
		s := string(node.Content)
		parse, _ := time.Parse(util.FormatStr, s)
		data = parse.Format(util.DateTime)
		break
	}
	return
}

func parseBigInt(bytes []byte) (*big.Int, error) {
	if err := checkInteger(bytes); err != nil {
		return nil, err
	}
	ret := new(big.Int)
	if len(bytes) > 0 && bytes[0]&0x80 == 0x80 {
		// This is a negative number.
		notBytes := make([]byte, len(bytes))
		for i := range notBytes {
			notBytes[i] = ^bytes[i]
		}
		ret.SetBytes(notBytes)
		ret.Add(ret, bigOne)
		ret.Neg(ret)
		return ret, nil
	}
	ret.SetBytes(bytes)
	return ret, nil
}

var bigOne = big.NewInt(1)

func checkInteger(bytes []byte) error {
	if len(bytes) == 0 {
		return asn1.StructuralError{"empty integer"}
	}
	if len(bytes) == 1 {
		return nil
	}
	if (bytes[0] == 0 && bytes[1]&0x80 == 0) || (bytes[0] == 0xff && bytes[1]&0x80 == 0x80) {
		return asn1.StructuralError{"integer not minimally-encoded"}
	}
	return nil
}

// parseBitString parses an ASN.1 bit string from the given byte slice and returns it.
func parseBitString(bytes []byte) (ret asn1.BitString, err error) {
	if len(bytes) == 0 {
		err = asn1.SyntaxError{"zero length BIT STRING"}
		return
	}
	paddingBits := int(bytes[0])
	if paddingBits > 7 ||
		len(bytes) == 1 && paddingBits > 0 ||
		bytes[len(bytes)-1]&((1<<bytes[0])-1) != 0 {
		err = asn1.SyntaxError{"invalid padding bits in BIT STRING"}
		return
	}
	ret.BitLength = (len(bytes)-1)*8 - paddingBits
	ret.Bytes = bytes[1:]
	return
}
