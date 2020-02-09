#!/usr/bin/env bats

make cli >/dev/null

setup() {
	dir="$BATS_TMPDIR/test"
	mkdir "$dir"
}

teardown() {
	rm -rf "$dir"
}

@test "basic test" {
	cd "$dir"

	run gud start
	[ "$status" -eq 0 ]
	[ -d ".gud" ]

	run gud branch
	[ "$status" -eq 0 ]
	[ "${lines[1]}" = "master" ]

	run gud status
	[ "$status" -eq 0 ]
	[ -z "$output" ]

	readonly file="f.txt"
	readonly abs_file="$dir/$file"
	readonly data1="test data"
	echo "$data1" >"$abs_file"

	run gud status
	[ "$status" -eq 0 ]
	[ "$output" = "non-update new:$file" ]

	run gud add "$abs_file"
	[ "$status" -eq 0 ]

	run gud status
	[ "$status" -eq 0 ]
	[ "$output" = "new: $file" ]

	readonly msg1="first version"
	run gud save -m "$msg1"
	[ "$status" -eq 0 ]

	run gud log
	[ "$status" -eq 0 ]
	[ "${lines[0]}" = "Message: $msg1" ]

	hash1="$(echo "${lines[2]}" | grep -Po '^Hash: \K([0-9a-f]{40})$')"
	readonly hash1
	[ -n "$hash1" ]

	run gud status
	[ "$status" -eq 0 ]
	[ -z "$output" ]

	readonly data2="new test data"
	echo "$data2" >"$abs_file"

	run gud status
	[ "$status" -eq 0 ]
	printf '"%s"' "$output"
	[ "$output" = "non-update new:$file" ]

	run gud add "$abs_file"
	[ "$status" -eq 0 ]

	run gud status
	[ "$status" -eq 0 ]
	[ "$output" = "new: $file" ]

	readonly msg2="second version"
	run gud save -m "$msg2"
	[ "$status" -eq 0 ]

	run gud log
	[ "$status" -eq 0 ]
	[ "${lines[0]}" = "Message: $msg2" ]

	run gud status
	[ "$status" -eq 0 ]
	[ -z "$output" ]

	run gud checkout "$hash1"
	[ "$status" -eq 0 ]
	[ "$(cat "$abs_file")" = "$data1" ]
}
