# Bali -  Minimalist Golang build and packaging tool

[![Master Branch Status](https://github.com/balibuild/bali/workflows/CI/badge.svg)](https://github.com/balibuild/bali/actions)


[简体中文](./README.zh-CN.md)

Bali is a *minimal Golang build and packaging tool* developed using Golang.

## Feature

Bali has some functions that I think are useful:

+ Build parameters support derivation of environment variables
+ Package, create compressed package, support `rpm`, `tar`, `zip`, `sh`.
+ The Windows platform supports embedded version information, icons, and application manifest.

rpm supported compression:
+ gzip
+ zstd
+ lzma
+ xz

tar supported compression:
+ none   --> pure tar
+ gzip   --> tar.gz
+ zstd   --> tar.zst
+ xz     --> tar.xz
+ bzip2  --> tar.bz2
+ brotli --> tar.br

sh supported compression:
+ none   --> pure tar
+ gzip   --> tar.gz
+ zstd   --> tar.zst
+ xz     --> tar.xz
+ bzip2  --> tar.bz2

zip supported compression:
+ deflate
+ zstd
+ bzip2
+ xz


Bali's command line help information is as follows:

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

bali build command:

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


## Instructions

Common build:

```shell
cd /path/to/project
bali
```

Create `Tar.gz` compressed package:

```shell
bali --pack=tar
```

Create `STGZ` installation package, mainly used on Linux/macOS platform:

```shell
bali --pack=sh --target=linux --arch=amd64
```

Output the installation package to the specified directory:

```shell
bali --pack=rpm --target=linux --arch=amd64 --dest=/tmp/output
```

Create multiple packages at once:

```shell
bali --target=linux --arch=arm64 '--pack=sh,rpm,tar' 
```

## Bali build file format

Project file `bali.toml`:

```toml
# https://toml.io/en/
name = "bali"
summary = "Bali - Minimalist Golang build and packaging tool"
description = "Bali - Minimalist Golang build and packaging tool"
package-name = "bali-dev"
version = "3.0.1"
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

Built-in environment variables:

+ `BUILD_VERSION` is filled by the `version` field of balisrc.json
+ `BUILD_TIME` is filled by the build time formatted according to `RFC3339`
+ `BUILD_COMMIT` is filled by the commit id of the repository (when it is a git repository)
+ `BUILD_GOVERSION` is filled by `go version` output (removed `go version` prefix)
+ `BUILD_BRANCH` is filled with the branch name of the repository (when it is a git repository)

Other environment variables can be used in goflags.

Program build file `crate.toml`:

```toml
name = "bali"
description = "Bali - Minimalist Golang build and packaging tool"
destination = "bin"
version = "3.0.1"
goflags = [
    "-ldflags",
    "-X 'main.VERSION=$BUILD_VERSION' -X 'main.BUILD_TIME=$BUILD_TIME' -X 'main.BUILD_BRANCH=$BUILD_BRANCH' -X 'main.BUILD_COMMIT=$BUILD_COMMIT'  -X 'main.BUILD_REFNAME=$BUILD_REFNAME' -X 'main.BUILD_GOVERSION=$BUILD_GOVERSION'",
]

```

Windows-related manifest files (crate.toml sibling)：`winres.toml:`

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


Bali integrates [`goversioninfo`](https://github.com/josephspurrier/goversioninfo). When the target is Windows, it can embed version information (`winres.toml`) into the executable program.

The benefits of adding a reference program manifest are self-evident. For example, Windows UAC privilege escalation, Windows 10 long path support (ie path support> 260 characters), Windows Vista style controls, TaskDialog, DPI settings, etc. all need to modify the application manifest.

## Bootstrap

Usually after installing and configuring the Golang environment, you can follow the following command to complete Bali's bootstrapping:

UNIX:

```shell
./script/bootstrap.sh
```

Windows:

```ps1
# powershell
pwsh ./script/bootstrap.ps1
# cmd
script/bootstrap.bat
```


## Github Actions Use bali

```
go install github.com/balibuild/bali/v3/cmd/bali@latest
```

## Thanks

Bali's ability to automatically add version information to PE files is inseparable from the contribution of open source projects. Thank you very much [akavel/rsrc](https://github.com/akavel/rsrc) and [josephspurrier/goversioninfo](https://github.com/josephspurrier/goversioninfo) Developer and maintainer of two projects.

The Bali Github organization and Bali's own icons come from [www.flaticon.com](https://www.flaticon.com/) The creator is [Smashicons](https://www.flaticon.com/authors/smashicons) .

