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
  - [다운로드 및 캐시 동작](#다운로드-및-캐시-동작)
  - [에러 처리 (Error Handling)](#에러-처리-error-handling)
- [소스에서 빌드](#소스에서-빌드)
- [라이선스](#라이선스)

## 주요 기능
- **단일 IP 확인**: 특정 IP가 어떤 클라우드 제공자에 속해 있는지 확인합니다.
- **다중 IP 확인**: 여러 IP 주소를 한 번에 검사할 수 있습니다.
- **IPv4 및 IPv6 지원**: IPv4와 IPv6 주소를 모두 지원합니다.
- **출력 형식**: `--format` 옵션을 사용해 출력 형식을 변경합니다.
- **제공자 업데이트 캐시**: 제공자 데이터 업데이트 확인은 기본적으로 24시간 동안 캐시됩니다.

### 현재 지원되는 클라우드 제공자
- AWS (Amazon Web Services)
- GCP (Google Cloud Platform)
- Azure (Microsoft Azure)
- Cloudflare (Cloudflare, Inc.)
- Vultr (The Constant Company)

## 설치
### Homebrew (macOS 전용)
```shell
brew tap jongwoo328/cloudip
brew install cloudip
```

### Arch Linux
> **참고**: AUR 패키지는 더 이상 관리되지 않습니다.

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

IP 조회에는 최소 하나의 IP 인자가 필요합니다: `cloudip IP [IP...]`. `--help`와 `version`은 IP 없이 사용할 수 있습니다.

- 버전 확인 (Version Check)
  ```shell
  cloudip version
  ```
  출력:
  ```text
  0.11.1
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

  - `table`: 표 형태로 결과를 표시합니다. 컬럼 너비에 맞춰 공백으로 여백을 채워 보기 쉽게 정렬합니다. 헤더는 항상 출력됩니다.
    ```shell
    cloudip 54.230.176.25 --format=table
    ```
    출력:
    ```text
    IP            Provider
    54.230.176.25 aws
    ```
    `--delimiter` 옵션을 `table` 형식에서도 사용할 수 있습니다.
    ```shell
    cloudip 54.230.176.25 --format=table --delimiter=' | '
    ```
    출력:
    ```text
    IP            | Provider
    54.230.176.25 | aws
    ```

  - `json`: JSON 형식으로 결과를 출력합니다. `jq`와 같은 도구로 파싱하기 용이합니다.
    ```shell
    cloudip 54.230.176.25 --format=json
    ```
    출력:
    ```json
    [{"ip":"54.230.176.25","provider":"aws","error":""}]
    ```
    JSON 출력은 `ip`, `provider`, `error` 소문자 키를 사용합니다. IP 검사에 실패하면 `provider`는 `error`가 되고 `error` 필드에 실패 원인이 들어갑니다.

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
  `--verbose` 또는 `-v` 옵션의 진단 로그는 stderr로, 조회 결과는 stdout으로 출력됩니다. 따라서 `--format=json --verbose`를 함께 사용해도 stdout은 JSON 형식을 유지합니다.
  ```shell
  cloudip --verbose 54.230.176.25
  ```
  stderr 예시 (캐시 상태에 따라 달라짐):
  ```text
  AWS IP ranges file does not exist.
  Downloading AWS IP ranges...
  AWS IP ranges updated [2024-12-27 04:12:30]
  ```
  stdout:
  ```text
  54.230.176.25 aws
  ```

- 제공자 데이터 업데이트 건너뛰기
  로컬 데이터 파일이 이미 있는 경우, `cloudip`는 기본적으로 제공자 데이터 업데이트 확인을 최대 24시간에 한 번만 수행합니다. `--no-update`를 사용하면 제공자 업데이트 확인을 건너뛰고 로컬 데이터만 사용합니다.
  ```shell
  cloudip --no-update 54.230.176.25
  ```
  필요한 로컬 제공자 데이터 파일이 없으면 `--no-update`는 파일을 다운로드하지 않고 에러를 반환합니다.

### 다운로드 및 캐시 동작
각 데이터 다운로드 요청에는 응답 본문 수신을 포함해 30초 제한이 적용됩니다. 이는 전체 명령의 실행 시간 제한이 아니며, 여러 제공자나 요청을 거치면 더 오래 걸릴 수 있습니다.

업데이트 다운로드나 데이터 검증에 실패해도 유효한 기존 캐시가 있으면 진단 메시지를 stderr로 출력하고 해당 캐시로 조회를 계속합니다. 실패한 다운로드나 잘못된 데이터로 기존 캐시를 덮어쓰지 않습니다. 사용할 캐시가 없거나 유효하지 않으면 해당 제공자의 초기화가 실패합니다.

여러 IP를 한 번에 조회할 때 초기화에 실패한 제공자는 같은 배치에서 다시 초기화하지 않습니다. 다른 정상 제공자의 조회는 계속하며, 다음 실행에서는 실패한 제공자를 다시 시도합니다.

### 에러 처리 (Error Handling)
IP 인자가 없거나 `--format`이 `text`, `table`, `json` 이외의 값(빈 값 포함)이면 제공자 초기화 및 조회 전에 오류를 반환하며 결과 행은 출력하지 않습니다. IP 인자가 없으면 사용법도 표시합니다.

캐시 경로의 접근 권한 부족, 잘못된 경로, 파일이 있어야 할 위치의 디렉터리는 파일 접근 오류로 처리합니다. 앱 저장 디렉터리 생성에 실패하면 진단 메시지를 stderr로 출력하고 non-zero 종료 코드로 종료합니다.

입력 검증을 통과한 뒤 하나 이상의 IP 검사에 실패하면 `cloudip`는 모든 결과 행을 출력한 뒤 non-zero 종료 코드를 반환합니다. `text`와 `table` 형식에서는 실패한 행의 provider 컬럼에 `ERROR`를 표시하고, 상세 에러 메시지는 stderr로 출력합니다. `json` 형식에서는 각 행의 `error` 필드에 에러 원인을 포함합니다.

---

## 소스에서 빌드
1. Go가 설치되어 있는지 확인하세요 (Go v1.25 이상 필요).
2. 프로젝트를 클론하고 빌드하세요:
   ```shell
   git clone https://github.com/jongwoo328/cloudip.git
   cd cloudip
   go mod tidy
   go build -trimpath -o build/cloudip --ldflags '-X cloudip/cmd.Version=0.11.1'
   ```
3. 바이너리 `cloudip`가 `build/` 디렉토리에 생성됩니다.

**메인테이너 참고사항**: 릴리즈 빌드의 경우 `make build` 명령어를 사용하세요. 이는 GoReleaser를 통해 크로스 플랫폼 컴파일 및 배포를 수행합니다.

## 개발 및 테스트

### 테스트
다음 명령어를 사용하여 테스트를 실행할 수 있습니다:

```bash
# 모든 테스트 실행
make test

# 상세 출력으로 테스트 실행
make test-verbose

# 커버리지 보고서와 함께 테스트 실행
make test-coverage

# 벤치마크 테스트 실행
make test-bench
```

---

## 라이선스
이 프로젝트는 [Apache License 2.0](../LICENSE)에 따라 라이선스가 부여됩니다.

Apache 2.0 라이선스의 조건에 따라 이 프로젝트를 사용할 수 있습니다. 자세한 내용은 LICENSE 파일을 참조하거나 [Apache License 공식 웹사이트](http://www.apache.org/licenses/LICENSE-2.0)를 방문하세요.
