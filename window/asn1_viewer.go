package window

import (
	. "HeTu/helper"
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
	value := TagToName[node.Tag]
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
		// 包装子节点为带缩进的Container
		indentedChildAccordion := container.NewHBox(
			widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{}), // 占位符保持布局
			container.NewMax(container.NewGridWrap(indentation), childAccordion),  // 缩进的Accordion
		)
		//return widget.NewAccordionItem(fmt.Sprintf("%s :", value), container.NewVBox(content, indentedChildAccordion))
		return widget.NewAccordionItem(fmt.Sprintf("%s :", value), container.NewVBox(indentedChildAccordion))
	}

	// 如果没有子节点，直接返回包含内容的AccordionItem，应用缩进
	indentedContent := container.NewHBox(
		widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{}),               // 占位符保持布局
		container.NewMax(container.NewGridWrap(indentation), widget.NewLabel(node.Content)), // 缩进的Label
	)

	return widget.NewAccordionItem(fmt.Sprintf("%s :", value), indentedContent)
}

func Asn1Structure() *fyne.Container {

	// 创建输入框，供用户输入Base64数据
	input := widget.NewMultiLineEntry()
	input.SetPlaceHolder("Please input base64/hex cert")
	//input.Text = "MIICETCCAbWgAwIBAgINKl81oFaaablKOp0YTjAMBggqgRzPVQGDdQUAMGExCzAJBgNVBAYMAkNOMQ0wCwYDVQQKDARCSkNBMSUwIwYDVQQLDBxCSkNBIEFueXdyaXRlIFRydXN0IFNlcnZpY2VzMRwwGgYDVQQDDBNUcnVzdC1TaWduIFNNMiBDQS0xMB4XDTIwMDgxMzIwMTkzNFoXDTIwMTAyNDE1NTk1OVowHjELMAkGA1UEBgwCQ04xDzANBgNVBAMMBuWGr+i9rDBZMBMGByqGSM49AgEGCCqBHM9VAYItA0IABAIF97Sqq0Rv616L2PjFP3xt16QGJLmi+W8Ht+NLHiXntgUey0Nz+ZVnSUKUMzkKuGTikY3h2v7la20b6lpKo8WjgZIwgY8wCwYDVR0PBAQDAgbAMB0GA1UdDgQWBBSxiaS6z4Uguz3MepS2zblkuAF/LTAfBgNVHSMEGDAWgBTMZyRCGsP4rSes0vLlhIEf6cUvrjBABgNVHSAEOTA3MDUGCSqBHIbvMgICAjAoMCYGCCsGAQUFBwIBFhpodHRwOi8vd3d3LmJqY2Eub3JnLmNuL2NwczAMBggqgRzPVQGDdQUAA0gAMEUCIG6n6PG0BOK1EdFcvetQlC+9QhpsTuTui2wkeqWiPKYWAiEAvqR8Z+tSiYR5DIs7SyHJPWZ+sa8brtQL/1jURvHGxU8="

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
	cancelButten := buildButton("清除", theme.CancelIcon(), func() {
		input.Text = ""
		input.Refresh()
	})
	// 布局
	allButton := container.New(layout.NewGridLayout(2), confirmButton, cancelButten)
	vbox := container.NewVBox(input, allButton)

	return container.NewBorder(vbox, nil, nil, nil, accordion)

}
