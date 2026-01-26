#!/bin/bash
# Windows安装包打包脚本
# 依赖: mingw-w64 (brew install mingw-w64), makensis (brew install makensis)

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

# 检查必要工具
check_dependencies() {
    print_info "检查构建依赖..."
    
    # 检查Go环境
    if ! command -v go &> /dev/null; then
        print_error "Go未安装或未在PATH中"
        exit 1
    fi
    
    # 检查mingw-w64
    if ! command -v x86_64-w64-mingw32-gcc &> /dev/null; then
        print_error "mingw-w64未安装，请运行: brew install mingw-w64"
        exit 1
    fi
    
    # 检查NSIS
    if ! command -v makensis &> /dev/null; then
        print_warning "makensis未安装，将只生成exe文件，不创建安装包"
        print_warning "如需创建安装包，请运行: brew install makensis"
        NSIS_AVAILABLE=false
    else
        NSIS_AVAILABLE=true
    fi
    
    print_success "依赖检查完成"
}

# 构建Windows可执行文件
build_executable() {
    local dir=$(pwd)
    
    print_info "准备构建Windows可执行文件..."
    
    # 清理依赖
    print_info "清理Go模块依赖..."
    go mod tidy
    
    # 删除旧文件
    if [ -f "$dir/HeTu.exe" ]; then
        print_info "删除旧的可执行文件: $dir/HeTu.exe"
        rm -f "$dir/HeTu.exe"
    fi
    
    # 创建构建目录
    mkdir -p "$dir/dist/windows"
    
    print_info "开始交叉编译Windows可执行文件..."
    GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
        go build -ldflags "-H windowsgui -s -w" -o "$dir/dist/windows/HeTu.exe"
    
    if [ $? -eq 0 ]; then
        print_success "Windows可执行文件构建完成: $dir/dist/windows/HeTu.exe"
    else
        print_error "Windows可执行文件构建失败"
        exit 1
    fi
}

# 转换PNG图标为ICO格式
convert_icon() {
    local dir=$(pwd)
    local png_icon="$dir/Icon.png"
    local ico_icon="$dir/dist/windows/Icon.ico"
    
    if [ -f "$png_icon" ]; then
        print_info "检查图标转换工具..."
        
        # 检查是否有ImageMagick
        if command -v convert &> /dev/null; then
            print_info "使用ImageMagick转换PNG为ICO格式..."
            convert "$png_icon" -resize 256x256 "$ico_icon"
            if [ $? -eq 0 ]; then
                print_success "图标转换完成: $ico_icon"
                return 0
            else
                print_warning "ImageMagick转换失败，将复制原PNG文件"
            fi
        else
            print_warning "ImageMagick未安装，建议安装: brew install imagemagick"
        fi
        
        # 如果转换失败或没有ImageMagick，直接复制PNG
        print_info "复制PNG图标文件..."
        cp "$png_icon" "$dir/dist/windows/Icon.png"
        return 1
    else
        print_warning "Icon.png文件不存在"
        return 1
    fi
}

