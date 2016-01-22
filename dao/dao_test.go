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

func TestMain(m *testing.M) {

	idgen = core.NewIDGen(1)

	// Connect to Cassandra
	cluster := gocql.NewCluster("192.168.1.10" /*"192.168.1.3"*/)
	cluster.Keyspace = "areyouin"
	cluster.Consistency = gocql.Quorum

	s, err := cluster.CreateSession()

	if err != nil {
		log.Println("Error connection to cassandra", err)
		return
	}

	session = s
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
