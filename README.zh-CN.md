# Bali - 极简的 Golang 构建打包工具

[![Master Branch Status](https://github.com/balibuild/bali/workflows/CI/badge.svg)](https://github.com/balibuild/bali/actions)

Bali 是一个使用 Golang 开发的*极简 Golang 构建打包工具*，[Bali(old)](https://github.com/fcharlie/bali) 最初使用 PowerShell 编写，用于解决项目构建过程中的打包和配置设置等问题。[Bali(old)](https://github.com/fcharlie/bali) 在构建时能够将，时间，git 的 commit 等信息填充到构建的二进制文件中，在追溯应用程序缺陷时有一定的用处。但基于 PowerShell 开发的 Bali 并没有直接支持创建 `STGZ` 安装包，`STGZ` 安装包本质上将 `Shell Script` 头与 `tar.gz` 合并在一起，然后分发，用户在安装时，由 `Shell Script` 调用解压缩命令解压完成安装，在 Shell 脚本中，还可以设置好 `post_install` 之类的脚本在解压后执行相关操作。`STGZ` 小巧而又强大，在近两年的开发过程中，我开发的 Linux 平台项目大多都实现了创建 STGZ 安装包的功能。实际上 Bali 完全可以实现这一功能，考虑到 Golang 内置了 `archive/tar` `archive/zip` `compress/gzip` 实际上如果使用 Golang 重新实现 Bali 要获得更大的受益。这便是 [Baligo](https://github.com/fcharlie/baligo)(Gitee: [Baligo](https://gitee.com/ipvb/baligo))。但我仍然觉得 Baligo 还有所欠缺，比如 Windows 平台不支持嵌入图标，版本信息，清单文件，因此这才有了新的 **Bali** 诞生。


## 功能

Bali 有一些功能是我觉得有些用处的：

+   构建参数支持环境变量推导
+   打包，创建压缩包，target 为 Windows 时创建 zip, target 为其他时，创建 tar.gz。
+   Windows 平台支持嵌入版本信息，图标，和应用程序清单。
+   UNIX 平台支持 STGZ 打包


bali 的命令行帮助信息如下：

```shell
Bali -  Minimalist Golang build and packaging tool
usage: ./build/bin/bali <option> args ...
  -h|--help        Show usage text and quit
  -v|--version     Show version number and quit
  -V|--verbose     Make the operation more talkative
  -F|--force       Turn on force mode. eg: Overwrite configuration file
  -w|--workdir     Specify bali running directory. (Position 0, default $PWD)
  -a|--arch        Build arch: amd64 386 arm arm64
  -t|--target      Build target: windows linux darwin ...
  -o|--out         Specify build output directory. default '$PWD/build'
  -d|--dest        Specify the path to save the package
  -z|--zip         Create archive file (UNIX: .tar.gz, Windows: .zip)
  -p|--pack        Create installation package (UNIX: STGZ, Windows: none)
  --cleanup        Cleanup build directory
  --no-rename      Disable file renaming (STGZ installation package, default: OFF)

```

## 使用方法

普通构建：

```shell
bali /path/to/project
```

创建 `Tar.gz` 压缩包：

```shell
bali /path/to/project -z
```

创建 `STGZ` 安装包，主要用于 Linux/macOS 平台：

```shell
bali /path/to/project -p
```

将安装包输出到指定目录：

```shell
bali /path/to/project -p -d /tmp/output
# # bali /path/to/project -p -d/tmp/output
# bali /path/to/project -p -d=/tmp/output
# bali /path/to/project -p --dist=/tmp/output
# bali /path/to/project -p --dist /tmp/output
```

## Bali 构建文件格式

Bali 同时支持 TOML 或者 JSON 格式的项目文件，JSON 使用内置解析不支持注释，TOML 支持注释。Bali 构建文件有两种，一种是项目文件 `bali.toml`，通常在项目根目录下，用于也可以在其他目录创建此文件，运行构建时，通过 `bali -w` 或者 `bali /path/to/buildroot` 指定 `bali.toml` 所在目录，也可以在那个目录下运行 `bali`；另一种构建文件是特定程序源码目录下的 `balisrc.toml` 文件，`balisrc.toml` 所在目录应当存在 `main` 包，bali 通过解析 `bali.toml` 的 `dirs` 解析 `balisrc.toml`，与 `cmake` 的 `add_subdirectory` 指令类似。二者示例如下：

项目文件 `bali.toml`:

```toml
# https://toml.io/en/
name = "bali"
version = "2.1.1"
dirs = [
    "cmd/bali", # dirs
]

[[files]]
path = "LICENSE"
destination = "share"
newname = "LICENSE.bali"
norename = true

```

程序构建文件 `balisrc.toml`:

```toml
name = "bali"
description = "Bali - Minimalist Golang build and packaging tool"
destination = "bin"
version = "2.1.1"
versioninfo = "res/versioninfo.json"
icon = "res/bali.ico"
manifest = "res/bali.manifest"
links = ["bin/baligo"]
goflags = [
    "-ldflags",
    "-X 'main.VERSION=$BUILD_VERSION' -X 'main.BUILDTIME=$BUILD_TIME' -X 'main.BUILDBRANCH=$BUILD_BRANCH' -X 'main.BUILDCOMMIT=$BUILD_COMMIT' -X 'main.GOVERSION=$BUILD_GOVERSION'",
]

```

内置环境变量：

+   `BUILD_VERSION` 由 balisrc.toml 的 `version` 字段填充
+   `BUILD_TIME` 由构建时间按照 `RFC3339` 格式化后填充
+   `BUILD_COMMIT` 由存储库（为 git 存储库时） 的 commit id 填充
+   `BUILD_GOVERSION` 由 `go version` 输出（删除了 `go version` 前缀）填充
+   `BUILD_BRANCH` 由存储库（为 git 存储库时） 的分支名填充

可以在 goflags 中使用其他环境变量。

`versioninfo.json:`

```json
{
	"FixedFileInfo": {
		"FileVersion": {
			"Major": 0,
			"Minor": 0,
			"Patch": 0,
			"Build": 0
		},
		"ProductVersion": {
			"Major": 0,
			"Minor": 0,
			"Patch": 0,
			"Build": 0
		},
		"FileFlagsMask": "3f",
		"FileFlags ": "00",
		"FileOS": "40004",
		"FileType": "01",
		"FileSubType": "00"
	},
	"StringFileInfo": {
		"Comments": "",
		"CompanyName": "Bali Team",
		"FileDescription": "Bali - Minimalist Golang build and packaging tool",
		"FileVersion": "",
		"InternalName": "bali.exe",
		"LegalCopyright": "Copyright \u00A9 2022. Bali contributors",
		"LegalTrademarks": "",
		"OriginalFilename": "bali.exe",
		"PrivateBuild": "",
		"ProductName": "Bali",
		"ProductVersion": "1.0",
		"SpecialBuild": ""
	},
	"VarFileInfo": {
		"Translation": {
			"LangID": "0409",
			"CharsetID": "04B0"
		}
	}
}
```

Bali 整合了 [`goversioninfo`](https://github.com/josephspurrier/goversioninfo)，在目标为 Windows 时，能够将版本信息嵌入到可执行程序中，`versioninfo` 字段与 `goversioninfo` 项目类似，但更宽松，一些特定的值，比如版本，描述会使用 `bali.toml/balisrc.toml` 的值填充过去，`icon`/`manifest` 则会覆盖 `versioninfo.json`。

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

