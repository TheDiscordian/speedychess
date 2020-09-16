#!/usr/bin/python

import sys

if len(sys.argv) != 2:
	print("Invalid number of arguments, expects 1 (server: bool).")
	quit(1)

if sys.argv[1] != "true" and sys.argv[1] != "false":
	print("Only arguments expected are true/false.")
	quit(2)

output = """package flag

const SERVER = %s""" % sys.argv[1]

open('../flags/compile.go', 'w').write(output)
