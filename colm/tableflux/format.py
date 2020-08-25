import csv
import sys
import re
import string

csv_f = csv.reader(sys.stdin)

pattern = re.compile("^([0-9.\+\-]+)+$")

def print_table(rows, widths, dirs):
	underline = 1
	for row in rows:
		colsep = ""
		line = ""
		for j in range(len(row)):
			cell = row[j]
			line = line + colsep
			line = line + "{0: {dir}{width}}".format(cell,
					dir = dirs[j], width = widths[j] )
			colsep = " | "

		sys.stdout.write( line )
		sys.stdout.write( "\n" )

		if underline != 0:
			sys.stdout.write( "".join( ["="] * len(line) ) )
			sys.stdout.write( "\n" );

		underline = 0

widths = []
dirs = []
rows = []
table_sep = ""

for row in csv_f:
	if len(row) == 0:
		if len(rows) > 0:
			sys.stdout.write( table_sep )
			print_table( rows, widths, dirs )
				
		table_sep = "\n"
		widths = []
		dirs = []
		rows = []
	else:
		row = row[2:]
		if len(row) > len(widths):
			additional = len(row) - len(widths)
			widths.extend( [0] * additional )
			dirs.extend( ["<"] * additional )

		for j in range(len(row)):
			if len(row[j]) > widths[j]:
				widths[j] = len(row[j])
			if pattern.match(row[j]):
				dirs[j] = ">"
				
		rows.append( row )



