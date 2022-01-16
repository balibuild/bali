# Bali -  Minimalist Golang build and packaging tool

[![Master Branch Status](https://github.com/balibuild/bali/workflows/CI/badge.svg)](https://github.com/balibuild/bali/actions)


[简体中文](./README.zh-CN.md)

Bali is a *minimal Golang build and packaging tool* developed using Golang. [Bali(old)](https://github.com/fcharlie/bali) was originally written in PowerShell and used to solve packaging and configuration issues during the project build process.

[Bali(old)](https://github.com/fcharlie/bali) can fill information such as time, git commit, etc. into the built binary file at build time, which is of some use when tracing application defects.

However, Bali developed based on PowerShell does not support the creation of `STGZ` installation package. The `STGZ` installation package essentially merges the `Shell Script` header with `tar.gz` and distributes it. When the user installs, the `Shell Script` calls the decompression command to decompress to complete the installation. In the shell script, you can also set up a script such as `post_install` to perform related operations after decompression. `STGZ` is small and powerful. In the past two years of development, most of the Linux platform projects that I have developed have realized the function of creating STGZ installation packages.

In fact, Bali can fully achieve this function, considering that Golang has built-in `archive/tar`, `archive/zip`, `compress/gzip`. In fact, if you use Golang to re-implement Bali, you will get greater benefits. This is [Baligo](https://github.com/fcharlie/baligo) (Gitee: [Baligo](https://gitee.com/ipvb/baligo)). But I still think that Baligo still lacks, such as the Windows platform does not support embedded `icons`, `version information`, `manifest files`, so this is where the new **Bali** was born.

## Feature

Bali has some functions that I think are useful:

+ Build parameters support derivation of environment variables
+ Package, create compressed package, create zip when target is Windows, and create tar.gz when target is other.
+ The Windows platform supports embedded version information, icons, and application manifest.
+ UNIX platform supports STGZ packaging


Bali's command line help information is as follows:

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

## Instructions

Common build:

```shell
bali /path/to/project
```

Create `Tar.gz` compressed package:

```shell
bali /path/to/project -z
```

Create `STGZ` installation package, mainly used on Linux/macOS platform:

```shell
bali /path/to/project -p
```

Output the installation package to the specified directory:

```shell
bali /path/to/project -p -d /tmp/output
# # bali /path/to/project -p -d/tmp/output
# bali /path/to/project -p -d=/tmp/output
# bali /path/to/project -p --dist=/tmp/output
# bali /path/to/project -p --dist /tmp/output
```

## Bali build file format

You can choose to write project files in json or toml format.. There are two types of Bali build files. One is the project file `bali.toml`, which is usually in the root directory of the project. It can also be used to create this file in other directories. When running the build, use `bali -w` or `bali /path/to/buildroot` specifies the directory where `bali.toml` is located, you can also run `bali` in that directory; another build file is the `balisrc.toml` file under the specific program source code directory, `balisrc.toml` There should be a `main` package in the directory where bali resolves `balisrc.toml` by parsing `dirs` of `bali.toml`, similar to the `add_subdirectory` instruction of `cmake`. Examples of both are as follows:


Project file `bali.toml`:

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

Built-in environment variables:

+ `BUILD_VERSION` is filled by the `version` field of balisrc.json
+ `BUILD_TIME` is filled by the build time formatted according to `RFC3339`
+ `BUILD_COMMIT` is filled by the commit id of the repository (when it is a git repository)
+ `BUILD_GOVERSION` is filled by `go version` output (removed `go version` prefix)
+ `BUILD_BRANCH` is filled with the branch name of the repository (when it is a git repository)

Other environment variables can be used in goflags.

Program build file `balisrc.toml`:

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

Bali integrates [`goversioninfo`](https://github.com/josephspurrier/goversioninfo). When the target is Windows, it can embed version information into the executable program. The `versioninfo` field is similar to the `goversioninfo` project. But more loosely, some specific values, such as version, description will be filled with the value of `bali.toml/balisrc.toml`, and `icon`/`manifest` will override `versioninfo.json`.

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
go install github.com/balibuild/bali/v2/cmd/bali@latest
```

peassets: 

```
go install github.com/balibuild/bali/v2/cmd/peassets@latest
```

## Thanks

Bali's ability to automatically add version information to PE files is inseparable from the contribution of open source projects. Thank you very much [akavel/rsrc](https://github.com/akavel/rsrc) and [josephspurrier/goversioninfo](https://github.com/josephspurrier/goversioninfo) Developer and maintainer of two projects.

The Bali Github organization and Bali's own icons come from [www.flaticon.com](https://www.flaticon.com/) The creator is [Smashicons](https://www.flaticon.com/authors/smashicons) .

