# 配布方法

<!-- TODO: 配布方法（go install、バイナリ配布、バージョンの取り方、cgoの有無など）を書く -->

## 決まっていること

- Goの**単一バイナリ**で配る（`docs/design/architecture.md`）。ライセンスはMIT
- 依存ライブラリがあれば、そのライセンス文を配布物に同梱する（`docs/design/architecture_ja.md`「実装」）

## リリースの流れ

ビルド済みのバイナリなどを GitHub Releases で配るときは、次の流れにする。

1. **人間が**タグ（`v0.1.0`の形）を打って push する
2. タグで動く workflow（`.github/workflows/release.yml`）が、テストを通してから配布物を作り、**下書きの Release**（`--draft`）に添付する
3. **人間が**下書きの中身を見て、公開する

- **Claude Codeは、タグも Release も作らない。** `gh release create`は`.claude/settings.json`で拒否（deny）されている。Release を作るのは workflow の中だけ
- `release.yml`と、配布物の作り方（goreleaser、または`qsokufile`の自前の項目）は、プロジェクトが選ぶ。このテンプレートには置いていない
- ビルド済みのバイナリは、リポジトリにコミットしない（`.claude/rules/directory-structure.md`「配布物」）
