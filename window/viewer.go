package window

import (
	"HeTu/util"
	"net/url"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	_ "github.com/lengzhao/font/autoload" //这个可以让你识别中文
)

func NewWindow() {
	myApp := app.New()
	// 设置应用主题为深色主题
	myApp.Settings().SetTheme(theme.DefaultTheme())

	// 创建一个窗口对象
	myWindow := myApp.NewWindow("🔐 HeTu - 密码学工具箱")
	// 设置窗口图标（可选）
	// myWindow.SetIcon(resourceIconPng)

	body := newBody()
	myWindow.SetContent(body)
	myWindow.Resize(fyne.Size{1000, 700}) // 增大窗口尺寸
	myWindow.CenterOnScreen()             // 窗口居中显示
	myWindow.ShowAndRun()
}

func newBody() *fyne.Container {
	// 创建美化的表头区域
	headerContainer := createHeader()

	// 创建全局共享的输入框
	sharedInput := createSharedInput()

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
	input.SetPlaceHolder("📝 请输入 Base64/Hex 格式的数据进行解析...")
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

	// 创建美化的标签页
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("🔄 编码转换", theme.ZoomInIcon(), CoderStructure(sharedInput)),
		container.NewTabItemWithIcon("🏆 证书解析", theme.InfoIcon(), CertificateStructure(sharedInput)),
		container.NewTabItemWithIcon("🌳 ASN.1结构", theme.ZoomInIcon(), Asn1Structure(sharedInput)),
		container.NewTabItemWithIcon("🗝️ 密钥工具", theme.ColorChromaticIcon(), KeyStructure(sharedInput)),
		container.NewTabItemWithIcon("📦 信封解析", theme.FolderIcon(), SM2EnvelopedPfxStructure(sharedInput)),
		container.NewTabItemWithIcon("🎫 P12证书", theme.AccountIcon(), SM2PfxStructure(sharedInput)),
		container.NewTabItemWithIcon("📜 CRL列表", theme.AccountIcon(), CrlStructure(sharedInput)),
	)

	// 设置标签页样式
	tabs.SetTabLocation(container.TabLocationTop)

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
	versionLabel := widget.NewLabel("v1.0.0")
	versionLabel.TextStyle = fyne.TextStyle{Italic: true}

	// 状态信息
	statusLabel := widget.NewLabel("✅ 就绪")
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

func refreshTimeSeconds() *widget.Label {
	//填充当前时间
	nowTime := widget.NewLabel(time.Now().Format(util.DateTime))
	//异步线程更新时间
	go func() {
		for range time.Tick(time.Second) {
			nowTime.SetText(time.Now().Format(util.DateTime))
		}
	}()
	return nowTime
}
