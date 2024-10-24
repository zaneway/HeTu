package window

import (
	. "HeTu/helper"
	"HeTu/util"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/zaneway/cain-go/sm2"
	"math/big"
	"strings"
	"time"
)

// 将ASN.1结构转换为Accordion的递归函数，并加入缩进
func buildAccordion(node ASN1Node, level int) *widget.AccordionItem {
	// 缩进根据层级来决定
	indentation := fyne.NewSize(float32(level*30), 0) // 通过level决定缩进量
	tag := getRealTag(node.Tag)

	name := tag.TypeName
	//标签名称
	value := fmt.Sprintf("%s (0x%s)", name, util.HexEncodeIntToString(node.Tag))
	// 节点的内容展示
	content := widget.NewLabel(fmt.Sprintf("%s :", value))
	content.Resize(fyne.NewSize(600, content.MinSize().Height))

	// 如果有子节点，递归生成子节点的Accordion
	var childrenAccordionItems []*widget.AccordionItem
	for _, child := range node.Children {
		childrenAccordionItems = append(childrenAccordionItems, buildAccordion(*child, level+1))
	}

	// 如果有子节点，将这些子节点放入到容器中，并应用缩进
	if len(childrenAccordionItems) > 0 {
		childAccordion := widget.NewAccordion(childrenAccordionItems...)
		//禁止折叠
		//childAccordion.MultiOpen = true
		// 包装子节点为带缩进的Container
		indentedChildAccordion := container.NewHBox(
			widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{}), // 占位符保持布局
			container.NewMax(container.NewGridWrap(indentation), childAccordion),  // 缩进的Accordion
		)
		//return widget.NewAccordionItem(fmt.Sprintf("%s :", value), container.NewVBox(content, indentedChildAccordion))
		return widget.NewAccordionItem(fmt.Sprintf("%s :", value), container.NewVBox(indentedChildAccordion))
	}
	data := hex.EncodeToString(node.Content)
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
		data = identifier.String()
		break
		//UTF8String
	case 12:
		data = string(node.Content)
		break
		//UTC time
	case 23:
		s := string(node.Content)
		formatStr := "060102150405Z0700"
		parse, _ := time.Parse(formatStr, s)
		data = parse.Format(DateTime)
		break
	}

	// 		value = string(node.Content)如果没有子节点，直接返回包含内容的AccordionItem，应用缩进
	indentedContent := container.NewHBox(
		widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{}), // 占位符保持布局
		//实际的值,todo 根据实际的Tag进行类型转换
		container.NewMax(container.NewGridWrap(indentation), widget.NewLabel(data)), // 缩进的Label
	)

	item := widget.NewAccordionItem(fmt.Sprintf("%s :", value), indentedContent)
	return item
}

func Asn1Structure() *fyne.Container {

	// 创建输入框，供用户输入Base64数据
	input := widget.NewMultiLineEntry()
	input.SetPlaceHolder("Please input base64/hex cert")
	input.Text = "MIICETCCAbWgAwIBAgINKl81oFaaablKOp0YTjAMBggqgRzPVQGDdQUAMGExCzAJBgNVBAYMAkNOMQ0wCwYDVQQKDARCSkNBMSUwIwYDVQQLDBxCSkNBIEFueXdyaXRlIFRydXN0IFNlcnZpY2VzMRwwGgYDVQQDDBNUcnVzdC1TaWduIFNNMiBDQS0xMB4XDTIwMDgxMzIwMTkzNFoXDTIwMTAyNDE1NTk1OVowHjELMAkGA1UEBgwCQ04xDzANBgNVBAMMBuWGr+i9rDBZMBMGByqGSM49AgEGCCqBHM9VAYItA0IABAIF97Sqq0Rv616L2PjFP3xt16QGJLmi+W8Ht+NLHiXntgUey0Nz+ZVnSUKUMzkKuGTikY3h2v7la20b6lpKo8WjgZIwgY8wCwYDVR0PBAQDAgbAMB0GA1UdDgQWBBSxiaS6z4Uguz3MepS2zblkuAF/LTAfBgNVHSMEGDAWgBTMZyRCGsP4rSes0vLlhIEf6cUvrjBABgNVHSAEOTA3MDUGCSqBHIbvMgICAjAoMCYGCCsGAQUFBwIBFhpodHRwOi8vd3d3LmJqY2Eub3JnLmNuL2NwczAMBggqgRzPVQGDdQUAA0gAMEUCIG6n6PG0BOK1EdFcvetQlC+9QhpsTuTui2wkeqWiPKYWAiEAvqR8Z+tSiYR5DIs7SyHJPWZ+sa8brtQL/1jURvHGxU8="

	// Tree数据存储
	treeData := make(map[string]ASN1Node)

	// 创建Accordion组件
	accordion := widget.NewAccordion()

	// 解析按钮
	confirmButton := widget.NewButtonWithIcon("确认", theme.ConfirmIcon(), func() {
		inputData := strings.TrimSpace(input.Text)
		decodedData, err := base64.StdEncoding.DecodeString(inputData)
		if err != nil {
			decodedData, err = hex.DecodeString(inputData)
			if err != nil {
				fyne.LogError("解析请求错误", err)
				return
			}
		}
		treeData = make(map[string]ASN1Node)

		// 解析ASN.1数据并构建Accordion
		rootNode := ParseAsn1(decodedData, treeData)
		rootAccordionItem := buildAccordion(rootNode, 0) // 初始层级为0
		//清除上次数据
		accordion.RemoveIndex(0)
		accordion.Append(rootAccordionItem)
	})
	//清除按钮
	cancelButton := buildButton("清除", theme.CancelIcon(), func() {
		input.Text = ""
		input.Refresh()
	})
	// 布局
	allButton := container.New(layout.NewGridLayout(2), confirmButton, cancelButton)
	vbox := container.NewVBox(input, allButton)

	return container.NewBorder(vbox, nil, nil, nil, accordion)

}

func getRealTag(tag int) ASN1Content {
	prefix := ""
	//32 = 0x20, ASN1中小于0x20的都是通用简单类型

	//0x20 到 0x40 通用,结构类型
	if 32 <= tag && tag < 64 {
		//prefix = "Universal Structure "
		tag -= 32
	} else if 64 <= tag && tag < 96 {
		prefix = "Application Simple "
		tag -= 64
	} else if 96 <= tag && tag < 128 {
		prefix = "Application Structure "
		tag -= 96
	} else if 128 <= tag && tag < 160 {
		prefix = "Context Specific Simple "
		tag -= 128
	} else if 160 <= tag && tag < 192 {
		prefix = "Context Specific Structure "
		tag -= 160
	} else if 192 <= tag && tag < 224 {
		prefix = "Private Simple "
		tag -= 192
	} else if 224 <= tag && tag < 256 {
		prefix = "Private Structure "
		tag -= 224
	}
	content := TagToName[tag]
	content.TypeName = prefix + content.TypeName
	return content
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
