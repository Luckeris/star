# Build script for Star VCS binary distribution
Write-Host "Compiling Star binaries..." -ForegroundColor Cyan

New-Item -ItemType Directory -Force -Path "bin" | Out-Null

# 1. Build Windows binary
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o bin/star.exe ./cmd/star
Write-Host "Built bin/star.exe (Windows)" -ForegroundColor Green

# 2. Build Linux binary
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o bin/star-linux ./cmd/star
Write-Host "Built bin/star-linux (Linux)" -ForegroundColor Green

# 3. Build macOS binary
$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -o bin/star-darwin ./cmd/star
Write-Host "Built bin/star-darwin (macOS)" -ForegroundColor Green

# Reset environment variables
Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue

Write-Host "Build complete! All binaries generated in bin folder." -ForegroundColor Cyan
