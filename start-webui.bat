@echo off
title db-router Web UI
echo.
echo  db-router Web UI Test Panel
echo  ----------------------------
echo  Building...
go build -o webui.exe ./cmd/webui
if %errorlevel% neq 0 ( echo  Build failed! & pause & exit /b 1 )

echo  Starting on http://localhost:8080
echo  Press Ctrl+C to stop.
echo.
webui.exe
