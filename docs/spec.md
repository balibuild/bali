# Bali Specification

|Attributes|Description|
|---|---|
|Status|Draft|
|Author|Force Charlie|
|Date|2020-04-27|

## Overview

Bali is a minimalist build and packaging tool for Golang projects based on Golang. Bali chose JSON as the file format. The advantage of using JSON is that Golang has built-in support for parsing, and it can be formatted using an editor. There are two types of Bali build files. One is the project file `bali.json`, which is usually in the root directory of the project. It can also be used to create this file in other directories. When running the build, use `bali -w` or `bali /path/to/buildroot` specifies the directory where `bali.json` is located, you can also run `bali` in that directory; another build file is the `balisrc.json` file under the specific program source code directory, `balisrc.json` There should be a `main` package in the directory where bali resolves `balisrc.json` by parsing `dirs` of `bali.json`, similar to the `add_subdirectory` instruction of `cmake`. 

|FileName|Location|Description|
|---|---|---|
|`bali.json`|Usually in the root directory of the project.||
|`balisrc.json`|Program source code directory|Parse its path through `bali.json`|

## bali.json Description

|Field|Type|Description|
|---|---|---|
|name|string|Project Name|
|version|string|Project Version|
|files|object|install files|
|dirs|string array|build executables|

File object:

|Field|Type|Description|
|---|---|---|
|path|string|file relative path|
|destination|string|install destination|
|newname|optional string|rename when install|
|norename|bool|no rename when package|

## balisrc.json Description

|Field|Type|Description|
|---|---|---|
|name|string|Executable Name|
|description|string|Executable description|
|destination|string|Executable install destination|
|version|string|Executable version|
|links|optional string array|Executable symlink when install and create package|
|goflags|optional string array|golang build flags|
|versioninfo|optional string|Executable versioninfo file|
|icon|optional string|EXE icon path|
|manifest|optional string|EXE manifest path|


`versioninfo.json` example:

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