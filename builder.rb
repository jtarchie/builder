# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Builder < Formula
  desc ""
  homepage ""
  version "0.0.23"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/jtarchie/builder/releases/download/v0.0.23/builder_darwin_x86_64.tar.gz"
      sha256 "af2ffc4c53ccbecaa96267fc59959a9aebf8d87a89ac344a8d6d34e2eae3f655"

      def install
        bin.install "builder"
      end
    end
    if Hardware::CPU.arm?
      url "https://github.com/jtarchie/builder/releases/download/v0.0.23/builder_darwin_arm64.tar.gz"
      sha256 "db14de3eec3dc89d346e3fe6fec8a821198b6af5c06a4e4b96b6383aa6e5dbc6"

      def install
        bin.install "builder"
      end
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/jtarchie/builder/releases/download/v0.0.23/builder_linux_x86_64.tar.gz"
      sha256 "d625ad6150885aea4267293682785e1982c4f4aae0428dcadcd7903779f09466"

      def install
        bin.install "builder"
      end
    end
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/jtarchie/builder/releases/download/v0.0.23/builder_linux_arm64.tar.gz"
      sha256 "05f88c3ba1c7d61242c79550852159f03d4ff8858331be3ae6f93de4461e31df"

      def install
        bin.install "builder"
      end
    end
  end

  test do
    system "#{bin}/builder --help"
  end
end
