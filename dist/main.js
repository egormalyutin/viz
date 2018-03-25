"use strict";

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

function _asyncToGenerator(fn) { return function () { var gen = fn.apply(this, arguments); return new Promise(function (resolve, reject) { function step(key, arg) { try { var info = gen[key](arg); var value = info.value; } catch (error) { reject(error); return; } if (info.done) { resolve(value); } else { return Promise.resolve(value).then(function (value) { step("next", value); }, function (err) { step("throw", err); }); } } return step("next"); }); }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

//################
//### HELPERS ####
//################
var $chunks, $cont, $spacer, $table, ALL_PX_SIZE, App, CHUNK_PX_SIZE, CHUNK_SIZE, TOTAL_CHUNKS, TOTAL_SIZE, app, byId, debug;

byId = function byId() {
  var _document;

  return (_document = document).getElementById.apply(_document, arguments);
};

debug = function debug() {
  var _console;

  return (_console = console).log.apply(_console, ["[DEBUG]"].concat(Array.prototype.slice.call(arguments)));
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
CHUNK_SIZE = 100;

CHUNK_PX_SIZE = 0;

TOTAL_SIZE = 0;

TOTAL_CHUNKS = 0;

ALL_PX_SIZE = 0;

(function () {
  // calculate chunk pixel size
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
  return debug("chunk pixel size: " + CHUNK_PX_SIZE);
})();

App = function () {
  //############
  //### APP ####
  //############
  var App = function () {
    function App() {
      _classCallCheck(this, App);

      // server replaces {{.}} into WebSocket URL
      this.ws = new WebSocket("{{.}}");
      this.$activeChunks = {};
      this.readyState = 2;
      this.data = {
        received: false
      };
    }

    _createClass(App, [{
      key: "bind",
      value: function bind() {
        var self;
        self = this;
        $cont.onscroll = function () {
          return self.onscroll.apply(self, arguments);
        };
        this.ws.onmessage = function () {
          return self.commands.message.apply(self, arguments);
        };
        this.ws.onopen = function () {
          return self.commands.open.apply(self, arguments);
        };
        this.ws.onerror = function () {
          return self.commands.error.apply(self, arguments);
        };
        return this.ws.onclose = function () {
          return self.commands.close.apply(self, arguments);
        };
      }
    }, {
      key: "onscroll",
      value: function onscroll() {
        var bottom, chunk, chunkBottom, chunkTop, loadChunk, name, ref, results, self, top;
        if (this.readyState <= 0) {
          // calculate top and bottom of container
          top = $cont.scrollTop;
          bottom = top + $cont.clientHeight;
          self = this;
          loadChunk = function () {
            var _ref = _asyncToGenerator( /*#__PURE__*/regeneratorRuntime.mark(function _callee(currentChunk) {
              var $chunk, arr, cend, cstart, data, html, item, j, k, len, len1, line, ref;
              return regeneratorRuntime.wrap(function _callee$(_context) {
                while (1) {
                  switch (_context.prev = _context.next) {
                    case 0:
                      if (!(self.$activeChunks[currentChunk] == null)) {
                        _context.next = 16;
                        break;
                      }

                      // positions
                      $chunk = self.$activeChunks[currentChunk] = document.createElement("tbody");
                      $chunk.style.position = "absolute";
                      $chunk.style.top = currentChunk * CHUNK_PX_SIZE + "px";
                      // load chunk content
                      cstart = currentChunk * CHUNK_SIZE;
                      cend = cstart + CHUNK_SIZE;
                      self.get(cstart, cend);
                      _context.next = 9;
                      return self.receive();

                    case 9:
                      data = _context.sent;

                      // fill chunk element with CSV data
                      html = "";
                      ref = data.split('\n');
                      for (j = 0, len = ref.length; j < len; j++) {
                        line = ref[j];
                        arr = parseCSV(data, ";");
                        html += "<tr class='row'>";
                        for (k = 0, len1 = arr.length; k < len1; k++) {
                          item = arr[k];
                          html += "<td>" + line + "</td>";
                        }
                        html += "</tr>";
                      }
                      $chunk.innerHTML = html;
                      // add chunk to table
                      $table.appendChild($chunk);
                      // message
                      return _context.abrupt("return", debug("loaded new chunk", $chunk));

                    case 16:
                    case "end":
                      return _context.stop();
                  }
                }
              }, _callee, this);
            }));

            return function loadChunk(_x) {
              return _ref.apply(this, arguments);
            };
          }();
          // load chunks, bounding to user view
          loadChunk(Math.floor(top / CHUNK_PX_SIZE));
          loadChunk(Math.floor(bottom / CHUNK_PX_SIZE));
          ref = this.$activeChunks;
          // delete invisible chunks
          results = [];
          for (name in ref) {
            chunk = ref[name];
            try {
              // calculate bottom and top of chunk
              chunkTop = parseInt(chunk.style.top);
              chunkBottom = chunkTop + CHUNK_PX_SIZE;
              // check, is chunk fits into container
              if (!(chunkTop <= top && chunkBottom >= top || chunkTop <= bottom && chunkBottom >= bottom)) {
                // remove chunk from active chunks
                delete this.$activeChunks[name];
                // remove chunk from table
                $table.removeChild(chunk);
                // message
                results.push(debug("deleted invisible chunk", chunk));
              } else {
                results.push(void 0);
              }
            } catch (error) {}
          }
          return results;
        }
      }
    }, {
      key: "calculateSize",
      value: function calculateSize(size) {
        TOTAL_SIZE = size;
        TOTAL_CHUNKS = Math.ceil(TOTAL_SIZE / CHUNK_SIZE);
        ALL_PX_SIZE = CHUNK_PX_SIZE * TOTAL_CHUNKS;
        return $spacer.style.height = ALL_PX_SIZE + "px";
      }
    }, {
      key: "push",
      value: function push(data) {
        this.data.data = data;
        return this.data.received = true;
      }
    }, {
      key: "receive",
      value: function receive() {
        var self;
        self = this;
        return new Promise(function (r) {
          return Object.defineProperty(self.data, "received", {
            set: function set() {
              return r(self.data.data);
            }
          });
        });
      }
    }, {
      key: "get",
      value: function get(start, end) {
        return this.ws.send(start + ":" + end);
      }
    }]);

    return App;
  }();

  ;

  App.prototype.commands = {
    lines: function lines(n) {
      return this.calculateSize(n);
    },
    message: function message(msg) {
      var data, splitted;
      data = msg.data;
      splitted = data.split(":");
      if (splitted[0] === "lines") {
        this.commands.lines.call(this, parseInt(splitted[1]));
        debug("lines message:", splitted[1]);
        return this.readyState--;
      } else {
        debug("data message");
        return this.push.call(this, data);
      }
    },
    open: function open() {
      debug("connection open");
      this.ws.send("lines");
      return this.readyState--;
    },
    error: function error(err) {
      return debug("connection error:", err);
    },
    close: function close() {
      return debug("connection closed");
    }
  };

  return App;
}.call(undefined);

app = new App();

app.bind();