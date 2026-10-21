*[日本語](freshness_ja.md) | **English***

# Freshness and coverage

## The idea

Since reading uses no AI, the biggest risk is trusting a description that has gone stale. crube ties descriptions to paths, so comparing the real directories with the structure is enough to check freshness, without AI.

`crube status` does this check. It only walks the files, and uses no outside information such as Git. status only reports: it changes neither the structure nor the files.

It checks only the one structure file used for the target language (for how the language is chosen, see the [data model](data-model.md#languages-one-independent-file-per-language)).

## What status reports

| Kind | How it is found | What to do |
| --- | --- | --- |
| untracked | It exists, but no entry with a desc (patterns included) matches it. Reported with its type (dir / file) | Add a description with `set` |
| orphan | No real path matches the entry (for a pattern: it matches no name directly in its directory) | `mv` or `rm` |
| invalid | A file edited by hand has a desc / detail over the limit | Fix the value |

Moves are not guessed. After a move, the old path is reported as an orphan and the new path as untracked. Which is the same file as which is for the user (a person or AI) to decide, and `crube mv` moves the description.

## Coverage: no default

"How many descriptions are enough" differs by project and by situation. So crube has no policy, and only reports whether each of these two levels is met (with the number of untracked paths).

| Level | Met when |
| --- | --- |
| all match | Every file and directory in scope has a description |
| dirs match | Every directory in scope has a description (files do not count) |

When a pass or fail is needed, the user picks the level with `crube status --require all|dirs`. Without it, untracked paths do not change the exit code.

For example, CI can stop only missing descriptions of the structure with `--require dirs`, while a person aims at all match.

## Rules

- A path matched by a pattern entry counts as described.
- A path matched by a line without a desc (a line for structure only) counts as not described.
- Empty directories and ignored paths are out of scope. An empty directory is one with no file in scope anywhere below it (no file or symbolic link that is not ignored). A directory with only ignored or empty things in it is empty too.
- A description of the parent directory does not count for its children (otherwise one line at the root would pass everything).
- A symbolic link counts as a file (not in scope of dirs match; in scope of all match). A directory entry (ending in `/`) does not match a link.
- Ignored paths and empty directories do exist, so an entry that matches one is not an orphan.
- Entries below an orphan, and below an ignored directory whose inside was not looked at, are not reported as orphans. Moving one directory reports only the line on top as an orphan.
- A path matched by an invalid entry still counts as described (invalid does not change coverage; the exit code is 1).
- A directory that cannot be read stops status with an error naming its path (exit code 2).

## Updating with AI

`crube status --json` outputs the places that need a description in a form that can be fed to AI as it is. The AI writes descriptions from that list, and runs status again to make sure no problem is left. The steps are in [examples/ai-workflow.md](../examples/ai-workflow.md).
