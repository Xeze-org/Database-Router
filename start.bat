@echo off
echo ====================================
echo Go Database Manager
echo ====================================
echo.
echo Installing dependencies...
go mod tidy
echo.
echo Starting server...
echo.
go build -o db-router.exe ./cmd/ && db-router.exe
