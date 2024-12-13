package window

import (
	"encoding/base64"
	"encoding/hex"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"strings"
)

func CoderStructure() *fyne.Container {
	// 创建输入框，供用户输入数据
	input := widget.NewMultiLineEntry()
	output := widget.NewMultiLineEntry()
	output.Hide()
	input.SetPlaceHolder("Please input base64/hex data")
	// 解析按钮
	confirmButton := widget.NewButtonWithIcon("确认", theme.ConfirmIcon(), func() {
		inputData := strings.TrimSpace(input.Text)
		output.Text = ""
		decodedData, err := hex.DecodeString(inputData)
		if err == nil {
			output.Text = base64.StdEncoding.EncodeToString(decodedData)
		} else {
			hexDecodeString, _ := base64.StdEncoding.DecodeString(inputData)
			output.Text = hex.EncodeToString(hexDecodeString)
		}
		output.Show()
		output.Refresh()
	})
	//清除按钮
	cancelButton := buildButton("清除", theme.CancelIcon(), func() {
		input.Text = ""
		output.Text = ""
		input.Refresh()
		output.Refresh()
	})
	// 布局
	allButton := container.New(layout.NewGridLayout(2), confirmButton, cancelButton)
	vbox := container.NewVBox(input, allButton, output)

	return vbox

}
