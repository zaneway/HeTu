package window

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func CoderStructure(input *widget.Entry) *fyne.Container {
	// 创建输出框，供用户输入数据
	output := widget.NewMultiLineEntry()
	output.Wrapping = fyne.TextWrapWord
	// 设置输出框的最小高度，确保长文本能够正常显示
	output.Resize(fyne.NewSize(400, 120))
	output.Hide()
	input.SetPlaceHolder("Please input base64/hex data")
	// 为公共输入框也设置最小高度
	input.Wrapping = fyne.TextWrapWord
	dataLenPrint := widget.NewLabel("")
	// 解析按钮
	confirmButton := widget.NewButtonWithIcon("确认", theme.ConfirmIcon(), func() {
		inputData := strings.TrimSpace(input.Text)
		output.Text = ""
		decodedData, err := hex.DecodeString(inputData)
		if err == nil {
			output.Text = base64.StdEncoding.EncodeToString(decodedData)
		} else {
			decodedData, err = base64.StdEncoding.DecodeString(inputData)
			if err == nil {
				output.Text = hex.EncodeToString(decodedData)
			} else {
				decodedData = []byte(inputData)
				output.Text = hex.EncodeToString(decodedData)
			}
		}
		dataLen := len(decodedData)
		dataLenPrint.Text = fmt.Sprintf("%s%d%s", "数据长度为:", dataLen, "字节(优先按照HEX解析,后按照Base64解析,普通字符串兜底)")

		dataLenPrint.Refresh()
		output.Show()
		output.Refresh()
	})

	//清除按钮
	cancelButton := buildButton("清除", theme.CancelIcon(), func() {
		input.Text = ""
		output.Text = ""
		dataLenPrint.Text = ""
		input.Refresh()
		output.Refresh()
	})
	// 布局 - 不添加全局输入框，它已经在主界面的固定位置
	allButton := container.New(layout.NewGridLayout(2), confirmButton, cancelButton)
	vbox := container.NewVBox(allButton, dataLenPrint, output)
	// 使用带滚动条的容器包装
	scrollContainer := container.NewScroll(vbox)
	scrollContainer.SetMinSize(fyne.NewSize(400, 300))

	return container.NewMax(scrollContainer)

}
