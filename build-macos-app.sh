#!/bin/bash
# HeTu 密码学工具箱 macOS 应用打包脚本
# 依赖: fyne CLI 工具

set -e  # 遇到错误立即退出

# 颜色输出函数
print_info() {
    echo -e "\033[36m[INFO]\033[0m $1"
}

print_success() {
    echo -e "\033[32m[SUCCESS]\033[0m $1"
}

print_error() {
    echo -e "\033[31m[ERROR]\033[0m $1"
}

print_warning() {
    echo -e "\033[33m[WARNING]\033[0m $1"
}

# 项目信息
APP_NAME="HeTu"
APP_VERSION="1.0.8"
APP_DISPLAY_NAME="HeTu 密码学工具箱"
APP_BUNDLE_ID="com.hetu.cryptotoolbox"

# 获取项目路径和名称
project_path=$(cd `dirname $0`; pwd)
project_name="${project_path##*/}"

print_info "🔐 开始构建 $APP_DISPLAY_NAME macOS 应用"
print_info "项目路径: $project_path"
print_info "应用版本: $APP_VERSION"

# 检查必要工具
check_dependencies() {
    print_info "检查构建依赖..."
    
    # 检查Go环境
    if ! command -v go &> /dev/null; then
        print_error "Go未安装或未在PATH中"
        exit 1
    fi
    
    # 检查fyne CLI
    if ! command -v fyne &> /dev/null; then
        print_error "fyne CLI未安装，请运行: go install fyne.io/fyne/v2/cmd/fyne@latest"
        exit 1
    fi
    
    print_success "依赖检查完成"
}

# 清理旧文件
cleanup_old_files() {
    print_info "清理旧的构建文件..."
    
    # 删除旧的app文件
    if [ -d "$project_name.app" ]; then
        print_info "删除旧的应用包: $project_name.app"
        rm -rf "$project_name.app"
    fi
    
    # 删除旧的dmg文件
    if [ -f "$project_name.dmg" ]; then
        print_info "删除旧的DMG文件: $APP_NAME*.dmg"
        rm -f "$project_name.dmg"
    fi
    
    # 删除旧的构建目录
    if [ -d "MacOS-App" ]; then
        print_info "删除旧的构建目录: MacOS-App"
        rm -rf "MacOS-App"
    fi
    
    print_success "清理完成"
}

