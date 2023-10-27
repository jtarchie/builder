# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Builder < Formula
  desc ""
  homepage ""
  version "0.0.28"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/jtarchie/builder/releases/download/v0.0.28/builder_darwin_arm64.tar.gz"
      sha256 "a0ff082ee15fd286edf860af9a6e4040e13649fadff35ee5e3bf23e06f3e18dd"

      def install
        bin.install "builder"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/jtarchie/builder/releases/download/v0.0.28/builder_darwin_x86_64.tar.gz"
      sha256 "3e0b9b1c0f3a47bec1bf715a594e005166f3b74b33d11f54cd88652ac910a16e"

      def install
        bin.install "builder"
      end
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/jtarchie/builder/releases/download/v0.0.28/builder_linux_x86_64.tar.gz"
      sha256 "64505cbd2714f7a8dff8d37c17780da1e028b0c32ba7c614b4fc5b0cf4cf25aa"

      def install
        bin.install "builder"
      end
    end
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/jtarchie/builder/releases/download/v0.0.28/builder_linux_arm64.tar.gz"
      sha256 "a5f1da1c118ffb3bc5d344d1d412847361ac03b802a1fbb998767bd8f817ab92"

      def install
        bin.install "builder"
      end
    end
  end

  test do
    system "#{bin}/builder --help"
  end
end
