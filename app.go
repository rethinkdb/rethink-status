package main

import (
	r "github.com/dancannon/gorethink"
	"github.com/googollee/go-socket.io"
	"log"
	"net/http"
)

func main() {
	server, _ := socketio.NewServer(nil)

	conn, _ := r.Connect(r.ConnectOpts{
		Address:  "localhost:28015",
		Database: "rethinkdb",
	})

	stats, _ := r.Table("stats").Filter(
		r.Row.Field("id").AtIndex(0).Eq("cluster")).Changes().Run(conn)

	go func() {
		var change r.WriteChanges
		for stats.Next(&change) {
			server.BroadcastTo("monitor", "stats", change.NewValue)
		}
	}()

	servers, _ := r.Table("server_status").Changes().Run(conn)

	go func() {
		var change r.WriteChanges
		for servers.Next(&change) {
			server.BroadcastTo("monitor", "servers", change)
		}
	}()

	server.On("connection", func(so socketio.Socket) {
		so.Join("monitor")

		result, _ := r.Table("server_status").Run(conn)

		var output []interface{}
		result.All(&output)

		so.Emit("servers", output)
	})

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("public")))

	log.Println("Starting server on port 8091")
	log.Fatal(http.ListenAndServe(":8091", nil))
}
