"use strict";

var CHUNK_SIZE, all, byId, cont, current, getCurrents, getLines, parseCSV, render, spc, text, triggered, ws;

ws = new WebSocket("{{.}}");

// CSV PARSER
parseCSV = function parseCSV(str) {
  var comma = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : ",";

  var arr, c, cc, col, nc, quote, row;
  arr = [];
  quote = false;
  // true means we're inside a quoted field
  // iterate over each character, keep track of current row and column (of the returned array)
  row = col = c = 0;
  while (c < str.length) {
    cc = str[c];
    nc = str[c + 1];
    // current character, next character
    arr[row] = arr[row] || [];
    // create a new row if necessary
    arr[row][col] = arr[row][col] || '';
    // create a new column (start with empty string) if necessary
    // If the current character is a quotation mark, and we're inside a
    // quoted field, and the next character is also a quotation mark,
    // add a quotation mark to the current column and skip the next character
    if (cc === '"' && quote && nc === '"') {
      arr[row][col] += cc;
      ++c;
      c++;
      continue;
    }
    // If it's just one quotation mark, begin/end quoted field
    if (cc === '"') {
      quote = !quote;
      c++;
      continue;
    }
    // If it's a comma and we're not in a quoted field, move on to the next column
    if (cc === comma && !quote) {
      ++col;
      c++;
      continue;
    }
    // If it's a newline (CRLF) and we're not in a quoted field, skip the next character
    // and move on to the next row and move to column 0 of that new row
    if (cc === "\r" && nc === '\n' && !quote) {
      ++row;
      col = 0;
      ++c;
      c++;
      continue;
    }
    // If it's a newline (LF or CR) and we're not in a quoted field,
    // move on to the next row and move to column 0 of that new row
    if (cc === '\n' && !quote) {
      ++row;
      col = 0;
      c++;
      continue;
    }
    if (cc === "\r" && !quote) {
      ++row;
      col = 0;
      c++;
      continue;
    }
    // Otherwise, append the current character to the current column
    arr[row][col] += cc;
    c++;
  }
  return arr;
};

CHUNK_SIZE = 50;

current = 0;

byId = function byId() {
  var _document;

  return (_document = document).getElementById.apply(_document, arguments);
};

getLines = function getLines(start, end) {
  return ws.send(start + ":" + end);
};

all = 0;

getCurrents = function getCurrents() {
  all += CHUNK_SIZE;
  return getLines(current * CHUNK_SIZE, (current + 1) * CHUNK_SIZE);
};

text = byId("text");

cont = byId("container");

spc = byId("spacer");

render = function render(data) {
  var html, i, item, j, len, len1, line, lines;
  html = "";
  lines = parseCSV(data, ";");
  for (i = 0, len = lines.length; i < len; i++) {
    line = lines[i];
    html += "<tr>";
    for (j = 0, len1 = line.length; j < len1; j++) {
      item = line[j];
      html += "<td>" + item + "</td>";
    }
    html += "</tr>";
  }
  return html;
};

triggered = false;

ws.onmessage = function (msg) {
  var data, linesCount, splitted;
  data = msg.data;
  splitted = data.split(":");
  if (splitted[0] === "lines") {
    linesCount = parseInt(splitted[1]);
    return spc.style.height = 33 * (linesCount - all) + "px";
  } else {
    text.innerHTML += render(data);
    return triggered = false;
  }
};

ws.onopen = function () {
  getCurrents();
  return ws.send("lines");
};

cont.onscroll = function () {
  var sc;
  sc = 33 * all;
  if (cont.scrollTop + cont.clientHeight >= sc - 300 && !triggered) {
    triggered = true;
    current += 1;
    return getCurrents();
  }
};