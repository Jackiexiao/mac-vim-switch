class MacVimSwitch < Formula
  desc "Input method switcher for Vim users on macOS"
  homepage "https://github.com/jackiexiao/mac-vim-switch"
  url "https://github.com/jackiexiao/mac-vim-switch/archive/v1.0.0.tar.gz"
  sha256 "YOUR_SHA256_HERE" # 需要在发布后更新
  license "MIT"
  
  depends_on "go" => :build
  depends_on "laishulu/macism/macism"

  def install
    system "go", "build", "-tags", "cgo", *std_go_args(ldflags: "-s -w")
    
    # Install and load launchd service
    plist_path = "#{prefix}/homebrew.mxcl.mac-vim-switch.plist"
    cp "mac-vim-switch.plist", plist_path
    
    # Replace ${USER} with actual username in plist
    inreplace plist_path, "${USER}", ENV["USER"]
  end

  service do
    run [opt_bin/"mac-vim-switch"]
    keep_alive true
    log_path "#{ENV["HOME"]}/.mac-vim-switch.log"
    error_log_path "#{ENV["HOME"]}/.mac-vim-switch.log"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/mac-vim-switch --version")
  end
end 