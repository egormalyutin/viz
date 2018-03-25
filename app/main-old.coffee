ws = new WebSocket "{{.}}"

# CSV PARSER
parseCSV = (str, comma = ",") ->
	arr = []
	quote = false
	# true means we're inside a quoted field
	# iterate over each character, keep track of current row and column (of the returned array)
	row = col = c = 0
	while c < str.length
		cc = str[c]
		nc = str[c + 1]
		# current character, next character
		arr[row] = arr[row] or []
		# create a new row if necessary
		arr[row][col] = arr[row][col] or ''
		# create a new column (start with empty string) if necessary
		# If the current character is a quotation mark, and we're inside a
		# quoted field, and the next character is also a quotation mark,
		# add a quotation mark to the current column and skip the next character
		if cc == '"' and quote and nc == '"'
			arr[row][col] += cc
			++c
			c++
			continue
		# If it's just one quotation mark, begin/end quoted field
		if cc == '"'
			quote = !quote
			c++
			continue
		# If it's a comma and we're not in a quoted field, move on to the next column
		if cc == comma and !quote
			++col
			c++
			continue
		# If it's a newline (CRLF) and we're not in a quoted field, skip the next character
		# and move on to the next row and move to column 0 of that new row
		if cc == '\u000d' and nc == '\n' and !quote
			++row
			col = 0
			++c
			c++
			continue
		# If it's a newline (LF or CR) and we're not in a quoted field,
		# move on to the next row and move to column 0 of that new row
		if cc == '\n' and !quote
			++row
			col = 0
			c++
			continue
		if cc == '\u000d' and !quote
			++row
			col = 0
			c++
			continue
		# Otherwise, append the current character to the current column
		arr[row][col] += cc
		c++
	arr

CHUNK_SIZE = 50

current = 0

byId = -> document.getElementById arguments...

getLines = (start, end) ->
	ws.send start + ":" + end

all = 0

getCurrents = ->
	all += CHUNK_SIZE
	getLines current * CHUNK_SIZE, (current + 1) * CHUNK_SIZE

text = byId "text"
cont = byId "container"
spc  = byId "spacer"

render = (data) ->
	html = ""
	lines = parseCSV data, ";"
	for line in lines
		html += "<tr>"
		for item in line
			html += "<td>" + item + "</td>"
		html += "</tr>"

	return html

triggered = false

ws.onmessage = (msg) ->
	data = msg.data
	splitted = data.split(":")
	if splitted[0] == "lines"
		linesCount = parseInt splitted[1]
		spc.style.height = (33 * (linesCount - all)) + "px"
	else
		text.innerHTML += render data
		triggered = false

ws.onopen = ->
	getCurrents()
	ws.send "lines"

cont.onscroll = ->
	sc = 33 * all
	if ((cont.scrollTop + cont.clientHeight) >= (sc - 300)) and not triggered
		triggered = true
		current += 1
		getCurrents()
