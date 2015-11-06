package main

import (
	"fmt"
	//	"io"
	"github.com/gocql/gocql"
	"log"
	"net"
	"runtime"
	"time"
)

const max_messages = 60000

type StatMessage struct {
	opcode int
}

func handleConnection(client net.Conn, c chan StatMessage) {

	buffer := make([]byte, 1024) // 1 Kb

	checkErrors := func(err error) bool {
		if err == nil {
			return false
		}
		/*if err == io.EOF {
			c <- StatMessage{1} // Closed connection
		} else {
			c <- StatMessage{2} // Error!
		}*/
		return true
	}

	for {
		// Read request
		n, err := client.Read(buffer)
		if checkErrors(err) {
			break
		}
		// Echo request
		n, err = client.Write(buffer[:n])
		if checkErrors(err) {
			break
		}
	}

	client.Close()
}

func manage_stats(c chan StatMessage) {

	var connections, closed, errors uint32
	current_time := time.Now()

	for {
		// Receive value
		select {
		case m := <-c: // Compute stats
			if m.opcode == 0 { // New connection
				connections++
			} else if m.opcode == 1 { // Closed
				connections--
				closed++
			} else if m.opcode == 2 { // Error
				connections--
				errors++
			}
		default: // Show info each 2 seconds
			if time.Since(current_time).Seconds() > 2 {
				fmt.Println("connections", connections, "closed", closed, "errors", errors)
				current_time = time.Now()
			}
		}
	}
}

func cassandra_example() {

	// connect to the cluster
	cluster := gocql.NewCluster("192.168.1.2")
	cluster.Keyspace = "example"
	cluster.Consistency = gocql.Quorum
	session, _ := cluster.CreateSession()
	defer session.Close()

	// insert a tweet
	if err := session.Query(`INSERT INTO tweet (timeline, id, text) VALUES (?, ?, ?)`,
		"me", gocql.TimeUUID(), "hello world").Exec(); err != nil {
		log.Fatal(err)
	}

	var id gocql.UUID
	var text string

	/* Search for a specific set of records whose 'timeline' column matches
	 * the value 'me'. The secondary index that we created earlier will be
	 * used for optimizing the search */
	if err := session.Query(`SELECT id, text FROM tweet WHERE timeline = ? LIMIT 1`,
		"me").Consistency(gocql.One).Scan(&id, &text); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Tweet:", id, text)

	// list all tweets
	iter := session.Query(`SELECT id, text FROM tweet WHERE timeline = ?`, "me").Iter()
	for iter.Scan(&id, &text) {
		fmt.Println("Tweet:", id, text)
	}
	if err := iter.Close(); err != nil {
		log.Fatal(err)
	}
}

func main() {

	fmt.Println("GOMAXPROCS is", runtime.GOMAXPROCS(0))
	cassandra_example()

	// Connect to cassandra
	/*cluster := gocql.NewCluster("192.168.1.2") //, "192.168.1.2", "192.168.1.3")
	cluster.Keyspace = "peeple"
	session, err := cluster.CreateSession()

	if err != nil {
		log.Fatal(err)
	}*/

	// Setup routes

	// Start web server

	/*server, err := net.Listen("tcp", ":4000")

	if err != nil {
		panic("Couldn't start listening: " + err.Error())
	}

	c := make(chan StatMessage, max_messages)
	go manage_stats(c)

	for {
		client, err := server.Accept()

		if err != nil {
			fmt.Println("Couldn't accept:", err.Error())
			continue
		}

		c <- StatMessage{0} // New connection
		go handleConnection(client, c)
	}*/
}
