# crube ドキュメント

crube(しるべの略)は、プロジェクト内のファイルとディレクトリに説明を付けて `.crube/` に保存し、説明付きの `ls` / `tree` / `find` で目的の場所へすぐたどり着けるようにするツールです。説明の作成・更新には AI を使えますが、参照するときに AI は必要ありません。

```
$ cd src/auth
$ crube tree
  src/             アプリケーションのソースコード
>   auth/          認証(ログイン、セッション、トークン更新)
      session.ts   セッションの保存と有効期限の管理
      *_test.go    同名ファイルのテスト
    main.go        エントリポイント
  data/            データ格納場所
    *.csv          CSVデータ
```

## 目次

### design/ — 設計

なぜこの形なのかを説明します。

- [概要と基本方針](design/overview.md) — 目的、範囲、基本方針、用語
- [アーキテクチャ](design/architecture.md) — コアと CLI / MCP の構成、実装言語とライブラリ
- [データモデル](design/data-model.md) — ツリー形式、パターン、多言語、シンボリックリンク、長さの上限
- [鮮度と充足度](design/freshness.md) — status の判定と、充足度を使う側に委ねる考え方
- [決定事項](design/decisions.md) — 決めたことと理由の一覧

### reference/ — リファレンス

正確な仕様です。

- [CLI](reference/cli.md) — コマンド、オプション、環境変数 `CRUBE_LANG`、終了コード
- [MCP](reference/mcp.md) — ツール一覧と引数
- [ファイル形式](reference/file-format.md) — `structure.txt` の文法、言語ごとのファイル、`ignore`

### examples/ — 例

- [はじめての crube](examples/getting-started.md) — 初期化から参照までの流れ
- [AI エージェントによる説明の更新](examples/ai-workflow.md) — status --json を使った更新手順、CI
- [サンプルの .crube/](examples/sample-crube/) — `structure.txt`、`structure_ja.txt`、`ignore`

## 基本情報

| 項目 | 内容 |
| --- | --- |
| 実装言語 | Go(単一バイナリ) |
| データ形式 | 独自のツリー形式テキスト(依存ライブラリなし) |
| 書き込み先 | `.crube/` の中だけ |
| 依存 | `golang.org/x/text`、`golang.org/x/term`、MCP の公式 Go SDK だけ(Git は不要) |
| ライセンス | MIT |
