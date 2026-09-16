<#
.SYNOPSIS
    批量构建 dia-platform 的四个前端，产物就地输出到各自 frontend/dist。

.DESCRIPTION
    前端清单（部署子路径 -> 源码目录）:
      business_base        -> base/frontend
      business_application -> business/application/frontend
      business_member      -> business/member/frontend
      business_portal      -> business/portal/frontend

    每个前端执行 `npm run build`（vue-tsc && vite build），产物落盘到该前端的
    frontend/dist，并校验 dist/index.html 里的资源前缀是否与 .env 的 VITE_BASE_PATH
    一致（例如 /business_portal/assets/...），不一致即判失败。

.PARAMETER Only
    只构建指定的一到多个前端（按上表的部署子路径名），可写多个，用逗号分隔。

.PARAMETER Install
    构建前先执行 npm install（默认直接用现有 node_modules）。

.PARAMETER SkipVerify
    跳过 dist/index.html 资源前缀校验。

.EXAMPLE
    powershell -ExecutionPolicy Bypass -File .\build-frontends.ps1

.EXAMPLE
    .\build-frontends.ps1 -Only business_portal,business_member

.EXAMPLE
    .\build-frontends.ps1 -Install -Only business_base
#>
[CmdletBinding()]
param(
    [string[]]$Only = @(),
    [switch]$Install,
    [switch]$SkipVerify
)

$ErrorActionPreference = 'Stop'
$OutputEncoding = [System.Text.Encoding]::UTF8

# 脚本所在目录 = 仓库根目录
$Root = Split-Path -Parent $MyInvocation.MyCommand.Path

$Targets = @(
    [pscustomobject]@{ Name = 'business_base';        Dir = 'base\frontend'                 ; Base = '/business_base/' }
    [pscustomobject]@{ Name = 'business_application'; Dir = 'business\application\frontend' ; Base = '/business_application/' }
    [pscustomobject]@{ Name = 'business_member';      Dir = 'business\member\frontend'      ; Base = '/business_member/' }
    [pscustomobject]@{ Name = 'business_portal';      Dir = 'business\portal\frontend'      ; Base = '/business_portal/' }
)

if ($Only.Count -gt 0) {
    $Targets = @($Targets | Where-Object { $Only -contains $_.Name })
    if ($Targets.Count -eq 0) {
        Write-Host "参数 -Only 未匹配到任何前端。可用值: business_base, business_application, business_member, business_portal" -ForegroundColor Red
        exit 1
    }
}

# 定位 npm
$NpmCmd = Get-Command npm.cmd -ErrorAction SilentlyContinue
if (-not $NpmCmd) { $NpmCmd = Get-Command npm -ErrorAction SilentlyContinue }
if (-not $NpmCmd) {
    Write-Host "找不到 npm，请先安装 Node.js 并把 npm 加入 PATH。" -ForegroundColor Red
    exit 1
}
$Npm = $NpmCmd.Path
if (-not $Npm) { $Npm = $NpmCmd.Source }

Write-Host ""
Write-Host "dia-platform 前端批量构建" -ForegroundColor Green
Write-Host ("仓库根目录: {0}" -f $Root)
Write-Host ("node: {0} / npm: {1}" -f (& node -v), (& $Npm -v))
Write-Host ("待构建前端: {0}" -f (($Targets | ForEach-Object { $_.Name }) -join ', '))
if ($Install) { Write-Host "已开启 -Install: 每个前端构建前会执行 npm install" -ForegroundColor Yellow }

$Results = New-Object System.Collections.ArrayList
$Total = [System.Diagnostics.Stopwatch]::StartNew()

