package window

import (
	. "HeTu/helper"
	"HeTu/util"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"strings"
)

// 将ASN.1结构转换为Accordion的递归函数，并加入缩进
func buildAccordion(node ASN1Node, level int) *widget.AccordionItem {
	// 缩进根据层级来决定
	indentation := fyne.NewSize(float32(level*30), 0) // 通过level决定缩进量
	//根据节点Tag获取指定类型
	tag := getRealTag(node.Tag)

	name := tag.TypeName
	//标签名称
	value := fmt.Sprintf("%s (0x%s)", name, util.HexEncodeIntToString(node.Tag))
	// 展示标签
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

	// 		value = string(node.Content)如果没有子节点，直接返回包含内容的AccordionItem，应用缩进
	indentedContent := container.NewHBox(
		widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{}), // 占位符保持布局
		//实际的值,todo 根据实际的Tag进行类型转换
		container.NewMax(container.NewGridWrap(indentation), widget.NewLabel(node.Value)), // 缩进的Label
	)

	item := widget.NewAccordionItem(fmt.Sprintf("%s :", value), indentedContent)
	return item
}

func Asn1Structure() *fyne.Container {

	// 创建输入框，供用户输入Base64数据
	input := widget.NewMultiLineEntry()
	input.SetPlaceHolder("Please input base64/hex cert")
	input.Text = "MIICETCCAbWgAwIBAgINKl81oFaaablKOp0YTjAMBggqgRzPVQGDdQUAMGExCzAJBgNVBAYMAkNOMQ0wCwYDVQQKDARCSkNBMSUwIwYDVQQLDBxCSkNBIEFueXdyaXRlIFRydXN0IFNlcnZpY2VzMRwwGgYDVQQDDBNUcnVzdC1TaWduIFNNMiBDQS0xMB4XDTIwMDgxMzIwMTkzNFoXDTIwMTAyNDE1NTk1OVowHjELMAkGA1UEBgwCQ04xDzANBgNVBAMMBuWGr+i9rDBZMBMGByqGSM49AgEGCCqBHM9VAYItA0IABAIF97Sqq0Rv616L2PjFP3xt16QGJLmi+W8Ht+NLHiXntgUey0Nz+ZVnSUKUMzkKuGTikY3h2v7la20b6lpKo8WjgZIwgY8wCwYDVR0PBAQDAgbAMB0GA1UdDgQWBBSxiaS6z4Uguz3MepS2zblkuAF/LTAfBgNVHSMEGDAWgBTMZyRCGsP4rSes0vLlhIEf6cUvrjBABgNVHSAEOTA3MDUGCSqBHIbvMgICAjAoMCYGCCsGAQUFBwIBFhpodHRwOi8vd3d3LmJqY2Eub3JnLmNuL2NwczAMBggqgRzPVQGDdQUAA0gAMEUCIG6n6PG0BOK1EdFcvetQlC+9QhpsTuTui2wkeqWiPKYWAiEAvqR8Z+tSiYR5DIs7SyHJPWZ+sa8brtQL/1jURvHGxU8="

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

		// 解析ASN.1数据并构建Accordion
		rootNode := ParseAsn1(decodedData)
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
