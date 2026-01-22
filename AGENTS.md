# Repository Guidelines

## Project Overview & Architecture
Daily Summary is a macOS-focused Go CLI that records work entries, schedules reminders, and generates daily AI summaries. It supports Codex (default) and Claude Code providers, stores per-day JSON records, and writes Markdown summaries. A launchd service keeps it running in the background; a file lock prevents duplicate processes, and a reset signal file is used for inter-process coordination.

## Project Structure & Module Organization
- Entry point: `main.go`; module deps: `go.mod`, `go.sum`.
- Core packages under `internal/`: `cli`, `scheduler`, `summary`, `storage`, `dialog`, `models`.
- Config: `config.yaml` (see `config.example.yaml` and `config.pomodoro.yaml`).
- Runtime data: `run/data`, `run/summaries`, `run/logs`, plus `run/.reset_signal` and `run/daily_summary.lock`.
- Supporting materials: `templates/` (prompt templates), `scripts/` (install/test helpers), `docs/` (guides), `launchd/` (service plist).

## Build, Test, and Development Commands
- `go build -o daily_summary`: build local binary.
- `GOOS=darwin GOARCH=amd64 go build -o daily_summary` (or `arm64`): cross-compile if needed.、
- `go run main.go --config ./config.yaml`: run with explicit config.
- `go test ./...`: run all tests; `go test -run TestName ./internal/scheduler/` for focused runs.
- `go mod tidy` or `go mod download`: dependency cleanup or fetch.
- `./scripts/quick_test.sh`, `./scripts/test_config.sh`, `./scripts/test_dialog.sh`, `./scripts/test_minute_interval.sh`: targeted checks.
- Logs: `tail -f ./run/logs/app.log` (app), `./run/logs/stdout.log`, `./run/logs/stderr.log` (launchd).
/Users/bytedance/.claude/skills/Users/bytedance/.claude/skills/Users/bytedance/.claude/skills
## Coding Style & Naming Conventions
- Use `gofmt`; keep package names lowercase.
- Tests use `*_test.go`; configs use snake_case YAML (e.g., `config.example.yaml`).
- Keep CLI subcommands in `internal/cli` aligned with routing in `main.go`.

## Testing Guidelines
- Tests live alongside packages (e.g., `internal/scheduler/scheduler_test.go`).
- No explicit coverage threshold; add tests for scheduler logic, CLI commands, and summary generation changes.

## Commit & Pull Request Guidelines
- Recent history uses `feat:`, `fix:`, and `update` prefixes.
- Keep commits scoped; PRs should include a short summary, testing notes (commands run), and screenshots if UX changes.

## Configuration & Security Notes
- Default config path: `~/.config/daily_summary/config.yaml` (YAML or JSON).
- Key settings: `data_dir`, `summary_dir`, `hourly_interval` or `minute_interval`, `summary_time`, `claude_code_path`, `dialog_timeout`.
- `run/` contains local artifacts only; do not commit generated data or logs.
