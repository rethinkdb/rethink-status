
var express = require("express");
var sockio = require("socket.io");
var r = require("rethinkdb");

var app = express();
app.use(express.static(__dirname + "/public"));

var io = sockio.listen(app.listen(8099), {log: false});
console.log("Server started on port " + 8099);

r.connect({db: "rethinkdb"}).then(function(c) {
  r.table("stats").get(["cluster"]).changes().run(c)
    .then(function(cursor) {
      cursor.each(function(err, item) {
        io.sockets.emit("stats", item);
      });
    });

  r.table("server_status").changes().run(c)
    .then(function(cursor) {
      cursor.each(function(err, item) {
        io.sockets.emit("servers", item);
      });
    });
});

io.sockets.on("connection", function(socket) {
  var conn;
  r.connect({db: "rethinkdb"}).then(function(c) {
    conn = c;
    return r.table("server_status").run(conn);
  })
  .then(function(cursor) { return cursor.toArray(); })
  .then(function(result) {
    socket.emit("servers", result);
  })
  .error(function(err) { console.log("Failure:", err); })
  .finally(function() {
    if (conn)
      conn.close();
  });
});
