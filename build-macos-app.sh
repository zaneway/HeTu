#!/bin/bash

# you need install some tools
#  go get fyne.io/fyne/v2/cmd/fyne@latest

echo start reflash go.sum
go mod tidy
echo start package
fyne package -os darwin -icon Icon.png
echo package over
echo ---------------------------------

project_path=$(cd `dirname $0`; pwd)
project_name="${project_path##*/}"

echo 当前项目名称: $project_name


appName=$project_name.app

# 自定义目录名称
MacOSDirName=MacOS-App

echo 开始删除无效目录: $MacOSDirName
rm -fr $MacOSDirName

rm -f $project_name.dmg



# build new package
mkdir $MacOSDirName

echo 开始适配安装目录
# 相关打包数据，拷贝到同一个目录下
cp -R $appName $MacOSDirName
ln -s /Applications ./$MacOSDirName/Applications
# 开始打包
echo start build $project_name.dmg
hdiutil create -volname $project_name -srcfolder $MacOSDirName -ov -format UDZO "./$project_name.dmg"

echo over