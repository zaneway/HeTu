#!/bin/bash
# 需要安装 mingw-w64
dir=$(pwd)
go mod tidy
echo 准备删除 $dir/CertViewer.exe
rm -f $dir/CertViewer.exe
echo 删除完成,开始打包
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -ldflags "-H windowsgui" -o CertViewer.exe

echo over