<#
.SYNOPSIS
    批量编译 dia-platform 的四个 Go 后端，可执行文件就地输出到各自 backend 目录。

.DESCRIPTION
    后端清单（模块 -> 源码目录 -> 产物）:
      base        -> base\backend                     -> base\backend\base
      application -> business\application\backend     -> business\application\backend\application
      member      -> business\member\backend          -> business\member\backend\member
      portal      -> business\portal\backend          -> business\portal\backend\portal

    与各 backend 目录下的 comp.bat 口径一致：`set goos=linux` + `go build -o <名> main.go`，
    默认交叉编译 Linux/amd64、CGO_ENABLED=0（静态二进制，适用于生产 Linux 机）。
    编译后校验产物存在且为 ELF（并读取 e_machine 校验目标架构）。

.PARAMETER Only
    只编译指定的一到多个后端（按上表的模块名），可写多个，用逗号分隔。

.PARAMETER Arch
    目标架构，默认 amd64（可选 arm64、386）。

.PARAMETER Strip
    追加 -trimpath -ldflags "-s -w"（去掉符号表/调试信息，体积明显变小，便于传输，
    但与 comp.bat 的历史产物不完全一致，需要时再用）。

.PARAMETER Vet
    编译前先执行 `go vet ./...`，静态检查不通过则跳过该后端的编译。

.PARAMETER SkipVerify
    跳过 ELF 产物校验。

.EXAMPLE
    powershell -ExecutionPolicy Bypass -File .\build-backends.ps1

.EXAMPLE
    .\build-backends.ps1 -Only portal,member

.EXAMPLE
    .\build-backends.ps1 -Arch arm64 -Strip -Vet
#>
[CmdletBinding()]
param(
    [string[]]$Only = @(),
    [ValidateSet('amd64', 'arm64', '386')]
    [string]$Arch = 'amd64',
    [switch]$Strip,
    [switch]$Vet,
    [switch]$SkipVerify
)

$ErrorActionPreference = 'Stop'
$OutputEncoding = [System.Text.Encoding]::UTF8

# 脚本所在目录 = 仓库根目录
$Root = Split-Path -Parent $MyInvocation.MyCommand.Path

$Targets = @(
    [pscustomobject]@{ Name = 'base';        Dir = 'base\backend'                ; Out = 'base' }
    [pscustomobject]@{ Name = 'application'; Dir = 'business\application\backend'; Out = 'application' }
    [pscustomobject]@{ Name = 'member';      Dir = 'business\member\backend'     ; Out = 'member' }
    [pscustomobject]@{ Name = 'portal';      Dir = 'business\portal\backend'     ; Out = 'portal' }
)

if ($Only.Count -gt 0) {
    $Targets = @($Targets | Where-Object { $Only -contains $_.Name })
    if ($Targets.Count -eq 0) {
        Write-Host "参数 -Only 未匹配到任何后端。可用值: base, application, member, portal" -ForegroundColor Red
        exit 1
    }
}

$GoCmd = Get-Command go -ErrorAction SilentlyContinue
if (-not $GoCmd) {
    Write-Host "找不到 go，请先安装 Go 并把 go 加入 PATH。" -ForegroundColor Red
    exit 1
}
$Go = $GoCmd.Path
if (-not $Go) { $Go = $GoCmd.Source }

$BuildArgs = @('build')
if ($Strip) { $BuildArgs += @('-trimpath', '-ldflags', '-s -w') }

# ELF e_machine -> 期望值（用于校验 GOARCH 是否真的生效）
$ExpectedElfArch = @{ amd64 = 'x86-64'; arm64 = 'aarch64'; '386' = 'i386' }[$Arch]

# 保存并在结束时还原环境变量
$OldGoos = $env:GOOS
$OldGoarch = $env:GOARCH
$OldCgo = $env:CGO_ENABLED

Write-Host ""
Write-Host "dia-platform 后端批量编译" -ForegroundColor Green
Write-Host ("仓库根目录: {0}" -f $Root)
Write-Host ("go: {0}" -f (& $Go version))
Write-Host ("目标平台: linux/{0}  CGO_ENABLED=0{1}" -f $Arch, $(if ($Strip) { '  已开启 -Strip（去符号表）' } else { '' }))
Write-Host ("待编译后端: {0}" -f (($Targets | ForEach-Object { $_.Name }) -join ', '))
if ($Vet) { Write-Host "已开启 -Vet: 每个后端编译前先执行 go vet ./..." -ForegroundColor Yellow }

$Results = New-Object System.Collections.ArrayList
$Total = [System.Diagnostics.Stopwatch]::StartNew()

$env:GOOS = 'linux'
$env:GOARCH = $Arch
$env:CGO_ENABLED = '0'

