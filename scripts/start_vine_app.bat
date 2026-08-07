@echo off
setlocal EnableExtensions EnableDelayedExpansion

set "SCRIPT_DIR=%~dp0"
for %%I in ("%SCRIPT_DIR%..") do set "DEFAULT_PROJECT_ROOT=%%~fI"

if not defined VINE_PROJECT_ROOT set "VINE_PROJECT_ROOT=%DEFAULT_PROJECT_ROOT%"
if not defined VINE_WEB_DIR set "VINE_WEB_DIR=%VINE_PROJECT_ROOT%\src\web"
if not defined VINE_SERVER_PACKAGE set "VINE_SERVER_PACKAGE=./src/server/cmd/demo"
if not defined VINE_PREPARE_PACKAGE set "VINE_PREPARE_PACKAGE="
if not defined VINE_HOST set "VINE_HOST=127.0.0.1"
if not defined VINE_VITE_PORT set "VINE_VITE_PORT=5174"
if not defined VINE_PUBLIC_PORT set "VINE_PUBLIC_PORT=7288"
if not defined VINE_DASHBOARD_PORT set "VINE_DASHBOARD_PORT=7299"
if not defined VINE_STARTUP_TIMEOUT set "VINE_STARTUP_TIMEOUT=45"

set "INSTALL=0"
set "CHECK_ONLY=0"
set "EXIT_CODE=0"
set "TOKEN=%RANDOM%%RANDOM%"
set "FRONTEND_TITLE=VineFrontend-%TOKEN%"
set "BACKEND_TITLE=VineServer-%TOKEN%"
set "FRONTEND_STARTED=0"
set "BACKEND_STARTED=0"

:parse_args
if "%~1"=="" goto args_done
if /I "%~1"=="--install" (
    set "INSTALL=1"
    shift
    goto parse_args
)
if /I "%~1"=="--check" (
    set "CHECK_ONLY=1"
    shift
    goto parse_args
)
if /I "%~1"=="-h" goto help
if /I "%~1"=="--help" goto help
echo [vine-start] error: unknown argument: %~1 1>&2
set "EXIT_CODE=2"
goto help_with_code

:args_done
if not exist "%VINE_PROJECT_ROOT%\go.mod" (
    echo [vine-start] error: go.mod is missing from project root: %VINE_PROJECT_ROOT% 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)
if not exist "%VINE_PROJECT_ROOT%\skel\" (
    echo [vine-start] error: hand-maintained contract directory is missing: %VINE_PROJECT_ROOT%\skel 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)
if not exist "%VINE_PROJECT_ROOT%\src\server\seed\hub.yaml" (
    echo [vine-start] error: Hub seed configuration is missing: %VINE_PROJECT_ROOT%\src\server\seed\hub.yaml 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)
if not exist "%VINE_PROJECT_ROOT%\skeled\golang\" (
    echo [vine-start] error: generated Go directory is missing: %VINE_PROJECT_ROOT%\skeled\golang 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)
if not exist "%VINE_PROJECT_ROOT%\skeled\typescript\" (
    echo [vine-start] error: generated TypeScript directory is missing: %VINE_PROJECT_ROOT%\skeled\typescript 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)
if not exist "%VINE_PROJECT_ROOT%\src\server\app\app.go" (
    echo [vine-start] error: Vine App definition is missing: %VINE_PROJECT_ROOT%\src\server\app\app.go 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)
if not exist "%VINE_PROJECT_ROOT%\src\server\core\" (
    echo [vine-start] error: business core directory is missing: %VINE_PROJECT_ROOT%\src\server\core 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)
if not exist "%VINE_PROJECT_ROOT%\src\server\impl\" (
    echo [vine-start] error: capability adapter directory is missing: %VINE_PROJECT_ROOT%\src\server\impl 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)
if not exist "%VINE_PROJECT_ROOT%\src\server\repo\" (
    echo [vine-start] error: persistence directory is missing: %VINE_PROJECT_ROOT%\src\server\repo 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)
set "VINE_SERVER_PATH=%VINE_SERVER_PACKAGE:/=\%"
if not exist "%VINE_PROJECT_ROOT%\%VINE_SERVER_PATH%\main.go" (
    echo [vine-start] error: Vine process entry is missing: %VINE_PROJECT_ROOT%\%VINE_SERVER_PATH%\main.go 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)
if not exist "%VINE_WEB_DIR%\src\" (
    echo [vine-start] error: frontend source directory is missing: %VINE_WEB_DIR%\src 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)
if not exist "%VINE_WEB_DIR%\package.json" (
    echo [vine-start] error: frontend package.json is missing: %VINE_WEB_DIR%\package.json 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)
