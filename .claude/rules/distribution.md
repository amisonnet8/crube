# 配布方法

## 決まっていること

- Goの**単一バイナリ**で配る（`docs/design/architecture_ja.md`）。ライセンスはMIT
- 入れ方は2つ：`go install github.com/amisonnet8/crube/cmd/crube@latest`と、GitHub Releases のビルド済みのアーカイブ
- **cgo を使わない**（`CGO_ENABLED=0`）。どの OS でも、ほかのライブラリなしに動く
- 配布物は linux・darwin・windows の amd64・arm64 の6つ。linux・darwin は`.tar.gz`、windows は`.zip`。名前は`crube_<版>_<OS>_<アーキテクチャ>`（版は`v`なし）。`checksums.txt`を添える
- アーカイブに入れるもの：`crube`（`crube.exe`）、`LICENSE`、`README.md`・`README_ja.md`、`licenses/`（依存のライセンス文。下）
- **版**：goreleaser が`-X main.version=<タグ>`（`v0.1.0`の形）を埋める。`go install`で入れたものは、Go がバイナリに記録したモジュールの版を出す（`cli.BuildVersion`。`docs/reference/cli_ja.md`「使い方と版」）。どちらでも`crube v0.1.0`になる
- **依存のライセンス文**を配布物に同梱する（`docs/design/architecture_ja.md`「実装」）。`tools/licenses`が、3つの OS で crube に組み込まれるモジュールを`go list -deps`で求め、そのライセンスのファイルと Go の標準ライブラリの`LICENSE`を`bin/licenses/<モジュールのパス>/`に写す（goreleaser の`before`の手順）。ライセンスのファイルが無いモジュールがあれば失敗し、リリースが止まる
  - 依存を足す・外すと、`tools/licenses`の`TestRun`の一覧が落ちる。新しいモジュールのライセンスを確かめてから一覧を直す（Trivy の`qsoku trivy`も通す）

## リリースの流れ

ビルド済みのバイナリなどを GitHub Releases で配るときは、次の流れにする。

1. **人間が**タグ（`v0.1.0`の形）を打って push する
2. タグで動く workflow（`.github/workflows/release.yml`）が、テストを通してから配布物を作り、**下書きの Release**（`--draft`）に添付する
3. **人間が**下書きの中身を見て、公開する

- **Claude Codeは、タグも Release も作らない。** `gh release create`は`.claude/settings.json`で拒否（deny）されている。Release を作るのは workflow の中だけ
- `release.yml`は、まず`ci.yml`をそのまま呼んで（`workflow_call`）普段の CI の検査をすべて通し、そのあと goreleaser（`.goreleaser.yaml`。`goreleaser-action`の`version: "~> v2"`で、v2 の最新）で配布物を作る。書き込みの権限（`contents: write`）は、Release を作るジョブだけに与える
- 同じタグでもう一度動かすと、前の下書きを置き換える（`replace_existing_draft`）
- **タグを打つ前に、手元で`qsoku snapshot`を動かして中身を見る**。タグなしで`dist/`に同じ配布物を作る（版は`0.0.1-snapshot`、バイナリの版は`v0.0.0`）。goreleaser が要る：`go install github.com/goreleaser/goreleaser/v2@latest`（`proxy.golang.org`と`storage.googleapis.com`への接続が要る）。`.goreleaser.yaml`を変えたら`qsoku snapshot`で確かめる
- ビルド済みのバイナリは、リポジトリにコミットしない（`.claude/rules/directory-structure.md`「配布物」）
