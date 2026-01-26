package window

import (
	"HeTu/util"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/corvus-ch/shamir"
)

// ShamirStructure 构造Shamir门限算法图形模块
func ShamirStructure(input *widget.Entry) *fyne.Container {
	structure := container.NewVBox()

	// --- 拆分部分 ---
	splitGroup := widget.NewCard("🔐 秘密拆分 (Split)", "将秘密拆分为多个部分", nil)

	// 参数输入
	partsEntry := widget.NewEntry()
	partsEntry.SetPlaceHolder("总份数 (N)")
	partsEntry.Text = "5"

	thresholdEntry := widget.NewEntry()
	thresholdEntry.SetPlaceHolder("恢复所需份数 (K)")
	thresholdEntry.Text = "3"

	// 结果显示
	splitResult := widget.NewMultiLineEntry()
	splitResult.SetPlaceHolder("拆分结果将显示在这里...")
	splitResult.Wrapping = fyne.TextWrapWord
	splitResult.SetMinRowsVisible(5)

	// 拆分按钮
	splitBtn := widget.NewButtonWithIcon("拆分秘密", theme.ContentCutIcon(), func() {
		secretStr := input.Text
		if secretStr == "" {
			dialog.ShowError(fmt.Errorf("请输入要拆分的秘密"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		parts, err := strconv.Atoi(partsEntry.Text)
		if err != nil || parts < 2 || parts > 255 {
			dialog.ShowError(fmt.Errorf("总份数必须在 2-255 之间"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		threshold, err := strconv.Atoi(thresholdEntry.Text)
		if err != nil || threshold < 2 || threshold > parts {
			dialog.ShowError(fmt.Errorf("阈值必须在 2 到总份数之间"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		// 执行拆分
		shares, err := shamir.Split([]byte(secretStr), parts, threshold)
		if err != nil {
			dialog.ShowError(fmt.Errorf("拆分失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		// 格式化输出
		var sb strings.Builder
		for k, v := range shares {
			sb.WriteString(fmt.Sprintf("Index: %d, Share: %s\n", k, hex.EncodeToString(v)))
		}
		splitResult.SetText(sb.String())

		// 保存到历史记录
		util.GetHistoryDB().AddHistory("🧩 Shamir", secretStr)
	})

	splitContent := container.NewVBox(
		container.NewGridWithColumns(2,
			widget.NewForm(widget.NewFormItem("总份数 (N)", partsEntry)),
			widget.NewForm(widget.NewFormItem("阈值 (K)", thresholdEntry)),
		),
		splitBtn,
		splitResult,
	)
	splitGroup.Content = splitContent

	// --- 合并部分 ---
	combineGroup := widget.NewCard("🔓 秘密恢复 (Combine)", "使用分片恢复秘密", nil)

	// 分片输入
	sharesInput := widget.NewMultiLineEntry()
	sharesInput.SetPlaceHolder("请输入分片数据，格式：Index:ShareHex (每行一个)\n例如:\n1: a1b2...\n2: c3d4...")
	sharesInput.Wrapping = fyne.TextWrapWord
	sharesInput.SetMinRowsVisible(5)

	// 恢复结果
	combineResult := widget.NewEntry()
	combineResult.SetPlaceHolder("恢复的秘密将显示在这里...")
	combineResult.TextStyle = fyne.TextStyle{Bold: true}

	// 合并按钮
	combineBtn := widget.NewButtonWithIcon("恢复秘密", theme.ConfirmIcon(), func() {
		inputStr := sharesInput.Text
		if inputStr == "" {
			dialog.ShowError(fmt.Errorf("请输入分片数据"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		lines := strings.Split(inputStr, "\n")
		shares := make(map[byte][]byte)

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				// 尝试解析纯Hex可能不太好，因为需要Index。这里严格要求格式。
				continue
			}

			indexStr := strings.TrimSpace(parts[0])
			shareHex := strings.TrimSpace(parts[1])

			// 去除可能的 "Index: " 前缀如果用户直接复制了上面的输出
			indexStr = strings.TrimPrefix(indexStr, "Index")
			indexStr = strings.TrimSpace(indexStr)

			// 去除 "Share: " 前缀
			shareHex = strings.TrimPrefix(shareHex, "Share")
			shareHex = strings.TrimPrefix(shareHex, ":")
			shareHex = strings.TrimSpace(shareHex)

			idx, err := strconv.Atoi(indexStr)
			if err != nil || idx < 0 || idx > 255 {
				continue // 跳过无效行
			}

			shareData, err := hex.DecodeString(shareHex)
			if err != nil {
				continue // 跳过无效Hex
			}

			shares[byte(idx)] = shareData
		}

		if len(shares) < 2 {
			dialog.ShowError(fmt.Errorf("有效分片数量不足，请检查输入格式"), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		// 执行恢复
		secret, err := shamir.Combine(shares)
		if err != nil {
			dialog.ShowError(fmt.Errorf("恢复失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		combineResult.SetText(string(secret))
	})

	combineContent := container.NewVBox(
		sharesInput,
		combineBtn,
		combineResult,
	)
	combineGroup.Content = combineContent

	// 组装整体布局
	structure.Add(splitGroup)
	structure.Add(layout.NewSpacer()) // 增加一点间距
	structure.Add(combineGroup)

	// 使用滚动容器
	scrollContainer := container.NewScroll(structure)
	scrollContainer.SetMinSize(fyne.NewSize(600, 500))

	return container.NewMax(scrollContainer)
}
