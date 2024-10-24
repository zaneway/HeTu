package helper

import (
	"HeTu/util"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"log"
	"reflect"
	"time"
)

// 定义tag和名称的映射关系
//var TagToName = map[int]string{
//	0:  "",
//	1:  "BOOLEAN",
//	2:  "INTEGER",
//	3:  "BIT STRING",
//	4:  "OCTET STRING",
//	5:  "NULL",
//	6:  "OBJECT IDENTIFIER",
//	8:  "EXTERNAL",
//	10: "ENUMERATED",
//	12: "UTF8String",
//	16: "SEQUENCE",
//	17: "SET",
//	19: "PrintableString",
//	20: "T61String",
//	22: "IA5String",
//	23: "UTCTime",
//	24: "GeneralizedTime",
//	27: "TagGeneralString",
//	30: "TagBMPString",
//}

type ASN1Content struct {
	TypeName string
	RealType interface{}
}

var TagToName = map[int]ASN1Content{
	0:  {},
	1:  {"BOOLEAN", reflect.TypeOf(true)},
	2:  {"INTEGER", reflect.TypeOf(1)},
	3:  {"BIT STRING", asn1.BitString{}},
	4:  {"OCTET STRING", reflect.TypeOf(asn1.TagOctetString)},
	5:  {"NULL", reflect.TypeOf(nil)},
	6:  {"OBJECT IDENTIFIER", asn1.ObjectIdentifier{}},
	10: {"ENUMERATED", reflect.TypeOf(asn1.Enumerated(1))},
	12: {"UTF8String", reflect.TypeOf("")},
	16: {"SEQUENCE", nil},
	17: {"SET", nil},
	19: {"PrintableString", reflect.TypeOf("")},
	20: {"T61String", reflect.TypeOf("")},
	22: {"IA5String", reflect.TypeOf("")},
	23: {"UTCTime", reflect.TypeOf(time.DateTime)},
	24: {"GeneralizedTime", reflect.TypeOf(time.DateTime)},
	27: {"TagGeneralString", reflect.TypeOf("")},
	30: {"TagBMPString", reflect.TypeOf("")},
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
	SHA256             string
	Content, FullBytes []byte
}

func ParseAsn1(data []byte, resultMap map[string]ASN1Node) ASN1Node {
	var thisNode ASN1Node
	var node asn1.RawValue

	//将输入转为base64
	_, err := asn1.Unmarshal(data, &node)
	if err != nil {
		log.Fatalln(err)
	}
	//tag to hex
	//todo 后期转为asn1的tag类型描述,并对应string\oid等转换
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
		fmt.Println("thisNodeValue:", hex.EncodeToString(thisNodeValue))
		for len(thisNodeValue) > 0 {
			childrenNode := ParseAsn1(thisNodeValue, resultMap)
			thisNode.Children = append(thisNode.Children, &childrenNode)
			//上面可能只截取了第一段结构
			thisNodeValue = thisNodeValue[childrenNode.Length:]
		}

	} else {
		thisNode.Content = node.Bytes
	}
	thisNode.FullBytes = node.FullBytes
	//节点hash
	digest := sha256.New()
	digest.Write(util.Serialize(thisNode))
	hashBytes := digest.Sum(nil)
	thisNode.SHA256 = hex.EncodeToString(hashBytes)
	resultMap[thisNode.SHA256] = thisNode
	return thisNode
}
