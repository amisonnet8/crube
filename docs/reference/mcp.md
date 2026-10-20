# MCP リファレンス

`crube mcp` で stdio の MCP サーバーとして起動します。CLI と同じコアを呼び、結果は JSON で返します。

## クライアントへの登録例

```json
{
  "mcpServers": {
    "crube": {
      "command": "crube",
      "args": ["mcp", "--allow-write"],
      "env": { "CRUBE_LANG": "ja" }
    }
  }
}
```

サーバーは起動したディレクトリから上位へ `.crube/` を探します。言語は、サーバー起動時の環境変数 `CRUBE_LANG` で決まります(クライアント設定の `env` で指定できます)。書き込み系ツールも、書き込むのは `.crube/` の中だけです。`remove_description` / `move_description` は説明を消す・付け替えるだけで、ファイルには触れません。

## ツール一覧

書き込み系ツールは、`crube mcp --allow-write` で起動したときだけ公開します。参照だけにしたい場合は、`--allow-write` を付けずに起動します。

| 種別 | ツール | 引数 | 返す内容 |
| --- | --- | --- | --- |
| 参照 | `tree` | path, depth | CLI の `crube tree --json` と同じ入れ子の構造。現在地の印は付けない(`here` は最上位が `""`、各エントリが `false`) |
| 参照 | `ls` | path | 直下の一覧と説明 |
| 参照 | `find` | query, limit | CLI の `crube find --json` と同じ(一致したパスと説明、省いた件数) |
| 参照 | `show` | path | CLI の `crube show` と同じ(desc、detail、当たったエントリのパス、シンボリックリンクならリンク先) |
| 保守 | `status` | kinds(絞り込み) | `crube status --json` と同じ内容 |
| 書き込み | `set_description` | path, desc, detail | 更新後のエントリ |
| 書き込み | `remove_description` | path | 削除結果 |
| 書き込み | `move_description` | from, to | 移動結果 |

## ツール名

参照系と保守系のツールは、CLI のコマンドと同じ名前・同じ内容にしています。

書き込み系だけは、CLI の `set` / `rm` / `mv` とは別の名前(`set_description` / `remove_description` / `move_description`)にしています。エージェントはツール一覧の名前から用途を判断することがあり、`rm` や `mv` だとファイル自体を削除・移動するツールだと誤解される恐れがあるためです。名前から「説明だけを操作する」ことが伝わるようにしています。

## 設計上の注意

- ツール説明に「ファイルを探す前に、まず tree / find を使う」旨を書き、エージェントが先に参照するよう促す。
- tree の出力は structure ファイルとほぼ同じ大きさ。大きいプロジェクトでは depth か path で絞る。
- 書き込み系ツールの検証(長さの上限、改行の禁止)は CLI の set と同じ。
- 書き込みは CLI と同じく、ファイル全体を正規形で出力し直す。
- エラーはツールのエラーとして返す。内容は CLI のエラーメッセージ(`crube: ` の後ろ)と同じ。
- MCP の実装には、公式の Go SDK(`github.com/modelcontextprotocol/go-sdk`)を使う。
