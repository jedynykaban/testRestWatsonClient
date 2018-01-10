@ECHO OFF
SETLOCAL ENABLEDELAYEDEXPANSION

SET outdir=.\_build
SET debugdir=.\cmd
SET output=%outdir%\testRestWatsonClient.exe
SET entrypoint=.\cmd

IF "%~1"=="build" (
    CALL :build
    EXIT /B 0
) ELSE IF "%~1"=="clean" (
    CALL :clean
    EXIT /B 0
) ELSE IF "%~1"=="fmt" (
    CALL :fmt
    EXIT /B 0
) ELSE IF "%~1"=="vet" (
    CALL :vet
    EXIT /B 0
) ELSE IF "%~1"=="test" (
    CALL :test
    EXIT /B 0
) ELSE GOTO all

:clean
echo Cleaning output dir
IF EXIST %outdir% rmdir %outdir% /s /q
govendor clean +local
mkdir %outdir%
EXIT /B 0

:fmt
govendor fmt +local > CON
EXIT /B 0

:vet
govendor vet +local > CON
EXIT /B 0

:test
echo Testing
govendor test --cover -v +local
EXIT /B 0

:build
echo Building testRestWatsonClient
go build -o %output% %entrypoint% > CON
echo Build complete
EXIT /B 0

:all
CALL :clean
CALL :fmt
CALL :vet
CALL :test
CALL :build
EXIT /B 0
