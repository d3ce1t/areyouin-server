package dao

import (
	"flag"
	"github.com/gocql/gocql"
	"log"
	"os"
	core "peeple/areyouin/common"
	"testing"
)

var session *gocql.Session
var idgen *core.IDGen
var participants100 []*core.EventParticipant
var eventIds10000 []uint64
var eventIds10000unsorted []uint64

func TestMain(m *testing.M) {

	idgen = core.NewIDGen(1)

	// Connect to Cassandra
	cluster := gocql.NewCluster("192.168.1.10", "192.168.1.11" /*"192.168.1.3"*/)
	cluster.Keyspace = "areyouin"
	cluster.Consistency = gocql.Quorum

	s, err := cluster.CreateSession()

	if err != nil {
		log.Println("Error connection to cassandra", err)
		return
	}

	session = s

	// Create 100 randome participants
	participants100 = make([]*core.EventParticipant, 0, 100)

	for i := 0; i < 100; i++ {
		participants100 = append(participants100, CreateRandomParticipant())
	}

	// Load 1000 events Ids
	eventIds10000 = GetEventIDs(10000)

	// Unsort
	eventIds10000unsorted = make([]uint64, len(eventIds10000))

	initial_offset := 0
	final_offset := len(eventIds10000) - 1

	for i := 0; i < len(eventIds10000); i += 2 {
		eventIds10000unsorted[initial_offset] = eventIds10000[i]
		eventIds10000unsorted[final_offset] = eventIds10000[i+1]
		initial_offset++
		final_offset--
	}

	//core.Create100Users(NewUserDAO(s))
	//core.Delete100Users(NewUserDAO(s))

	flag.Parse()
	os.Exit(m.Run())
}

func CreateParticipantsList(author_id uint64, participants_id []uint64) []*core.EventParticipant {

	result := make([]*core.EventParticipant, 0, len(participants_id))

	dao := NewUserDAO(session)

	for _, user_id := range participants_id {
		if ok, _ := dao.AreFriends(author_id, user_id); ok {
			if uac, _ := dao.Load(user_id); uac != nil {
				result = append(result, uac.AsParticipant())
			} else {
				log.Println("createParticipantList() participant", user_id, "does not exist")
			}
		} else {
			log.Println("createParticipantList() Not friends", author_id, "and", user_id, "or doesn't exist")
		}
	}

	return result
}

func CreateRandomParticipant() *core.EventParticipant {

	participant := &core.EventParticipant{
		UserId:    idgen.GenerateID(),
		Name:      "Prueba",
		Response:  core.AttendanceResponse_NO_RESPONSE,
		Delivered: core.MessageStatus_NO_DELIVERED,
	}

	return participant
}

func GetEventIDs(limit int) []uint64 {

	stmt := `SELECT DISTINCT event_id FROM event LIMIT ?`

	iter := session.Query(stmt, limit).Iter()

	list_ids := make([]uint64, 0, limit)
	var event_id uint64

	for iter.Scan(&event_id) {
		list_ids = append(list_ids, event_id)
	}

	if err := iter.Close(); err != nil {
		log.Println("GetEventIDS Error:", err)
		return nil
	}

	return list_ids
}
