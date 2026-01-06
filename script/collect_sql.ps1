$OutputEncoding = [System.Text.Encoding]::UTF8
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
# 收集 SQL 文件的脚本
# 使用方法：在 PowerShell 中运行 powershell -ExecutionPolicy Bypass -File collect_sql.ps1

# 脚本所在目录的父目录（即 adp 目录）
$ScriptDir = $PSScriptRoot
$AdpDir = Split-Path -Parent $ScriptDir
Write-Host "工作目录: $AdpDir"

# 来源目录列表
$SrcDirs = @(
    "ontology\ontology-manager\migrations",
    "vega\data-connection\migrations",
    "vega\mdl-data-model\migrations",
    "vega\vega-gateway\migrations",
    "vega\vega-metadata\migrations"
)

# 数据库类型列表
$DbTypes = @("dm8", "mariadb")

# 目标目录（与 script 同级）
$DstDir = Join-Path $AdpDir "sql"
if (-not (Test-Path $DstDir)) {
    New-Item -ItemType Directory -Path $DstDir | Out-Null
}

# 为每个数据库类型创建临时文件
$TmpFiles = @{}
$utf8NoBom = New-Object System.Text.UTF8Encoding $false
foreach ($dbType in $DbTypes) {
    $TmpFiles[$dbType] = [System.IO.Path]::GetTempFileName()
    # 清空临时文件内容
    [System.IO.File]::WriteAllText($TmpFiles[$dbType], "", $utf8NoBom)
}

try {
    # 遍历每个来源目录
    foreach ($dir in $SrcDirs) {
        $fullDir = Join-Path $AdpDir $dir
        Write-Host "处理目录: $fullDir"
        
        # 遍历每个数据库类型
        foreach ($dbType in $DbTypes) {
            $dbDir = Join-Path $fullDir $dbType
            Write-Host "  数据库目录: $dbDir"
            Write-Host "  目录是否存在: $(Test-Path $dbDir)"
            
            if (-not (Test-Path $dbDir)) {
                Write-Host "  警告: 目录不存在 $dbDir" -ForegroundColor Yellow
                continue
            }
            
            # 找到版本号最大的文件夹
            $allDirs = Get-ChildItem -Path $dbDir -Directory -ErrorAction SilentlyContinue
            $versionDirs = $allDirs | Where-Object { $_.Name -match '^\d+\.\d+\.\d+$' }
            
            if ($versionDirs.Count -eq 0) {
                Write-Host "  警告: 在 $dbDir 中未找到版本目录" -ForegroundColor Yellow
                continue
            }
            
            # 使用版本号排序
            $versionArray = $versionDirs | Sort-Object -Property { [version]$_.Name } -Descending
            $latest = $versionArray[0]
            
            Write-Host "  找到最新版本: $($latest.Name)"
            
            # 检查 init.sql 是否在 pre/ 子目录中
            $initSql = Join-Path $latest.FullName "pre\init.sql"
            
            if (-not (Test-Path $initSql)) {
                Write-Host "  错误: 在 $($latest.FullName)\pre\init.sql 中未找到 init.sql 文件" -ForegroundColor Red
                exit 1
            }
            
            Write-Host "  合并文件: $initSql"
            
            # 将绝对路径转换为相对路径
            $relativePath = $initSql.Substring($AdpDir.Length + 1)
            
            # 写入对应的临时文件（使用 UTF-8 without BOM 编码）
            # 如果文件不为空，先添加一个换行符
            if ([System.IO.File]::Exists($TmpFiles[$dbType])) {
                $fileInfo = Get-Item $TmpFiles[$dbType]
                if ($fileInfo.Length -gt 0) {
                    [System.IO.File]::AppendAllText($TmpFiles[$dbType], "`n", $utf8NoBom)
                }
            }
            [System.IO.File]::AppendAllText($TmpFiles[$dbType], "-- Source: $relativePath`n", $utf8NoBom)
            $content = [System.IO.File]::ReadAllText($initSql, [System.Text.Encoding]::UTF8)
            $content = $content.Replace("`r`n", "`n")
            [System.IO.File]::AppendAllText($TmpFiles[$dbType], $content, $utf8NoBom)
        }
    }
    
    # 合并结果写入目标文件
    foreach ($dbType in $DbTypes) {
        $destFile = "$DstDir/${dbType}_init.sql"
        Move-Item -Path $TmpFiles[$dbType] -Destination $destFile -Force
        Write-Host "已生成 $destFile" -ForegroundColor Green
    }
} finally {
    # 清理临时文件
    foreach ($dbType in $DbTypes) {
        if (Test-Path $TmpFiles[$dbType]) {
            Remove-Item $TmpFiles[$dbType] -Force
        }
    }
}
