substitution='s:b1101:rvolosatovs:'
for f in `find -iname '*.go' -o -iname 'Makefile' | grep -v 'vendor'`; do sed -i $substitution $f; done

repo=`git remote get-url origin`
[ `grep 'b1101' <<< $repo` ] && "git remote set-url origin `sed -e $substitution <<< $repo`"
