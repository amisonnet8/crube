#!/usr/bin/env bash
set -euo pipefail

# wget, gnupg,
# lsb-release: Adding the Trivy apt repository below.
# gcc:         qsoku race (CGO_ENABLED=1 go test -race) needs a C compiler.
#              The container itself runs with CGO_ENABLED=0 (devcontainer.json).
# jq:          Used by .claude/hooks/build.sh, .claude/hooks/require-bcuz.sh (and
#              its test, qsoku hook) and the allowWrite block below; also for
#              inspecting --json output.
# ShellCheck:  Static analysis of tracked *.sh and *.bash files (qsoku shellcheck).
#              Comment lines must not start with the lowercase directive word,
#              or ShellCheck parses them as directives (SC1072/SC1073).
# fish:        One of the four shells that crube completion has a script for (the
#              scripts are run by the real shells in e2e/completion_test.go), and
#              one of the four that qsoku's shell integration supports. The other
#              three come from elsewhere: bash and zsh are in the base image, and
#              PowerShell is the "powershell" feature in devcontainer.json.
# gh:          GitHub CLI (issues, pull requests, Actions runs) is the "github-cli"
#              feature in devcontainer.json, not installed here. The features are left
#              unpinned (no "version" option): each rebuild installs the latest release.
sudo apt-get update
sudo apt-get install -y wget gnupg lsb-release gcc jq shellcheck fish

# The Bash sandbox (.claude/settings.json "sandbox") needs bubblewrap (bwrap)
# and socat on Linux; without them it silently stays off even with
# "enabled": true. Install only what is missing, since some base images
# already have them. Project-independent (.claude/rules/testing.md).
missing_sandbox_deps=()
command -v bwrap >/dev/null 2>&1 || missing_sandbox_deps+=(bubblewrap)
command -v socat >/dev/null 2>&1 || missing_sandbox_deps+=(socat)
if [ "${#missing_sandbox_deps[@]}" -gt 0 ]; then
  sudo apt-get update
  sudo apt-get install -y "${missing_sandbox_deps[@]}"
fi

# Trivy: known vulnerabilities (CVE) and license compatibility of dependencies
# (qsoku trivy, .claude/rules/testing.md). Installed from the official apt repository.
wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | gpg --dearmor | sudo tee /usr/share/keyrings/trivy.gpg >/dev/null
echo "deb [signed-by=/usr/share/keyrings/trivy.gpg] https://aquasecurity.github.io/trivy-repo/deb $(lsb_release -sc) main" | sudo tee /etc/apt/sources.list.d/trivy.list >/dev/null
sudo apt-get update
sudo apt-get install -y trivy

