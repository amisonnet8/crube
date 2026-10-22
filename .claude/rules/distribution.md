# 配布方法

## 決まっていること

- Goの**単一バイナリ**で配る（`docs/design/architecture_ja.md`）。ライセンスはMIT
- 入れ方は2つ：`go install github.com/amisonnet8/crube/cmd/crube@latest`と、GitHub Releases のビルド済みのアーカイブ
- **cgo を使わない**（`CGO_ENABLED=0`）。どの OS でも、ほかのライブラリなしに動く
- 配布物は linux・darwin・windows の amd64・arm64 の6つ。linux・darwin は`.tar.gz`、windows は`.zip`。名前は`crube_<版>_<OS>_<アーキテクチャ>`（版は`v`なし）。`checksums.txt`を添える
- アーカイブに入れるもの：`crube`（`crube.exe`）、`LICENSE`、`README.md`・`README_ja.md`、`licenses/`（依存のライセンス文。下）
- **版**：goreleaser が`-X main.version=<タグ>`（`v0.1.0`の形）を埋める。`go install`で入れたものは、Go がバイナリに記録したモジュールの版を出す（`cli.BuildVersion`。`docs/reference/cli_ja.md`「使い方と版」）。どちらでも`crube v0.1.0`になる
- **依存のライセンス文**を配布物に同梱する（`docs/design/architecture_ja.md`「実装」）。`tools/licenses`が、3つの OS で crube に組み込まれるモジュールを`go list -deps`で求め、そのライセンスのファイルと Go の標準ライブラリの`LICENSE`を`bin/licenses/<モジュールのパス>/`に写す（goreleaser の`before`の手順で`go run ./tools/licenses bin/licenses ./cmd/crube`。配るコマンドの main パッケージを引数で渡す）。ライセンスのファイルが無いモジュールがあれば失敗し、リリースが止まる
  - `.goreleaser.yaml`の`archives.files`には、`bin/licenses`を**ディレクトリのまま**`src`に書く（`dst: licenses`）。グロブ（`bin/licenses/**/*`）にすると、当たるファイルが1つだけのとき（依存の無いプロジェクトは`go/LICENSE`だけ）goreleaser が親のディレクトリを落とし、`licenses/LICENSE`になる
  - 依存を足す・外すと、`tools/licenses`の`TestRun`の一覧が落ちる。新しいモジュールのライセンスを確かめてから一覧を直す（Trivy の`qsoku trivy`も通す）

## リリースの流れ

ビルド済みのバイナリなどを GitHub Releases で配るときは、次の流れにする。どのプロジェクトも同じ。

1. **人間が**タグ（`v0.1.0`の形）を打って push する。人間がするのはこれだけ
2. タグで動く workflow（`.github/workflows/release.yml`）が、検査を通してから配布物を作り、Release に添付して**公開する**（本文は goreleaser が前のタグからのコミットで作る）

- **Claude Codeは、タグも Release も作らない。** `gh release create`は`.claude/settings.json`で拒否（deny）されている。Release を作るのは workflow の中だけ
- `release.yml`は、まず`ci.yml`をそのまま呼んで（`workflow_call`）普段の CI の検査をすべて通し、そのあと goreleaser（`.goreleaser.yaml`。`goreleaser-action`の`version: "~> v2"`で、v2 の最新）で配布物を作る。書き込みの権限（`contents: write`）は、Release を作るジョブだけに与える
- goreleaser はいったん**下書き**に添付し、すべて添付し終えてから、最後の手順（`gh release edit --draft=false`）で公開する。途中で落ちると下書きだけが残り、公開された Release に配布物が欠けることはない。同じタグでもう一度動かすと、その下書きを置き換える（`replace_existing_draft`）。落ちたときは、原因を直したコミットに新しいタグを打つか、同じコミットのままなら workflow を再実行する
- **公開済みの Release があるタグでは、何も変えずに失敗する**（最初の手順）。出し直すときは、新しいタグを打つ
- **タグを打つ前に、手元で`qsoku snapshot`を動かして中身を見る**。タグなしで`dist/`に同じ配布物を作る（版は前のタグのパッチを1つ上げた`<版>-snapshot`、バイナリの版は前のタグ。v0.1.0 のあとは`0.1.1-snapshot`と`v0.1.0`。タグが1つも無ければ`0.0.1-snapshot`と`v0.0.0`）。goreleaser は`.devcontainer/postCreate.sh`が入れる。コンテナを作り直す前に要るときは、`GOTOOLCHAIN=auto go install github.com/goreleaser/goreleaser/v2@latest`（goreleaser がコンテナより新しい Go を求めるとき、その Go を取ってきてビルドする。Bashサンドボックスの中では、`proxy.golang.org`と`storage.googleapis.com`への接続が要る）。`.goreleaser.yaml`を変えたら`qsoku snapshot`で確かめる
- ビルド済みのバイナリは、リポジトリにコミットしない（`.claude/rules/directory-structure.md`「配布物」）
