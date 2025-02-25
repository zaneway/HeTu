# HeTu(河图)
* 河图与洛书是中国古代传说中上天授予的祥瑞之兆.



# 简单介绍
    本工程使用go语言开发, 基于使用 fyne 构建的一个可视化工具.
        当前提供以下几种模式:
             1. 解析X509证书
             2. 解析Asn1编码(正在完善)
             3. Base64/Hex转换
             4. 支持产生密钥，P10，证书，摘要
             5. 验证证书合规性
             6. C1C3C2转换
             7. 尝试解析各种信封
             8. pfx
             x. ...(构思中)


# 快速开始
1. 安装依赖库：go get fyne.io/fyne/v2/cmd/fyne@latest
2. 安装包构建脚本：
```shell
# 提供了Windows、MacOS的打包脚本,直接执行即可.
sh build-macos-app.sh

sh build-windows-app.sh

```
3. 从指定目录获取服务包完成安装