# The Bash sandbox (.claude/settings.json) only honours an allowWrite path that
# already exists, and ~/.cache itself is read-only there. Create every
# filesystem.allowWrite path up front, or the first qsoku lint / qsoku trivy in
# a fresh container fails with "read-only file system" (.claude/rules/testing.md).
# The list is read from settings.json, so this block needs no change when
# allowWrite does, and it is project-independent. "~/x" and "$HOME/x" are created
# under the home directory; other absolute paths (e.g. /go) are created if they do
# not exist yet, and a failure only warns (no sudo is used). Relative paths, "."
# and $TMPDIR, and the special paths /dev, /proc and /sys are left alone.
sandbox_settings="$(dirname "$0")/../.claude/settings.json"
tilde='~'
if [ -f "$sandbox_settings" ]; then
  while IFS= read -r allow_path; do
    case "$allow_path" in
      "$tilde/"*) allow_path="$HOME/${allow_path#"$tilde/"}" ;;
      /dev/* | /proc/* | /sys/*) continue ;;
      /*) ;;
      *) continue ;;
    esac
    mkdir -p "$allow_path" || echo "warning: cannot create allowWrite path $allow_path" >&2
  done < <(jq -r '.sandbox.filesystem.allowWrite[]?' "$sandbox_settings")
fi

# golangci-lint: lint (qsoku check, .golangci.yaml). Built from source with the
# container's Go (go install), not the prebuilt binary of install.sh. @latest,
# not pinned: a golangci-lint reads the export data of the standard library only
# up to the version of its own x/tools, so a pinned one stops working when the
# container's Go moves on ("export data version 5 is greater than maximum
# supported version 4"; building the same version again does not help). A newer
# lint reads it. If lint still fails that way right after a Go update, the lint
# is not out yet: `qsoku lint` says so (qsokufile, .claude/rules/testing.md).
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

go install golang.org/x/tools/gopls@latest
go install golang.org/x/tools/cmd/goimports@latest

# qsoku: build/check/test entry points (qsokufile).
go install github.com/amisonnet8/qsoku/cmd/qsoku@latest

# mtqg: memo/todo/qa/bug/glossary/rule journal (.mtqg/, CLAUDE.md, .mcp.json).
go install github.com/amisonnet8/mtqg/cmd/mtqg@latest

# srwr: the AI's edits go through its look / edit / new tools, each with a
# reason (why), and are recorded on a tape (.claude/rules/srwr-usage.md).
# @latest, like qsoku and mtqg above. srwr is 0.x, so after a rebuild that
# picked up a new version, run `srwr init` once: it makes .claude/settings.json
# follow any change in the tools' names.
go install github.com/amisonnet8/srwr/cmd/srwr@latest

# bcuz: every command Claude Code runs through the Bash tool goes through
# `bcuz "<reason>" <command>`, which records when / why / what / result in
# .bcuz/history.jsonl (.claude/rules/bcuz-usage.md; enforced by the PreToolUse
# hook .claude/hooks/require-bcuz.sh). @latest, like the tools above.
go install github.com/amisonnet8/bcuz/cmd/bcuz@latest

# Create .bcuz/ at the repository root (idempotent: prints "already exists"
# and succeeds if it is there). Without it, commands still run but are not recorded.
(cd "$(dirname "$0")/.." && bcuz --init)

# goreleaser: builds the release archives (.goreleaser.yaml). Needed here only
# for `qsoku snapshot`, which makes them without a tag to look at before tagging
# (.claude/rules/distribution.md); the release workflow installs its own.
# @latest, like the tools above.
go install github.com/goreleaser/goreleaser/v2@latest

# crube: built from this repository's own source, so a fresh devcontainer has
# a `crube` command and its completion (below) to use interactively. It is a
# snapshot of the source at creation, for people. Claude Code does not use it:
# it uses `qsoku crube` and the crube / crube-ja MCP servers of .mcp.json, which
# build the current source each time (.claude/rules/crube-usage.md).
(cd "$(dirname "$0")/.." && go install ./cmd/crube)

# Wire up qsoku's shell integration (working-directory carry-back and
# completion) for bash, zsh, fish and pwsh. Idempotent: skipped if already
# present, so re-running postCreate.sh does not duplicate the line. The single
# quotes are intentional -- the line is meant to land in the rc file unexpanded.
# shellcheck disable=SC2016
grep -qF 'qsoku .shell bash' ~/.bashrc 2>/dev/null || echo 'eval "$(qsoku .shell bash)"' >>~/.bashrc
# shellcheck disable=SC2016
grep -qF 'qsoku .shell zsh' ~/.zshrc 2>/dev/null || echo 'eval "$(qsoku .shell zsh)"' >>~/.zshrc
mkdir -p ~/.config/fish
grep -qF 'qsoku .shell fish' ~/.config/fish/config.fish 2>/dev/null || echo 'qsoku .shell fish | source' >>~/.config/fish/config.fish
# pwsh has no $PROFILE of its own until asked; query it rather than hardcode
# the path. The single quotes below are intentional too -- $PROFILE is pwsh's
# variable, not bash's.
# shellcheck disable=SC2016
pwsh_profile=$(pwsh -NoLogo -NoProfile -Command '$PROFILE')
mkdir -p "$(dirname "$pwsh_profile")"
grep -qF 'qsoku .shell pwsh' "$pwsh_profile" 2>/dev/null || echo 'Invoke-Expression (& qsoku .shell pwsh | Out-String)' >>"$pwsh_profile"

# The completion of crube and mtqg, for interactive use in this container, in
# all four shells. Each install path is the one that their docs name (for
# crube: docs/reference/cli.md "crube completion").
# shellcheck disable=SC2016
zsh_fpath=$(zsh -c 'echo $fpath[1]')
# zsh's fpath[1] in this image is a root-owned system directory, unlike the
# user-local completion directories of bash and fish.
sudo mkdir -p "$zsh_fpath"
mkdir -p ~/.local/share/bash-completion/completions ~/.config/fish/completions
for tool in crube mtqg; do
  "$tool" completion bash >~/.local/share/bash-completion/completions/"$tool"
  "$tool" completion zsh | sudo tee "$zsh_fpath/_$tool" >/dev/null
  "$tool" completion fish >~/.config/fish/completions/"$tool".fish
  # pwsh has no completions directory: the script is appended to $PROFILE,
  # so re-running this needs the same marker-line check as qsoku above.
  grep -qF "Register-ArgumentCompleter -Native -CommandName $tool" "$pwsh_profile" 2>/dev/null ||
    "$tool" completion powershell >>"$pwsh_profile"
done
