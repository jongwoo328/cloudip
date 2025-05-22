# cloudip

[English](./README.md) | [한국어](./docs/README_ko.md)

`cloudip` is a CLI tool that identifies which cloud provider manages the given IP address. You can input a single or multiple IPs, and the results can be displayed in various formats, such as table, json.

**🚨 Warning 🚨**

This project is currently under development, and features and options may change without notice until the official release. 

## Table of Contents
- [Features](#features)
- [Currently Supported Cloud Providers](#currently-supported-cloud-providers)
- [Installation](#installation)
  - [Arch Linux](#arch-linux)
  - [Binary Download](#binary-download)
- [Usage](#usage)
  - [Basic Usage](#basic-usage)
  - [Output Options](#output-options)
    - [Custom Delimiters](#custom-delimiters)
    - [Output Formats](#output-formats)
  - [Other Options](#other-options)
- [Build from Source](#build-from-source)
- [License](#license)

## Features
- **Single IP Check**: Identify which cloud provider owns a specific IP.
- **Multiple IP Check**: Check multiple IP addresses at once.
- **IPv4 and IPv6 Support**: Supports both IPv4 and IPv6 addresses.
- **Format Output**: Display results in various formats using the `--format` option.

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

### Basic Usage
- Version Check
  ```shell
  cloudip version
  ```
  Output:
  ```text
  0.6.0
  ```

- Single IP Check
  ```shell
  cloudip 54.230.176.25
  ```
  Output:
  ```text
  54.230.176.25 aws
  ```

- Multiple IP Check
  ```shell
  cloudip 54.230.176.25 54.230.176.30 54.230.176.45
  ```
  Output:
  ```text
  54.230.176.25 aws
  54.230.176.30 aws
  54.230.176.45 aws
  ```

### Output Options
- #### Custom Delimiters
  You can specify a custom delimiter for the output. The default delimiter is a space.
  - Comma (,) Delimited
    ```shell
    cloudip 54.230.176.25 --delimiter=','
    ```
    Output:
    ```text
    54.230.176.25,aws
    ```

  - Tab (\t) Delimited
    ```shell
    cloudip 54.230.176.25 --delimiter=$'\t'
    ```
    Output:
    ```text
    54.230.176.25   aws
    ```
  and any other custom delimiters can be used.

- #### Output Formats
  Use the `--format` option to specify the output format. Supported formats include:

  - `text` (default): Displays results as simple text.
    ```shell
    cloudip 54.230.176.25 --format=text
    ```
    Output:
    ```text
    54.230.176.25 aws
    ```
    Use the `--header` option to include a header row.
    ```shell
    cloudip 54.230.176.25 --format=text --header
    ```
    Output:
    ```text
    IP Provider
    54.230.176.25 aws
    ```

  - `table`: Displays results in a table format.
    ```shell
    cloudip 54.230.176.25 --format=table --header
    ```
    Output:
    ```text
    IP              PROVIDER 
    54.230.176.25   aws
    ```

  - `json`: Outputs results in JSON format, suitable for parsing with tools like `jq`.
    ```shell
    cloudip 54.230.176.25 --format=json
    ```
    Output:
    ```json
    [{"IP":"54.230.176.25","Provider":"aws"}]
    ```

  - `csv`: This tool does not have a direct `--format=csv` option. 
    However, you can produce CSV-like output by combining `--format=text` with `--delimiter=','`.
    To include a header row, also add the `--header` option.
    ```shell
    cloudip 54.230.176.25 --format=text --delimiter=',' --header
    ```
    Output:
    ```csv
    IP,Provider
    54.230.176.25,aws
    ```

### Other Options
- Verbose Output
  You can use the `--verbose`, `-v` option to display detailed information.
  ```shell
  cloudip --verbose 54.230.176.25
  ```
  Output:
  ```text
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
