package window

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	_ "github.com/lengzhao/font/autoload" //这个可以让你识别中文
	"time"
)

func NewWindow() {
	myApp := app.New()
	// 创建一个窗口对象
	myWindow := myApp.NewWindow("Cert Viewer")
	body := newBody()
	myWindow.SetContent(body)
	myWindow.ShowAndRun()

}

func newBody() *fyne.Container {
	// 表头
	helloGay := widget.NewLabel("欢迎访问全球最大的同性交友网站： https://github.com/zaneway/CertViewer ")
	body := container.NewVBox(
		helloGay,
		Structure(),
		refreshTimeSeconds(),
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
