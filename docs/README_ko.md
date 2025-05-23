# cloudip

[English](../README.md) | [한국어](./README_ko.md)

`cloudip`는 입력된 IP 주소가 속하는 클라우드 제공자를 식별하는 CLI 도구입니다. 단일 또는 여러 IP를 입력할 수 있으며, 결과를 json, table과 같은 다양한 형식으로 출력할 수 있습니다.

**🚨 경고 🚨**

이 프로젝트는 현재 개발 중에 있으며, 정식 릴리즈 전까지 기능과 옵션이 예고 없이 변경될 수 있습니다.

## 목차
- [주요 기능](#주요-기능)
- [현재 지원되는 클라우드 제공자](#현재-지원되는-클라우드-제공자)
- [설치](#설치)
  - [Arch Linux](#arch-linux)
  - [바이너리 다운로드](#바이너리-다운로드)
- [사용 방법 (Usage)](#사용-방법-usage)
  - [기본 사용법 (Basic Usage)](#기본-사용법-basic-usage)
  - [출력 옵션 (Output Options)](#출력-옵션-output-options)
    - [구분자 지정 (Delimiter Specification)](#구분자-지정-delimiter-specification)
    - [출력 형식 (Output Formats)](#출력-형식-output-formats)
  - [기타 옵션 (Other Options)](#기타-옵션-other-options)
- [소스에서 빌드](#소스에서-빌드)
- [라이선스](#라이선스)

## 주요 기능
- **단일 IP 확인**: 특정 IP가 어떤 클라우드 제공자에 속해 있는지 확인합니다.
- **다중 IP 확인**: 여러 IP 주소를 한 번에 검사할 수 있습니다.
- **IPv4 및 IPv6 지원**: IPv4와 IPv6 주소를 모두 지원합니다.
- **출력 형식**: `--format` 옵션을 사용해 출력 형식을 변경합니다.

### 현재 지원되는 클라우드 제공자
- AWS (Amazon Web Services)
- GCP (Google Cloud Platform)
- Azure (Microsoft Azure)

## 설치
### Arch Linux
[cloudip](https://aur.archlinux.org/packages/cloudip)은 AUR에서 패키지로 제공됩니다.
AUR 도우미(e.g., `yay`)를 사용하여 설치할 수 있습니다:
```shell
yay -S cloudip
```

### 바이너리 다운로드
[Releases](https://github.com/jongwoo328/cloudip/releases) 페이지에서 최신 바이너리를 다운로드하세요.

**지원 플랫폼**:
- Linux (x86)
- Linux (x86_64)
- Linux (ARM64)
- macOS (x86_64)
- macOS (ARM64)


## 사용 방법 (Usage)

### 기본 사용법 (Basic Usage)
- 버전 확인 (Version Check)
  ```shell
  cloudip version
  ```
  출력:
  ```text
  0.6.0
  ```

- 단일 IP 확인 (Single IP Check)
  ```shell
  cloudip 54.230.176.25
  ```
  출력:
  ```text
  54.230.176.25 aws
  ```

- 다중 IP 확인 (Multiple IP Check)
  ```shell
  cloudip 54.230.176.25 54.230.176.30 54.230.176.45
  ```
  출력:
  ```text
  54.230.176.25 aws
  54.230.176.30 aws
  54.230.176.45 aws
  ```

### 출력 옵션 (Output Options)
- #### 구분자 지정 (Delimiter Specification)
  출력에 사용할 구분자를 지정할 수 있습니다. 기본 구분자는 공백입니다.
  - 쉼표(,) 구분 (Comma (,) Delimited)
    ```shell
    cloudip 54.230.176.25 --delimiter=','
    ```
    출력:
    ```text
    54.230.176.25,aws
    ```

  - 탭(\t) 구분 (Tab (\t) Delimited)
    ```shell
    cloudip 54.230.176.25 --delimiter=$'\t'
    ```
    출력:
    ```text
    54.230.176.25   aws
    ```
  그리고 다른 사용자 정의 구분자도 사용할 수 있습니다.

- #### 출력 형식 (Output Formats)
  `--format` 옵션으로 출력 형식을 지정합니다. 지원되는 형식은 다음과 같습니다:
  - `text` (기본값): 간단한 텍스트로 결과를 표시합니다.
    ```shell
    cloudip 54.230.176.25 --format=text
    ```
    출력:
    ```text
    54.230.176.25 aws
    ```
    `--header` 옵션으로 헤더를 추가할 수 있습니다.
    ```shell
    cloudip 54.230.176.25 --format=text --header
    ```
    출력:
    ```text
    IP Provider
    54.230.176.25 aws
    ```

  - `table`: 표 형태로 정렬하여 결과를 표시합니다.
    ```shell
    cloudip 54.230.176.25 --format=table --header
    ```
    출력:
    ```text
    IP              PROVIDER 
    54.230.176.25   aws
    ```

  - `json`: JSON 형식으로 결과를 출력합니다. `jq`와 같은 도구로 파싱하기 용이합니다.
    ```shell
    cloudip 54.230.176.25 --format=json
    ```
    출력:
    ```json
    [{"IP":"54.230.176.25","Provider":"aws"}]
    ```

  - `csv`: CSV 형식은 `--format=csv` 옵션을 직접 지원하지 않습니다. 
    대신, `--format=text` 와 `--delimiter=','` 옵션을 함께 사용하여 CSV와 유사한 형식으로 출력할 수 있습니다. 헤더를 포함하려면 `--header` 옵션을 추가합니다.
    ```shell
    cloudip 54.230.176.25 --format=text --delimiter=',' --header
    ```
    출력:
    ```csv
    IP,Provider
    54.230.176.25,aws
    ```

### 기타 옵션 (Other Options)
- 상세 출력 (Verbose Output)
`--verbose` 또는 `-v` 옵션을 사용하여 상세한 출력을 확인할 수 있습니다.
  ```shell
  cloudip --verbose 54.230.176.25
  ```
  출력:
  ```text
  AWS IP ranges file not exists.
  Downloading AWS IP ranges...
  AWS IP ranges updated [2024-12-27 04:12:30]
  54.230.176.25 aws
  ```

---

## 소스에서 빌드
1. Go가 설치되어 있는지 확인하세요 (Go v1.20 이상 권장).
2. 다음 명령어를 사용하여 프로젝트를 빌드하세요:
   ```shell
   git clone https://github.com/jongwoo328/cloudip.git
   cd cloudip
   go mod tidy
   go build -o dist --ldflags '-X cloudip/cmd.Version=0.6.0'
   ```
3. 빌드된 바이너리는 `build/` 디렉토리에 생성됩니다.

---

## 라이선스
이 프로젝트는 [Apache License 2.0](../LICENSE)에 따라 라이선스가 부여됩니다.

Apache 2.0 라이선스의 조건에 따라 이 프로젝트를 사용할 수 있습니다. 자세한 내용은 LICENSE 파일을 참조하거나 [Apache License 공식 웹사이트](http://www.apache.org/licenses/LICENSE-2.0)를 방문하세요.
