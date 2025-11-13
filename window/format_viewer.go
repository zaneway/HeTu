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

	// 创建状态标签和进度条
	statusLabel := widget.NewLabel("准备格式化...")
	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	// 格式化函数
	performFormatting := func(inputData string, detail *fyne.Container, statusLabel *widget.Label, progressBar *widget.ProgressBar) {
		// 保存到历史记录
		if inputData != "" {
			util.GetHistoryDB().AddHistory("📄 JSON/XML", inputData)

			// 刷新历史记录下拉框
			if historyManager := GetGlobalHistoryManager(); historyManager != nil {
				historyManager.LoadHistoryForTab("📄 JSON/XML")
			}
		}

		// 清除旧内容并显示进度
		detail.RemoveAll()
		statusLabel.SetText("正在检查数据类型...")
		progressBar.Show()
		progressBar.SetValue(0.1)
		detail.Add(statusLabel)
		detail.Add(progressBar)
		detail.Refresh()

		// 在后台 goroutine 中执行格式化操作
		go func() {
			// 先快速检查数据类型
			var dataType string
			var isJSON, isXML bool

			fyne.Do(func() {
				statusLabel.SetText("正在检查数据类型...")
				progressBar.SetValue(0.2)
			})

			// 检查数据类型
			isJSON = util.IsJSON(inputData)
			isXML = util.IsXML(inputData)

			if !isJSON && !isXML {
				fyne.Do(func() {
					progressBar.Hide()
					dialog.ShowError(fmt.Errorf("输入的数据既不是有效的JSON也不是有效的XML"), fyne.CurrentApp().Driver().AllWindows()[0])
					statusLabel.SetText("格式检查失败")
				})
				return
			}

			// 执行格式化
			var formattedData string
			var err error

			if isJSON {
				dataType = "JSON"
				fyne.Do(func() {
					statusLabel.SetText("正在格式化JSON数据...")
					progressBar.SetValue(0.5)
				})

				formattedData, err = util.FormatJSON(inputData)
			} else {
				dataType = "XML"
				fyne.Do(func() {
					statusLabel.SetText("正在格式化XML数据...")
					progressBar.SetValue(0.5)
				})

				formattedData, err = util.FormatXML(inputData)
			}

			if err != nil {
				fyne.Do(func() {
					progressBar.Hide()
					dialog.ShowError(fmt.Errorf("%s格式化失败: %v", dataType, err), fyne.CurrentApp().Driver().AllWindows()[0])
					statusLabel.SetText(fmt.Sprintf("%s格式化失败", dataType))
				})
				return
			}

			// 更新UI显示结果
			fyne.Do(func() {
				statusLabel.SetText("正在显示结果...")
				progressBar.SetValue(0.9)

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

				// 清除进度条，显示结果
				detail.RemoveAll()
				detail.Add(label)
				detail.Add(resultScroll)

				progressBar.Hide()
				detail.Refresh()
			})
		}()
	}

	// 确认按钮
	confirm := widget.NewButtonWithIcon("格式化", theme.ConfirmIcon(), func() {
		inputData := strings.TrimSpace(input.Text) // 清理输入数据
		if inputData == "" {
			dialog.ShowError(fmt.Errorf("请输入JSON或XML数据"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		// 检查数据大小，如果太大给出警告
		dataSize := len(inputData)
		if dataSize > 10*1024*1024 { // 10MB
			dialog.ShowConfirm("数据较大", fmt.Sprintf("输入数据较大（%d KB），格式化可能需要较长时间，是否继续？", dataSize/1024),
				func(confirmed bool) {
					if !confirmed {
						return
					}
					performFormatting(inputData, detail, statusLabel, progressBar)
				}, fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		performFormatting(inputData, detail, statusLabel, progressBar)
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