if not exist "%VINE_WEB_DIR%\tsconfig.json" (
    echo [vine-start] error: frontend tsconfig.json is missing: %VINE_WEB_DIR%\tsconfig.json 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)

where go.exe >nul 2>nul
if errorlevel 1 (
    echo [vine-start] error: required command is unavailable on PATH: go 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)
where pnpm >nul 2>nul
if errorlevel 1 (
    echo [vine-start] error: required command is unavailable on PATH: pnpm 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)

call :validate_positive_number "VINE_VITE_PORT" "%VINE_VITE_PORT%" 65535
if errorlevel 1 goto preflight_failed
call :validate_positive_number "VINE_PUBLIC_PORT" "%VINE_PUBLIC_PORT%" 65535
if errorlevel 1 goto preflight_failed
call :validate_positive_number "VINE_DASHBOARD_PORT" "%VINE_DASHBOARD_PORT%" 65535
if errorlevel 1 goto preflight_failed
call :validate_positive_number "VINE_STARTUP_TIMEOUT" "%VINE_STARTUP_TIMEOUT%" 86400
if errorlevel 1 goto preflight_failed

if not exist "%VINE_WEB_DIR%\node_modules" if "%INSTALL%"=="0" (
    echo [vine-start] error: frontend dependencies are missing. Rerun with --install or run pnpm install in %VINE_WEB_DIR% 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)

if "%VINE_VITE_PORT%"=="%VINE_PUBLIC_PORT%" (
    echo [vine-start] error: Vite and Portal ports must be distinct 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)
if "%VINE_VITE_PORT%"=="%VINE_DASHBOARD_PORT%" (
    echo [vine-start] error: Vite and Dashboard ports must be distinct 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)
if "%VINE_PUBLIC_PORT%"=="%VINE_DASHBOARD_PORT%" (
    echo [vine-start] error: Portal and Dashboard ports must be distinct 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)

call :require_available_port "%VINE_VITE_PORT%"
if errorlevel 1 goto preflight_failed
call :require_available_port "%VINE_PUBLIC_PORT%"
if errorlevel 1 goto preflight_failed
call :require_available_port "%VINE_DASHBOARD_PORT%"
if errorlevel 1 goto preflight_failed

if "%CHECK_ONLY%"=="1" (
    echo [vine-start] preflight passed
    goto cleanup
)

if "%INSTALL%"=="1" (
    echo [vine-start] run: pnpm install
    pushd "%VINE_WEB_DIR%"
    call pnpm install
    if errorlevel 1 (
        popd
        echo [vine-start] error: pnpm install failed 1>&2
        set "EXIT_CODE=1"
        goto cleanup
    )
    popd
)

if defined VINE_PREPARE_PACKAGE (
    echo [vine-start] run: go run %VINE_PREPARE_PACKAGE%
    pushd "%VINE_PROJECT_ROOT%"
    go run "%VINE_PREPARE_PACKAGE%"
    if errorlevel 1 (
        popd
        echo [vine-start] error: preparation command failed 1>&2
        set "EXIT_CODE=1"
        goto cleanup
    )
    popd
)

echo [vine-start] start frontend: pnpm run dev
start "%FRONTEND_TITLE%" /D "%VINE_WEB_DIR%" cmd.exe /D /C "call pnpm run dev"
set "FRONTEND_STARTED=1"
call :wait_for_port "frontend" "%VINE_VITE_PORT%"
if errorlevel 1 (
    set "EXIT_CODE=1"
    goto cleanup
)

echo [vine-start] start Vine server: go run %VINE_SERVER_PACKAGE%
start "%BACKEND_TITLE%" /D "%VINE_PROJECT_ROOT%" cmd.exe /D /C "go run %VINE_SERVER_PACKAGE%"
set "BACKEND_STARTED=1"
call :wait_for_page
if errorlevel 1 (
    set "EXIT_CODE=1"
    goto cleanup
)

echo [vine-start] application: http://%VINE_HOST%:%VINE_PUBLIC_PORT%/
echo [vine-start] vRPC: http://%VINE_HOST%:%VINE_PUBLIC_PORT%/api/invoke
echo [vine-start] Dashboard: http://%VINE_HOST%:%VINE_DASHBOARD_PORT%/
echo [vine-start] press S to stop both processes.

:monitor
tasklist /V /FI "WINDOWTITLE eq %FRONTEND_TITLE%" 2>nul | findstr /C:"%FRONTEND_TITLE%" >nul
if errorlevel 1 (
    echo [vine-start] error: frontend process exited; stopping Vine server 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)
