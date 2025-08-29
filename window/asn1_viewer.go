package window

import (
	. "HeTu/helper"
	"HeTu/util"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// 将ASN.1结构转换为Accordion的递归函数，并加入缩进
func buildAccordion(node ASN1Node, level int) *widget.AccordionItem {
	// 防止过深的嵌套
	if level > 15 {
		return widget.NewAccordionItem("⚠️ 嵌套过深...", widget.NewLabel("为了性能考虑，停止在第15层解析"))
	}

	// 根据节点Tag获取指定类型
	name := getRealTag(node.Tag)

	// 标签名称，添加更多信息和状态图标
	var value string
	var statusIcon string
	if node.Error != "" {
		statusIcon = "❌"
		value = fmt.Sprintf("%s %s (Tag:0x%s) - %s", statusIcon, name, util.HexEncodeIntToString(node.Tag), node.Error)
	} else {
		if len(node.Children) > 0 {
			statusIcon = "📁" // 复合类型
		} else {
			statusIcon = "📄" // 简单类型
		}

		// 根节点使用更突出的显示
		if level == 0 {
			value = fmt.Sprintf("🌟 根节点: %s %s (Tag:0x%s) [%d bytes] - 深度:%d", statusIcon, name, util.HexEncodeIntToString(node.Tag), node.Length, node.Depth)
		} else {
			value = fmt.Sprintf("%s %s (Tag:0x%s) [%d bytes]", statusIcon, name, util.HexEncodeIntToString(node.Tag), node.Length)
		}
	}

	// 如果有子节点，递归生成子节点的Accordion
	if len(node.Children) > 0 {
		var childrenAccordionItems []*widget.AccordionItem

		// 动态限制显示的子节点数量，根据层级调整
		maxDisplay := 100 - level*10 // 越深层级，显示越少子节点
		if maxDisplay < 10 {
			maxDisplay = 10
		}

		for i, child := range node.Children {
			if i >= maxDisplay {
				remaining := len(node.Children) - maxDisplay
				truncateItem := widget.NewAccordionItem(
					fmt.Sprintf("⚠️ 已截断 - 还有 %d 个子节点", remaining),
					widget.NewRichTextFromMarkdown(fmt.Sprintf("为了性能考虑，在第%d层只显示前%d个子节点\n\n总子节点数: %d\n已显示: %d\n已隐藏: %d",
						level+1, maxDisplay, len(node.Children), maxDisplay, remaining)))
				childrenAccordionItems = append(childrenAccordionItems, truncateItem)
				break
			}
			childrenAccordionItems = append(childrenAccordionItems, buildAccordion(*child, level+1))
		}

		childAccordion := widget.NewAccordion(childrenAccordionItems...)

		// 根节点的子节点容器特殊处理
		if level == 0 {
			// 根节点简化显示
			content := container.NewVBox(
				container.NewPadded(childAccordion),
			)

			return widget.NewAccordionItem(value, content)
		} else {
			// 非根节点的正常显示
			//statsLabel := widget.NewLabel(fmt.Sprintf("子节点数量: %d", len(node.Children)))
			//statsLabel.TextStyle = fyne.TextStyle{Italic: true}

			content := container.NewVBox(
				//statsLabel,
				container.NewPadded(childAccordion),
			)

			return widget.NewAccordionItem(value, content)
		}
	}

	// 如果没有子节点，直接返回包含内容的AccordionItem
	// 创建详细的内容显示
	contentText := node.Value
	maxDisplayLength := 1500 // 增加显示长度
	if len(contentText) > maxDisplayLength {
		contentText = contentText[:maxDisplayLength] + fmt.Sprintf("\n\n... 已截断 (总长度: %d 字符)", len(node.Value))
	}

	// 构建详细信息
	var infoBuilder strings.Builder
	if level == 0 {
		// 根节点显示更详细的信息
		infoBuilder.WriteString("🌟 **根节点详细信息**\n\n")
		infoBuilder.WriteString(fmt.Sprintf("**类型**: %s\n", name))
		infoBuilder.WriteString(fmt.Sprintf("**标签**: 0x%s (十进制: %d)\n", util.HexEncodeIntToString(node.Tag), node.Tag))
		infoBuilder.WriteString(fmt.Sprintf("**类别**: %d\n", node.Class))
		infoBuilder.WriteString(fmt.Sprintf("**数据长度**: %d bytes\n", len(node.Content)))
		infoBuilder.WriteString(fmt.Sprintf("**节点深度**: %d\n", node.Depth))
		if node.Error != "" {
			infoBuilder.WriteString(fmt.Sprintf("**错误**: %s\n", node.Error))
		}
		infoBuilder.WriteString("\n---\n\n**数据内容**:\n")
		infoBuilder.WriteString(contentText)
	} else {
		// 子节点只显示内容，不显示技术参数
		infoBuilder.WriteString(contentText)
	}

	// 创建可复制的内容显示
	contentEntry := widget.NewMultiLineEntry()
	contentEntry.SetText(infoBuilder.String())
	contentEntry.Wrapping = fyne.TextWrapWord

	// 根据层级调整显示大小
	if level == 0 {
		// 根节点使用更大的显示区域
		contentEntry.Resize(fyne.NewSize(600, 200))
	} else {
		// 子节点使用标准大小
		contentEntry.Resize(fyne.NewSize(500, 120))
	}

	// 添加复制按钮
	copyBtn := widget.NewButtonWithIcon("📋 复制内容", theme.ContentCopyIcon(), func() {
		// 使用系统剪贴板复制内容
		clipboard := fyne.CurrentApp().Driver().AllWindows()[0].Clipboard()
		clipboard.SetContent(contentEntry.Text)

		// 显示复制成功的提示
		//dialog.ShowInformation("复制成功", "内容已复制到剪贴板", fyne.CurrentApp().Driver().AllWindows()[0])
	})

	// 根节点添加额外的操作按钮
	if level == 0 {
		// 导出按钮
		exportBtn := widget.NewButtonWithIcon("💾 导出数据", theme.DocumentSaveIcon(), func() {
			// 这里可以添加导出功能
			dialog.ShowInformation("导出功能", "导出功能待实现", fyne.CurrentApp().Driver().AllWindows()[0])
		})

		buttonContainer := container.NewHBox(copyBtn, exportBtn, layout.NewSpacer())

		content := container.NewVBox(
			contentEntry,
			widget.NewSeparator(),
			buttonContainer,
		)

		return widget.NewAccordionItem(value, content)
	} else {
		// 子节点只有复制按钮
		buttonContainer := container.NewHBox(copyBtn, layout.NewSpacer())

		content := container.NewVBox(
			contentEntry,
			buttonContainer,
		)

		return widget.NewAccordionItem(value, content)
	}
}

func Asn1Structure(input *widget.Entry) *fyne.Container {
	// 为公共输入框设置最优配置
	input.Wrapping = fyne.TextWrapWord

	// 创建状态显示标签
	statusLabel := widget.NewLabel("准备解析ASN.1数据...")
	statusLabel.TextStyle = fyne.TextStyle{Italic: true}

	// 创建进度条
	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	// 创建统计信息显示区域
	statsContainer := container.NewVBox()
	statsContainer.Hide()

	//创建Accordion组件
	accordion := widget.NewAccordion()
	var rootAccordionItem *widget.AccordionItem
	// 异步解析按钮，添加进度提示
	confirmButton := widget.NewButtonWithIcon("🔍 解析ASN.1", theme.ConfirmIcon(), func() {
		inputData := strings.TrimSpace(input.Text)
		if inputData == "" {
			dialog.ShowError(fmt.Errorf("请输入ASN.1数据"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		// 更新状态
		statusLabel.SetText("正在预处理数据...")
		progressBar.Show()
		progressBar.SetValue(0.1)

		// 异步处理以避免UI阻塞
		go func() {
			time.Sleep(time.Millisecond * 100) // 给UI时间更新

			// 预处理检查
			if len(inputData) > 5*1024*1024 {
				dialog.ShowError(fmt.Errorf("输入数据过大（%d 字符）", len(inputData)), fyne.CurrentApp().Driver().AllWindows()[0])
				statusLabel.SetText("解析失败：数据过大")
				progressBar.Hide()
				return
			}

			statusLabel.SetText("正在清理输入数据...")
			progressBar.SetValue(0.2)

			cleanedInput := cleanInputForASN1(inputData)
			if cleanedInput == "" {
				dialog.ShowError(fmt.Errorf("清理后的数据为空"), fyne.CurrentApp().Driver().AllWindows()[0])
				statusLabel.SetText("解析失败：数据无效")
				progressBar.Hide()
				return
			}

			statusLabel.SetText("正在解码数据...")
			progressBar.SetValue(0.4)

			decodedData, err := base64.StdEncoding.DecodeString(cleanedInput)
			if err != nil {
				decodedData, err = hex.DecodeString(cleanedInput)
				if err != nil {
					dialog.ShowError(fmt.Errorf("无法解码输入数据\nBase64错误: %v\nHex错误: %v", err, err), fyne.CurrentApp().Driver().AllWindows()[0])
					statusLabel.SetText("解析失败：解码错误")
					progressBar.Hide()
					return
				}
			}

			if len(decodedData) < 2 || len(decodedData) > 2*1024*1024 {
				dialog.ShowError(fmt.Errorf("解码后数据大小异常（%d 字节）", len(decodedData)), fyne.CurrentApp().Driver().AllWindows()[0])
				statusLabel.SetText("解析失败：数据大小异常")
				progressBar.Hide()
				return
			}

			statusLabel.SetText("正在解析ASN.1结构...")
			progressBar.SetValue(0.7)

			rootNode := ParseAsn1(decodedData)

			statusLabel.SetText("正在构建树状视图...")
			progressBar.SetValue(0.9)

			// 回到主线程更新UI
			rootAccordionItem = buildAccordion(rootNode, 0)
			if accordion.Items != nil && len(accordion.Items) > 0 {
				accordion.RemoveIndex(0)
			}
			accordion.Append(rootAccordionItem)

			// 显示统计信息
			childrenCount := countChildren(rootNode)
			maxDepth := getMaxDepth(rootNode)

			statsInfo := widget.NewRichTextFromMarkdown(fmt.Sprintf(
				"### 📊 解析统计\n\n"+
					"- **数据大小**: %d 字节\n"+
					"- **节点总数**: %d\n"+
					"- **最大深度**: %d\n"+
					"- **根节点类型**: %s\n",
				len(decodedData), childrenCount, maxDepth, getRealTag(rootNode.Tag)))

			statsContainer.RemoveAll()
			statsContainer.Add(statsInfo)
			statsContainer.Show()

			statusLabel.SetText("✅ 解析完成")
			progressBar.SetValue(1.0)
			time.Sleep(time.Second)
			progressBar.Hide()

			if rootNode.Error != "" {
				statusLabel.SetText(fmt.Sprintf("⚠️ 解析完成但有警告: %s", rootNode.Error))
			}
		}()
	})
	//清除按钮
	cancelButton := buildButton("🗑️ 清除", theme.CancelIcon(), func() {
		input.SetText("")
		if accordion.Items != nil && len(accordion.Items) > 0 {
			accordion.RemoveIndex(0)
		}
		statsContainer.Hide()
		statusLabel.SetText("准备解析ASN.1数据...")
		progressBar.Hide()
	})

	// 布局
	buttonContainer := container.New(layout.NewGridLayout(2), confirmButton, cancelButton)
	topContainer := container.NewVBox(
		widget.NewLabelWithStyle("🔐 ASN.1 结构解析器", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		input,
		buttonContainer,
		statusLabel,
		progressBar,
		statsContainer,
		widget.NewSeparator(),
	)

	// 使用带滚动条的容器包装整个布局
	mainContainer := container.NewBorder(topContainer, nil, nil, nil, accordion)
	scrollContainer := container.NewScroll(mainContainer)
	scrollContainer.SetMinSize(fyne.NewSize(700, 500))

	return container.NewMax(scrollContainer)

}

func getRealTag(tag int) string {
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
	if len(prefix) > 0 {
		prefix = fmt.Sprintf("%s :", prefix)
	}
	return prefix + TagToName[tag]
}

// cleanInputForASN1 清理ASN1输入数据，移除可能影响解析的字符
func cleanInputForASN1(input string) string {
	// 移除所有空格、换行符、制表符等空白字符
	cleaned := strings.ReplaceAll(input, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "\n", "")
	cleaned = strings.ReplaceAll(cleaned, "\r", "")
	cleaned = strings.ReplaceAll(cleaned, "\t", "")
	return strings.TrimSpace(cleaned)
}

// countChildren 计算节点总数
func countChildren(node ASN1Node) int {
	count := 1 // 当前节点
	for _, child := range node.Children {
		count += countChildren(*child)
	}
	return count
}

// getMaxDepth 获取最大深度
func getMaxDepth(node ASN1Node) int {
	maxDepth := node.Depth
	for _, child := range node.Children {
		childDepth := getMaxDepth(*child)
		if childDepth > maxDepth {
			maxDepth = childDepth
		}
	}
	return maxDepth
}
