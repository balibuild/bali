# Bali - 极简的 Golang 构建打包工具

[![Master Branch Status](https://github.com/balibuild/bali/workflows/CI/badge.svg)](https://github.com/balibuild/bali/actions)

Bali 是一个使用 Golang 开发的*极简 Golang 构建打包工具*，[Bali(old)](https://github.com/fcharlie/bali) 

## 功能

Bali 有一些功能是我觉得有些用处的：

+   构建参数支持环境变量推导
+   打包，创建压缩包，支持 `rpm`, `tar`, `zip`, `sh` 等。
+   Windows 平台支持嵌入版本信息，图标，和应用程序清单。


bali 的命令行帮助信息如下：

```txt
Usage: bali <command> [flags]

Bali - Minimalist Golang build and packaging tool

Flags:
  -h, --help             Show context-sensitive help.
  -M, --module="."       Explicitly specify a module directory
  -B, --build="build"    Explicitly specify a build directory
  -V, --verbose          Make the operation more talkative
  -v, --version          Print version information and quit

Commands:
  build     Compile the current module (default)
  update    Update dependencies as recorded in the go.mod
  clean     Remove generated artifacts

Run "bali <command> --help" for more information on a command.

```

bali 构建命令帮助：

```txt
Usage: bali build [flags]

Compile the current module (default)

Flags:
  -h, --help                  Show context-sensitive help.
  -M, --module="."            Explicitly specify a module directory
  -B, --build="build"         Explicitly specify a build directory
  -V, --verbose               Make the operation more talkative
  -v, --version               Print version information and quit

  -T, --target="windows"      Target OS for which the code is compiled
  -A, --arch="amd64"          Target architecture for which the code is compiled
      --release=STRING        Specifies the rpm package tag version
  -D, --destination="dest"    Specify the package save destination
      --pack=PACK,...         Packaged in a specific format. supported: zip,
                              tar, sh, rpm
      --compression=STRING    Specifies the compression method
```

rpm 支持的压缩算法:
+ gzip
+ zstd
+ lzma
+ xz

tar 支持的压缩算法:
+ none   --> pure tar
+ gzip   --> tar.gz
+ zstd   --> tar.zst
+ xz     --> tar.xz
+ bzip2  --> tar.bz2
+ brotli --> tar.br

sh 支持的压缩算法:
+ none   --> pure tar
+ gzip   --> tar.gz
+ zstd   --> tar.zst
+ xz     --> tar.xz
+ bzip2  --> tar.bz2

zip 支持的压缩算法:
+ deflate
+ zstd
+ bzip2
+ xz


## 使用方法

普通构建：

```shell
cd /path/to/project
bali
```

创建 `Tar.gz` 压缩包：

```shell
bali --pack=tar
```

创建 `STGZ` 安装包，主要用于 Linux/macOS 平台：

```shell
bali --pack=sh --target=linux --arch=amd64
```

将安装包输出到指定目录：

```shell
bali --pack=rpm --target=linux --arch=amd64 --dest=/tmp/output
```

一次性创建多种包：

```shell
bali --target=linux --arch=arm64 '--pack=sh,rpm,tar' 
```

## Bali 构建文件格式

项目文件 `bali.toml`:

```toml
# https://toml.io/en/
name = "bali"
summary = "Bali - Minimalist Golang build and packaging tool"
description = "Bali - Minimalist Golang build and packaging tool"
package-name = "bali-dev"
version = "3.1.0"
license = "MIT"
prefix = "/usr/local"
crates = [
    "cmd/bali",     # crates
    "cmd/peassets",
]

[[include]]
path = "LICENSE"
destination = "share"
rename = "BALI-COPYRIGHT.txt"
permissions = "0664"

```

程序构建文件 `crate.toml`:

```toml
name = "bali"
description = "Bali - Minimalist Golang build and packaging tool"
destination = "bin"
version = "3.1.0"
goflags = [
    "-ldflags",
    "-X 'main.VERSION=$BUILD_VERSION' -X 'main.BUILD_TIME=$BUILD_TIME' -X 'main.BUILD_BRANCH=$BUILD_BRANCH' -X 'main.BUILD_COMMIT=$BUILD_COMMIT'  -X 'main.BUILD_REFNAME=$BUILD_REFNAME' -X 'main.BUILD_GOVERSION=$BUILD_GOVERSION'",
]


```

内置环境变量：

+   `BUILD_VERSION` 由 balisrc.toml 的 `version` 字段填充
+   `BUILD_TIME` 由构建时间按照 `RFC3339` 格式化后填充
+   `BUILD_COMMIT` 由存储库（为 git 存储库时） 的 commit id 填充
+   `BUILD_GOVERSION` 由 `go version` 输出（删除了 `go version` 前缀）填充
+   `BUILD_BRANCH` 由存储库（为 git 存储库时） 的分支名填充

可以在 goflags 中使用其他环境变量。

Windows 相关清单文件（crate.toml 同级）：`winres.toml:`

```toml
icon = "res/bali.ico" # data:base64-content
manifest = """data:<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<assembly xmlns="urn:schemas-microsoft-com:asm.v1" manifestVersion="1.0" xmlns:asmv3="urn:schemas-microsoft-com:asm.v3">
  <description>Bali</description>
  <trustInfo xmlns="urn:schemas-microsoft-com:asm.v3">
    <security>
      <requestedPrivileges>
        <requestedExecutionLevel level="asInvoker" uiAccess="false" />
      </requestedPrivileges>
    </security>
  </trustInfo>
  <compatibility xmlns="urn:schemas-microsoft-com:compatibility.v1">
    <application>
      <!-- Windows 10 -->
      <supportedOS Id="{8e0f7a12-bfb3-4fe8-b9a5-48fd50a15a9a}"/>
    </application>
  </compatibility>
  <asmv3:application>
    <asmv3:windowsSettings xmlns="http://schemas.microsoft.com/SMI/2005/WindowsSettings">
      <longPathAware xmlns="http://schemas.microsoft.com/SMI/2016/WindowsSettings">true</longPathAware>
    </asmv3:windowsSettings>
  </asmv3:application>
</assembly>
"""

[FixedFileInfo]
FileFlagsMask = "3f"
FileFlags = "00"
FileOS = "40004"
FileType = "01"
FileSubType = "00"

[FixedFileInfo.FileVersion]
Major = 0
Minor = 0
Patch = 0
Build = 0

[FixedFileInfo.ProductVersion]
Major = 0
Minor = 0
Patch = 0
Build = 0

[StringFileInfo]
Comments = ""
CompanyName = "Bali Team"
FileDescription = "Bali - Minimalist Golang build and packaging tool"
FileVersion = ""
InternalName = "bali.exe"
LegalCopyright = "Copyright © 2024. Bali contributors"
LegalTrademarks = ""
OriginalFilename = "bali.exe"
PrivateBuild = ""
ProductName = "Bali"
ProductVersion = ""
SpecialBuild = ""

[VarFileInfo]
[VarFileInfo.Translation]
LangID = "0409"
CharsetID = "04B0"

```

Bali 整合了 [`goversioninfo`](https://github.com/josephspurrier/goversioninfo)，在目标为 Windows 时，能够将版本信息 **winres.toml** 嵌入到可执行程序中。

添加引用程序清单的好处不言而喻，比如 Windows 的 UAC 提权，Windows 10 长路经支持（即路径支持 >260 字符），Windows Vista 风格控件，TaskDialog，DPI 设置等都需要修改应用程序清单。

## 自举

通常在安装配置好 Golang 环境后，你可以按照下面的命令完成 Bali 的自举：

UNIX:

```shell
./script/bootstrap.sh
```

Windows:

```ps1
# 使用 powershell 运行
pwsh ./script/bootstrap.ps1
# 或者在 cmd 中运行
script/bootstrap.bat
```

## 感谢

Bali 自动添加版本信息到 PE 文件的功能离不开开源项目的贡献，在这里非常感谢 [akavel/rsrc](https://github.com/akavel/rsrc) 和 [josephspurrier/goversioninfo](https://github.com/josephspurrier/goversioninfo) 两个项目的开发者和维护者。

Bali Github 组织和 Bali 自身的图标来源于 [www.flaticon.com](https://www.flaticon.com/) 制作者为 [Smashicons](https://www.flaticon.com/authors/smashicons)。

