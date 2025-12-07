@echo off
chcp 65001
REM 设置窗口标题
title Wails Production Build - Optimized

REM ----------------------------------------------------------------------
REM ** 1. 极致压缩和优化参数 **
REM ----------------------------------------------------------------------
REM -ldflags "-s -w": 移除符号表和调试信息，最大限度减小Go二进制文件体积。
REM -trimpath: 移除生成的可执行文件中的所有文件系统路径。
REM -upx: 使用 UPX 压缩最终的二进制文件。确保您的系统已安装 UPX。
REM -upxflags "--best": 告诉 UPX 使用最佳压缩级别。
REM ----------------------------------------------------------------------

SET WAILS_BUILD_FLAGS=-ldflags "-s -w" -trimpath -upx -upxflags "--best" -platform windows/amd64

echo.
echo ==========================================================
echo    🚀 正在执行 Wails 生产构建 (极致优化)
echo ==========================================================
echo.
echo 构建参数: %WAILS_BUILD_FLAGS%
echo.

REM 执行 Wails 构建命令
wails build %WAILS_BUILD_FLAGS%

IF %ERRORLEVEL% NEQ 0 (
    echo.
    echo ==========================================================
    echo    ❌ 构建失败！请检查错误信息。
    echo ==========================================================
    goto :end
)

echo.
echo ==========================================================
echo    ✅ Wails 应用已成功构建和压缩！
echo    最终文件在: build\bin\
echo ==========================================================

:end
echo.
pause