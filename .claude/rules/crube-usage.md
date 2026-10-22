# crubeの使い方（このリポジトリの開発で）

このリポジトリは、crube自身で自分のファイルとディレクトリに説明を付けている（`.crube/`。dogfooding）。**Claude Codeは、場所を探すときも、ファイルを足す・動かす・消すときも、crubeを使う。** 使ってみて分かりにくい・おかしいと感じたことは、利用者が同じ目にあうので、`mtqg m add`（不具合なら`b add`）に残す。

## 入口

| 入口 | 使い方 |
|---|---|
| MCP（`.mcp.json`の`crube`・`crube-ja`） | `find`・`tree`・`ls`・`show`・`status`と、書き込みの`set_description`・`move_description`・`remove_description`。`crube`は英語（`.crube/structure.txt`）、`crube-ja`は日本語（`.crube/structure_ja.txt`）。言語はサーバーごとに1つ |
| CLI（`qsoku crube <引数>`） | `bcuz "理由" qsoku crube find …`。日本語は`env CRUBE_LANG=ja qsoku crube …`。毎回、今のソースから作って動かす（`.claude/rules/qsoku-usage.md`） |

- MCPのサーバーも`qsoku crube mcp --allow-write`で起動するので、**起動したときのソース**で動く。crubeのコードを変えたあと、MCPは古いままなので、変えた動きは`qsoku crube`で確かめる（MCPを新しくするには、Claude Codeの`/mcp`でつなぎ直す）
- `postCreate.sh`が入れる`crube`コマンドは、コンテナを作ったときのソースの写し。人間が手で使うためのもので、Claude Codeは使わない
- `.claude/settings.json`の`enabledMcpjsonServers`に`crube`・`crube-ja`が無いと、MCPは現れない（人間が管理する）。そのときはCLIで進める

## 場所を探すとき

- **まず crube で引く**：`find <キーワード>`（パス・desc・detail）、`tree <ディレクトリのエントリのパス>`、`show <パス>`。目的の場所の見当をつけてから、`Grep`・`Read`で中身を読む
- `Grep`は、コードの中身（関数名・文字列）を探すときに使う。「何がどこにあるか」は crube で引く
- 引けなかった・説明が分かりにくかったときは、その場で説明を直す（下の「説明を保つ」）

## 説明を保つ

- **ファイル・ディレクトリを足す・動かす・消すときは、同じ変更の中で説明を直す**：足したら`set`（`set_description`）、動かしたら`mv`（`move_description`）、消したら`rm`（`remove_description`）
- **英語と日本語の両方を直す**。言語ごとのファイルは独立で、片方だけ直すともう片方が古くなる
- desc は1行で「何がそこにあるか」。詳しいこと（いつ使うか、ほかとの関係）は detail に書く。上限は`docs/reference/file-format_ja.md`「長さの上限」
- `.crube/`の中は、srwrの`edit`で直接直してもよい（タブのインデントを崩さない）
- コミットしないものを`.gitignore`に足したら、`.crube/ignore`にも足す（crube は`.gitignore`を読まない）

## 気づくための仕組み

- **srwrの`new`のあと**：フック`.claude/hooks/crube-untracked.sh`が、作ったファイルと、その上のディレクトリのうち説明の無いものを、言語ごとに知らせる（止めはしない）。知らされたら、続けて両方の言語で説明を付ける
- **コミットの前**：`qsoku crube status --require all`と`env CRUBE_LANG=ja qsoku crube status --require all`の両方を通す（Skill`commit-with-tape`の手順0）。Bashで作った・動かした・消したもの（`git rm`、ツールが書いたもの）は、フックが見ないので、ここで見つかる

## 落とし穴

- **`tree`・`set`・`rm`・`mv`に渡すのはエントリのパス**で、ディレクトリは末尾に`/`を付ける（`tree .claude/hooks/`）。付けないと`crube: no entry .claude/hooks`になる。`ls`・`show`に渡すのは実在するパス（`/`は要らない）
- crube（MCP・CLI）が書いた`.crube/`の変更は、srwrのテープには理由のない`external`として載る。それでよい（理由は、作業のテープの題とコミットメッセージに書く）
- **`.crube/`は入れ子にできない。** 上のディレクトリに`.crube/`があると、crube はそちらを使う（`init`も「もうある」で止まる）。別のリポジトリの中にこのリポジトリを置いて作業しない
