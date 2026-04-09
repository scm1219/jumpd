@echo off
setlocal enabledelayedexpansion

:: jumpd build script for Windows
:: Usage: build.bat [output_name]

:: Locate project root (script directory)
set "ROOT=%~dp0"
cd /d "%ROOT%"

set "OUTPUT=%~1"
if "%OUTPUT%"=="" set "OUTPUT=jumpd.exe"
set "FULL_OUTPUT=%ROOT%%OUTPUT%"

:: Step 1: Build with optimizations
echo [1/2] Building %FULL_OUTPUT% ...
go build -trimpath -ldflags="-s -w" -o "%FULL_OUTPUT%" .
if %errorlevel% neq 0 (
    echo [ERROR] Build failed.
    exit /b 1
)
echo [1/2] Build successful: %FULL_OUTPUT%

:: Step 2: Check UPX and compress
where upx >nul 2>&1
if %errorlevel% neq 0 (
    echo [2/2] UPX not found, skipping compression.
    goto :done
)

echo [2/2] Compressing with UPX ...
upx --best --lzma "%FULL_OUTPUT%"
if %errorlevel% neq 0 (
    echo [WARN] UPX compression failed, keeping uncompressed binary.
    goto :done
)

:done
echo [DONE] Build complete: %FULL_OUTPUT%
for %%F in ("%FULL_OUTPUT%") do echo        Size: %%~zF bytes
pause
endlocal
