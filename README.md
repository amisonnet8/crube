# crube

crube（しるべの略）は、プロジェクト内のファイルとディレクトリに説明を付けて`.crube/`に保存し、説明付きの`ls`／`tree`／`find`で目的の場所へすぐたどり着けるようにするツール。説明の作成・更新にはAIを使えるが、参照するときにAIは要らない。書き込むのは`.crube/`の中だけで、Git にも依存しない。

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

**開発中**（CLI と MCP サーバーは動く。まだリリースしておらず、配布物は無い）。

## ドキュメント

[docs/README.md](docs/README.md)が目次。

- [docs/design/](docs/design/)：設計（概要と基本方針、アーキテクチャ、データモデル、鮮度と充足度、決定事項）
- [docs/reference/](docs/reference/)：仕様（CLI、MCP、ファイル形式）
- [docs/examples/](docs/examples/)：使い方の例

## 開発

devcontainer で開く。`.devcontainer/postCreate.sh`が、Go・[mtqg](https://github.com/amisonnet8/mtqg)・[qsoku](https://github.com/amisonnet8/qsoku)・[srwr](https://github.com/amisonnet8/srwr)・[bcuz](https://github.com/amisonnet8/bcuz)・golangci-lint・ShellCheck・Trivy を入れる。ビルドと検査は`qsoku check`（名前の一覧は`qsokufile`）。AIエージェント向けの決まりは[CLAUDE.md](CLAUDE.md)と`.claude/rules/`。

## ライセンス

MIT（[LICENSE](LICENSE)）
