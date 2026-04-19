@echo off
echo Starting Pomclaw Gateway with OpenClaw UI Integration...
echo.

REM 设置基本环境变量
set POMCLAW_GATEWAY_HOST=0.0.0.0
set POMCLAW_GATEWAY_PORT=8080

REM 启动gateway
pomclaw.exe gateway

pause
