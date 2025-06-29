# bash completion for cu                                   -*- shell-script -*-

__cu_debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE:-} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    fi
}

# Homebrew on Macs have version 1.3 of bash-completion which doesn't include
# _init_completion. This is a very minimal version of that function.
__cu_init_completion()
{
    COMPREPLY=()
    _get_comp_words_by_ref "$@" cur prev words cword
}

__cu_index_of_word()
{
    local w word=$1
    shift
    index=0
    for w in "$@"; do
        [[ $w = "$word" ]] && return
        index=$((index+1))
    done
    index=-1
}

__cu_contains_word()
{
    local w word=$1; shift
    for w in "$@"; do
        [[ $w = "$word" ]] && return
    done
    return 1
}

__cu_handle_go_custom_completion()
{
    __cu_debug "${FUNCNAME[0]}: cur is ${cur}, words[*] is ${words[*]}, #words[@] is ${#words[@]}"

    local shellCompDirectiveError=1
    local shellCompDirectiveNoSpace=2
    local shellCompDirectiveNoFileComp=4
    local shellCompDirectiveFilterFileExt=8
    local shellCompDirectiveFilterDirs=16

    local out requestComp lastParam lastChar comp directive args

    # Prepare the command to request completions for the program.
    # Calling ${words[0]} instead of directly cu allows handling aliases
    args=("${words[@]:1}")
    # Disable ActiveHelp which is not supported for bash completion v1
    requestComp="CU_ACTIVE_HELP=0 ${words[0]} __completeNoDesc ${args[*]}"

    lastParam=${words[$((${#words[@]}-1))]}
    lastChar=${lastParam:$((${#lastParam}-1)):1}
    __cu_debug "${FUNCNAME[0]}: lastParam ${lastParam}, lastChar ${lastChar}"

    if [ -z "${cur}" ] && [ "${lastChar}" != "=" ]; then
        # If the last parameter is complete (there is a space following it)
        # We add an extra empty parameter so we can indicate this to the go method.
        __cu_debug "${FUNCNAME[0]}: Adding extra empty parameter"
        requestComp="${requestComp} \"\""
    fi

    __cu_debug "${FUNCNAME[0]}: calling ${requestComp}"
    # Use eval to handle any environment variables and such
    out=$(eval "${requestComp}" 2>/dev/null)

    # Extract the directive integer at the very end of the output following a colon (:)
    directive=${out##*:}
    # Remove the directive
    out=${out%:*}
    if [ "${directive}" = "${out}" ]; then
        # There is not directive specified
        directive=0
    fi
    __cu_debug "${FUNCNAME[0]}: the completion directive is: ${directive}"
    __cu_debug "${FUNCNAME[0]}: the completions are: ${out}"

    if [ $((directive & shellCompDirectiveError)) -ne 0 ]; then
        # Error code.  No completion.
        __cu_debug "${FUNCNAME[0]}: received error from custom completion go code"
        return
    else
        if [ $((directive & shellCompDirectiveNoSpace)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __cu_debug "${FUNCNAME[0]}: activating no space"
                compopt -o nospace
            fi
        fi
        if [ $((directive & shellCompDirectiveNoFileComp)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __cu_debug "${FUNCNAME[0]}: activating no file completion"
                compopt +o default
            fi
        fi
    fi

    if [ $((directive & shellCompDirectiveFilterFileExt)) -ne 0 ]; then
        # File extension filtering
        local fullFilter filter filteringCmd
        # Do not use quotes around the $out variable or else newline
        # characters will be kept.
        for filter in ${out}; do
            fullFilter+="$filter|"
        done

        filteringCmd="_filedir $fullFilter"
        __cu_debug "File filtering command: $filteringCmd"
        $filteringCmd
    elif [ $((directive & shellCompDirectiveFilterDirs)) -ne 0 ]; then
        # File completion for directories only
        local subdir
        # Use printf to strip any trailing newline
        subdir=$(printf "%s" "${out}")
        if [ -n "$subdir" ]; then
            __cu_debug "Listing directories in $subdir"
            __cu_handle_subdirs_in_dir_flag "$subdir"
        else
            __cu_debug "Listing directories in ."
            _filedir -d
        fi
    else
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${out}" -- "$cur")
    fi
}

__cu_handle_reply()
{
    __cu_debug "${FUNCNAME[0]}"
    local comp
    case $cur in
        -*)
            if [[ $(type -t compopt) = "builtin" ]]; then
                compopt -o nospace
            fi
            local allflags
            if [ ${#must_have_one_flag[@]} -ne 0 ]; then
                allflags=("${must_have_one_flag[@]}")
            else
                allflags=("${flags[*]} ${two_word_flags[*]}")
            fi
            while IFS='' read -r comp; do
                COMPREPLY+=("$comp")
            done < <(compgen -W "${allflags[*]}" -- "$cur")
            if [[ $(type -t compopt) = "builtin" ]]; then
                [[ "${COMPREPLY[0]}" == *= ]] || compopt +o nospace
            fi

            # complete after --flag=abc
            if [[ $cur == *=* ]]; then
                if [[ $(type -t compopt) = "builtin" ]]; then
                    compopt +o nospace
                fi

                local index flag
                flag="${cur%=*}"
                __cu_index_of_word "${flag}" "${flags_with_completion[@]}"
                COMPREPLY=()
                if [[ ${index} -ge 0 ]]; then
                    PREFIX=""
                    cur="${cur#*=}"
                    ${flags_completion[${index}]}
                    if [ -n "${ZSH_VERSION:-}" ]; then
                        # zsh completion needs --flag= prefix
                        eval "COMPREPLY=( \"\${COMPREPLY[@]/#/${flag}=}\" )"
                    fi
                fi
            fi

            if [[ -z "${flag_parsing_disabled}" ]]; then
                # If flag parsing is enabled, we have completed the flags and can return.
                # If flag parsing is disabled, we may not know all (or any) of the flags, so we fallthrough
                # to possibly call handle_go_custom_completion.
                return 0;
            fi
            ;;
    esac

    # check if we are handling a flag with special work handling
    local index
    __cu_index_of_word "${prev}" "${flags_with_completion[@]}"
    if [[ ${index} -ge 0 ]]; then
        ${flags_completion[${index}]}
        return
    fi

    # we are parsing a flag and don't have a special handler, no completion
    if [[ ${cur} != "${words[cword]}" ]]; then
        return
    fi

    local completions
    completions=("${commands[@]}")
    if [[ ${#must_have_one_noun[@]} -ne 0 ]]; then
        completions+=("${must_have_one_noun[@]}")
    elif [[ -n "${has_completion_function}" ]]; then
        # if a go completion function is provided, defer to that function
        __cu_handle_go_custom_completion
    fi
    if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
        completions+=("${must_have_one_flag[@]}")
    fi
    while IFS='' read -r comp; do
        COMPREPLY+=("$comp")
    done < <(compgen -W "${completions[*]}" -- "$cur")

    if [[ ${#COMPREPLY[@]} -eq 0 && ${#noun_aliases[@]} -gt 0 && ${#must_have_one_noun[@]} -ne 0 ]]; then
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${noun_aliases[*]}" -- "$cur")
    fi

    if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
        if declare -F __cu_custom_func >/dev/null; then
            # try command name qualified custom func
            __cu_custom_func
        else
            # otherwise fall back to unqualified for compatibility
            declare -F __custom_func >/dev/null && __custom_func
        fi
    fi

    # available in bash-completion >= 2, not always present on macOS
    if declare -F __ltrim_colon_completions >/dev/null; then
        __ltrim_colon_completions "$cur"
    fi

    # If there is only 1 completion and it is a flag with an = it will be completed
    # but we don't want a space after the =
    if [[ "${#COMPREPLY[@]}" -eq "1" ]] && [[ $(type -t compopt) = "builtin" ]] && [[ "${COMPREPLY[0]}" == --*= ]]; then
       compopt -o nospace
    fi
}

# The arguments should be in the form "ext1|ext2|extn"
__cu_handle_filename_extension_flag()
{
    local ext="$1"
    _filedir "@(${ext})"
}

__cu_handle_subdirs_in_dir_flag()
{
    local dir="$1"
    pushd "${dir}" >/dev/null 2>&1 && _filedir -d && popd >/dev/null 2>&1 || return
}

__cu_handle_flag()
{
    __cu_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    # if a command required a flag, and we found it, unset must_have_one_flag()
    local flagname=${words[c]}
    local flagvalue=""
    # if the word contained an =
    if [[ ${words[c]} == *"="* ]]; then
        flagvalue=${flagname#*=} # take in as flagvalue after the =
        flagname=${flagname%=*} # strip everything after the =
        flagname="${flagname}=" # but put the = back
    fi
    __cu_debug "${FUNCNAME[0]}: looking for ${flagname}"
    if __cu_contains_word "${flagname}" "${must_have_one_flag[@]}"; then
        must_have_one_flag=()
    fi

    # if you set a flag which only applies to this command, don't show subcommands
    if __cu_contains_word "${flagname}" "${local_nonpersistent_flags[@]}"; then
      commands=()
    fi

    # keep flag value with flagname as flaghash
    # flaghash variable is an associative array which is only supported in bash > 3.
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        if [ -n "${flagvalue}" ] ; then
            flaghash[${flagname}]=${flagvalue}
        elif [ -n "${words[ $((c+1)) ]}" ] ; then
            flaghash[${flagname}]=${words[ $((c+1)) ]}
        else
            flaghash[${flagname}]="true" # pad "true" for bool flag
        fi
    fi

    # skip the argument to a two word flag
    if [[ ${words[c]} != *"="* ]] && __cu_contains_word "${words[c]}" "${two_word_flags[@]}"; then
        __cu_debug "${FUNCNAME[0]}: found a flag ${words[c]}, skip the next argument"
        c=$((c+1))
        # if we are looking for a flags value, don't show commands
        if [[ $c -eq $cword ]]; then
            commands=()
        fi
    fi

    c=$((c+1))

}

__cu_handle_noun()
{
    __cu_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    if __cu_contains_word "${words[c]}" "${must_have_one_noun[@]}"; then
        must_have_one_noun=()
    elif __cu_contains_word "${words[c]}" "${noun_aliases[@]}"; then
        must_have_one_noun=()
    fi

    nouns+=("${words[c]}")
    c=$((c+1))
}

__cu_handle_command()
{
    __cu_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    local next_command
    if [[ -n ${last_command} ]]; then
        next_command="_${last_command}_${words[c]//:/__}"
    else
        if [[ $c -eq 0 ]]; then
            next_command="_cu_root_command"
        else
            next_command="_${words[c]//:/__}"
        fi
    fi
    c=$((c+1))
    __cu_debug "${FUNCNAME[0]}: looking for ${next_command}"
    declare -F "$next_command" >/dev/null && $next_command
}

__cu_handle_word()
{
    if [[ $c -ge $cword ]]; then
        __cu_handle_reply
        return
    fi
    __cu_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"
    if [[ "${words[c]}" == -* ]]; then
        __cu_handle_flag
    elif __cu_contains_word "${words[c]}" "${commands[@]}"; then
        __cu_handle_command
    elif [[ $c -eq 0 ]]; then
        __cu_handle_command
    elif __cu_contains_word "${words[c]}" "${command_aliases[@]}"; then
        # aliashash variable is an associative array which is only supported in bash > 3.
        if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
            words[c]=${aliashash[${words[c]}]}
            __cu_handle_command
        else
            __cu_handle_noun
        fi
    else
        __cu_handle_noun
    fi
    __cu_handle_word
}

_cu_auth_login()
{
    last_command="cu_auth_login"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--token=")
    two_word_flags+=("--token")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    local_nonpersistent_flags+=("-t")
    flags+=("--workspace=")
    two_word_flags+=("--workspace")
    two_word_flags+=("-w")
    local_nonpersistent_flags+=("--workspace")
    local_nonpersistent_flags+=("--workspace=")
    local_nonpersistent_flags+=("-w")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_auth_logout()
{
    last_command="cu_auth_logout"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--workspace=")
    two_word_flags+=("--workspace")
    two_word_flags+=("-w")
    local_nonpersistent_flags+=("--workspace")
    local_nonpersistent_flags+=("--workspace=")
    local_nonpersistent_flags+=("-w")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_auth_status()
{
    last_command="cu_auth_status"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_auth()
{
    last_command="cu_auth"

    command_aliases=()

    commands=()
    commands+=("login")
    commands+=("logout")
    commands+=("status")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_bulk_close()
{
    last_command="cu_bulk_close"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--yes")
    flags+=("-y")
    local_nonpersistent_flags+=("--yes")
    local_nonpersistent_flags+=("-y")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_bulk_delete()
{
    last_command="cu_bulk_delete"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--yes")
    flags+=("-y")
    local_nonpersistent_flags+=("--yes")
    local_nonpersistent_flags+=("-y")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_bulk_update()
{
    last_command="cu_bulk_update"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--add-assignee=")
    two_word_flags+=("--add-assignee")
    local_nonpersistent_flags+=("--add-assignee")
    local_nonpersistent_flags+=("--add-assignee=")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--priority=")
    two_word_flags+=("--priority")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--priority")
    local_nonpersistent_flags+=("--priority=")
    local_nonpersistent_flags+=("-p")
    flags+=("--remove-assignee=")
    two_word_flags+=("--remove-assignee")
    local_nonpersistent_flags+=("--remove-assignee")
    local_nonpersistent_flags+=("--remove-assignee=")
    flags+=("--status=")
    two_word_flags+=("--status")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--status")
    local_nonpersistent_flags+=("--status=")
    local_nonpersistent_flags+=("-s")
    flags+=("--tag=")
    two_word_flags+=("--tag")
    local_nonpersistent_flags+=("--tag")
    local_nonpersistent_flags+=("--tag=")
    flags+=("--yes")
    flags+=("-y")
    local_nonpersistent_flags+=("--yes")
    local_nonpersistent_flags+=("-y")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_bulk()
{
    last_command="cu_bulk"

    command_aliases=()

    commands=()
    commands+=("close")
    commands+=("delete")
    commands+=("update")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_completion()
{
    last_command="cu_completion"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")
    local_nonpersistent_flags+=("--help")
    local_nonpersistent_flags+=("-h")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    must_have_one_noun+=("bash")
    must_have_one_noun+=("fish")
    must_have_one_noun+=("powershell")
    must_have_one_noun+=("zsh")
    noun_aliases=()
}

_cu_config_get()
{
    last_command="cu_config_get"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_config_list()
{
    last_command="cu_config_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_config_set()
{
    last_command="cu_config_set"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_config()
{
    last_command="cu_config"

    command_aliases=()

    commands=()
    commands+=("get")
    commands+=("list")
    commands+=("set")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_export_tasks()
{
    last_command="cu_export_tasks"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--assignee=")
    two_word_flags+=("--assignee")
    local_nonpersistent_flags+=("--assignee")
    local_nonpersistent_flags+=("--assignee=")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    local_nonpersistent_flags+=("-f")
    flags+=("--list=")
    two_word_flags+=("--list")
    two_word_flags+=("-l")
    local_nonpersistent_flags+=("--list")
    local_nonpersistent_flags+=("--list=")
    local_nonpersistent_flags+=("-l")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    local_nonpersistent_flags+=("-o")
    flags+=("--priority=")
    two_word_flags+=("--priority")
    local_nonpersistent_flags+=("--priority")
    local_nonpersistent_flags+=("--priority=")
    flags+=("--space=")
    two_word_flags+=("--space")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--space")
    local_nonpersistent_flags+=("--space=")
    local_nonpersistent_flags+=("-s")
    flags+=("--status=")
    two_word_flags+=("--status")
    local_nonpersistent_flags+=("--status")
    local_nonpersistent_flags+=("--status=")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_export()
{
    last_command="cu_export"

    command_aliases=()

    commands=()
    commands+=("tasks")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_help()
{
    last_command="cu_help"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    has_completion_function=1
    noun_aliases=()
}

_cu_interactive()
{
    last_command="cu_interactive"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_list_default()
{
    last_command="cu_list_default"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_list_list()
{
    last_command="cu_list_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--archived")
    local_nonpersistent_flags+=("--archived")
    flags+=("--folder=")
    two_word_flags+=("--folder")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--folder")
    local_nonpersistent_flags+=("--folder=")
    local_nonpersistent_flags+=("-f")
    flags+=("--space=")
    two_word_flags+=("--space")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--space")
    local_nonpersistent_flags+=("--space=")
    local_nonpersistent_flags+=("-s")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_list()
{
    last_command="cu_list"

    command_aliases=()

    commands=()
    commands+=("default")
    commands+=("list")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_me()
{
    last_command="cu_me"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_space_list()
{
    last_command="cu_space_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_space()
{
    last_command="cu_space"

    command_aliases=()

    commands=()
    commands+=("list")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_task_close()
{
    last_command="cu_task_close"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_task_create()
{
    last_command="cu_task_create"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--assignee=")
    two_word_flags+=("--assignee")
    two_word_flags+=("-a")
    local_nonpersistent_flags+=("--assignee")
    local_nonpersistent_flags+=("--assignee=")
    local_nonpersistent_flags+=("-a")
    flags+=("--description=")
    two_word_flags+=("--description")
    two_word_flags+=("-d")
    local_nonpersistent_flags+=("--description")
    local_nonpersistent_flags+=("--description=")
    local_nonpersistent_flags+=("-d")
    flags+=("--due=")
    two_word_flags+=("--due")
    local_nonpersistent_flags+=("--due")
    local_nonpersistent_flags+=("--due=")
    flags+=("--list=")
    two_word_flags+=("--list")
    two_word_flags+=("-l")
    local_nonpersistent_flags+=("--list")
    local_nonpersistent_flags+=("--list=")
    local_nonpersistent_flags+=("-l")
    flags+=("--name=")
    two_word_flags+=("--name")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    local_nonpersistent_flags+=("-n")
    flags+=("--priority=")
    two_word_flags+=("--priority")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--priority")
    local_nonpersistent_flags+=("--priority=")
    local_nonpersistent_flags+=("-p")
    flags+=("--status=")
    two_word_flags+=("--status")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--status")
    local_nonpersistent_flags+=("--status=")
    local_nonpersistent_flags+=("-s")
    flags+=("--tag=")
    two_word_flags+=("--tag")
    local_nonpersistent_flags+=("--tag")
    local_nonpersistent_flags+=("--tag=")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_task_interactive()
{
    last_command="cu_task_interactive"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_task_list()
{
    last_command="cu_task_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--assignee=")
    two_word_flags+=("--assignee")
    local_nonpersistent_flags+=("--assignee")
    local_nonpersistent_flags+=("--assignee=")
    flags+=("--due=")
    two_word_flags+=("--due")
    local_nonpersistent_flags+=("--due")
    local_nonpersistent_flags+=("--due=")
    flags+=("--folder=")
    two_word_flags+=("--folder")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--folder")
    local_nonpersistent_flags+=("--folder=")
    local_nonpersistent_flags+=("-f")
    flags+=("--limit=")
    two_word_flags+=("--limit")
    local_nonpersistent_flags+=("--limit")
    local_nonpersistent_flags+=("--limit=")
    flags+=("--list=")
    two_word_flags+=("--list")
    two_word_flags+=("-l")
    local_nonpersistent_flags+=("--list")
    local_nonpersistent_flags+=("--list=")
    local_nonpersistent_flags+=("-l")
    flags+=("--order=")
    two_word_flags+=("--order")
    local_nonpersistent_flags+=("--order")
    local_nonpersistent_flags+=("--order=")
    flags+=("--page=")
    two_word_flags+=("--page")
    local_nonpersistent_flags+=("--page")
    local_nonpersistent_flags+=("--page=")
    flags+=("--priority=")
    two_word_flags+=("--priority")
    local_nonpersistent_flags+=("--priority")
    local_nonpersistent_flags+=("--priority=")
    flags+=("--sort=")
    two_word_flags+=("--sort")
    local_nonpersistent_flags+=("--sort")
    local_nonpersistent_flags+=("--sort=")
    flags+=("--space=")
    two_word_flags+=("--space")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--space")
    local_nonpersistent_flags+=("--space=")
    local_nonpersistent_flags+=("-s")
    flags+=("--status=")
    two_word_flags+=("--status")
    local_nonpersistent_flags+=("--status")
    local_nonpersistent_flags+=("--status=")
    flags+=("--tag=")
    two_word_flags+=("--tag")
    local_nonpersistent_flags+=("--tag")
    local_nonpersistent_flags+=("--tag=")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_task_reopen()
{
    last_command="cu_task_reopen"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--status=")
    two_word_flags+=("--status")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--status")
    local_nonpersistent_flags+=("--status=")
    local_nonpersistent_flags+=("-s")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_task_search()
{
    last_command="cu_task_search"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--include-description")
    local_nonpersistent_flags+=("--include-description")
    flags+=("--limit=")
    two_word_flags+=("--limit")
    local_nonpersistent_flags+=("--limit")
    local_nonpersistent_flags+=("--limit=")
    flags+=("--list=")
    two_word_flags+=("--list")
    two_word_flags+=("-l")
    local_nonpersistent_flags+=("--list")
    local_nonpersistent_flags+=("--list=")
    local_nonpersistent_flags+=("-l")
    flags+=("--space=")
    two_word_flags+=("--space")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--space")
    local_nonpersistent_flags+=("--space=")
    local_nonpersistent_flags+=("-s")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_task_update()
{
    last_command="cu_task_update"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--add-assignee=")
    two_word_flags+=("--add-assignee")
    local_nonpersistent_flags+=("--add-assignee")
    local_nonpersistent_flags+=("--add-assignee=")
    flags+=("--description=")
    two_word_flags+=("--description")
    two_word_flags+=("-d")
    local_nonpersistent_flags+=("--description")
    local_nonpersistent_flags+=("--description=")
    local_nonpersistent_flags+=("-d")
    flags+=("--due=")
    two_word_flags+=("--due")
    local_nonpersistent_flags+=("--due")
    local_nonpersistent_flags+=("--due=")
    flags+=("--name=")
    two_word_flags+=("--name")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    local_nonpersistent_flags+=("-n")
    flags+=("--priority=")
    two_word_flags+=("--priority")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--priority")
    local_nonpersistent_flags+=("--priority=")
    local_nonpersistent_flags+=("-p")
    flags+=("--remove-assignee=")
    two_word_flags+=("--remove-assignee")
    local_nonpersistent_flags+=("--remove-assignee")
    local_nonpersistent_flags+=("--remove-assignee=")
    flags+=("--status=")
    two_word_flags+=("--status")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--status")
    local_nonpersistent_flags+=("--status=")
    local_nonpersistent_flags+=("-s")
    flags+=("--tag=")
    two_word_flags+=("--tag")
    local_nonpersistent_flags+=("--tag")
    local_nonpersistent_flags+=("--tag=")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_task_view()
{
    last_command="cu_task_view"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_task()
{
    last_command="cu_task"

    command_aliases=()

    commands=()
    commands+=("close")
    commands+=("create")
    commands+=("interactive")
    commands+=("list")
    commands+=("reopen")
    commands+=("search")
    commands+=("update")
    commands+=("view")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_user_list()
{
    last_command="cu_user_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_user()
{
    last_command="cu_user"

    command_aliases=()

    commands=()
    commands+=("list")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_version()
{
    last_command="cu_version"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_cu_root_command()
{
    last_command="cu"

    command_aliases=()

    commands=()
    commands+=("auth")
    commands+=("bulk")
    commands+=("completion")
    commands+=("config")
    commands+=("export")
    commands+=("help")
    commands+=("interactive")
    commands+=("list")
    commands+=("me")
    commands+=("space")
    commands+=("task")
    commands+=("user")
    commands+=("version")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--debug")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

__start_cu()
{
    local cur prev words cword split
    declare -A flaghash 2>/dev/null || :
    declare -A aliashash 2>/dev/null || :
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -s || return
    else
        __cu_init_completion -n "=" || return
    fi

    local c=0
    local flag_parsing_disabled=
    local flags=()
    local two_word_flags=()
    local local_nonpersistent_flags=()
    local flags_with_completion=()
    local flags_completion=()
    local commands=("cu")
    local command_aliases=()
    local must_have_one_flag=()
    local must_have_one_noun=()
    local has_completion_function=""
    local last_command=""
    local nouns=()
    local noun_aliases=()

    __cu_handle_word
}

if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_cu cu
else
    complete -o default -o nospace -F __start_cu cu
fi

# ex: ts=4 sw=4 et filetype=sh
