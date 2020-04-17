# Bali - 极简的 Golang 构建打包工具

在开发基于 Golang 的项目时，虽然使用 go build 便可以简单的完成构建，但在打包等操作时，还是遇到了一些麻烦，因此，我实现了一个基于 PowerShell Core 开发的跨平台工具 bali，用来简化这一过程，最早在 2017 年，bali 诞生，而今已经到了 2020 年，我对项目构建打包也有了新的认识，因此为了改进基于 PowerShell 编写的 bali，因此在这个项目中使用 Golang 重写了 bali.

baligo 的命令行帮助信息如下：

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

## 特性

+   构建参数支持环境变量推导
+   打包，Windows zip, UNIX tar.gz
+   Windows 平台支持嵌入版本信息，图标，和应用程序清单。
+   UNIX 平台支持 STGZ 打包

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

Bali 选择了 JSON 作为文件格式，使用 JSON 的好处在于 Golang 内置支持解析，并且可以使用编辑器的格式化。Bali 构建文件有两种，一种是项目文件 `bali.json`，通常在项目根目录下，用于也可以在其他目录创建此文件，运行构建时，通过 `bali -w` 或者 `bali /path/to/buildroot` 指定 `bali.json` 所在目录，也可以在那个目录下运行 `bali`；另一种构建文件是特定程序源码目录下的 `balisrc.json` 文件，`balisrc.json` 所在目录应当存在 `main` 包，bali 通过解析 `bali.json` 的 `dirs` 解析 `balisrc.json`，与 `cmake` 的 `add_subdirectory` 指令类似。二者示例如下：

项目文件 `bali.json`:

```js
{
  // name 用于项目打包命名
    "name": "bali",
    // 用于打包的版本
    "version": "1.0.0",
    // 主要用于提示 bali 安装配置文件。
    "files": [
        {
            "path": "config/bali.json",
            "destination": "config"
        },
        {
            "path": "LICENSE",
            // 安装目录
            "destination": "share",
            // 安装/配置时重命名文件
            "newname": "LICENSE.bali",
            // 创建 STGZ 安装包时，不改名，即安装时如果存在相应文件则会覆盖，默认不会覆盖
            "norename": true
        }
    ],
    // 相对目录
    "dirs": [
        "cmd/bali"
    ]
}
```

程序构建文件 `balisrc.json`:

```js
{
    // 二进制文件名称，不存在时使用目录名
    "name": "bali",
    // 描述信息，默认填充到 PE 文件版本信息的 FileDescription
    "description": "Bali -  Minimalist Golang build and packaging tool",
    // 安装目录
    "destination": "bin",
    // 版本信息，在 goflags 中，可以推导 $BUILD_VERSION
    "version": "1.0.0",
    // 二进制的符号链接，比如在 GCC/Clang 编译后，程序为 GCC-9 然后会创建 GCC  的符号链接。
    "links": [
        "bin/baligo"
    ],
    // Go 编译器的参数，这些参数会使用 ExpandEnv 展开
    "goflags": [
        "-ldflags",
        "-X 'main.VERSION=$BUILD_VERSION' -X 'main.BUILDTIME=$BUILD_TIME' -X 'main.BUILDBRANCH=$BUILD_BRANCH' -X 'main.BUILDCOMMIT=$BUILD_COMMIT' -X 'main.GOVERSION=$BUILD_GOVERSION'"
    ],
    // 构建 Windows 目标，PE 文件的版本信息
    "versioninfo": "res/versioninfo.json",
    // 构建 Windows 目标，PE 文件的图标
    "icon": "res/bali.ico",
    // 构建 Windows 目标，PE 文件的应用程序清单
    "manifest": "res/bali.manifest"
}
```


## 感谢

Bali 自动添加版本信息到 PE 文件的功能离不开开源项目的贡献，在这里非常感谢 [akavel/rsrc](https://github.com/akavel/rsrc) 和 [josephspurrier/goversioninfo](https://github.com/josephspurrier/goversioninfo) 两个项目的开发者和维护者。

Bali Github 组织和 Bali 自身的图标来源于 [www.flaticon.com](https://www.flaticon.com/) 制作者为 [Smashicons](https://www.flaticon.com/authors/smashicons)。