tasklist /V /FI "WINDOWTITLE eq %BACKEND_TITLE%" 2>nul | findstr /C:"%BACKEND_TITLE%" >nul
if errorlevel 1 (
    echo [vine-start] error: Vine server process exited; stopping frontend 1>&2
    set "EXIT_CODE=1"
    goto cleanup
)
choice /C SC /N /T 1 /D C >nul
if errorlevel 2 goto monitor
goto cleanup

:preflight_failed
set "EXIT_CODE=1"
goto cleanup

:validate_positive_number
echo(%~2| findstr /R /X "[1-9][0-9]*" >nul
if errorlevel 1 (
    echo [vine-start] error: %~1 must be a positive integer 1>&2
    exit /b 1
)
set /A "VALIDATED_NUMBER=%~2"
if !VALIDATED_NUMBER! GTR %~3 (
    echo [vine-start] error: %~1 exceeds the allowed maximum %~3 1>&2
    exit /b 1
)
exit /b 0

:require_available_port
netstat -ano -p tcp | findstr /R /C:":%~1 .*LISTENING" >nul
if not errorlevel 1 (
    echo [vine-start] error: required listener %VINE_HOST%:%~1 is already in use 1>&2
    exit /b 1
)
exit /b 0

:wait_for_port
set "WAIT_NAME=%~1"
set "WAIT_PORT=%~2"
set /A "WAIT_REMAINING=%VINE_STARTUP_TIMEOUT%"
:wait_for_port_loop
netstat -ano -p tcp | findstr /R /C:":%WAIT_PORT% .*LISTENING" >nul
if not errorlevel 1 (
    echo [vine-start] %WAIT_NAME% is listening on %VINE_HOST%:%WAIT_PORT%
    exit /b 0
)
if !WAIT_REMAINING! LEQ 0 (
    echo [vine-start] error: timed out waiting for %WAIT_NAME% on %VINE_HOST%:%WAIT_PORT% 1>&2
    exit /b 1
)
timeout /T 1 /NOBREAK >nul
set /A "WAIT_REMAINING-=1"
goto wait_for_port_loop

:wait_for_page
set /A "WAIT_REMAINING=%VINE_STARTUP_TIMEOUT%"
where curl.exe >nul 2>nul
if errorlevel 1 goto wait_for_page_port_only
:wait_for_page_loop
curl.exe --fail --silent --max-time 1 "http://%VINE_HOST%:%VINE_PUBLIC_PORT%/" >nul 2>nul
if not errorlevel 1 (
    echo [vine-start] Portal page is ready: http://%VINE_HOST%:%VINE_PUBLIC_PORT%/
    exit /b 0
)
if !WAIT_REMAINING! LEQ 0 (
    echo [vine-start] error: timed out waiting for Portal page on %VINE_HOST%:%VINE_PUBLIC_PORT% 1>&2
    exit /b 1
)
timeout /T 1 /NOBREAK >nul
set /A "WAIT_REMAINING-=1"
goto wait_for_page_loop

:wait_for_page_port_only
echo [vine-start] curl.exe is unavailable; checking only the Portal listener.
call :wait_for_port "Vine server" "%VINE_PUBLIC_PORT%"
exit /b %ERRORLEVEL%

:cleanup
if "%BACKEND_STARTED%"=="1" (
    echo [vine-start] stop Vine server process tree
    taskkill /FI "WINDOWTITLE eq %BACKEND_TITLE%" /T /F >nul 2>nul
)
if "%FRONTEND_STARTED%"=="1" (
    echo [vine-start] stop frontend process tree
    taskkill /FI "WINDOWTITLE eq %FRONTEND_TITLE%" /T /F >nul 2>nul
)
endlocal & exit /b %EXIT_CODE%

:help
set "EXIT_CODE=0"
:help_with_code
echo Usage: scripts\start_vine_app.bat [--install] [--check]
echo.
echo Starts a browser-enabled Vine Standalone project:
echo   http://127.0.0.1:7288/            Portal WEBGW frontend
echo   http://127.0.0.1:7288/api/invoke  Portal RPCGW vRPC
echo   http://127.0.0.1:7299/            Standalone Dashboard
echo   http://127.0.0.1:5174/            Vite development upstream
echo.
echo Options:
echo   --install  Run pnpm install before startup.
echo   --check    Validate files, commands, dependencies, and ports without starting.
echo   -h, --help Show this help.
echo.
echo Optional environment overrides:
echo   VINE_PROJECT_ROOT, VINE_WEB_DIR, VINE_SERVER_PACKAGE,
echo   VINE_PREPARE_PACKAGE, VINE_HOST, VINE_VITE_PORT,
echo   VINE_PUBLIC_PORT, VINE_DASHBOARD_PORT, VINE_STARTUP_TIMEOUT
echo.
echo Set VINE_PREPARE_PACKAGE=./src/server/cmd/migrate only when explicitly required.
goto cleanup
