# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Builder < Formula
  desc ""
  homepage ""
  version "0.0.29"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/jtarchie/builder/releases/download/v0.0.29/builder_darwin_arm64.tar.gz"
      sha256 "5b60434afab84e4006d2856d71763e43b98e00efae38bcedbf19ec18d4e68566"

      def install
        bin.install "builder"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/jtarchie/builder/releases/download/v0.0.29/builder_darwin_x86_64.tar.gz"
      sha256 "69a798955ffb08a0c7a5d8a3716a84b141d23d9a8f72ed090250dbeeb145029e"

      def install
        bin.install "builder"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/jtarchie/builder/releases/download/v0.0.29/builder_linux_arm64.tar.gz"
      sha256 "5d65b90b15816b6f5784de9d5ca73188424f637aca50a84c3868a66374dccdc9"

      def install
        bin.install "builder"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/jtarchie/builder/releases/download/v0.0.29/builder_linux_x86_64.tar.gz"
      sha256 "c8c38a4bd532ee487f469d3529b7c372c29a3455829452789bb66c6473a3a8ee"

      def install
        bin.install "builder"
      end
    end
  end

  test do
    system "#{bin}/builder --help"
  end
end
