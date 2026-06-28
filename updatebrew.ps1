param(
    [Parameter(Mandatory = $true)]
    [string]$Version
)

$repo = "llm-supermarket/cli-deepseekv4-flash-go"
$platforms = @("darwin-amd64", "darwin-arm64", "linux-amd64", "linux-arm64")
$formulaPath = "$PSScriptRoot/Formula/cli-deepseekv4-flash-go.rb"
$base = "https://github.com/$repo/releases/download/v$Version"

$hash = @{}
foreach ($platform in $platforms) {
    $fileName = "cli-deepseekv4-flash-go-$platform.tar.gz"
    $url = "$base/$fileName"
    $tempFile = Join-Path ([System.IO.Path]::GetTempPath()) $fileName

    Write-Host "Downloading $url ..."
    Invoke-WebRequest -Uri $url -OutFile $tempFile

    $hash[$platform] = (Get-FileHash -Path $tempFile -Algorithm SHA256).Hash.ToLower()
    Write-Host "SHA256 for ${platform}: $($hash[$platform])"

    Remove-Item $tempFile
}

$formula = @"
class CliDeepseekv4FlashGo < Formula
  desc "CLI tool for rclone-compatible file encryption and decryption"
  homepage "https://github.com/$repo"
  version "$Version"

  on_macos do
    if Hardware::CPU.arm?
      url "$base/cli-deepseekv4-flash-go-darwin-arm64.tar.gz"
      sha256 "$($hash['darwin-arm64'])"
    else
      url "$base/cli-deepseekv4-flash-go-darwin-amd64.tar.gz"
      sha256 "$($hash['darwin-amd64'])"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "$base/cli-deepseekv4-flash-go-linux-arm64.tar.gz"
      sha256 "$($hash['linux-arm64'])"
    else
      url "$base/cli-deepseekv4-flash-go-linux-amd64.tar.gz"
      sha256 "$($hash['linux-amd64'])"
    end
  end

  def install
    bin.install "cli-deepseekv4-flash-go-darwin-arm64" => "cli-deepseekv4-flash-go" if OS.mac? && Hardware::CPU.arm?
    bin.install "cli-deepseekv4-flash-go-darwin-amd64" => "cli-deepseekv4-flash-go" if OS.mac? && !Hardware::CPU.arm?
    bin.install "cli-deepseekv4-flash-go-linux-arm64" => "cli-deepseekv4-flash-go" if OS.linux? && Hardware::CPU.arm?
    bin.install "cli-deepseekv4-flash-go-linux-amd64" => "cli-deepseekv4-flash-go" if OS.linux? && !Hardware::CPU.arm?
  end

  test do
    assert_match "cli-deepseekv4-flash-go #{version}", shell_output("#{bin}/cli-deepseekv4-flash-go --version")
  end
end
"@

Set-Content -Path $formulaPath -Value $formula -NoNewline
Write-Host "Wrote $formulaPath for version $Version"
