//################
//### HELPERS ####
//################
var $chunks, $cont, $spacer, $table, ALL_PX_SIZE, App, CHUNK_PX_SIZE, CHUNK_SIZE, TOTAL_CHUNKS, TOTAL_SIZE, app, byId, debug, graph;

byId = function() {
  return document.getElementById(...arguments);
};

debug = function() {
  return console.log("[DEBUG]", ...arguments);
};

//#################
//### ELEMENTS ####
//#################

// $ here means vanilla element, not JQuery element
$chunks = byId("chunks");

$cont = byId("container");

$spacer = byId("spacer");

$table = byId("table");

//##################
//### CONSTANTS ####
//##################
CHUNK_SIZE = 50;

CHUNK_PX_SIZE = 0;

TOTAL_SIZE = 0;

TOTAL_CHUNKS = 0;

ALL_PX_SIZE = 0;

(function() {  // calculate chunk pixel size
  var $tbody, html, i;
  // new fake tbody
  $tbody = document.createElement("tbody");
  // fill tbody with CHUNKS_SIZE lines of fake content
  i = 0;
  html = "";
  while (i < CHUNK_SIZE) {
    html += "<tr class='row'><td>" + i + "</td></tr>";
    i++;
  }
  $tbody.innerHTML = html;
  // add tbody to table
  $table.appendChild($tbody);
  // check height
  CHUNK_PX_SIZE = $tbody.clientHeight;
  // remove from table
  $table.removeChild($tbody);
  // message
  return debug("chunk pixel size:", CHUNK_PX_SIZE);
})();

graph = [];

App = (function() {
  class App {
    constructor() {
      // server replaces {{.}} into WebSocket URL
      this.ws = new WebSocket("{{.}}");
      this.$activeChunks = {};
      this.readyState = 2;
      this.datas = {};
      this.i = 0;
    }

    bind() {
      var self;
      self = this;
      $cont.onscroll = function() {
        return self.onscroll.apply(self, arguments);
      };
      this.ws.onmessage = function() {
        return self.commands.message.apply(self, arguments);
      };
      this.ws.onopen = function() {
        return self.commands.open.apply(self, arguments);
      };
      this.ws.onerror = function() {
        return self.commands.error.apply(self, arguments);
      };
      return this.ws.onclose = function() {
        return self.commands.close.apply(self, arguments);
      };
    }

    onscroll() {
      var bottom, chunk, chunkBottom, chunkTop, gr, loadChunk, name, ref, self, top;
      if (this.readyState <= 0) {
        // calculate top and bottom of container
        top = $cont.scrollTop;
        bottom = top + $cont.clientHeight;
        self = this;
        loadChunk = async function(currentChunk) {
          var $chunk, arr, cend, cstart, data, html, item, j, k, len, len1, line;
          // if chunk is inactive
          if (self.$activeChunks[currentChunk] == null) {
            // positions
            $chunk = self.$activeChunks[currentChunk] = document.createElement("tbody");
            $chunk.style.position = "absolute";
            $chunk.style.top = (currentChunk * CHUNK_PX_SIZE) + "px";
            // load chunk content
            cstart = currentChunk * CHUNK_SIZE;
            cend = cstart + CHUNK_SIZE;
            self.i++;
            self.get(cstart, cend, self.i);
            data = (await self.receive(self.i));
            // fill chunk element with CSV data
            arr = parseCSV(data, ";");
            html = "";
            for (j = 0, len = arr.length; j < len; j++) {
              line = arr[j];
              html += "<tr class='row'>";
              for (k = 0, len1 = line.length; k < len1; k++) {
                item = line[k];
                html += "<td>" + item + "</td>";
              }
              html += "</tr>";
            }
            $chunk.innerHTML = html;
            $chunk.chunkData = arr;
            // add chunk to table
            $table.appendChild($chunk);
            // message
            return debug("loaded new chunk", $chunk);
          }
        };
        // load chunks, bounding to user view
        loadChunk(Math.floor(top / CHUNK_PX_SIZE));
        loadChunk(Math.floor(bottom / CHUNK_PX_SIZE));
        gr = [];
        ref = this.$activeChunks;
        // delete invisible chunks
        for (name in ref) {
          chunk = ref[name];
          try {
            // calculate bottom and top of chunk
            chunkTop = parseInt(chunk.style.top);
            chunkBottom = chunkTop + CHUNK_PX_SIZE;
            // check, is chunk fits into container
            if (!(((chunkTop <= top) && (chunkBottom >= top)) || ((chunkTop <= bottom) && (chunkBottom >= bottom)))) {
              // remove chunk from active chunks
              delete this.$activeChunks[name];
              // remove chunk from table
              $table.removeChild(chunk);
              // message
              debug("deleted invisible chunk", chunk);
            } else {
              gr.push(chunk.chunkData);
            }
          } catch (error) {}
        }
        return graph = gr;
      }
    }

    calculateSize(size) {
      TOTAL_SIZE = size;
      TOTAL_CHUNKS = Math.ceil(TOTAL_SIZE / CHUNK_SIZE);
      ALL_PX_SIZE = CHUNK_PX_SIZE * TOTAL_CHUNKS;
      return $spacer.style.height = ALL_PX_SIZE + "px";
    }

    push(data, id) {
      this.datas[id].data = data;
      return this.datas[id].received = true;
    }

    receive(id) {
      var self;
      this.datas[id] = {
        received: false
      };
      self = this;
      return new Promise(function(r) {
        return Object.defineProperty(self.datas[id], "received", {
          set: function() {
            var data;
            data = self.datas[id];
            delete self.datas[id];
            return r(data.data);
          }
        });
      });
    }

    get(start, end, id) {
      return this.ws.send('{"type":"read","start":' + start + ',"end":' + end + ',"id":' + id + '}');
    }

  };

  App.prototype.commands = {
    lines: function(n) {
      return this.calculateSize(n);
    },
    message: function(msg) {
      var data;
      data = JSON.parse(msg.data);
      switch (data.type) {
        case "linesCount":
          this.commands.lines.call(this, data.linesCount);
          debug("lines message:", data.linesCount);
          return this.readyState--;
        case "read":
          debug("data message");
          return this.push.call(this, data.lines, data.id);
        case "error":
          return alert("ERROR: " + data.error);
      }
    },
    open: function() {
      debug("connection open");
      this.ws.send('{"type":"linesCount"}');
      return this.readyState--;
    },
    error: function(err) {
      return debug("connection error:", err);
    },
    close: function() {
      return debug("connection closed");
    }
  };

  return App;

}).call(this);

app = new App;

app.bind();

//# sourceMappingURL=maps/main.js.map
