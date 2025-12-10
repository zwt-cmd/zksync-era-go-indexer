@echo off
chcp 65001 >nul

echo 🚀 启动 Docker 服务 (Windows)
echo.

docker compose -f docker\docker-compose.windows.yml up -d

echo.
echo ✅ 启动完成！
echo.
echo MySQL:         localhost:3307 (用户: scanner / 密码: scannerpass)
echo Redis:         localhost:6380
echo phpMyAdmin:    http://localhost:8090
echo Redis管理:     http://localhost:8091
echo.

pause

