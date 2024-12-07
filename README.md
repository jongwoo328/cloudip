# cloudip

[English](./README.md) | [한국어](./docs/README_ko.md)

`cloudip` is a CLI tool that identifies which cloud provider (e.g., AWS, GCP, Azure, etc.) manages the given IP address. You can input a single or multiple IPs, and the results can be displayed in either plain text or a well-formatted table.

**🚨 Warning 🚨**

This project is currently under development and supports **AWS only**. Additional cloud providers will be supported in future updates.


## Features
- **Single IP Check**: Identify which cloud provider owns a specific IP.
- **Multiple IP Check**: Check multiple IP addresses at once.
- **IPv4 and IPv6 Support**: Supports both IPv4 and IPv6 addresses.
- **Table Format Output**: Display results in a formatted table with the `--pretty` option.


## Installation
### Binary Download
Download the latest binary from the Releases page.

**Supported Platforms**:
- Linux (x86)
- Linux (x86_64)
- Linux (ARM64)
- macOS (x86_64)
- macOS (ARM64)


## Usage
### Single IP Check
```zsh
cloudip 54.230.176.25
```
Output:
```
54.230.176.25 aws
```

### Multiple IP Check
```zsh
cloudip 54.230.176.25 54.230.176.30 54.230.176.45
```
Output:
```
54.230.176.25 aws
54.230.176.30 aws
54.230.176.45 aws
```

### Custom Delimiters
#### Comma (,) Delimited
```zsh
cloudip 54.230.176.25 --delimiter ,
```
Output:
```
54.230.176.25,aws
```

#### Tab (\t) Delimited
```zsh
cloudip 54.230.176.25 --delimiter $'\t'
```
Output:
```
54.230.176.25   aws
```

### Table Format Output
```zsh
cloudip 54.230.176.25 --pretty
```
Output:
```
+---------------+----------+
|      IP       | PROVIDER |
+---------------+----------+
| 54.230.176.25 | aws      |
+---------------+----------+
```

---

## Build from Source
1. Ensure that Go is installed (Go v1.20 or later is recommended).
2. Use the `make` command to build the project:
   ```zsh
   git clone https://github.com/jongwoo328/cloudip.git
   go mod tidy
   cd cloudip
   make build -j
   ```
3. The binary will be generated in the `build/` directory.

---

## License
This project is licensed under the [Apache License 2.0](./LICENSE).

You may use this project in compliance with the terms and conditions of the Apache 2.0 License. For more details, see the LICENSE file or visit the [official Apache License website](http://www.apache.org/licenses/LICENSE-2.0).
