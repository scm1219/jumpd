@echo off
setlocal enabledelayedexpansion

:: jumpd build and install script (requires Administrator)
:: Usage: build-install.bat [output_name]

:: Locate project root (script directory)
set "ROOT=%~dp0"
cd /d "%ROOT%"

:: Check for Administrator privileges
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Administrator privileges required.
    echo         Please right-click and select "Run as administrator".
    pause
    exit /b 1
)

set "OUTPUT=%~1"
if "%OUTPUT%"=="" set "OUTPUT=jumpd.exe"
set "FULL_OUTPUT=%ROOT%%OUTPUT%"
set "DEST=C:\Windows\%OUTPUT%"

:: Step 1: Build with optimizations
echo [1/3] Building %FULL_OUTPUT% ...
go build -trimpath -ldflags="-s -w" -o "%FULL_OUTPUT%" .
if %errorlevel% neq 0 (
    echo [ERROR] Build failed.
    pause
    exit /b 1
)
echo [1/3] Build successful: %FULL_OUTPUT%

:: Step 2: Check UPX and compress
where upx >nul 2>&1
if %errorlevel% neq 0 (
    echo [2/3] UPX not found, skipping compression.
    goto :install
)

echo [2/3] Compressing with UPX ...
upx --best --lzma "%FULL_OUTPUT%"
if %errorlevel% neq 0 (
    echo [WARN] UPX compression failed, keeping uncompressed binary.
    goto :install
)
echo [2/3] Compression successful.
goto :install

:install
:: Step 3: Copy to C:\Windows
echo [3/3] Installing to %DEST% ...
copy /y "%FULL_OUTPUT%" "%DEST%" >nul
if %errorlevel% neq 0 (
    echo [ERROR] Failed to copy to %DEST%.
    pause
    exit /b 1
)

echo [DONE] Installed: %DEST%
for %%F in ("%DEST%") do echo        Size: %%~zF bytes
pause
endlocal
