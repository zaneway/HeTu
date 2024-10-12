package helper

import (
	"HeTu/util"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"log"
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

//asn1编码中，高位两位（第八位、第七位）的取值范围
var ClassToNum = map[int]int{
	0: 0,
	1: 64,  //0x40,第7位为1
	2: 128, //0x80,第8位为1
	3: 192, //0xc0,第8位和第7位为1
}

//asn1结构，tlv结构。Children=子节点，Content=实际报文，Sha256报文摘要，唯一标识
type ASN1Node struct {
	//this Tag is real Number in asn1
	Tag, Class, Length int
	Children           []*ASN1Node
	Content, SHA256    string
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
		thisNode.Content = hex.EncodeToString(node.Bytes)
	}

	//节点hash
	digest := sha256.New()
	digest.Write(util.Serialize(thisNode))
	hashBytes := digest.Sum(nil)
	thisNode.SHA256 = hex.EncodeToString(hashBytes)
	resultMap[thisNode.SHA256] = thisNode
	return thisNode
}
