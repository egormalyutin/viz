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

CHUNK_SIZE = 100
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
	debug "chunk pixel size: " + CHUNK_PX_SIZE

#############
#### APP ####
#############

class App
	constructor: ->
		# server replaces {{.}} into WebSocket URL
		@ws = new WebSocket "{{.}}"
		@$activeChunks = {}
		@readyState = 2
		@data = received: false

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
			data = msg.data
			splitted = data.split(":")

			if splitted[0] == "lines"
				@commands.lines.call @, parseInt splitted[1]
				debug "lines message:", splitted[1]
				@readyState--
			else
				debug "data message"
				@push.call @, data

		open: ->
			debug "connection open"
			@ws.send "lines"
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
					self.get cstart, cend

					data = await self.receive()

					# fill chunk element with CSV data
					html = ""
					for line in data.split '\n'
						arr = parseCSV data, ";"

						html += "<tr class='row'>"
						for item in arr
							html += "<td>" + line + "</td>"
						html += "</tr>"

					$chunk.innerHTML = html

					# add chunk to table
					$table.appendChild $chunk

					# message
					debug "loaded new chunk", $chunk

			# load chunks, bounding to user view
			loadChunk Math.floor top    / CHUNK_PX_SIZE
			loadChunk Math.floor bottom / CHUNK_PX_SIZE

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

	calculateSize: (size) ->
		TOTAL_SIZE   = size
		TOTAL_CHUNKS = Math.ceil TOTAL_SIZE / CHUNK_SIZE
		ALL_PX_SIZE = CHUNK_PX_SIZE * TOTAL_CHUNKS

		$spacer.style.height = ALL_PX_SIZE + "px"

	push: (data) ->
		@data.data = data
		@data.received = true

	receive: ->
		self = @
		return new Promise (r) ->
			Object.defineProperty self.data, "received",
				set: ->
					r self.data.data

	get: (start, end) ->
		@ws.send start + ":" + end

app = new App
app.bind()
