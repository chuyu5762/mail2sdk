# Mail2 SDK v1.0.0 Release 发布脚本
# 请在新的 PowerShell 窗口中运行此脚本

Write-Host "=== Mail2 SDK v1.0.0 Release 发布 ===" -ForegroundColor Green
Write-Host ""

# 1. 检查 gh 是否可用
Write-Host "1. 检查 GitHub CLI..." -ForegroundColor Yellow
try {
    gh --version
    Write-Host "✓ GitHub CLI 已安装" -ForegroundColor Green
} catch {
    Write-Host "✗ 请重启 PowerShell 后再运行此脚本" -ForegroundColor Red
    exit 1
}

Write-Host ""

# 2. 检查是否已登录
Write-Host "2. 检查登录状态..." -ForegroundColor Yellow
$authStatus = gh auth status 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "需要登录 GitHub..." -ForegroundColor Yellow
    Write-Host "请选择登录方式：" -ForegroundColor Cyan
    Write-Host "  1. 使用浏览器登录（推荐）" -ForegroundColor Cyan
    Write-Host "  2. 使用 Token 登录" -ForegroundColor Cyan
    $choice = Read-Host "请输入选择 (1/2)"
    
    if ($choice -eq "1") {
        gh auth login -w
    } else {
        gh auth login
    }
} else {
    Write-Host "✓ 已登录 GitHub" -ForegroundColor Green
}

Write-Host ""

# 3. 发布 Release
Write-Host "3. 发布 Release..." -ForegroundColor Yellow

# 删除旧 release（如果存在）
Write-Host "检查并删除旧 Release..." -ForegroundColor Cyan
gh release delete v1.0.0 -y 2>$null

Write-Host "正在发布 v1.0.0..." -ForegroundColor Cyan

# 使用 --notes-file 而不是 --notes 变量
gh release create v1.0.0 --title "Mail2 SDK v1.0.0" --notes-file "RELEASE_NOTES_v1.0.0.md" --latest

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✓ Release 发布成功！" -ForegroundColor Green
    Write-Host ""
    Write-Host "查看 Release：https://github.com/chuyu5762/mail2sdk/releases/tag/v1.0.0" -ForegroundColor Cyan
} else {
    Write-Host ""
    Write-Host "✗ 发布失败，请检查错误信息" -ForegroundColor Red
}

Write-Host ""
Write-Host "按任意键退出..." -ForegroundColor Gray
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
