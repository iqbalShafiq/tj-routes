# Test script for bulk upload endpoint
# This script demonstrates how to use the bulk upload API

Write-Host "=== Bulk Upload API Test ===" -ForegroundColor Cyan
Write-Host ""

# Step 1: Login (replace with your admin credentials)
Write-Host "Step 1: Logging in..." -ForegroundColor Yellow
$loginBody = @{
    email = "admin@example.com"  # Replace with your admin email
    password = "yourpassword"    # Replace with your admin password
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" `
        -Method POST `
        -ContentType "application/json" `
        -Body $loginBody
    
    $token = $loginResponse.data.access_token
    Write-Host "✓ Login successful!" -ForegroundColor Green
    Write-Host "Token: $($token.Substring(0, 20))..." -ForegroundColor Gray
    Write-Host ""
} catch {
    Write-Host "✗ Login failed: $_" -ForegroundColor Red
    Write-Host "Make sure you have an admin account registered." -ForegroundColor Yellow
    exit 1
}

# Step 2: Upload CSV file
Write-Host "Step 2: Uploading CSV file..." -ForegroundColor Yellow
$csvFile = "examples\bulk_upload\routes_example.csv"

if (-not (Test-Path $csvFile)) {
    Write-Host "✗ CSV file not found: $csvFile" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
}

try {
    $form = @{
        file = Get-Item $csvFile
    }
    
    $uploadResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/bulk-upload/route" `
        -Method POST `
        -Headers $headers `
        -Form $form
    
    Write-Host "✓ Upload successful!" -ForegroundColor Green
    Write-Host "Upload ID: $($uploadResponse.data.id)" -ForegroundColor Cyan
    Write-Host "Status: $($uploadResponse.data.status)" -ForegroundColor Cyan
    Write-Host ""
    
    $uploadId = $uploadResponse.data.id
    
    # Step 3: Check upload status
    Write-Host "Step 3: Checking upload status..." -ForegroundColor Yellow
    Start-Sleep -Seconds 2
    
    $statusResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/bulk-upload/$uploadId" `
        -Method GET `
        -Headers $headers
    
    Write-Host "✓ Status retrieved!" -ForegroundColor Green
    Write-Host "Status: $($statusResponse.data.status)" -ForegroundColor Cyan
    Write-Host "Total Rows: $($statusResponse.data.total_rows)" -ForegroundColor Cyan
    Write-Host "Success: $($statusResponse.data.success_count)" -ForegroundColor Green
    Write-Host "Duplicates: $($statusResponse.data.duplicate_count)" -ForegroundColor Yellow
    Write-Host "Errors: $($statusResponse.data.error_count)" -ForegroundColor $(if ($statusResponse.data.error_count -gt 0) { "Red" } else { "Green" })
    
} catch {
    Write-Host "✗ Upload failed: $_" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response: $responseBody" -ForegroundColor Yellow
    }
    exit 1
}

Write-Host ""
Write-Host "=== Test Complete ===" -ForegroundColor Cyan

