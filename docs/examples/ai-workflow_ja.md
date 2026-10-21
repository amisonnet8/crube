*[English](ai-workflow.md) | **日本語***

# AI エージェントによる説明の更新

crube 本体は AI を呼び出しません。AI エージェントが CLI または MCP を使って説明を書き込みます。AI の種類を問わず、同じ手順で使えます。

## 手順

1. `crube status --json` で、説明が必要な箇所を取得する。
2. AI が該当ファイルの中身や、利用可能な情報源(設計資料、作業記録など)を読む。
3. AI が `crube set` / `mv` / `rm`(または MCP の書き込みツール)で structure を更新する。
4. もう一度 `crube status` を実行し、問題がなくなったことを確認する。
5. 人間が `.crube/` の差分をレビューする。

## status --json の出力例

```json
{
  "structure": ".crube/structure.txt",
  "items": [
    { "kind": "untracked", "type": "dir",  "path": "src/db/" },
    { "kind": "untracked", "type": "file", "path": "src/db/conn.go" },
    { "kind": "orphan",    "path": "src/legacy/" }
  ],
  "coverage": {
    "all":  { "ok": false, "missing": 2 },
    "dirs": { "ok": false, "missing": 1 }
  }
}
```

各フィールドと並び順は [CLI リファレンス](../reference/cli_ja.md#crube-status) を参照してください。

## エージェント向けの指示の例

`.claude/` などのエージェント側ルールに、次のような指示を書いておくと運用しやすくなります。

```markdown
## ファイルを探すとき
- まず `crube find <キーワード>` か `crube tree` で場所を確認する。

## 作業を終えるとき
- `crube status --json` を実行する。
- untracked / orphan があれば、ファイルを読んで
  `crube set` / `crube mv` / `crube rm` で説明を更新する。
- 自分がファイルを移動・改名した場合は、`crube mv <旧パス> <新パス>` で説明を付け替える。
- desc は 1 行で、何がそこにあるかを書く。詳しい内容は --detail に書く。
- 複数の言語のファイルがある場合は、`CRUBE_LANG` を切り替えて、言語ごとに同じ手順を行う。
```

## 運用のタイミング

- AI エージェントの作業終了時、またはコミット前のフックで手順 1〜4 を実行する。
- CI では `crube status --require dirs` などの終了コードで説明漏れを検出する。どの段階を求めるかはプロジェクトで決める。

```yaml
# 例: CI のステップ
- name: Check crube descriptions
  run: crube status --require dirs
```
