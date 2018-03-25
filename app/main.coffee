#################
#### HELPERS ####
#################

byId = -> document.getElementById arguments...
debug = ->
	console.log "[DEBUG]", arguments...

##################
#### ELEMENTS ####
##################

# $ here means vanilla element, not JQuery element
$chunks = byId "chunks"
$cont   = byId "container"
$spacer = byId "spacer"
$table  = byId "table"

###################
#### CONSTANTS ####
###################

CHUNK_SIZE = 50
CHUNK_PX_SIZE = 0

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

	# remove from table
	$table.removeChild $tbody

	# message
	debug "chunk pixel size:", CHUNK_PX_SIZE

graph = []

class App
	constructor: ->
		# server replaces {{.}} into WebSocket URL
		@ws = new WebSocket "{{.}}"
		@$activeChunks = {}
		@readyState = 2
		@datas = {}
		@i = 0

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
			loadChunk Math.floor bottom / CHUNK_PX_SIZE

			gr = []

			# delete invisible chunks
			for name, chunk of @$activeChunks
				try
					# calculate bottom and top of chunk
					chunkTop = parseInt chunk.style.top
					chunkBottom = chunkTop + CHUNK_PX_SIZE

					# check, is chunk fits into container
					unless (((chunkTop <= top) && (chunkBottom >= top)) or ((chunkTop <= bottom) && (chunkBottom >= bottom)))
						# remove chunk from active chunks
						delete @$activeChunks[name]

						# remove chunk from table
						$table.removeChild chunk

						# message
						debug "deleted invisible chunk", chunk
					else
						gr.push chunk.chunkData

			graph = gr



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
