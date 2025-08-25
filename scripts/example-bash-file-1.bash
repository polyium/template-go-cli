#!/usr/bin/env bash

# -*-  Coding: UTF-8  -*- #
# -*-  System: Linux  -*- #
# -*-  Usage:   *.*   -*- #

#
# Advanced Bash Script Example
#

#
# The following script makes use of the following concepts:
#   - shellcheck
#   - setopt (set-options)
#   - dictionaries/hashmaps
#   - bash builtins
#   - standard-output & standard-error redirects
#   - jq projections

#
# Shellcheck Ignore List
#
# shellcheck disable=SC1073
# shellcheck disable=SC2120
# shellcheck disable=SC2071
# shellcheck disable=SC2086
# shellcheck disable=SC2048
#

set -o pipefail
set -o pipeline
set -o errexit
set -o xtrace

function unset-aws-environment-variables() {
    unset AWS_ACCESS_KEY_ID
    unset AWS_SECRET_ACCESS_KEY
    unset AWS_SESSION_TOKEN

    unset AWS_PROFILE
    unset AWS_DEFAULT_PROFILE
}

# a helper function that more-or-less acts as an argument-parser and validator for position based parameters.
function validate() {
    [[ (( ${#FUNCNAME[*]} < 2 )) ]] && echo -e "Fatal Error - Cannot Call \"${FUNCNAME[0]}\" || \"${FUNCNAME[1]}\" From Global Namespace." > /dev/stderr
    [[ (( ${#FUNCNAME[*]} < 2 )) ]] && return 1

    echo "[Debug] (${FUNCNAME[1]}) Evaluating Input. Arguments-Count: (${#}), Function-Count: (${#FUNCNAME[*]})" > /dev/stderr

    if ! [[ ${#} -eq 2 ]]; then
        echo "[Error] (${FUNCNAME[0]}) Invalid Arguments Received. Caller: \"${FUNCNAME[1]}\". Expected Total Arguments Required (1), and Total Caller Arguments (2)." > /dev/stderr
        return 1
    fi

    return 0
}

function generate-versioned-object-delete-input() {
    validate 1 ${#} || echo "[Error] (${FUNCNAME[0]}) Invalid Arguments Received. Expected Bucket-Name (1)." > /dev/stderr

    [[ -z "${1}" ]] && echo "[Error] (${FUNCNAME[0]}) Invalid Bucket-Name (1). Function: \"${FUNCNAME[0]}\", Caller: \"${FUNCNAME[1]}\"" > /dev/stderr
    [[ -z "${1}" ]] && return 1

    declare bucket
    declare input
    declare output

    readonly bucket="${1}"

    readonly input="input.delete-objects.${bucket}.json"
    readonly output="output.delete-objects.${bucket}.json"

    jq --indent 4 ". += {Quiet: false}" > "${output}" <(aws s3api list-object-versions --bucket "${bucket}" --output "json" --query "{Objects: Versions[].{Key:Key,VersionId:VersionId}}")

    echo "[Debug] (${FUNCNAME[0]}) Generated Object-Delete-Input. Format: JSON, File: \"${input}\"" > /dev/stderr

    printf "%s" "${output}" > /dev/stdout

    return 0
}

function delete-versioned-object() {
    validate 2 ${#} || echo "[Error] (${FUNCNAME[0]}) Invalid Arguments Received. Expected Bucket-Name (1), Delete-Object-Input-File (2)." > /dev/stderr

    [[ -z "${1}" ]] && echo "[Error] (${FUNCNAME[0]}) Invalid Bucket-Name (1). Caller: \"${FUNCNAME[1]}\"" > /dev/stderr
    [[ -z "${1}" ]] && return 1

    [[ -z "${2}" ]] && echo "[Error] (${FUNCNAME[0]}) Invalid Delete-Object-Input-File (2). Caller: \"${FUNCNAME[1]}\"" > /dev/stderr
    [[ -z "${2}" ]] && return 1

    declare bucket && readonly bucket="${1}"
    declare file && readonly file="${2}"

    echo "[Debug] (${FUNCNAME[0]}) Deleting AWS S3 Versioned Objects. File: \"${file}\"" > /dev/stderr

    aws s3api delete-objects --bucket "${bucket}" --delete "file://${file}"

    echo "[Log] (${FUNCNAME[0]}) Successfully Deleted All Versioned Objects" > /dev/stderr

    return 0
}

function delete-s3-bucket() {
    validate 1 ${#} || echo "[Error] (${FUNCNAME[0]}) Invalid Arguments Received. Expected Bucket-Name (1)." > /dev/stderr

    [[ -z "${1}" ]] && echo "[Error] (${FUNCNAME[0]}) Invalid Bucket-Name (1). Caller: \"${FUNCNAME[1]}\"" > /dev/stderr
    [[ -z "${1}" ]] && return 1

    declare bucket

    readonly bucket="${1}"

    echo "[Debug] (${FUNCNAME[0]}) Deleting AWS S3 Bucket. Bucket-Name: \"${bucket}\"" > /dev/stderr

    aws s3 rb "${bucket}" --force

    echo "[Log] (${FUNCNAME[0]}) Successfully Deleted All Versioned Objects" > /dev/stderr

    return 0
}

function pre-flight-checks() {
    declare name

    name="jq"
    if [[ -z $(command -v "${name}") ]]; then
        echo "[Error] (${FUNCNAME[0]}) Command Not Found. Target: \"${name}\"" > /dev/stderr
        return 1
    fi

    name="aws"
    if [[ -z $(command -v "${name}") ]]; then
        echo "[Error] (${FUNCNAME[0]}) Command Not Found. Target: \"${name}\"" > /dev/stderr
        return 1
    fi
}

function main() {
    echo "[Log] (${FUNCNAME[0]}) Deleting (${#*}) Total Bucket(s)" > /dev/stderr

    for item in ${*}; do
        declare -A mapping=( ["bucket-name"]="${item}" )

        echo "[Debug] (${FUNCNAME[0]}) Evaluating. Bucket-Name: \"${mapping["bucket-name"]}\""

        mapping["object-delete-input"]="$(generate-versioned-object-delete-input "${mapping["bucket-name"]}")"
        delete-versioned-object "${mapping["bucket-name"]}" "${mapping["object-delete-input"]}"
    done
}

pre-flight-checks

declare -a buckets=(
    "example-bucket-1"
    "example-bucket-2"
    "example-bucket-3"
)

main "${buckets[@]}"
