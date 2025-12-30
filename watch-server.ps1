# Watch for new builds and restart server automatically
$serverPath = "bin\server.exe"
$process = $null

function Start-Server {
    if ($script:process -and !$script:process.HasExited) {
        Write-Host "Stopping server..." -ForegroundColor Yellow
        Stop-Process -Id $script:process.Id -Force -ErrorAction SilentlyContinue
        Start-Sleep -Milliseconds 500
    }

    if (Test-Path $serverPath) {
        Write-Host "Starting server..." -ForegroundColor Green
        $script:process = Start-Process -FilePath $serverPath -PassThru -NoNewWindow
        Write-Host "Server running (PID: $($script:process.Id))" -ForegroundColor Cyan
    } else {
        Write-Host "Waiting for $serverPath..." -ForegroundColor Gray
    }
}

# Initial start
Start-Server

# Get initial file info
$lastWrite = if (Test-Path $serverPath) { (Get-Item $serverPath).LastWriteTime } else { $null }

Write-Host ""
Write-Host "Watching for new builds... (Ctrl+C to stop)" -ForegroundColor Magenta
Write-Host ""

try {
    while ($true) {
        Start-Sleep -Seconds 1

        if (Test-Path $serverPath) {
            $currentWrite = (Get-Item $serverPath).LastWriteTime

            if ($lastWrite -eq $null -or $currentWrite -gt $lastWrite) {
                Write-Host ""
                Write-Host "New build detected! ($currentWrite)" -ForegroundColor Yellow
                $lastWrite = $currentWrite
                Start-Sleep -Milliseconds 500  # Wait for file to be fully written
                Start-Server
            }
        }
    }
} finally {
    if ($process -and !$process.HasExited) {
        Write-Host "Shutting down server..." -ForegroundColor Yellow
        Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
    }
    Write-Host "Done." -ForegroundColor Green
}
