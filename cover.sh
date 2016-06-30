#!/bin/sh
# Generate test coverage statistics for Go packages.
#
# Works around the fact that `go test -coverprofile` currently does not work
# with multiple packages, see https://code.google.com/p/go/issues/detail?id=6909
#
# Usage: script/coverage [--html|--coveralls]
#
#     --html      Additionally create HTML report and open it in browser
#     --coveralls Push coverage statistics to coveralls.io
#

# Adopted from https://github.com/mlafeldt/chef-runner/raw/v0.7.0/script/coverage (Apache v2)
# And https://github.com/golang/go/issues/6909 (mmindenhall comment)
# Edited by Romans Volosatovs <b1101@riseup.net> on 29.06.2016

set -e

workdir=".cover"
profile="cover.out"
mode="atomic"

generate_cover_data() {
    mkdir -p "$workdir"

    # For each package with test files, run with full coverage (including other packages)
    go list -f '{{if gt (len .TestGoFiles) 0}}"go test -covermode='${mode}' -coverprofile='${workdir}'/{{.Name}}.coverprofile -coverpkg=./... {{.ImportPath}}"{{end}}' ./... | xargs -I {} bash -c {}

    # Merge the generated cover profiles into a single file
    gocovmerge `ls $workdir/*.coverprofile` > $profile

    rm -rf "$workdir"
}

show_cover_report() {
    go tool cover -${1}="$profile"
}

push_to_coveralls() {
    if [ $# = 0 ]; then
        goveralls -coverprofile="$profile"
    else
        goveralls -coverprofile="$profile" -service=${1}
    fi
}

generate_cover_data
show_cover_report func

case "$1" in
    "")
        ;;
    --html)
        show_cover_report html ;;
    --coveralls)
        push_to_coveralls $2;;
    *)
        echo >&2 "error: invalid option: $1"; exit 1 ;;
esac
