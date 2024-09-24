package window

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	_ "github.com/lengzhao/font/autoload" //这个可以让你识别中文
	"time"
)

func NewWindow() {
	myApp := app.New()
	// 创建一个窗口对象
	myWindow := myApp.NewWindow("证书解析客户端")
	body := newBody()
	myWindow.SetContent(body)
	myWindow.Resize(fyne.Size{800, 600})
	myWindow.ShowAndRun()

}

func newBody() *fyne.Container {
	// 表头
	//url, _ := url.Parse("https://github.com/zaneway/CertViewer/tree/bcja/v1.0.0-pqc")
	//link := widget.NewHyperlink("^-^  欢迎使用证书解析客户端  ^-^", url)
	////超链接显示在中间
	//centerLink := container.NewCenter(link)
	//时间显示在最右侧
	rightTime := container.NewHBox(layout.NewSpacer(), refreshTimeSeconds())

	//填充布局
	body := container.NewVBox(
		//centerLink,
		rightTime,
		Structure(),
	)
	return body

}

const DateTime = "2006-01-02 15:04:05"

func refreshTimeSeconds() *widget.Label {
	//填充当前时间
	nowTime := widget.NewLabel(time.Now().Format(DateTime))
	//异步线程更新时间
	go func() {
		for range time.Tick(time.Second) {
			nowTime.SetText(time.Now().Format(DateTime))
		}
	}()
	return nowTime
}
