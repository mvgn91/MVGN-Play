@echo off
cd /d "%~dp0.."
echo.
echo  === MVGN Observer ===
echo  Terminal + Dashboard visual en http://localhost:8080
echo  Ctrl+C para cerrar.
echo.
.\mvgn-scan\mvgn-scan.exe -watch -port 8080
pause
