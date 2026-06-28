class CliDeepseekv4FlashGo < Formula
  desc "CLI tool for rclone-compatible file encryption and decryption"
  homepage "https://github.com/llm-supermarket/cli-deepseekv4-flash-go"
  version "0.1.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/llm-supermarket/cli-deepseekv4-flash-go/releases/download/v0.1.0/cli-deepseekv4-flash-go-darwin-arm64.tar.gz"
      sha256 "ab7999549a2f3eeb9acc789bc5ad0c5538f46331b8fd89086b7749c1ed60a7fb"
    else
      url "https://github.com/llm-supermarket/cli-deepseekv4-flash-go/releases/download/v0.1.0/cli-deepseekv4-flash-go-darwin-amd64.tar.gz"
      sha256 "402ff90c2a6f95b6e6a001e5d3b8686ac132fd2ad5206f3d428691d4f6885765"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/llm-supermarket/cli-deepseekv4-flash-go/releases/download/v0.1.0/cli-deepseekv4-flash-go-linux-arm64.tar.gz"
      sha256 "7225b22a093e4b7ba2c98c991174fc80244386896df3d3d381b99b435a68166e"
    else
      url "https://github.com/llm-supermarket/cli-deepseekv4-flash-go/releases/download/v0.1.0/cli-deepseekv4-flash-go-linux-amd64.tar.gz"
      sha256 "c48c2ee6c6061d435331ed9582743da5052f8bde33876202225a17ed050c1b10"
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