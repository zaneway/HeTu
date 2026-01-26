package window

import (
	"HeTu/util"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/zaneway/otp/totp"
	"strings"
)

// FormatStructure 构造JSON/XML格式化核心图形模块
func OTPStructure(input *widget.Entry) *fyne.Container {

	structure := container.NewVBox()
	detail := container.NewVBox()

	// 创建状态标签和进度条
	statusLabel := widget.NewLabel("准备生成OTP")
	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	// 格式化函数
	performFormatting := func(inputData string, detail *fyne.Container, statusLabel *widget.Label, progressBar *widget.ProgressBar) {
		// 保存到历史记录
		if inputData != "" {
			util.GetHistoryDB().AddHistory("📄 Secret", inputData)

			// 刷新历史记录下拉框
			if historyManager := GetGlobalHistoryManager(); historyManager != nil {
				historyManager.LoadHistoryForTab("📄 Secret")
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

		// 在后台 goroutine 中执行操作
		go func() {

			// 更新UI显示结果
			fyne.Do(func() {
				statusLabel.SetText("正在显示结果...")
				progressBar.SetValue(0.9)
				otp, err := totp.Generate(inputData, 30)
				if err != nil {
					return
				}
				// 显示格式化后的数据
				resultEntry := widget.NewMultiLineEntry()
				resultEntry.Wrapping = fyne.TextWrapWord
				resultEntry.SetText(otp)

				// 固定可见行数为15行，取消自动调整
				resultEntry.SetMinRowsVisible(15)

				// 将结果框包装在滚动容器中以确保滚动功能
				resultScroll := container.NewScroll(resultEntry)
				resultScroll.SetMinSize(fyne.NewSize(0, 300)) // 固定高度300像素

				// 添加标签

				// 清除进度条，显示结果
				detail.RemoveAll()
				detail.Add(resultScroll)

				progressBar.Hide()
				detail.Refresh()
			})
		}()
	}

	// 确认按钮
	confirm := widget.NewButtonWithIcon("生成", theme.ConfirmIcon(), func() {
		inputData := strings.TrimSpace(input.Text) // 清理输入数据
		if inputData == "" {
			dialog.ShowError(fmt.Errorf("请输入密钥"), fyne.CurrentApp().Driver().AllWindows()[0])
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
