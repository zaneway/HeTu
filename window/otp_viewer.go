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
	"time"
)

// FormatStructure 构造JSON/XML格式化核心图形模块
func OTPStructure(input *widget.Entry) *fyne.Container {

	structure := container.NewVBox()
	detail := container.NewVBox()
	var stopChan chan struct{}

	// 创建状态标签和进度条
	statusLabel := widget.NewLabel("准备生成OTP")
	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	// 格式化函数
	performFormatting := func(inputData string, detail *fyne.Container, statusLabel *widget.Label, progressBar *widget.ProgressBar) {
		// 保存到历史记录
		if inputData != "" {
			util.GetHistoryDB().AddHistory("📄 TOTP", inputData)

			// 刷新历史记录下拉框
			if historyManager := GetGlobalHistoryManager(); historyManager != nil {
				historyManager.LoadHistoryForTab("📄 TOTP")
			}
		}

		// 停止之前的倒计时
		if stopChan != nil {
			close(stopChan)
		}
		stopChan = make(chan struct{})
		currentStopChan := stopChan

		// 清除旧内容
		detail.RemoveAll()

		// OTP显示组件
		otpLabel := widget.NewLabel("")
		otpLabel.TextStyle = fyne.TextStyle{Bold: true}
		otpLabel.Alignment = fyne.TextAlignCenter

		// 复制按钮
		copyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
			if otpLabel.Text != "" && !strings.HasPrefix(otpLabel.Text, "Error") {
				win := fyne.CurrentApp().Driver().AllWindows()[0]
				win.Clipboard().SetContent(otpLabel.Text)
			}
		})

		otpContainer := container.NewBorder(nil, nil, nil, copyBtn, otpLabel)

		// 倒计时标签
		countDownLabel := widget.NewLabel("")
		countDownLabel.Alignment = fyne.TextAlignCenter

		// 倒计时进度条
		timerBar := widget.NewProgressBar()
		timerBar.Min = 0
		timerBar.Max = 30
		timerBar.TextFormatter = func() string { return "" }

		// 布局
		content := container.NewVBox(
			layout.NewSpacer(),
			//widget.NewLabelWithStyle("OTP Code:", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			otpContainer,
			layout.NewSpacer(),
			countDownLabel,
			timerBar,
			layout.NewSpacer(),
		)

		detail.Add(content)
		detail.Refresh()

		// 在后台 goroutine 中执行操作
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			updateFunc := func() {
				otp, err := totp.Generate(inputData, 30)
				now := time.Now().Unix()
				remaining := 30 - (now % 30)

				fyne.Do(func() {
					if err != nil {
						otpLabel.SetText("Error: " + err.Error())
					} else {
						otpLabel.SetText(otp)
					}
					countDownLabel.SetText(fmt.Sprintf("更新倒计时: %d 秒", remaining))
					timerBar.SetValue(float64(remaining))
				})
			}

			// 立即执行一次
			updateFunc()

			for {
				select {
				case <-currentStopChan:
					return
				case <-ticker.C:
					updateFunc()
				}
			}
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
		if stopChan != nil {
			close(stopChan)
			stopChan = nil
		}
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
