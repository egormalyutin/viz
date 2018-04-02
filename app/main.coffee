################
#### CONFIG ####
################

alert "hi"

config =
	WS: '{{ws}}'
	types: '{{types}}'
	headers: '{{headers}}'

	animation: false
	animateLastChunk: true

#################
#### HELPERS ####
#################

byId = -> document.getElementById arguments...

debug = ->
	console.log "[DEBUG]", arguments...

endTimeout = null
onend = (cb) ->
	clearTimeout endTimeout
	endTimeout = setTimeout cb, 1000

##################
#### ELEMENTS ####
##################

# $ here means vanilla element, not JQuery element
$chunks   = byId "chunks"
$cont     = byId "container"
$spacer   = byId "spacer"
$table    = byId "table"
$top      = byId "top"

##################
#### GRAPHICS ####
##################

cacheGraph = []

charts = []

[colors, getColor] = do ->
	clr = [ "red", "green", "blue", "purple", "pink" ]
	i = 0
	get = ->
		ret = clr[i]
		i++
		return ret

	return [clr, get]

timeColumn = 0

addChart = (type, header) ->
	if type == "time"
		charts.push false
		return

	$ = document.createElement 'canvas'
	$.classList.add "chart"
	$.width  = 265
	$.height = 265
	$top.appendChild $

	ctx = $.getContext '2d'
	color = getColor()

	graphics = new Chart ctx,
		type: 'line'
		data: {
			labels: []
			datasets: [{
				label: header
				backgroundColor: color
				borderColor: color
				data: []
				fill: false
			}]
		},
		options: {
			responsive: false
			animation: config.animation
			title: {
				display: true
				text: header
			},
			tooltips: {
				mode: 'index'
				intersect: false
			},
			hover: {
				mode: 'nearest'
				intersect: true
			},
			scales: {
				xAxes: [{
					display: true
					scaleLabel: {
						display: true
						labelString: 'X'
					}
				}]
				yAxes: [{
					display: true
					scaleLabel: {
						display: true
						labelString: 'Y'
					}
				}]
			}
		}

	charts.push { type, chart: graphics }

parseGraphics = (chunks) ->
	arrs = []

	for chunk in chunks
		for num, line of chunk
			for num, chart of charts
				arrs[num] ?= labels: [], data: []
				arrs[num].labels.push line[timeColumn].split(" ")[1]

				if chart.type == "float"
					arrs[num].data.push parseFloat line[num]

				if chart.type == "int"
					arrs[num].data.push parseInt line[num]

	for num, arr of arrs when arr != undefined
		if charts[num] != false
			charts[num].chart.data.datasets[0].data = arr.data
			charts[num].chart.data.labels = arr.labels
			charts[num].chart.update()

###################
#### CONSTANTS ####
###################

CHUNK_SIZE = 50

CHUNK_PX_SIZE = 0
ROW_PX_SIZE   = 0

TOTAL_SIZE   = 0
TOTAL_CHUNKS = 0
ALL_PX_SIZE  = 0

# calculate chunk pixel size
do ->
	# new fake tbody
	$tbody = document.createElement "tbody"

	# fill tbody with CHUNKS_SIZE lines of fake content
	i = 0
	html = ""
	while i < CHUNK_SIZE
		html += "<tr class='row'><td>" + i + "</td></tr>"
		i++
	$tbody.innerHTML = html

	# add tbody to table
	$table.appendChild $tbody

	# check height
	CHUNK_PX_SIZE = $tbody.clientHeight
	ROW_PX_SIZE   = $tbody.children[0].clientHeight

	# remove from table
	$table.removeChild $tbody

	# message
	debug "chunk pixel size:", CHUNK_PX_SIZE
	debug   "row pixel size:",   ROW_PX_SIZE

graph = []

