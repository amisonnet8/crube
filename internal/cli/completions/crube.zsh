#compdef crube
# zsh completion for crube.
#
# It has no list of commands. At every TAB it asks `crube __complete` what can
# come next, so it is never out of date with the crube that is installed, and
# it can offer the entry paths of the project the line is in.
#
# Install it with:  crube completion zsh > "${fpath[1]}/_crube"
# or try it with:   source <(crube completion zsh)
_crube() {
    local line value item
    local -a dirs others
    # The words before the one being typed, and that one as --word, without
    # their quotes: an empty word would be dropped by some shells, and
    # "--word=" never is.
    for line in "${(@f)$(command crube __complete "--word=${(Q)words[CURRENT]}" -- "${(@Q)words[2,CURRENT-1]}" 2>/dev/null)}"; do
        [[ -n $line ]] || continue
        # "value<TAB>desc" is "value:desc" for _describe, which wants a colon
        # in the value written as \:.
        value=${line%%$'\t'*}
        item=${value//:/\\:}
        [[ $line == *$'\t'* ]] && item+=":${line#*$'\t'}"
        # A directory entry goes on: no space after it.
        if [[ $value == */ ]]; then
            dirs+=("$item")
        else
            others+=("$item")
        fi
    done
    if (( ${#dirs} + ${#others} == 0 )); then
        # Where nothing is offered (ls, show, -C), files are.
        _files
        return
    fi
    # zsh quotes what it puts in, so a pattern (data/*.csv) reaches crube as
    # it is.
    (( ${#dirs} )) && _describe -t candidates 'crube' dirs -S ''
    (( ${#others} )) && _describe -t candidates 'crube' others
    return 0
}

if [ "$funcstack[1]" = "_crube" ]; then
    _crube "$@"
else
    compdef _crube crube
fi
