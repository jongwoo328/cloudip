# cloudip

[English](./README.md) | [한국어](./docs/README_ko.md)

`cloudip` is a CLI tool that identifies cloud provider manages the given IP address. You can input a single or multiple IPs, and the results can be displayed in various formats, such as table, json.

**🚨 Warning 🚨**

This project is currently under development, and features and options may change without notice until the official release. 


## Features
- **Single IP Check**: Identify which cloud provider owns a specific IP.
- **Multiple IP Check**: Check multiple IP addresses at once.
- **IPv4 and IPv6 Support**: Supports both IPv4 and IPv6 addresses.
- **Format Output**: Display results in a formatted with the `--format` option.

### Currently Supported Cloud Providers
- **AWS**: Amazon Web Services
- **GCP**: Google Cloud Platform
- **Azure**: Microsoft Azure

## Installation
### Arch Linux
[cloudip](https://aur.archlinux.org/packages/cloudip) is available as a package on the AUR.
You can install it using an AUR helper (e.g., `yay`):
```shell
yay -S cloudip
```

### Binary Download
Download the latest binary from the [Releases](https://github.com/jongwoo328/cloudip/releases) page.

**Supported Platforms**:
- Linux (x86)
- Linux (x86_64)
- Linux (ARM64)
- macOS (x86_64)
- macOS (ARM64)


## Usage
### Version Check
```shell
cloudip version
```
Output:
```
0.6.0
```

### Single IP Check
```shell
cloudip 54.230.176.25
```
Output:
```
54.230.176.25 aws
```

### Multiple IP Check
```shell
cloudip 54.230.176.25 54.230.176.30 54.230.176.45
```
Output:
```
54.230.176.25 aws
54.230.176.30 aws
54.230.176.45 aws
```

### Custom Delimiters
You can specify a custom delimiter for the output. The default delimiter is a space.
#### Comma (,) Delimited
```shell
cloudip 54.230.176.25 --delimiter=','
```
Output:
```
54.230.176.25,aws
```

#### Tab (\t) Delimited
```shell
cloudip 54.230.176.25 --delimiter=$'\t'
```
Output:
```
54.230.176.25   aws
```
and any other custom delimiters can be used.

### Output Formats
The output format can be specified using the `--format` option. The following formats are supported:
- `text` (default)
- `table`
- `json`

#### text
Text is the default output format.
```shell
cloudip 54.230.176.25
```
```shell
cloudip 54.230.176.25 --format=text
```
Output:
```
54.230.176.25 aws
```
Use `--header` option to display the header.
```shell
cloudip 54.230.176.25 --header
```
Output:
```
IP Provider
54.230.176.25 aws
```

#### table
```shell
cloudip 54.230.176.25 --format=table
```
Output:
```
54.230.176.25   aws
```
`--header` option can be used with the table format.
```shell
cloudip 54.230.176.25 --format=table --header
```
Output:
```
IP              PROVIDER 
54.230.176.25   aws
```

#### json
```shell
cloudip 54.230.176.25 --format=json
```
Output:
```json
[{"IP":"54.230.176.25","Provider":"aws"}]
```
You can use with `jq` to format the JSON output.
```shell
cloudip 54.230.176.25 --format=json | jq
```
Output:
```json
[
  {
    "IP": "54.230.176.25",
    "Provider": "aws"
  }
]
```

#### csv
csv format is not supported `--format` option.

You can get output in CSV format using `--format=text` and `--delimiter=','` options. 
If you want to get result with header, use `--header` option.
 
```shell
cloudip 54.230.176.25 --format=text --delimiter=',' --header
```
Output:
```
IP,Provider
54.230.176.25,aws
```

### Verbose Output
You can use the `--verbose`, `-v` option to display detailed information.
```shell
cloudip --verbose 54.230.176.25
```
Output:
```
AWS IP ranges file not exists.
Downloading AWS IP ranges...
AWS IP ranges updated [2024-12-27 04:12:30]
54.230.176.25 aws
```


---

## Build from Source
1. Ensure that Go is installed (Go v1.20 or later is recommended).
2. Use the `make` command to build the project:
   ```shell
   git clone https://github.com/jongwoo328/cloudip.git
   go mod tidy
   cd cloudip
   go build -o dist --ldflags '-X cloudip/cmd.Version=0.6.0'
   ```
3. The binary will be generated in the `build/` directory.

---

## License
This project is licensed under the [Apache License 2.0](./LICENSE).

You may use this project in compliance with the terms and conditions of the Apache 2.0 License. For more details, see the LICENSE file or visit the [official Apache License website](http://www.apache.org/licenses/LICENSE-2.0).
