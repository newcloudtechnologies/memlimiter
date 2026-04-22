#!/usr/bin/awk -f
# Purpose: pretty-print Makefile targets with descriptions from "##" comments.
# Usage: awk -f scripts/make_help.awk $(MAKEFILE_LIST)

# Global variables:
# - target: current target name (the block that owns found ## lines).
# - desc_count: number of description lines attached to target.
BEGIN {
	target = ""
	desc_count = 0
}

# print_entry prints current target with collected descriptions.
# Format: cyan target name, then each comment on a new line.
function print_entry() {
	if (target == "")
		return

	printf "\033[36m%s\033[0m\n", target

	if (desc_count == 0) {
		printf "  (description not provided)\n"
	} else {
		for (i = 1; i <= desc_count; i++) {
			printf "  %s\n", desc[i]
		}
	}

	# Empty line for visual separation between blocks.
	printf "\n"
}

# reset_desc clears collected descriptions
# (AWK has no built-in clear(), so delete items one by one).
function reset_desc() {
	for (i = 1; i <= desc_count; i++) {
		delete desc[i]
	}

	desc_count = 0
}

# Match target definitions:
# - start of line;
# - valid identifier (letters/digits/._-/);
# - ':' is not part of assignment (exclude foo:=, foo::=, etc.).
/^[[:alnum:]_][[:alnum:]_.\/-]*:([^=]|$)/ {
	match($0, /^[[:alnum:]_][[:alnum:]_.\/-]*/)
	name = substr($0, RSTART, RLENGTH)

	if (target == name)
		next

	if (target != "") {
		print_entry()
		reset_desc()
	} else {
		reset_desc()
	}

	# Important: save name after printing previous entry, or context is lost.
	target = name
	next
}

# Description line (starts with ## after optional spaces).
/^[[:space:]]*##[[:space:]]+/ {
	if (target == "")
		next

	line = $0
	sub(/^[[:space:]]*##[[:space:]]+/, "", line)
	desc[++desc_count] = line
	next
}

# At EOF print final block (if at least one target was found).
END {
	if (target != "")
		print_entry()
}