try {
    foreach ($t in $Targets) {
        $Full = Join-Path $Root $t.Dir
        $Bin = Join-Path $Full $t.Out

        Write-Host ""
        Write-Host ("=" * 66) -ForegroundColor DarkGray
        Write-Host ("[{0}]  {1}  ->  {2}" -f $t.Name, $t.Dir, ($t.Dir + '\' + $t.Out)) -ForegroundColor Cyan
        Write-Host ("=" * 66) -ForegroundColor DarkGray

        $Status = 'OK'
        $Note = ''
        $SizeMb = 0
        $ArchText = ''
        $Sw = [System.Diagnostics.Stopwatch]::StartNew()

        if (-not (Test-Path -LiteralPath $Full)) {
            $Status = 'FAIL'; $Note = "目录不存在: $($t.Dir)"
        }
        elseif (-not (Test-Path -LiteralPath (Join-Path $Full 'go.mod'))) {
            $Status = 'FAIL'; $Note = "缺少 go.mod: $($t.Dir)"
        }
        elseif (-not (Test-Path -LiteralPath (Join-Path $Full 'main.go'))) {
            $Status = 'FAIL'; $Note = "缺少 main.go: $($t.Dir)"
        }

        if ($Status -eq 'OK' -and $Vet) {
            Write-Host "go vet ./... ..." -ForegroundColor Yellow
            Push-Location -LiteralPath $Full
            try {
                & $Go vet ./...
                $VetCode = $LASTEXITCODE
            }
            finally {
                Pop-Location
            }
            if ($VetCode -ne 0) {
                $Status = 'FAIL'
                $Note = "go vet 退出码 $VetCode"
            }
        }

        if ($Status -eq 'OK') {
            Write-Host ("go build -o {0} main.go ..." -f $t.Out) -ForegroundColor Yellow
            Push-Location -LiteralPath $Full
            try {
                & $Go @BuildArgs -o $t.Out main.go
                $BuildCode = $LASTEXITCODE
            }
            finally {
                Pop-Location
            }
            if ($BuildCode -ne 0) {
                $Status = 'FAIL'
                $Note = "go build 退出码 $BuildCode"
            }
        }

        if ($Status -eq 'OK') {
            if (-not (Test-Path -LiteralPath $Bin -PathType Leaf)) {
                $Status = 'FAIL'; $Note = "未生成产物 $($t.Out)"
            }
            else {
                $Item = Get-Item -LiteralPath $Bin
                $SizeMb = [math]::Round($Item.Length / 1MB, 1)
                if ($Item.Length -eq 0) {
                    $Status = 'FAIL'; $Note = "产物为 0 字节: $($t.Out)"
                }
                elseif (-not $SkipVerify) {
                    $Head = New-Object byte[] 20
                    $Fs = [System.IO.File]::OpenRead($Bin)
                    try { $null = $Fs.Read($Head, 0, 20) } finally { $Fs.Dispose() }
                    $IsElf = ($Head[0] -eq 0x7F -and $Head[1] -eq 0x45 -and $Head[2] -eq 0x4C -and $Head[3] -eq 0x46)
                    if (-not $IsElf) {
                        $Status = 'FAIL'
                        $Note = "产物不是 Linux ELF（检查 GOOS/GOARCH 或 comp.bat 口径）"
                    }
                    else {
                        switch ([BitConverter]::ToUInt16($Head, 18)) {
                            0x3E { $ArchText = 'x86-64' }
                            0xB7 { $ArchText = 'aarch64' }
                            0x03 { $ArchText = 'i386' }
                            default { $ArchText = 'machine=0x{0:X}' -f [BitConverter]::ToUInt16($Head, 18) }
                        }
                        if ($ArchText -ne $ExpectedElfArch) {
                            $Status = 'FAIL'
                            $Note = "产物架构为 $ArchText，与 -Arch $Arch 不符"
                        }
                    }
                }
            }
        }

        $Sw.Stop()
        $Seconds = [math]::Round($Sw.Elapsed.TotalSeconds, 1)

        [void]$Results.Add([pscustomobject]@{
            Name    = $t.Name
            Status  = $Status
            Seconds = $Seconds
            SizeMb  = $SizeMb
            Arch    = $ArchText
            Bin     = $t.Dir + '\' + $t.Out
            Note    = $Note
        })

        if ($Status -eq 'OK') {
            Write-Host ("OK  {0}  ({1}s, {2} MB{3})  ->  {4}" -f $t.Name, $Seconds, $SizeMb, $(if ($ArchText) { ", ELF $ArchText" } else { '' }), ($t.Dir + '\' + $t.Out)) -ForegroundColor Green
        }
        else {
            Write-Host ("FAIL  {0}: {1}" -f $t.Name, $Note) -ForegroundColor Red
        }
    }
}
finally {
    $env:GOOS = $OldGoos
    $env:GOARCH = $OldGoarch
    $env:CGO_ENABLED = $OldCgo
}

$Total.Stop()

Write-Host ""
Write-Host ("=" * 66) -ForegroundColor DarkGray
Write-Host "编译结果汇总" -ForegroundColor Green
Write-Host ("=" * 66) -ForegroundColor DarkGray
foreach ($r in $Results) {
    $Line = "{0,-12} {1,-5} {2,7}s {3,7} MB {4,-10} {5}" -f $r.Name, $r.Status, $r.Seconds, $r.SizeMb, $r.Arch, $r.Bin
    if ($r.Status -eq 'OK') { Write-Host $Line -ForegroundColor Green } else { Write-Host $Line -ForegroundColor Red }
    if ($r.Note) { Write-Host ("    -> {0}" -f $r.Note) -ForegroundColor Red }
}

$Failed = @($Results | Where-Object { $_.Status -ne 'OK' })
Write-Host ""
Write-Host ("总计 {0} 个后端，成功 {1} 个，失败 {2} 个，耗时 {3}s" -f $Results.Count, ($Results.Count - $Failed.Count), $Failed.Count, [math]::Round($Total.Elapsed.TotalSeconds, 1)) -ForegroundColor $(if ($Failed.Count -eq 0) { 'Green' } else { 'Red' })
Write-Host ("产物为 linux/{0} 静态二进制，已就地输出到各后端目录（上传服务器时按需带 config.yaml 与 uploads\）。" -f $Arch)

if ($Failed.Count -gt 0) { exit 1 }
exit 0