# 创建NSIS安装脚本
create_nsis_script() {
    local dir=$(pwd)
    local nsis_script="$dir/dist/windows/installer.nsi"
    local has_ico_icon=false
    
    # 检查是否成功创建了ICO图标
    if [ -f "$dir/dist/windows/Icon.ico" ]; then
        has_ico_icon=true
        print_info "将使用ICO格式图标"
    fi
    
    print_info "创建NSIS安装脚本..."
    
    cat > "$nsis_script" << EOF
; HeTu 密码学工具箱 Windows 安装脚本
; 使用 NSIS 3.0+ 编译

; 安装包基本信息
!define APP_NAME "HeTu"
!define APP_VERSION "1.0.8"
!define APP_PUBLISHER "HeTu Development Team"
!define APP_URL "https://github.com/zaneway/HeTu"
!define APP_DESCRIPTION "河图密码学工具箱 - 可视化密码学操作平台"
!define APP_EXE "HeTu.exe"

; 安装包属性
Name "\${APP_NAME} \${APP_VERSION}"
OutFile "HeTu-\${APP_VERSION}-Setup.exe"
InstallDir "\$PROGRAMFILES64\\\${APP_NAME}"
InstallDirRegKey HKLM "Software\\\${APP_NAME}" "InstallDir"

; 请求管理员权限
RequestExecutionLevel admin

; 压缩算法
SetCompressor lzma

; 现代UI
!include "MUI2.nsh"
!include "FileFunc.nsh"

; UI设置
!define MUI_ABORTWARNING
EOF

    # 根据是否有ICO图标设置不同的图标路径
    if [ "$has_ico_icon" = true ]; then
        cat >> "$nsis_script" << EOF
!define MUI_ICON "Icon.ico"
!define MUI_UNICON "Icon.ico"
EOF
    else
        cat >> "$nsis_script" << EOF
!define MUI_ICON "\${NSISDIR}\\Contrib\\Graphics\\Icons\\modern-install.ico"
!define MUI_UNICON "\${NSISDIR}\\Contrib\\Graphics\\Icons\\modern-uninstall.ico"
EOF
    fi

    cat >> "$nsis_script" << 'EOF'

; 安装页面
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "license.txt"
!insertmacro MUI_PAGE_COMPONENTS
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

; 卸载页面
!insertmacro MUI_UNPAGE_WELCOME
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

; 语言文件
!insertmacro MUI_LANGUAGE "SimpChinese"
!insertmacro MUI_LANGUAGE "English"

; 版本信息
VIProductVersion "1.0.0.0"
VIAddVersionKey /LANG=${LANG_SIMPCHINESE} "ProductName" "${APP_NAME}"
VIAddVersionKey /LANG=${LANG_SIMPCHINESE} "Comments" "${APP_DESCRIPTION}"
VIAddVersionKey /LANG=${LANG_SIMPCHINESE} "CompanyName" "${APP_PUBLISHER}"
VIAddVersionKey /LANG=${LANG_SIMPCHINESE} "LegalCopyright" "© 2024 ${APP_PUBLISHER}"
VIAddVersionKey /LANG=${LANG_SIMPCHINESE} "FileDescription" "${APP_DESCRIPTION}"
VIAddVersionKey /LANG=${LANG_SIMPCHINESE} "FileVersion" "${APP_VERSION}"
VIAddVersionKey /LANG=${LANG_SIMPCHINESE} "ProductVersion" "${APP_VERSION}"
VIAddVersionKey /LANG=${LANG_SIMPCHINESE} "InternalName" "${APP_NAME}"
VIAddVersionKey /LANG=${LANG_SIMPCHINESE} "OriginalFilename" "${APP_EXE}"

; 路径处理函数
!include "WinMessages.nsh"

Function AddToPath
  Exch $0
  Push $1
  Push $2
  Push $3
  
  ; 读取系统PATH
  ReadRegStr $1 HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "PATH"
  Push "$1;"
  Push "$0;"
  Call StrStr
  Pop $2
  StrCmp $2 "" "" AddToPath_done
  
  ; 如果路径不存在，则添加
  StrCmp $1 "" AddToPath_NTdoIt
  StrCpy $2 "$1;$0"
  Goto AddToPath_NTdoIt
  
  AddToPath_NTdoIt:
    WriteRegExpandStr HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "PATH" $2
    SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=5000
    
  AddToPath_done:
    Pop $3
    Pop $2
    Pop $1
    Pop $0
FunctionEnd

Function un.RemoveFromPath
  Exch $0
  Push $1
  Push $2
  Push $3
  Push $4
  Push $5
  Push $6
  
  ; 读取系统PATH
  ReadRegStr $1 HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "PATH"
  StrCpy $5 $1 1 -1
  StrCmp $5 ";" +2
  StrCpy $1 "$1;"
  Push $1
  Push "$0;"
  Call un.StrStr
  Pop $2
  StrCmp $2 "" unRemoveFromPath_done
  
  ; 移除路径
  StrLen $3 "$0;"
  StrLen $4 $2
  StrCpy $5 $1 -$4
  StrCpy $6 $2 "" $3
  StrCpy $3 $5$6
  
  StrCpy $5 $3 1 -1
  StrCmp $5 ";" 0 +2
  StrCpy $3 $3 -1
  
  WriteRegExpandStr HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "PATH" $3
  SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=5000
  
  unRemoveFromPath_done:
    Pop $6
    Pop $5
    Pop $4
    Pop $3
    Pop $2
    Pop $1
    Pop $0
FunctionEnd

; 字符串查找函数
Function StrStr
  Exch $R1
  Exch
  Exch $R2
  Push $R3
  Push $R4
  Push $R5
  StrLen $R3 $R1
  StrCpy $R4 0
  loop:
    StrCpy $R5 $R2 $R3 $R4
    StrCmp $R5 $R1 done
    StrCmp $R5 "" done
    IntOp $R4 $R4 + 1
    Goto loop
  done:
  StrCpy $R1 $R2 "" $R4
  Pop $R5
  Pop $R4
  Pop $R3
  Pop $R2
  Exch $R1
FunctionEnd

Function un.StrStr
  Exch $R1
  Exch
  Exch $R2
  Push $R3
  Push $R4
  Push $R5
  StrLen $R3 $R1
  StrCpy $R4 0
  loop:
    StrCpy $R5 $R2 $R3 $R4
    StrCmp $R5 $R1 done
    StrCmp $R5 "" done
    IntOp $R4 $R4 + 1
    Goto loop
  done:
  StrCpy $R1 $R2 "" $R4
  Pop $R5
  Pop $R4
  Pop $R3
  Pop $R2
  Exch $R1
FunctionEnd

; 安装组件
Section "!核心程序" SecCore
    SectionIn RO  ; 只读，必须安装
    
    SetOutPath "$INSTDIR"
    
    ; 安装主程序
    File "HeTu.exe"
    
EOF

    # 根据是否有ICO图标添加相应的文件
    if [ "$has_ico_icon" = true ]; then
        cat >> "$nsis_script" << EOF
    ; 安装图标文件
    File "Icon.ico"
    File "Icon.png"
EOF
    else
        cat >> "$nsis_script" << EOF
    ; 安装图标文件
    File "Icon.png"
EOF
    fi

    cat >> "$nsis_script" << 'EOF'
    
    ; 创建卸载程序
    WriteUninstaller "$INSTDIR\Uninstall.exe"
    
    ; 写入注册表
    WriteRegStr HKLM "Software\${APP_NAME}" "InstallDir" "$INSTDIR"
    WriteRegStr HKLM "Software\${APP_NAME}" "Version" "${APP_VERSION}"
    
    ; 添加到控制面板的程序列表
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "DisplayName" "${APP_NAME} ${APP_VERSION}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "UninstallString" "$INSTDIR\Uninstall.exe"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "QuietUninstallString" "$INSTDIR\Uninstall.exe /S"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "InstallLocation" "$INSTDIR"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "DisplayIcon" "$INSTDIR\${APP_EXE}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "Publisher" "${APP_PUBLISHER}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "DisplayVersion" "${APP_VERSION}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "URLInfoAbout" "${APP_URL}"
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "NoModify" 1
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "NoRepair" 1
    
    ; 计算安装大小
    ${GetSize} "$INSTDIR" "/S=0K" $0 $1 $2
    IntFmt $0 "0x%08X" $0
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "EstimatedSize" "$0"
SectionEnd

Section "桌面快捷方式" SecDesktop
EOF

    # 根据是否有ICO图标设置不同的快捷方式创建方式
    if [ "$has_ico_icon" = true ]; then
        cat >> "$nsis_script" << 'EOF'
    ; 优先使用Icon.ico作为图标，如果不存在则使用exe图标
    IfFileExists "$INSTDIR\Icon.ico" 0 +3
        CreateShortCut "$DESKTOP\${APP_NAME}.lnk" "$INSTDIR\${APP_EXE}" "" "$INSTDIR\Icon.ico" 0
        Goto +2
    CreateShortCut "$DESKTOP\${APP_NAME}.lnk" "$INSTDIR\${APP_EXE}" "" "$INSTDIR\${APP_EXE}" 0
EOF
    else
        cat >> "$nsis_script" << 'EOF'
    ; 使用exe内置图标（PNG在快捷方式中显示效果不佳）
    CreateShortCut "$DESKTOP\${APP_NAME}.lnk" "$INSTDIR\${APP_EXE}" "" "$INSTDIR\${APP_EXE}" 0
EOF
    fi

    cat >> "$nsis_script" << 'EOF'
SectionEnd

Section "开始菜单快捷方式" SecStartMenu
    CreateDirectory "$SMPROGRAMS\${APP_NAME}"
EOF

    # 根据是否有ICO图标设置不同的快捷方式创建方式
    if [ "$has_ico_icon" = true ]; then
        cat >> "$nsis_script" << 'EOF'
    ; 优先使用Icon.ico作为图标，如果不存在则使用exe图标
    IfFileExists "$INSTDIR\Icon.ico" 0 +3
        CreateShortCut "$SMPROGRAMS\${APP_NAME}\${APP_NAME}.lnk" "$INSTDIR\${APP_EXE}" "" "$INSTDIR\Icon.ico" 0
        Goto +2
    CreateShortCut "$SMPROGRAMS\${APP_NAME}\${APP_NAME}.lnk" "$INSTDIR\${APP_EXE}" "" "$INSTDIR\${APP_EXE}" 0
EOF
    else
        cat >> "$nsis_script" << 'EOF'
    ; 使用exe内置图标（PNG在快捷方式中显示效果不佳）
    CreateShortCut "$SMPROGRAMS\${APP_NAME}\${APP_NAME}.lnk" "$INSTDIR\${APP_EXE}" "" "$INSTDIR\${APP_EXE}" 0
EOF
    fi

    cat >> "$nsis_script" << 'EOF'
    CreateShortCut "$SMPROGRAMS\${APP_NAME}\卸载 ${APP_NAME}.lnk" "$INSTDIR\Uninstall.exe" "" "$INSTDIR\Uninstall.exe" 0
SectionEnd

Section "添加到PATH环境变量" SecPath
    ; 添加到系统PATH
    Push "$INSTDIR"
    Call AddToPath
SectionEnd

; 组件描述
!insertmacro MUI_FUNCTION_DESCRIPTION_BEGIN
    !insertmacro MUI_DESCRIPTION_TEXT ${SecCore} "${APP_NAME} 核心程序文件（必需）"
    !insertmacro MUI_DESCRIPTION_TEXT ${SecDesktop} "在桌面创建快捷方式"
    !insertmacro MUI_DESCRIPTION_TEXT ${SecStartMenu} "在开始菜单创建程序组"
    !insertmacro MUI_DESCRIPTION_TEXT ${SecPath} "将程序目录添加到系统PATH环境变量"
!insertmacro MUI_FUNCTION_DESCRIPTION_END

; 安装前检查
Function .onInit
    ; 检查是否已安装
    ReadRegStr $R0 HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "UninstallString"
    StrCmp $R0 "" done
    
    MessageBox MB_OKCANCEL|MB_ICONEXCLAMATION "${APP_NAME} 已经安装。点击确定卸载旧版本，或点击取消退出安装。" /SD IDCANCEL IDOK uninst
    Abort
    
    uninst:
        ClearErrors
        ExecWait '$R0 /S _?=$INSTDIR'
        
        IfErrors no_remove_uninstaller done
        no_remove_uninstaller:
    
    done:
FunctionEnd

; 卸载程序
Section "Uninstall"
    ; 删除程序文件
    Delete "$INSTDIR\${APP_EXE}"
    Delete "$INSTDIR\Uninstall.exe"
    Delete "$INSTDIR\Icon.png"
    Delete "$INSTDIR\Icon.ico"
    
    ; 删除快捷方式
    Delete "$DESKTOP\${APP_NAME}.lnk"
    Delete "$SMPROGRAMS\${APP_NAME}\${APP_NAME}.lnk"
    Delete "$SMPROGRAMS\${APP_NAME}\卸载 ${APP_NAME}.lnk"
    RMDir "$SMPROGRAMS\${APP_NAME}"
    
    ; 从PATH中移除
    Push "$INSTDIR"
    Call un.RemoveFromPath
    
    ; 删除注册表项
    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}"
    DeleteRegKey HKLM "Software\${APP_NAME}"
    
    ; 删除安装目录
    RMDir "$INSTDIR"
    
    ; 如果安装目录为空则删除
    RMDir "$PROGRAMFILES64\${APP_NAME}"
