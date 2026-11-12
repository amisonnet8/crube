# crubeの使い方（このリポジトリの開発で）

このリポジトリは、crube自身で自分のファイルとディレクトリに説明を付けている（`.crube/`。dogfooding）。**Claude Codeは、場所を探すときも、ファイルを足す・動かす・消すときも、crubeを使う。** 使ってみて分かりにくい・おかしいと感じたことは、利用者が同じ目にあうので、`mtqg m add`（不具合なら`b add`）に残す。

## 入口

| 入口 | 使い方 |
|---|---|
| MCP（`.mcp.json`の`crube`） | `find`・`tree`・`ls`・`show`・`status`と、書き込みの`set_description`・`move_description`・`remove_description`。**サーバーは1つ**で、言語は道具の引数`lang`で渡す（日本語は`lang: "ja"`で`.crube/structure_ja.txt`。省くと英語の`.crube/structure.txt`） |
| CLI（`qsoku crube <引数>`） | `qsoku crube find …`。日本語は`qsoku crube find … --lang ja`。毎回、今のソースから作って動かす（`.claude/rules/qsoku-usage.md`） |

- MCPのサーバーも`qsoku crube mcp --allow-write`で起動するので、**起動したときのソース**で動く。crubeのコードを変えたあと、MCPは古いままなので、変えた動きは`qsoku crube`で確かめる（MCPを新しくするには、Claude Codeの`/mcp`でつなぎ直す）
- `postCreate.sh`が入れる`crube`コマンドは、コンテナを作ったときのソースの写し。人間が手で使うためのもので、Claude Codeは使わない
- **言語は、MCP でも CLI でもいつも明示する**（`lang: "en"`・`"ja"`、`--lang en`・`--lang ja`）。省くと、サーバーや手元の環境の`CRUBE_LANG`しだいで、思っていない言語のファイルに書く。`--lang`・`lang`を付ければ`CRUBE_LANG`は使われない
- `lang`の引数は、それを入れたあと（v0.2.0 から）のソースで起動したサーバーにだけある。古いサーバーは`lang`を知らない（`/mcp`でつなぎ直す）
- `.claude/settings.json`の`enabledMcpjsonServers`に`crube`が無いと、MCPは現れない（人間が管理する）。そのときはCLIで進める

## 場所を探すとき

- **まず crube で引く**：`find <キーワード>`（パス・desc・detail）、`tree <ディレクトリのエントリのパス>`、`show <パス>`。目的の場所の見当をつけてから、`Grep`・`Read`で中身を読む
- `Grep`は、コードの中身（関数名・文字列）を探すときに使う。「何がどこにあるか」は crube で引く
- 引けなかった・説明が分かりにくかったときは、その場で説明を直す（下の「説明を保つ」）

## 説明を保つ

- **ファイル・ディレクトリを足す・動かす・消すときは、同じ変更の中で説明を直す**：足したら`set`（`set_description`）、動かしたら`mv`（`move_description`）、消したら`rm`（`remove_description`）
- **英語と日本語の両方を直す**。言語ごとのファイルは独立で、片方だけ直すともう片方が古くなる
- desc は1行で「何がそこにあるか」。詳しいこと（いつ使うか、ほかとの関係）は detail に書く。上限は`docs/reference/file-format_ja.md`「長さの上限」
- `.crube/`の中は、`Edit`で直接直してもよい（タブのインデントを崩さない）
- コミットしないものを`.gitignore`に足したら、`.crube/ignore`にも足す（crube は`.gitignore`を読まない）

## 気づくための仕組み

- **`Write`のあと**：フック`.claude/hooks/crube-untracked.sh`が、作ったファイルと、その上のディレクトリのうち説明の無いものを、言語ごとに知らせる（止めはしない）。知らされたら、続けて両方の言語で説明を付ける
- **コミットの前**：`qsoku crube status --require all --lang en`と`qsoku crube status --require all --lang ja`の両方を通す。Bashで作った・動かした・消したもの（`git rm`、ツールが書いたもの）は、フックが見ないので、ここで見つかる

## 落とし穴

- **`tree`・`set`・`rm`・`mv`に渡すのはエントリのパス**（CLI は現在地から、MCP はルートから読む）。ディレクトリの末尾の`/`は省ける（`tree .claude/hooks`）。ただし、`set`で新しいディレクトリのエントリを作るときは`/`を付ける（付けないと`is a directory (use …/)`で止まる）。`ls`・`show`に渡すのは実在するパス
- 上の2つ（現在地から読む・`/`を省ける）は、MCP では、それを入れたあとのソースで起動したサーバーでだけ効く（上の「入口」。古いサーバーは`/`が無いと`no entry`になる）
- **CLI の`set`も、`-`で始まる desc はオプションとして読まれる**（`unknown option …`）。fnaze と同じく、`qsoku crube set --lang ja [--detail …] -- <path> <desc>`と、いつも`--`を付けて打つ。オプション（`--lang`・`--detail`）は`--`の前に置く（後ろに置くと引数として読まれ、`too many arguments`で止まる）。MCP の`set_description`は引数が分かれているので、この問題は無い
- **`.crube/`は入れ子にできない。** 上のディレクトリに`.crube/`があると、crube はそちらを使う（`init`も「もうある」で止まる）。別のリポジトリの中にこのリポジトリを置いて作業しない
