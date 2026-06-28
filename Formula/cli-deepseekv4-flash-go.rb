class CliDeepseekv4FlashGo < Formula
  desc "CLI tool for rclone-compatible file encryption and decryption"
  homepage "https://github.com/llm-supermarket/cli-deepseekv4-flash-go"
  version "0.0.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/llm-supermarket/cli-deepseekv4-flash-go/releases/download/v0.0.0/cli-deepseekv4-flash-go-darwin-arm64.tar.gz"
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
    else
      url "https://github.com/llm-supermarket/cli-deepseekv4-flash-go/releases/download/v0.0.0/cli-deepseekv4-flash-go-darwin-amd64.tar.gz"
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/llm-supermarket/cli-deepseekv4-flash-go/releases/download/v0.0.0/cli-deepseekv4-flash-go-linux-arm64.tar.gz"
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
    else
      url "https://github.com/llm-supermarket/cli-deepseekv4-flash-go/releases/download/v0.0.0/cli-deepseekv4-flash-go-linux-amd64.tar.gz"
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
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
