package asn1

import (
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
)

type ASN1Node struct {
	Tag      string
	Length   int
	Children []ASN1Node
	Content  string
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
	thisNode.Tag = strconv.FormatInt(int64(node.Tag), 16)
	//len
	thisNode.Length = len(node.FullBytes)
	//
	if node.IsCompound {
		thisNodeValue := node.Bytes
		fmt.Println("thisNodeValue:", hex.EncodeToString(thisNodeValue))
		for len(thisNodeValue) > 0 {
			childrenNode := ParseAsn1(thisNodeValue)
			thisNode.Children = append(thisNode.Children, childrenNode)
			//上面可能只截取了第一段结构
			thisNodeValue = thisNodeValue[childrenNode.Length:]
		}

	} else {
		thisNode.Content = hex.EncodeToString(node.Bytes)
	}
	return thisNode
}
