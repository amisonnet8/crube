# bash completion for crube.
#
# It has no list of commands. At every TAB it asks `crube __complete` what can
# come next, so it is never out of date with the crube that is installed, and
# it can offer the entry paths of the project the line is in.
#
# Install it with:  crube completion bash > ~/.local/share/bash-completion/completions/crube
# or try it with:   source <(crube completion bash)

# _crube_unquote sets REPLY to a word as bash keeps it, without its quotes and
# backslashes: crube wants the word, not how it was typed.
_crube_unquote() {
    REPLY=${1//\\/}
    REPLY=${REPLY//\'/}
    REPLY=${REPLY//\"/}
}

_crube() {
    local value i
    local -a words
    COMPREPLY=()
    for ((i = 1; i < COMP_CWORD; i++)); do
        _crube_unquote "${COMP_WORDS[i]}"
        words+=("$REPLY")
    done
    # The word being typed goes as --word: an empty word would be dropped by
    # some shells, and "--word=" never is.
    _crube_unquote "${COMP_WORDS[COMP_CWORD]}"
    while IFS=$'\t' read -r value _; do
        # A directory entry goes on: no space after it.
        [[ $value == */ ]] && compopt -o nospace 2>/dev/null
        # A pattern (data/*.csv), or a name with a space, is quoted so that it
        # reaches crube as it is.
        if [[ $value == *[][\ \*\?\'\"\\\$\&\;\|\<\>\(\)\!\{\}\#\`]* ]]; then
            value="'${value//\'/\'\\\'\'}'"
        fi
        COMPREPLY+=("$value")
    done < <(command crube __complete "--word=$REPLY" -- "${words[@]}" 2>/dev/null)
}

# -o default: where nothing is offered (ls, show, -C), files are.
complete -o default -F _crube crube
