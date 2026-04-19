# CLAUDE.md

## Build

- **Version injection**: Binary version is set via ldflags at build time: `go build -o build/cloudip --ldflags '-X cloudip/cmd.Version=0.9.1'`
- **Release builds**: `make build` uses GoReleaser for cross-platform compilation
- **Version bump**: Update version string in these locations: `AGENTS.md`, `README.md`, `docs/README_ko.md`

## Architecture

- **Provider initialization**: Providers have no `init()` — all setup (metadata, data loading, CIDR tree) happens lazily via `Initialize()`. Metadata is prepared in `EnsureDataFile()`, data loading in `loadFunc`. `checkCloudIp()` calls `Initialize()` before each IP check, which guarantees setup. New providers must follow this pattern.

## Tasks

- `.tasks/` 디렉토리에 리팩토링 가이드, 테스트 계획 등 개발 내부 작업 문서를 관리한다. `.gitignore`에 등록되어 git에 추적되지 않는다.
- 파일명 규칙: `{YYYYMMDD}_{주제}.md` (예: `20260405_refactor.md`)
- 작업 진행 중 초기 계획과 다르게 구현/스펙/의사결정이 변경되면, 해당 task 문서 마지막에 `구현 중 변경된 사항` 같은 별도 섹션을 추가해 변경 내용을 명시한다.
- 완료 시: `{YYYYMMDD}_{주제}.done.md`로 rename
