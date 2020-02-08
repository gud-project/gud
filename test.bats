#!/usr/bin/env bats

make client >/dev/null

setup() {
	dir="$BATS_TMPDIR/test"
	mkdir "$dir"
}

teardown() {
	rm -rf "$dir"
}

@test "lol" {
	cd "$dir"

	run gud start
	[ "$status" -eq 0 ]
	[ -d ".gud" ]

	run gud branch
	[ "$status" -eq 0 ]
	[ "${lines[1]}" = "master" ]

	run gud status
	[ "$status" -eq 0 ]

	local file="f.txt"
	local abs_file="$dir/$file"
	local data="test data"
	echo "$data" >"$abs_file"

	run gud status
	[ "$status" -eq 0 ]
	[ "$output" = "non-update new:$file" ]

	run gud add "$abs_file"
	[ "$status" -eq 0 ]

	run gud status
	[ "$status" -eq 0 ]
	[ "$output" = "new: $file" ]
}
