package cqldao

import (
	"time"

	"github.com/gocql/gocql"
)

func NewSession(keyspace string, cqlVersion int, hosts ...string) *GocqlSession {
	session := &GocqlSession{}
	session.cluster = gocql.NewCluster(hosts...)
	session.cluster.Keyspace = keyspace
	session.cluster.Consistency = gocql.LocalQuorum
	session.cluster.Timeout = 3 * time.Second
	session.cluster.ProtoVersion = cqlVersion
	session.cluster.DefaultTimestamp = true // is true by default
	return session
}

type GocqlSession struct {
	*gocql.Session
	cluster *gocql.ClusterConfig
}

func (self *GocqlSession) Connect() error {
	if session, err := self.cluster.CreateSession(); err == nil {
		self.Session = session
		return nil
	} else {
		return err
	}
}

func (self *GocqlSession) IsValid() bool {
	return self.Session != nil
}
