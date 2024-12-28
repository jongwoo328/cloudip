# cloudip

[English](../README.md) | [한국어](./README_ko.md)

`cloudip`는 입력된 IP 주소가 어떤 클라우드 제공자(예: AWS, GCP, Azure 등)에 의해 관리되는지 식별하는 CLI 도구입니다. 단일 또는 여러 IP를 입력할 수 있으며, 결과를 간단한 텍스트 형식 또는 보기 좋은 테이블 형식으로 출력할 수 있습니다.

**🚨 경고 🚨**

이 프로젝트는 현재 개발 중에 있으며, 정식 릴리즈 전까지 기능과 옵션이 예고 없이 변경될 수 있습니다. 현재 버전은 AWS와 GCP만 지원하며, 추후 업데이트에서 추가 클라우드 제공자를 지원할 예정입니다.


## 주요 기능
- **단일 IP 확인**: 특정 IP가 어떤 클라우드 제공자에 속해 있는지 확인합니다.
- **다중 IP 확인**: 여러 IP 주소를 한 번에 검사할 수 있습니다.
- **IPv4 및 IPv6 지원**: IPv4와 IPv6 주소를 모두 지원합니다.
- **출력 형식**: `--format` 옵션을 사용해 출력 형식을 변경합니다.


## 설치
### Arch Linux
[cloudip](https://aur.archlinux.org/packages/cloudip)은 AUR에서 패키지로 제공됩니다.
AUR 도우미(e.g., `yay`)를 사용하여 설치할 수 있습니다:
```shell
yay -S cloudip
```

### 바이너리 다운로드
Release 페이지에서 최신 바이너리를 다운로드하세요.

**지원 플랫폼**:
- Linux (x86)
- Linux (x86_64)
- Linux (ARM64)
- macOS (x86_64)
- macOS (ARM64)


## 사용 방법
### 버전 확인
```shell
cloudip -v
```
출력:
```
0.4.0
```

### 단일 IP 확인
```shell
cloudip 54.230.176.25
```
출력:
```
54.230.176.25 aws
```

### 다중 IP 확인
```shell
cloudip 54.230.176.25 54.230.176.30 54.230.176.45
```
출력:
```
54.230.176.25 aws
54.230.176.30 aws
54.230.176.45 aws
```

### 구분자 지정
출력에 사용할 구분자를 지정할 수 있습니다. 기본 구분자는 공백입니다.
#### 쉼표(,) 구분
```shell
cloudip 54.230.176.25 --delimiter=','
```
출력:
```
54.230.176.25,aws
```

#### 탭(\t) 구분
```shell
cloudip 54.230.176.25 --delimiter=$'\t'
```
출력:
```
54.230.176.25   aws
```
그리고 다른 사용자 정의 구분자도 사용할 수 있습니다.

### 출력 형식
`--format` 옵션을 사용하여 출력 형식을 지정할 수 있습니다. 현재 지원되는 형식은 다음과 같습니다:
- `text` (기본값)
- `table`
- `json`

#### text
텍스트는 기본 출력 형식입니다.
```shell
cloudip 54.230.176.25
```
```shell
cloudip 54.230.176.25 --format=text
```
출력:
```
54.230.176.25 aws
```
`--header` 옵션을 사용하여 헤더를 출력할 수 있습니다.
```shell
cloudip 54.230.176.25 --header
```
출력:
```
IP Provider
54.230.176.25 aws
```

#### table
```shell
cloudip 54.230.176.25 --format=table
```
출력:
```
54.230.176.25   aws
```
테이블 형식 출력에 `--header` 옵션을 사용할 수 있습니다.
```shell
cloudip 54.230.176.25 --format=table --header
```
출력:
```
IP              PROVIDER 
54.230.176.25   aws
```

#### json
```shell
cloudip 54.230.176.25 --format=json
```
출력:
```json
[{"IP":"54.230.176.25","Provider":"aws"}]
```
`jq`를 사용하여 JSON 출력을 포맷팅할 수 있습니다.
```shell
cloudip 54.230.176.25 --format=json | jq
```
출력:
```json
[
  {
    "IP": "54.230.176.25",
    "Provider": "aws"
  }
]
```

#### csv
csv 형식은 `--format` 옵션의 값으로 지원하지 않습니다.

대신 `--format=text`와 `--delimiter=','` 옵션을 사용하여 csv 형식으로 출력할 수 있습니다.
헤더를 출력하려면 `--header` 옵션을 사용하세요.
```shell
cloudip 54.230.176.25 --format=text --delimiter=',' --header
```
출력:
```
IP,Provider
54.230.176.25,aws
```

---

## 소스에서 빌드
1. Go가 설치되어 있는지 확인하세요 (Go v1.20 이상 권장).
2. 다음 명령어를 사용하여 프로젝트를 빌드하세요:
   ```shell
   git clone https://github.com/jongwoo328/cloudip.git
   cd cloudip
   go mod tidy
   make build -j
   ```
3. 빌드된 바이너리는 `build/` 디렉토리에 생성됩니다.

---

## 라이선스
이 프로젝트는 [Apache License 2.0](./LICENSE)에 따라 라이선스가 부여됩니다.

Apache 2.0 라이선스의 조건에 따라 이 프로젝트를 사용할 수 있습니다. 자세한 내용은 LICENSE 파일을 참조하거나 [Apache License 공식 웹사이트](http://www.apache.org/licenses/LICENSE-2.0)를 방문하세요.