class App
	constructor: ->
		# server replaces {{.}} into WebSocket URL
		@ws = new WebSocket config.WS

		@types   = JSON.parse config.types
		debug "types:", @types

		@headers = JSON.parse config.headers
		debug "headers:", @headers

		@timeColumn = 0

		for num of @types
			if @types[num] == "time"
				timeColumn = num

		for num of @types
			addChart @types[num], @headers[num]

		@$activeChunks = {}
		@readyState = 2
		@datas = {}
		@i = 0
		@onscrollCalled = false

	bind: ->
		self = @
		$cont.onscroll = -> self.onscroll.apply self, arguments

		@ws.onmessage = -> self.commands.message.apply self, arguments
		@ws.onopen    = -> self.commands.open.apply self,    arguments
		@ws.onerror   = -> self.commands.error.apply self,   arguments
		@ws.onclose   = -> self.commands.close.apply self,   arguments

	commands:
		lines: (n) ->
			@calculateSize n

		message: (msg) ->
			data = JSON.parse msg.data

			switch data.type
				when "linesCount"
					@commands.lines.call @, data.linesCount
					debug "lines message:", data.linesCount
					@readyState--
					unless @onscrollCalled
						@onscrollCalled = true
						@onscroll()

				when "read"
					debug "data message"
					@push.call @, data.lines, data.id

				when "error"
					alert "ERROR: " + data.error

		open: ->
			debug "connection open"
			@ws.send '{"type":"linesCount"}'
			@readyState--

		error: (err) ->
			debug "connection error:", err

		close: ->
			debug "connection closed"

	onscroll: ->
		if @readyState <= 0
			# calculate top and bottom of container
			top    = $cont.scrollTop
			bottom = top + $cont.clientHeight

			self = @
			loadChunk = (currentChunk) ->
				# if chunk is inactive
				unless self.$activeChunks[currentChunk]?
					# positions
					$chunk = self.$activeChunks[currentChunk] = document.createElement("tbody")
					$chunk.style.position = "absolute"
					$chunk.style.top = (currentChunk * CHUNK_PX_SIZE) + "px"

					# load chunk content
					cstart = currentChunk * CHUNK_SIZE
					cend   = cstart + CHUNK_SIZE
					self.i++
					self.get cstart, cend, self.i

					data = await self.receive self.i

					# fill chunk element with CSV data
					arr = parseCSV data, ";"

					html = ""
					for line in arr
						html += "<tr class='row'>"
						for item in line
							html += "<td>" + item + "</td>"
						html += "</tr>"

					$chunk.innerHTML = html
					$chunk.chunkData = arr

					# add chunk to table
					$table.appendChild $chunk

					# message
					debug "loaded new chunk", $chunk

			# load chunks, bounding to user view
			loadChunk Math.floor top    / CHUNK_PX_SIZE
			loadChunk Math.floor top    / CHUNK_PX_SIZE - 1
			loadChunk Math.floor bottom / CHUNK_PX_SIZE
			loadChunk Math.floor bottom / CHUNK_PX_SIZE + 1

			gr = []

			# delete invisible chunks
			self = @
			for name, chunk of @$activeChunks
				run = (chunkTop, chunkBottom, chunkTopG, chunkBottomG) ->
						# check, is chunk fits into container
						unless (((chunkTop <= top) && (chunkBottom >= top)) or ((chunkTop <= bottom) && (chunkBottom >= bottom)))
							# remove chunk from active chunks
							delete self.$activeChunks[name]

							# remove chunk from table
							$table.removeChild chunk

							# message
							debug "deleted invisible chunk", chunk

						if (((chunkTopG <= top) && (chunkBottomG >= top)) or ((chunkTopG <= bottom) && (chunkBottomG >= bottom)))
							gr.push chunk.chunkData

				# calculate bottom and top of chunk
				chunkTop = parseInt chunk.style.top
				chunkBottom = chunkTop + CHUNK_PX_SIZE
				run chunkTop - CHUNK_PX_SIZE, chunkBottom + CHUNK_PX_SIZE, chunkTop, chunkBottom

			if config.animateLastChunk
				parseGraphics [gr[gr.length - 1]]
			else
				parseGraphics gr

			onend -> 
				self.onscroll()


	calculateSize: (size) ->
		TOTAL_SIZE   = size
		TOTAL_CHUNKS = Math.ceil TOTAL_SIZE / CHUNK_SIZE
		ALL_PX_SIZE = CHUNK_PX_SIZE * TOTAL_CHUNKS

		$spacer.style.height = ALL_PX_SIZE + "px"

	push: (data, id) ->
		@datas[id].data = data
		@datas[id].received = true

	receive: (id) ->
		@datas[id] = received: false
		self = @
		return new Promise (r) ->
			Object.defineProperty self.datas[id], "received",
				set: ->
					data = self.datas[id]
					delete self.datas[id]
					r data.data

	get: (start, end, id) ->
		@ws.send '{"type":"read","start":'+start+',"end":'+end+',"id":'+id+'}'

app = new App
app.bind()
