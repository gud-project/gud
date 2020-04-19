#!/usr/bin/env bats

command -v gud >/dev/null || make cli

readonly dir="$BATS_TMPDIR/test"


setup() {
	mkdir "$dir"
}

teardown() {
	rm -rf "$dir"
}

@test "start, add and save" {
	cd "$dir"

	gud start
	[ -d '.gud' ]

	run gud status
	[ "$status" -eq 0 ]
	[ -z "$output" ]

	# add file
	local -r file='f.txt'
	local -r data1='test data'

	echo "$data1" >"$file"
	run gud status
	[ "$status" -eq 0 ]
	[ "$output" = "non-update new:$file" ]

	gud add "$file"

	run gud status
	[ "$status" -eq 0 ]
	[ "$output" = "new: $file" ]

	# save
	local -r msg1='first version'
	gud save -m "$msg1"

	run gud status
	[ "$status" -eq 0 ]
	[ -z "$output" ]

	# change file
	local -r data2='new test data'
	echo "$data2" >"$file"

	run gud status
	[ "$status" -eq 0 ]
	[ "$output" = "non-update modified: $file" ]

	run gud add "$file"
	[ "$status" -eq 0 ]

	run gud status
	[ "$status" -eq 0 ]
	[ "$output" = "modified: $file" ]

	# save change
	local -r msg2='second version'
	run gud save -m "$msg2"
	[ "$status" -eq 0 ]

	run gud status
	[ "$status" -eq 0 ]
	[ -z "$output" ]

	# remove file
	rm "$file"
	run gud status
	[ "$status" -eq 0 ]
	echo "$output"
	[ "$output" = "non-update deleted: $file" ]

	echo "$data2" >"$file"
	gud rm "$file"
	run gud status
	[ "$status" -eq 0 ]
	echo "$output"
	[ "$output" = "deleted: $file" ]

	gud save -m 'third version'
}

@test "checkout" {
	cd "$dir"
	gud start

	# add file
	local -r file='f.txt'
	local -r data1='test data'
	echo "$data1" >"$file"

	local -r msg1='first version'
	gud add "$file"
	gud save -m "$msg1"

	run gud log
	[ "$status" -eq 0 ]
	[ "${lines[4]}" = "Message: $msg1" ]

	hash1="$(echo "${lines[7]}" | grep -Po '^Hash: \K([0-9a-f]{40})$')"
	readonly hash1
	[ -n "$hash1" ]

	# change file
	local -r data2='new test data'
	echo "$data2" >"$file"
	run gud add "$file"
	[ "$status" -eq 0 ]

	# save change
	local -r msg2='second version'
	run gud save -m "$msg2"
	[ "$status" -eq 0 ]

	run gud log
	[ "$status" -eq 0 ]
	[ "${lines[8]}" = "Message: $msg2" ]

	# go back to previous version
	run gud checkout "$hash1"
	[ "$status" -eq 0 ]
	[ "$(cat "$file")" = "$data1" ]

	# go to new version
	run gud checkout 'master'
	[ "$status" -eq 0 ]
	[ "$(cat "$file")" = "$data2" ]
}

@test "merge two separate branches" {
	cd "$dir"
	gud start

	run gud branch
	[ "$status" -eq 0 ]
	[ "${lines[1]}" = 'master' ]

	local -r branch='secondary'
	gud branch create "$branch"

	run gud branch
	[ "$status" -eq 0 ]
	[ "${lines[1]}" = "$branch" ]

	gud checkout master
	run gud branch
	[ "$status" -eq 0 ]
	[ "${lines[1]}" = 'master' ]

	local -r file1='f.txt'
	local -r data1='test data'
	echo "$data1" >"$file1"
	gud add "$file1"

	local -r msg1='master version'
	gud save -m "$msg1"

	gud checkout "$branch"
	[ ! -e "$file1" ]

	local -r file2='g.txt'
	local -r data2='different data'
	echo "$data2" >"$file2"
	gud add "$file2"

	local -r msg2='secondary version'
	gud save -m "$msg2"

	gud checkout 'master'
	[ ! -e "$file2" ]
	[   -e "$file1" ]
	[ "$(cat "$file1")" = "$data1" ]

	gud merge "$branch"
	[ -e "$file1" ]
	[ -e "$file2" ]
	[ "$(cat "$file1")" = "$data1" ]
	[ "$(cat "$file2")" = "$data2" ]
}

# TODO: not sure how to set this up
#@test "server" {
#	local -r user='nitai'
#	local -r pass='gudpass'
#	psql -d 'gud' -v "user=$user" -v "pass=$pass" \
#		-c "INSERT INTO users (username, email, password, created_at) VALUES (':user', a@b.com, ':pass', NOW());"
#
#	[ "$(psql -d 'gud' -v "user=$user" -At  -c "SELECT EXISTS(SELECT 1 FROM users WHERE username = ':user');")" = 't' ]
#
#	cd "$dir"
#	gud start
#
#	printf "%s\n$%s\n" "$user" "$pass" | gud login
#}
