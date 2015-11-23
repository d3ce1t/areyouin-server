package main

import (
	proto "areyouin/protocol"
	"log"
	"net"
	"runtime"
	"time"
)

func ping_task(conn net.Conn) {
	// Keep Active
	for {
		// Send data
		ping_msg := proto.NewMessage().Ping().Marshal()
		conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
		_, err := conn.Write(ping_msg)

		if err != nil {
			con_error := err.(net.Error)
			if con_error.Timeout() {
				log.Println("Coudn't send PING: ", con_error)
				continue
			}
			log.Fatal("Write error: ", err)
		}

		log.Println("PING Sent", ping_msg)
		time.Sleep(5 * time.Second)
	}
}

func main() {
	log.Println("GOMAXPROCS is", runtime.GOMAXPROCS(0))

	// Open connection
	conn, err := net.Dial("tcp", "localhost:1822")

	if err != nil {
		return
	}

	go ping_task(conn)
	time.Sleep(120 * time.Second)
}
