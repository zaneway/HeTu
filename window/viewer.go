package window

import (
	"HeTu/util"
	"fmt"
	"net/url"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	_ "github.com/lengzhao/font/autoload" //这个可以让你识别中文
)

// 定义标签页名称常量
const (
	CoderTab       = "🔄 编码转换"
	CertificateTab = "🏆 证书解析"
	Asn1Tab        = "🌳 ASN.1结构"
	KeyTab         = "🗝️ 密钥工具"
	EnvelopTab     = "📦 信封解析"
	P10Tab         = "📝 P10请求"
	P12Tab         = "🎫 P12证书"
	P7bTab         = "🔗 P7B证书链"
	CrlTab         = "📜 CRL列表"
	FormatTab      = "📄 JSON/XML"
)

// 全局历史记录管理器引用
var (
	globalHistoryManager *HistoryManager
	historyManagerMutex  sync.RWMutex
)

// GetGlobalHistoryManager 获取全局历史记录管理器
func GetGlobalHistoryManager() *HistoryManager {
	historyManagerMutex.RLock()
	defer historyManagerMutex.RUnlock()
	return globalHistoryManager
}

// SetGlobalHistoryManager 设置全局历史记录管理器
func SetGlobalHistoryManager(manager *HistoryManager) {
	historyManagerMutex.Lock()
	defer historyManagerMutex.Unlock()
	globalHistoryManager = manager
}

func NewWindow() {
	myApp := app.New()
	// 设置应用主题为深色主题
	myApp.Settings().SetTheme(theme.DefaultTheme())

	// 创建一个窗口对象
	myWindow := myApp.NewWindow("🔐 HeTu - 密码学工具箱")
	// 设置窗口图标（可选）
	// myWindow.SetIcon(resourceIconPng)

	// 创建共享输入框，用于接收拖拽文件内容
	sharedInput := createSharedInput()

	// 设置文件拖拽处理函数
	myWindow.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		if len(uris) > 0 {
			filePath := uris[0].Path()
			// 读取文件内容
			content, err := util.ReadFileContent(filePath)
			if err != nil {
				sharedInput.SetText("文件读取错误: " + err.Error())
				return
			}

			// 判断内容是否为ASCII或汉字
			if util.IsASCIIOrChinese(content) {
				// 如果是ASCII或汉字，直接显示
				sharedInput.SetText(string(content))
			} else {
				// 否则进行base64编码
				encodedContent := util.Base64EncodeToString(content)
				sharedInput.SetText(encodedContent)
			}
		}
	})

	body := newBody(sharedInput)
	myWindow.SetContent(body)
	myWindow.Resize(fyne.Size{1000, 700}) // 增大窗口尺寸
	myWindow.CenterOnScreen()             // 窗口居中显示
	myWindow.ShowAndRun()
}

func newBody(sharedInput *widget.Entry) *fyne.Container {
	// 创建美化的表头区域
	headerContainer := createHeader()

	// 创建主要内容区域（传入共享输入框）
	mainContent := createMainContent(sharedInput)

	// 创建底部状态栏
	footerContainer := createFooter()

	// 整体布局采用边框布局
	body := container.NewBorder(
		headerContainer, // 顶部
		footerContainer, // 底部
		nil,             // 左侧
		nil,             // 右侧
		mainContent,     // 中心内容
	)

	return body
}

// 创建美化的表头
func createHeader() *fyne.Container {
	// 项目标题
	titleLabel := widget.NewLabelWithStyle("🔐 HeTu", fyne.TextAlignCenter, fyne.TextStyle{
		Bold: true,
	})
	titleLabel.TextStyle.Bold = true

	// 副标题
	//subTitle := widget.NewLabelWithStyle("可视化密码学操作平台", fyne.TextAlignCenter, fyne.TextStyle{
	//	Italic: true,
	//})

	// GitHub链接
	url, _ := url.Parse("https://github.com/zaneway/HeTu")
	githubLink := widget.NewHyperlink("🌟 访问全球最大的同性交友网站（项目主页）", url)

	// 时间显示
	timeLabel := refreshTimeSeconds()
	timeLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// 表头布局
	//headerTop := container.NewVBox(
	//	titleLabel,
	//	subTitle,
	//)

	headerBottom := container.NewBorder(
		nil, nil,
		container.NewCenter(githubLink),
		timeLabel,
		nil, // 移除中心位置的分隔线
	)

	headerContainer := container.NewVBox(
		//container.NewPadded(headerTop),
		headerBottom,
	)

	return headerContainer
}

