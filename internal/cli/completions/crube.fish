# fish completion for crube.
#
# It has no list of commands. At every TAB it asks `crube __complete` what can
# come next, so it is never out of date with the crube that is installed, and
# it can offer the entry paths of the project the line is in.
#
# Install it with:  crube completion fish > ~/.config/fish/completions/crube.fish
# or try it with:   crube completion fish | source

function __crube_complete
    # The words before the one being typed, without "crube" and their quotes.
    set -l words (commandline -opc)
    set -e words[1]
    # The word being typed, without its quotes and backslashes, as --word: an
    # empty word would be dropped, and "--word=" never is. It is quoted so
    # that the argument is there even where $cur is a list with nothing in it.
    set -l cur (commandline -ct | string replace -ra "[\\\\'\"]" '')
    crube __complete "--word=$cur" -- $words 2>/dev/null
end

# Where a file is wanted: after -C, and for the paths of ls and show.
function __crube_wants_file
    set -l words (commandline -opc)
    set -e words[1]
    test (count $words) -ge 1; and test "$words[-1]" = -C; and return 0
    contains -- "$words[1]" ls show
end

# fish puts in no space after a candidate that ends in "/", and escapes what
# it puts in, so a pattern (data/*.csv) reaches crube as it is.
complete -c crube -f -a '(__crube_complete)'
complete -c crube -F -n __crube_wants_file
