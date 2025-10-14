package window

import (
	"HeTu/util"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// FormatStructure 构造JSON/XML格式化核心图形模块
func FormatStructure(input *widget.Entry) *fyne.Container {
	structure := container.NewVBox()
	detail := container.NewVBox()

	// 确认按钮
	confirm := widget.NewButtonWithIcon("格式化", theme.ConfirmIcon(), func() {
		inputData := input.Text // 不使用TrimSpace，保持原始数据
		if strings.TrimSpace(inputData) == "" {
			dialog.ShowError(fmt.Errorf("请输入JSON或XML数据"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		// 保存到历史记录
		if inputData != "" {
			util.GetHistoryDB().AddHistory("📄 JSON/XML", inputData)
		}

		detail.RemoveAll()

		// 检查数据类型并格式化
		var formattedData string
		var err error
		var dataType string

		// 检查是否为JSON
		if util.IsJSON(inputData) {
			dataType = "JSON"
			formattedData, err = util.FormatJSON(inputData)
			if err != nil {
				dialog.ShowError(fmt.Errorf("JSON格式化失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
				return
			}
		} else if util.IsXML(inputData) {
			// 检查是否为XML
			dataType = "XML"
			formattedData, err = util.FormatXML(inputData)
			if err != nil {
				dialog.ShowError(fmt.Errorf("XML格式化失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
				return
			}
		} else {
			dialog.ShowError(fmt.Errorf("输入的数据既不是有效的JSON也不是有效的XML"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		// 显示格式化后的数据
		resultEntry := widget.NewMultiLineEntry()
		resultEntry.Wrapping = fyne.TextWrapWord
		resultEntry.SetText(formattedData)

		// 固定可见行数为15行，取消自动调整
		resultEntry.SetMinRowsVisible(15)

		// 将结果框包装在滚动容器中以确保滚动功能
		resultScroll := container.NewScroll(resultEntry)
		resultScroll.SetMinSize(fyne.NewSize(0, 300)) // 固定高度300像素

		// 添加标签
		label := widget.NewLabel(fmt.Sprintf("格式化后的%s数据:", dataType))
		label.TextStyle = fyne.TextStyle{Bold: true}

		detail.Add(label)
		detail.Add(resultScroll)
		detail.Refresh()
	})

	// 清除按钮
	clear := widget.NewButtonWithIcon("清除", theme.CancelIcon(), func() {
		input.Text = ""
		input.Refresh()
		detail.RemoveAll()
		detail.Refresh()
	})

	// 按钮布局
	buttons := container.New(layout.NewGridLayout(2), confirm, clear)
	structure.Add(buttons)
	structure.Add(detail)

	// 使用带滚动条的容器包装整个结构
	scrollContainer := container.NewScroll(structure)
	scrollContainer.SetMinSize(fyne.NewSize(600, 400))
	return container.NewMax(scrollContainer)
}