SectionEnd
EOF

    print_success "NSIS安装脚本创建完成: $nsis_script"
}

# 创建许可证文件
create_license_file() {
    local dir=$(pwd)
    local license_file="$dir/dist/windows/license.txt"
    
    print_info "创建许可证文件..."
    
    cat > "$license_file" << 'EOF'
HeTu 密码学工具箱
软件许可协议

版权所有 (c) 2024 HeTu Development Team

本软件按"原样"提供，不提供任何明示或暗示的保证，包括但不限于
对适销性、特定用途适用性和非侵权性的保证。

在任何情况下，作者或版权持有者均不对任何索赔、损害或其他责任负责，
无论是在合同、侵权或其他行为中产生的，由软件或软件的使用或其他
交易引起的或与之相关的。

使用条款：
1. 本软件仅供学习和研究使用
2. 禁止用于任何非法目的
3. 使用者需遵守当地法律法规
4. 作者不承担因使用本软件造成的任何损失

技术支持：
项目主页：https://github.com/zaneway/HeTu
EOF

    print_success "许可证文件创建完成: $license_file"
}

# 构建安装包
build_installer() {
    if [ "$NSIS_AVAILABLE" = true ]; then
        local dir=$(pwd)
        
        print_info "开始构建Windows安装包..."
        
        # 转换并复制图标文件
        convert_icon
        
        # 进入构建目录
        cd "$dir/dist/windows"
        
        # 执行NSIS编译
        makensis installer.nsi
        
        if [ $? -eq 0 ]; then
            print_success "Windows安装包构建完成: $dir/dist/windows/HeTu-1.0.0-Setup.exe"
            
            # 移动到根目录
            mv "HeTu-1.0.0-Setup.exe" "$dir/"
            print_success "安装包已移动到: $dir/HeTu-1.0.0-Setup.exe"
        else
            print_error "Windows安装包构建失败"
            exit 1
        fi
        
        # 返回原目录
        cd "$dir"
    else
        print_warning "跳过安装包构建（makensis未安装）"
    fi
}

# 显示构建结果
show_build_results() {
    local dir=$(pwd)
    
    echo
    print_success "🎉 Windows构建完成！"
    echo
    echo "构建产物："
    echo "  📦 可执行文件: $dir/dist/windows/HeTu.exe"
    
    if [ "$NSIS_AVAILABLE" = true ] && [ -f "$dir/HeTu-1.0.0-Setup.exe" ]; then
        echo "  📦 安装包: $dir/HeTu-1.0.0-Setup.exe"
    fi
    
    echo
    echo "使用方法："
    echo "  1. 直接运行: dist/windows/HeTu.exe"
    if [ "$NSIS_AVAILABLE" = true ] && [ -f "$dir/HeTu-1.0.0-Setup.exe" ]; then
        echo "  2. 安装后使用: 双击 HeTu-1.0.0-Setup.exe 进行安装"
    fi
    
}

# 主函数
main() {
    echo "🔐 HeTu 密码学工具箱 - Windows构建脚本"
    echo "========================================"
    echo
    
    check_dependencies
    build_executable
    
    if [ "$NSIS_AVAILABLE" = true ]; then
        convert_icon
        create_nsis_script
        create_license_file
        build_installer
    fi
    
    show_build_results
}

# 运行主函数
main "$@"