foreach ($t in $Targets) {
    $Full = Join-Path $Root $t.Dir

    Write-Host ""
    Write-Host ("=" * 66) -ForegroundColor DarkGray
    Write-Host ("[{0}]  {1}" -f $t.Name, $t.Dir) -ForegroundColor Cyan
    Write-Host ("=" * 66) -ForegroundColor DarkGray

    $Status = 'OK'
    $Note   = ''
    $SizeMb = 0
    $Files  = 0
    $Dist   = Join-Path $Full 'dist'
    $Sw = [System.Diagnostics.Stopwatch]::StartNew()

    if (-not (Test-Path -LiteralPath $Full)) {
        $Status = 'FAIL'; $Note = "目录不存在: $($t.Dir)"
    }
    elseif (-not (Test-Path -LiteralPath (Join-Path $Full 'package.json'))) {
        $Status = 'FAIL'; $Note = "缺少 package.json: $($t.Dir)"
    }
    elseif (-not (Test-Path -LiteralPath (Join-Path $Full 'node_modules'))) {
        if ($Install) {
            Write-Host "node_modules 缺失，执行 npm install ..." -ForegroundColor Yellow
            Push-Location -LiteralPath $Full
            try { & $Npm install; $InstallCode = $LASTEXITCODE } finally { Pop-Location }
            if ($InstallCode -ne 0) { $Status = 'FAIL'; $Note = "npm install 退出码 $InstallCode" }
        }
        else {
            $Status = 'FAIL'; $Note = "缺少 node_modules（先执行 npm install，或加 -Install 参数）"
        }
    }
    elseif ($Install) {
        Write-Host "npm install ..." -ForegroundColor Yellow
        Push-Location -LiteralPath $Full
        try { & $Npm install; $InstallCode = $LASTEXITCODE } finally { Pop-Location }
        if ($InstallCode -ne 0) { $Status = 'FAIL'; $Note = "npm install 退出码 $InstallCode" }
    }

    if ($Status -eq 'OK') {
        Write-Host "npm run build ..." -ForegroundColor Yellow
        Push-Location -LiteralPath $Full
        try {
            & $Npm run build
            $BuildCode = $LASTEXITCODE
        }
        finally {
            Pop-Location
        }
        if ($BuildCode -ne 0) {
            $Status = 'FAIL'
            $Note = "npm run build 退出码 $BuildCode"
        }
    }

    if ($Status -eq 'OK') {
        $IndexHtml = Join-Path $Dist 'index.html'
        if (-not (Test-Path -LiteralPath $IndexHtml)) {
            $Status = 'FAIL'
            $Note = '未生成 dist/index.html'
        }
        else {
            if (-not $SkipVerify) {
                $Html = [System.IO.File]::ReadAllText($IndexHtml)
                $Expected = $t.Base + 'assets/'
                if ($Html -notmatch [regex]::Escape($Expected)) {
                    $Status = 'FAIL'
                    $Note = "dist/index.html 资源前缀不是 $Expected（检查该前端 .env 的 VITE_BASE_PATH）"
                }
            }
            if ($Status -eq 'OK') {
                $FileList = @(Get-ChildItem -LiteralPath $Dist -Recurse -File -ErrorAction SilentlyContinue)
                $Files = $FileList.Count
                if ($Files -gt 0) {
                    $SizeMb = [math]::Round((($FileList | Measure-Object -Property Length -Sum).Sum / 1MB), 2)
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
        Files   = $Files
        SizeMb  = $SizeMb
        Dist    = $t.Dir + '\dist'
        Note    = $Note
    })

    if ($Status -eq 'OK') {
        Write-Host ("OK  {0}  ({1}s, {2} 个文件, {3} MB)  ->  {4}" -f $t.Name, $Seconds, $Files, $SizeMb, $t.Dir + '\dist') -ForegroundColor Green
    }
    else {
        Write-Host ("FAIL  {0}: {1}" -f $t.Name, $Note) -ForegroundColor Red
    }
}

$Total.Stop()

Write-Host ""
Write-Host ("=" * 66) -ForegroundColor DarkGray
Write-Host "构建结果汇总" -ForegroundColor Green
Write-Host ("=" * 66) -ForegroundColor DarkGray
foreach ($r in $Results) {
    $Line = "{0,-22} {1,-5} {2,7}s {3,6} 个文件 {4,8} MB   {5}" -f $r.Name, $r.Status, $r.Seconds, $r.Files, $r.SizeMb, $r.Dist
    if ($r.Status -eq 'OK') { Write-Host $Line -ForegroundColor Green } else { Write-Host $Line -ForegroundColor Red }
    if ($r.Note) { Write-Host ("    -> {0}" -f $r.Note) -ForegroundColor Red }
}

$Failed = @($Results | Where-Object { $_.Status -ne 'OK' })
Write-Host ""
Write-Host ("总计 {0} 个前端，成功 {1} 个，失败 {2} 个，耗时 {3}s" -f $Results.Count, ($Results.Count - $Failed.Count), $Failed.Count, [math]::Round($Total.Elapsed.TotalSeconds, 1)) -ForegroundColor $(if ($Failed.Count -eq 0) { 'Green' } else { 'Red' })
Write-Host "产物已就地输出到各前端目录下的 frontend\dist（与 .env 的 VITE_BASE_PATH 对应的部署子路径）。"

if ($Failed.Count -gt 0) { exit 1 }
exit 0