// 创建全局共享的输入框
func createSharedInput() *widget.Entry {
	input := widget.NewMultiLineEntry()
	input.SetPlaceHolder("📝 请输入 Base64/Hex 格式的数据进行解析，或拖拽文件到此处...")
	input.Wrapping = fyne.TextWrapWord
	// 设置固定尺寸，防止移位
	input.Resize(fyne.NewSize(0, 140)) // 宽度自适应，高度固定

	return input
}

// 创建主要内容区域
func createMainContent(sharedInput *widget.Entry) *fyne.Container {
	// 输入框标签
	inputLabel := widget.NewLabelWithStyle("📋 数据输入区域", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// 输入框容器 - 使用固定布局
	inputContainer := container.NewVBox(
		inputLabel,
		container.NewPadded(sharedInput),
		widget.NewSeparator(),
	)

	// 定义各标签页的占位符文本
	placeholders := map[string]string{
		CoderTab:       "📝 请输入 Base64/Hex 格式的数据进行编码转换，或拖拽文件到此处...",
		CertificateTab: "📝 请输入 Base64/Hex 格式的证书数据进行解析，或拖拽证书文件到此处...",
		Asn1Tab:        "📝 请输入 Base64/Hex 格式的 ASN.1 数据进行解析，或拖拽文件到此处...",
		KeyTab:         "📝 密钥生成工具 - 请在下方选择算法并生成密钥，或拖拽密钥文件到此处...",
		EnvelopTab:     "📝 请输入 Base64/Hex 格式的信封数据 (GMT-0009)，或拖拽文件到此处...",
		P10Tab:         "📝 请输入 Base64/Hex 格式的 P10 证书签名请求数据，或拖拽P10文件到此处...",
		P12Tab:         "📝 请输入 Base64/Hex 格式的证书数据生成 PFX 文件，或拖拽证书文件到此处...",
		P7bTab:         "📝 请输入 Base64/Hex 格式的 P7B 证书链数据，或拖拽P7B文件到此处...",
		CrlTab:         "📝 请输入 Base64/Hex 格式的 CRL 数据，或拖拽CRL文件到此处...",
		FormatTab:      "📝 请输入 JSON 或 XML 数据进行格式化，或拖拽文件到此处...",
	}

	// 创建历史记录下拉框
	historySelect := widget.NewSelect([]string{}, func(selected string) {
		// 历史记录选择功能将在HistoryManager中实现
	})
	historySelect.PlaceHolder = "📖 历史记录"

	// 创建历史记录管理器
	historyManager := NewHistoryManager(historySelect, sharedInput)

	// 设置全局历史记录管理器引用
	SetGlobalHistoryManager(historyManager)

	// 创建清除历史记录按钮
	clearHistoryBtn := widget.NewButtonWithIcon("🗑️", theme.DeleteIcon(), func() {
		// 清除当前标签页的历史记录
		dialog.ShowConfirm("确认清除", "确定要清除当前标签页的所有历史记录吗？",
			func(confirmed bool) {
				if confirmed {
					err := historyManager.ClearHistory()
					if err != nil {
						dialog.ShowError(fmt.Errorf("清除历史记录失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
						return
					}
					//dialog.ShowInformation("成功", "历史记录已清除", fyne.CurrentApp().Driver().AllWindows()[0])
				}
			}, fyne.CurrentApp().Driver().AllWindows()[0])
	})

	// 设置历史记录下拉框的回调函数
	historySelect.OnChanged = func(selected string) {
		historyManager.SelectHistory(selected)
	}

	// 创建历史记录容器
	historyContainer := container.NewBorder(
		nil,
		nil,
		clearHistoryBtn,
		nil,
		historySelect,
	)

	// 将历史记录容器添加到输入框容器中
	inputContainer.Add(historyContainer)
	inputContainer.Add(widget.NewSeparator())

	// 创建多行标签页（使用自定义布局）
	tabs := createMultiRowTabs(sharedInput, placeholders, historyManager)

	// 设置默认占位符（编码转换）
	sharedInput.SetPlaceHolder(placeholders[CoderTab])

	// 移除输入框内容变化时的自动保存逻辑
	// 当输入框内容发生变化时，不再自动保存到历史记录
	originalOnChanged := sharedInput.OnChanged
	sharedInput.OnChanged = func(s string) {
		// 调用原来的OnChanged处理函数（如果存在）
		if originalOnChanged != nil {
			originalOnChanged(s)
		}
		// 不再自动保存历史记录
	}

	// 主要内容区域 - 使用Border布局分离输入框和标签页
	mainContent := container.NewBorder(
		inputContainer, // 顶部固定输入框
		nil,            // 底部
		nil,            // 左侧
		nil,            // 右侧
		tabs,           // 中心标签页内容
	)

	return container.NewPadded(mainContent)
}

// 创建底部状态栏
func createFooter() *fyne.Container {
	// 版本信息
	versionLabel := widget.NewLabel("v1.0.6")
	versionLabel.TextStyle = fyne.TextStyle{Italic: true}

	// 状态信息
	statusLabel := widget.NewLabel("✅ 完美")
	statusLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// 底部布局
	footerContainer := container.NewBorder(
		widget.NewSeparator(), // 顶部分隔线
		nil,                   // 底部
		versionLabel,          // 左侧版本信息
		statusLabel,           // 右侧状态信息
		widget.NewLabel("河图洛书 - 探索密码学的奥秘"), // 中心文本
	)

	return container.NewPadded(footerContainer)
}

// createMultiRowTabs 创建多行显示的标签页
func createMultiRowTabs(sharedInput *widget.Entry, placeholders map[string]string, historyManager *HistoryManager) *fyne.Container {
	// 定义所有标签页的数据
	tabItems := []struct {
		name    string
		icon    fyne.Resource
		content func() *fyne.Container
	}{
		{CoderTab, theme.ZoomInIcon(), func() *fyne.Container { return CoderStructure(sharedInput) }},
		{CertificateTab, theme.InfoIcon(), func() *fyne.Container { return CertificateStructure(sharedInput) }},
		{Asn1Tab, theme.ZoomInIcon(), func() *fyne.Container { return Asn1Structure(sharedInput) }},
		{KeyTab, theme.ColorChromaticIcon(), func() *fyne.Container { return KeyStructure(sharedInput) }},
		{EnvelopTab, theme.FolderIcon(), func() *fyne.Container { return SM2EnvelopedPfxStructure(sharedInput) }},
		{P10Tab, theme.DocumentIcon(), func() *fyne.Container { return P10Structure(sharedInput) }},
		{P12Tab, theme.AccountIcon(), func() *fyne.Container { return SM2PfxStructure(sharedInput) }},
		{P7bTab, theme.InfoIcon(), func() *fyne.Container { return P7bStructure(sharedInput) }},
		{CrlTab, theme.AccountIcon(), func() *fyne.Container { return CrlStructure(sharedInput) }},
		{FormatTab, theme.DocumentIcon(), func() *fyne.Container { return FormatStructure(sharedInput) }},
	}

	// 创建内容容器
	contentContainer := container.NewStack()

	// 初始化显示第一个标签页的内容
	if len(tabItems) > 0 {
		contentContainer.Add(tabItems[0].content())
	}

	// 创建标签按钮容器（使用网格布局支持多行）
	var tabButtons []fyne.CanvasObject
	for i, item := range tabItems {
		index := i // 捕获索引
		tabName := item.name
		// 捕获内容函数，避免闭包问题
		contentFunc := item.content

		// 创建标签按钮
		tabBtn := widget.NewButtonWithIcon(tabName, item.icon, func() {
			// 移除旧内容
			contentContainer.RemoveAll()

			// 添加新内容
			contentContainer.Add(contentFunc())

			// 更新按钮样式（高亮当前选中的标签）
			for j, btn := range tabButtons {
				if button, ok := btn.(*widget.Button); ok {
					if j == index {
						button.Importance = widget.HighImportance
					} else {
						button.Importance = widget.MediumImportance
					}
					// 刷新按钮以更新显示
					button.Refresh()
				}
			}

			// 更新占位符
			if placeholder, exists := placeholders[tabName]; exists {
				sharedInput.SetPlaceHolder(placeholder)
				sharedInput.Refresh()
			}

			// 更新当前标签页
			historyManager.SetCurrentTab(tabName)

			// 加载该标签页的历史记录
			historyManager.LoadHistoryForTab(tabName)

			contentContainer.Refresh()
		})

		// 设置第一个按钮为高亮
		if i == 0 {
			tabBtn.Importance = widget.HighImportance
		} else {
			tabBtn.Importance = widget.MediumImportance
		}

		tabButtons = append(tabButtons, tabBtn)
	}

	// 使用网格布局，每行最多显示5个标签，自动换行
	// 使用 container.NewAdaptiveGrid 可以自动适应窗口大小
	tabButtonsContainer := container.NewGridWithColumns(5, tabButtons...)

	// 将标签按钮容器和内容容器组合
	result := container.NewBorder(
		tabButtonsContainer, // 顶部：标签按钮（多行显示）
		nil,                 // 底部
		nil,                 // 左侧
		nil,                 // 右侧
		contentContainer,    // 中心：标签页内容
	)

	return result
}

func refreshTimeSeconds() *widget.Label {
	//填充当前时间
	nowTime := widget.NewLabel(time.Now().Format(util.DateTime))
	//异步线程更新时间
	go func() {
		for range time.Tick(time.Second) {
			// 使用 fyne.Do 确保在正确的线程中更新 UI
			fyne.Do(func() {
				nowTime.SetText(time.Now().Format(util.DateTime))
			})
		}
	}()
	return nowTime
}

// HistoryManager 历史记录管理器
type HistoryManager struct {
	currentTab   string
	historyMap   map[string][]util.HistoryRecord
	selectWidget *widget.Select
	inputWidget  *widget.Entry
	// 为每个标签页维护一个显示文本到完整内容的映射
	displayTextMap map[string]map[string]string
}

// NewHistoryManager 创建历史记录管理器
func NewHistoryManager(selectWidget *widget.Select, inputWidget *widget.Entry) *HistoryManager {
	return &HistoryManager{
		currentTab:     CoderTab,
		historyMap:     make(map[string][]util.HistoryRecord),
		selectWidget:   selectWidget,
		inputWidget:    inputWidget,
		displayTextMap: make(map[string]map[string]string),
	}
}

// SetCurrentTab 设置当前标签页
func (hm *HistoryManager) SetCurrentTab(tabName string) {
	hm.currentTab = tabName
}

// LoadHistoryForTab 加载指定标签页的历史记录
func (hm *HistoryManager) LoadHistoryForTab(tabName string) {
	// 从数据库获取历史记录
	records, err := util.GetHistoryDB().GetHistory(tabName, 20) // 获取最近20条记录
	if err != nil {
		// 如果获取失败，不显示历史记录
		return
	}

	// 保存到内存缓存
	hm.historyMap[tabName] = records

	// 初始化当前标签页的显示文本映射
	hm.displayTextMap[tabName] = make(map[string]string)

	// 清空历史记录下拉框
	hm.selectWidget.Options = []string{"📖 历史记录"}

	// 添加历史记录到下拉框
	for _, record := range records {
		// 截取内容的前50个字符作为显示文本
		contentText := record.Content
		if len(contentText) > 50 {
			contentText = contentText[:50] + "..."
		}

		// 格式化时间（只显示月-日 时:分）
		timeText := record.CreatedAt.Format("01-02 15:04")

		// 构造显示文本（时间 + 内容）
		displayText := fmt.Sprintf("[%s] %s", timeText, contentText)

		// 处理重复显示文本的情况
		originalDisplayText := displayText
		counter := 1
		for {
			// 检查是否已存在相同的显示文本
			_, exists := hm.displayTextMap[tabName][displayText]
			if !exists {
				break
			}
			// 如果存在，添加计数器后缀
			counter++
			if len(originalDisplayText) > 45 {
				displayText = originalDisplayText[:45] + fmt.Sprintf("..%d", counter)
			} else {
				displayText = originalDisplayText + fmt.Sprintf(" #%d", counter)
			}
		}

		// 保存显示文本到完整内容的映射（只保存内容，不保存时间）
		hm.displayTextMap[tabName][displayText] = record.Content
		hm.selectWidget.Options = append(hm.selectWidget.Options, displayText)
	}

	// 更新下拉框
	hm.selectWidget.Refresh()
}

// SaveHistory 保存历史记录
func (hm *HistoryManager) SaveHistory(content string) {
	if content != "" {
		util.GetHistoryDB().AddHistory(hm.currentTab, content)
		// 重新加载历史记录
		hm.LoadHistoryForTab(hm.currentTab)
	}
}

// ClearHistory 清除当前标签页的历史记录
func (hm *HistoryManager) ClearHistory() error {
	err := util.GetHistoryDB().ClearHistory(hm.currentTab)
	if err != nil {
		return err
	}
	// 重新加载历史记录
	hm.LoadHistoryForTab(hm.currentTab)
	return nil
}

// SelectHistory 选择历史记录
func (hm *HistoryManager) SelectHistory(selected string) {
	if selected != "" && selected != "📖 历史记录" {
		// 直接使用显示文本映射获取完整内容
		if contentMap, exists := hm.displayTextMap[hm.currentTab]; exists {
			if content, exists := contentMap[selected]; exists {
				hm.inputWidget.SetText(content)
			}
		}
	}
}
