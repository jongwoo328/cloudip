# CLAUDE.md

## Build

- **Version injection**: Binary version is set via ldflags at build time: `go build -trimpath -o build/cloudip --ldflags '-X cloudip/cmd.Version=0.11.0'`
- **Release builds**: `make build` uses GoReleaser for cross-platform compilation
- **Version bump**: Update version string in these locations: `AGENTS.md`, `README.md`, `docs/README_ko.md`

## Architecture

- **Provider initialization**: Providers have no `init()` — all setup (metadata, data loading, CIDR tree) happens lazily via `Initialize()`. Metadata is prepared in `EnsureDataFile()`, data loading in `loadFunc`. `checkCloudIp()` calls `Initialize()` before each IP check, which guarantees setup. New providers must follow this pattern.

## Tasks

- `.tasks/` 디렉토리에 리팩토링 가이드, 테스트 계획 등 개발 내부 작업 문서를 관리한다. `.gitignore`에 등록되어 git에 추적되지 않는다.
- 파일명 규칙: `{YYYYMMDD}_{주제}.md` (예: `20260405_refactor.md`)
- `.tasks/`에 정리된 작업을 실제 코드/문서 변경으로 진행할 때는 먼저 해당 작업용 GitHub issue를 생성한다.
- 작업 브랜치는 생성된 issue 번호를 포함해 만든다. 기존 분류 규칙을 따른다: `feat/#번호`, `fix/#번호`, `refactor/#번호`, `test/#번호`, `docs/#번호`.
- 브랜치는 최신 `main` 기준으로 생성하고, 작업 범위에 해당하는 파일만 commit한다. `.tasks/` 문서는 내부 작업 문서이므로 별도 요청이 없으면 commit하지 않는다.
- PR 제목은 `{type}(#번호): {요약}` 형식을 사용하고, 본문에는 기존 PR 형식에 맞춰 `Close #번호`를 적는다.
- 작업 진행 중 초기 계획과 다르게 구현/스펙/의사결정이 변경되면, 해당 task 문서 마지막에 `구현 중 변경된 사항` 같은 별도 섹션을 추가해 변경 내용을 명시한다.
- 완료 시: `{YYYYMMDD}_{주제}.done.md`로 rename
