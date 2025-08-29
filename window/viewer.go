package window

import (
	"HeTu/util"
	"net/url"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	_ "github.com/lengzhao/font/autoload" //这个可以让你识别中文
)

func NewWindow() {
	myApp := app.New()
	// 创建一个窗口对象
	myWindow := myApp.NewWindow("zaneway`s Tools of HeTu")
	body := newBody()
	myWindow.SetContent(body)
	myWindow.Resize(fyne.Size{800, 600})
	myWindow.ShowAndRun()

}

func newBody() *fyne.Container {
	// 表头
	url, _ := url.Parse("https://github.com/zaneway/HeTu")
	link := widget.NewHyperlink("^-^  欢迎访问全球最大的同性交友网站  ^-^", url)
	//超链接显示在中间
	centerLink := container.NewCenter(link)
	//时间显示在最右侧
	rightTime := container.NewHBox(layout.NewSpacer(), refreshTimeSeconds())
	//搞一个公共输入框
	input := widget.NewMultiLineEntry()
	input.SetPlaceHolder("Please input base64/hex data")
	input.Wrapping = fyne.TextWrapWord
	// 设置输入框的最小高度，确保长文本能够正常显示
	input.Resize(fyne.NewSize(400, 120))
	//build tab
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("coder", theme.ZoomInIcon(), CoderStructure(input)),
		container.NewTabItemWithIcon("certificate", theme.InfoIcon(), CertificateStructure(input)),
		container.NewTabItemWithIcon("asn1", theme.ZoomInIcon(), Asn1Structure(input)),
		container.NewTabItemWithIcon("key", theme.ColorChromaticIcon(), KeyStructure(input)),
		container.NewTabItemWithIcon("envelop", theme.FolderIcon(), SM2EnvelopedPfxStructure(input)),
		container.NewTabItemWithIcon("p12", theme.AccountIcon(), SM2PfxStructure(input)),
		//container.NewTabItemWithIcon("timestamp", theme.AccountIcon(), TimestampStructure(input)),
		container.NewTabItemWithIcon("crl", theme.AccountIcon(), CrlStructure(input)),
	)
	//填充布局
	body := container.NewVBox(
		centerLink,
		rightTime,
		tabs,
	)
	return body

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