# 创建应用图标（如果不存在）
create_app_icon() {
    if [ ! -f "Icon.png" ]; then
        print_warning "未找到 Icon.png，将创建默认图标"
        
        # 创建一个简单的默认图标
        cat > "create_icon.py" << 'EOF'
from PIL import Image, ImageDraw, ImageFont
import os

# 创建512x512的图标
size = 512
img = Image.new('RGBA', (size, size), (52, 152, 219, 255))  # 蓝色背景
draw = ImageDraw.Draw(img)

# 绘制简单的密钥图标
# 绘制钥匙柄
handle_size = size // 4
handle_x = size // 2 - handle_size // 2
handle_y = size // 2 - handle_size // 2
draw.ellipse([handle_x, handle_y, handle_x + handle_size, handle_y + handle_size], 
            fill=(255, 255, 255, 255), outline=(0, 0, 0, 255), width=8)

# 绘制钥匙杆
shaft_width = size // 16
shaft_height = size // 3
shaft_x = size // 2 - shaft_width // 2
shaft_y = handle_y + handle_size
draw.rectangle([shaft_x, shaft_y, shaft_x + shaft_width, shaft_y + shaft_height],
              fill=(255, 255, 255, 255), outline=(0, 0, 0, 255), width=4)

# 绘制钥匙齿
tooth_width = size // 12
tooth_height = size // 24
tooth_x = shaft_x + shaft_width
tooth_y1 = shaft_y + shaft_height - tooth_height * 2
tooth_y2 = shaft_y + shaft_height - tooth_height

draw.rectangle([tooth_x, tooth_y1, tooth_x + tooth_width, tooth_y1 + tooth_height],
              fill=(255, 255, 255, 255), outline=(0, 0, 0, 255), width=2)
draw.rectangle([tooth_x, tooth_y2, tooth_x + tooth_width // 2, tooth_y2 + tooth_height],
              fill=(255, 255, 255, 255), outline=(0, 0, 0, 255), width=2)

# 保存图标
img.save('Icon.png', 'PNG')
print("默认图标已创建: Icon.png")
EOF

        # 尝试使用Python创建图标
        if command -v python3 &> /dev/null; then
            python3 -c "from PIL import Image, ImageDraw; img = Image.new('RGBA', (512, 512), (52, 152, 219, 255)); draw = ImageDraw.Draw(img); draw.ellipse([128, 128, 384, 384], fill=(255, 255, 255, 255), outline=(0, 0, 0, 255), width=16); img.save('Icon.png', 'PNG')"
            print_success "默认图标已创建"
        else
            print_warning "无法创建默认图标，将使用系统默认图标"
        fi
        
        # 清理临时文件
        [ -f "create_icon.py" ] && rm -f "create_icon.py"
    else
        print_success "找到应用图标: Icon.png"
    fi
}

# 构建应用
build_app() {
    print_info "更新Go模块依赖..."
    go mod tidy
    
    print_info "开始打包macOS应用..."
    
    # 设置构建参数
    local icon_param=""
    if [ -f "Icon.png" ]; then
        icon_param="-icon Icon.png"
    fi
    
    # 执行fyne打包
    fyne package -os darwin $icon_param -name "$APP_NAME" --source-dir . --app-id "$APP_BUNDLE_ID" --app-version "$APP_VERSION"
    
    if [ $? -eq 0 ] && [ -d "$APP_NAME.app" ]; then
        print_success "应用打包完成: $APP_NAME.app"
        
        # 显示应用信息
        app_size=$(du -sh "$APP_NAME.app" | cut -f1)
        print_info "应用大小: $app_size"
    else
        print_error "应用打包失败"
        exit 1
    fi
}

# 创建DMG安装包
create_dmg() {
    print_info "创建DMG安装包..."
    
    # 创建临时目录
    local dmg_dir="MacOS-App"
    mkdir -p "$dmg_dir"
    
    # 复制应用到临时目录
    cp -R "$APP_NAME.app" "$dmg_dir/"
    
    # 创建Applications快捷方式
    ln -s /Applications "$dmg_dir/Applications"
    
    # 创建DMG文件
    local dmg_name="$APP_NAME-$APP_VERSION-macOS.dmg"
    
    print_info "生成DMG文件: $dmg_name"
    hdiutil create -volname "$APP_DISPLAY_NAME" -srcfolder "$dmg_dir" -ov -format UDZO "$dmg_name"
    
    if [ $? -eq 0 ] && [ -f "$dmg_name" ]; then
        print_success "DMG安装包创建完成: $dmg_name"
        
        # 显示文件大小
        dmg_size=$(du -sh "$dmg_name" | cut -f1)
        print_info "DMG大小: $dmg_size"
        
        # 清理临时目录
        rm -rf "$dmg_dir"
    else
        print_error "DMG创建失败"
        exit 1
    fi
}

# 显示构建结果
show_build_results() {
    echo
    print_success "🎉 macOS构建完成！"
    echo
    echo "构建产物："
    echo "  📱 应用包: $APP_NAME.app"
    echo "  💿 安装包: $APP_NAME-$APP_VERSION-macOS.dmg"
    echo
    echo "使用方法："
    echo "  1. 直接运行: 双击 $APP_NAME.app"
    echo "  2. 安装使用: 双击 $APP_NAME-$APP_VERSION-macOS.dmg，拖拽到Applications文件夹"
    echo
    echo "📋 应用信息："
    echo "  名称: $APP_DISPLAY_NAME"
    echo "  版本: $APP_VERSION"
    echo "  Bundle ID: $APP_BUNDLE_ID"
}

# 主函数
main() {
    echo "🔐 HeTu 密码学工具箱 - macOS构建脚本"
    echo "======================================"
    echo
    
    check_dependencies
    cleanup_old_files
    create_app_icon
    build_app
    create_dmg
    show_build_results
}

# 运行主函数
main "$@"