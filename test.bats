#!/usr/bin/env bats

make -s cli

# test data
readonly file='f.txt'
readonly dir="$BATS_TMPDIR/test"
readonly abs_file="$dir/$file"

readonly data1='test data'
readonly data2='new test data'

readonly msg1='first version'
readonly msg2='second version'


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

	run gud branch
	[ "$status" -eq 0 ]
	[ "${lines[1]}" = 'master' ]

	run gud status
	[ "$status" -eq 0 ]
	[ -z "$output" ]

	# add file
	echo "$data1" >"$abs_file"
	run gud status
	[ "$status" -eq 0 ]
	[ "$output" = "non-update new:$file" ]

	gud add "$abs_file"

	run gud status
	[ "$status" -eq 0 ]
	[ "$output" = "new: $file" ]

	# save
	gud save -m "$msg1"

	run gud status
	[ "$status" -eq 0 ]
	[ -z "$output" ]

	# change file
	echo "$data2" >"$abs_file"

	run gud status
	[ "$status" -eq 0 ]
	[ "$output" = "non-update modified: $file" ]

	run gud add "$abs_file"
	[ "$status" -eq 0 ]

	run gud status
	[ "$status" -eq 0 ]
	[ "$output" = "modified: $file" ]

	# save change
	run gud save -m "$msg2"
	[ "$status" -eq 0 ]

	run gud status
	[ "$status" -eq 0 ]
	[ -z "$output" ]

	# go back to previous version
	run gud checkout "$hash1"
	[ "$status" -eq 0 ]
	[ "$(cat "$abs_file")" = "$data1" ]

	# go to new version
	run gud checkout 'master'
	[ "$status" -eq 0 ]
	[ "$(cat "$abs_file")" = "$data2" ]
}

@test "checkout" {
	cd "$dir"
	gud start

	# add file
	echo "$data1" >"$abs_file"

	gud add "$abs_file"
	gud save -m "$msg1"

	run gud log
	[ "$status" -eq 0 ]
	[ "${lines[0]}" = "Message: $msg1" ]

	hash1="$(echo "${lines[2]}" | grep -Po '^Hash: \K([0-9a-f]{40})$')"
	readonly hash1
	[ -n "$hash1" ]

	# change file
	echo "$data2" >"$abs_file"
	run gud add "$abs_file"
	[ "$status" -eq 0 ]

	# save change
	run gud save -m "$msg2"
	[ "$status" -eq 0 ]

	run gud log
	[ "$status" -eq 0 ]
	[ "${lines[3]}" = "Message: $msg2" ]

	# go back to previous version
	run gud checkout "$hash1"
	[ "$status" -eq 0 ]
	[ "$(cat "$abs_file")" = "$data1" ]

	# go to new version
	run gud checkout 'master'
	[ "$status" -eq 0 ]
	[ "$(cat "$abs_file")" = "$data2" ]
}
