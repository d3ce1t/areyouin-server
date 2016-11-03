package cqldao

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gocql/gocql"
)

const (
	CassandraHost     = "192.168.1.10"
	CassandraKeyspace = "areyouin_test"
	TestTableName     = "gocql_timestamp_test"
)

var (
	insert          = fmt.Sprintf("INSERT INTO %v (id, a, b) VALUES (?, ?, ?)", TestTableName)
	insertUsing     = fmt.Sprintf("INSERT INTO %v (id, a, b) VALUES (?, ?, ?) USING TIMESTAMP ?", TestTableName)
	insertColAUsing = fmt.Sprintf("INSERT INTO %v (id, a) VALUES (?, ?) USING TIMESTAMP ?", TestTableName)
	insertColBUsing = fmt.Sprintf("INSERT INTO %v (id, b) VALUES (?, ?) USING TIMESTAMP ?", TestTableName)
)

type checkValue struct {
	id  int64
	tsa int64
	tsb int64
}

func connectToCassandra() (*gocql.Session, error) {
	cluster := gocql.NewCluster(CassandraHost)
	cluster.Keyspace = CassandraKeyspace
	cluster.Consistency = gocql.LocalQuorum
	cluster.ProtoVersion = 3
	cluster.DefaultTimestamp = true // it is already true by default
	return cluster.CreateSession()
}

func createDatabase(session *gocql.Session) error {

	removeTableStmt := fmt.Sprintf("DROP TABLE IF EXISTS %v", TestTableName)

	createTableStmt := fmt.Sprintf(`CREATE TABLE %v (
		id bigint,
		a text,
		b text,
		PRIMARY KEY (id)
	)
	WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'}`,
		TestTableName)

	// Drop table
	if err := session.Query(removeTableStmt).Exec(); err != nil {
		return err
	}

	// Create
	if err := session.Query(createTableStmt).Exec(); err != nil {
		return err
	}

	return nil
}

func checkMultipleRowTS(session *gocql.Session, checkValues ...checkValue) error {
	for _, check := range checkValues {
		if err := checkRowTS(session, check); err != nil {
			return err
		}
	}
	return nil
}

func checkRowTS(session *gocql.Session, check checkValue) error {

	selectStmt := fmt.Sprintf(`select writetime(a) as ta, writetime(b) as tb
        from %v where id = ?`, TestTableName)

	var tsa int64
	var tsb int64

	q := session.Query(selectStmt, check.id)
	if err := q.Scan(&tsa, &tsb); err != nil {
		return err
	}

	if tsa != check.tsa || tsb != check.tsb {
		return errors.New("Read timestamp doesn't match written timestamp")
	}

	return nil
}

func TestGOCQL_CreateDatabase(t *testing.T) {
	// Connect to Cassandra
	session, err := connectToCassandra()
	if err != nil {
		t.Fatal(err)
	}

	// Create database
	if err := createDatabase(session); err != nil {
		t.Fatal(err)
	}
}

func TestGOCQL_Query_WithTimestamp(t *testing.T) {

	// Connect to Cassandra
	session, err := connectToCassandra()
	if err != nil {
		t.Fatal(err)
	}

	// Execute test
	ts := time.Now().UnixNano() / 1000
	query := session.Query(insert, 1, "foo", "bar").WithTimestamp(ts)
	if err := query.Exec(); err != nil {
		t.Fatal(err)
	}

	// Check
	check := checkValue{id: 1, tsa: ts, tsb: ts}
	if err := checkRowTS(session, check); err != nil {
		t.Fatal(err)
	}
}

func TestGOCQL_Query_UsingTimestamp(t *testing.T) {

	// Connect to Cassandra
	session, err := connectToCassandra()
	if err != nil {
		t.Fatal(err)
	}

	// Execute test
	ts := time.Now().UnixNano() / 1000
	query := session.Query(insertUsing, 2, "foo", "bar", ts)
	if err := query.Exec(); err != nil {
		t.Fatal(err)
	}

	// Check
	check := checkValue{id: 2, tsa: ts, tsb: ts}
	if err := checkRowTS(session, check); err != nil {
		t.Fatal(err)
	}
}

func TestGOCQL_Batch_Logged_WithTimestamp(t *testing.T) {

	// Connect to Cassandra
	session, err := connectToCassandra()
	if err != nil {
		t.Fatal(err)
	}

	// Execute test
	ts := time.Now().UnixNano() / 1000
	batch := session.NewBatch(gocql.LoggedBatch).WithTimestamp(ts)
	batch.Query(insert, 3, "foo", "bar")
	batch.Query(insert, 4, "foo", "bar")

	if err := session.ExecuteBatch(batch); err != nil {
		t.Fatal(err)
	}

	// Check
	check1 := checkValue{id: 3, tsa: ts, tsb: ts}
	check2 := checkValue{id: 4, tsa: ts, tsb: ts}
	if err := checkMultipleRowTS(session, check1, check2); err != nil {
		t.Fatal(err)
	}
}

func TestGOCQL_Batch_Unlogged_WithTimestamp(t *testing.T) {

	// Connect to Cassandra
	session, err := connectToCassandra()
	if err != nil {
		t.Fatal(err)
	}

	// Execute test
	ts := time.Now().UnixNano() / 1000
	batch := session.NewBatch(gocql.UnloggedBatch).WithTimestamp(ts)
	batch.Query(insert, 5, "foo", "bar")
	batch.Query(insert, 6, "foo", "bar")

	if err := session.ExecuteBatch(batch); err != nil {
		t.Fatal(err)
	}

	// Check
	check1 := checkValue{id: 5, tsa: ts, tsb: ts}
	check2 := checkValue{id: 6, tsa: ts, tsb: ts}
	if err := checkMultipleRowTS(session, check1, check2); err != nil {
		t.Fatal(err)
	}
}

func TestGOCQL_Batch_Logged_UsingTimestamp(t *testing.T) {

	// Connect to Cassandra
	session, err := connectToCassandra()
	if err != nil {
		t.Fatal(err)
	}

	// Execute test
	ts := time.Now().UnixNano() / 1000
	batch := session.NewBatch(gocql.LoggedBatch)
	batch.Query(insertUsing, 7, "foo", "bar", ts)
	batch.Query(insertUsing, 8, "foo", "bar", ts)

	if err := session.ExecuteBatch(batch); err != nil {
		t.Fatal(err)
	}

	// Check
	check1 := checkValue{id: 7, tsa: ts, tsb: ts}
	check2 := checkValue{id: 8, tsa: ts, tsb: ts}
	if err := checkMultipleRowTS(session, check1, check2); err != nil {
		t.Fatal(err)
	}
}

func TestGOCQL_Batch_Unlogged_UsingTimestamp(t *testing.T) {

	// Connect to Cassandra
	session, err := connectToCassandra()
	if err != nil {
		t.Fatal(err)
	}

	// Execute test
	ts := time.Now().UnixNano() / 1000
	batch := session.NewBatch(gocql.UnloggedBatch)
	batch.Query(insertUsing, 9, "foo", "bar", ts)
	batch.Query(insertUsing, 10, "foo", "bar", ts)

	if err := session.ExecuteBatch(batch); err != nil {
		t.Fatal(err)
	}

	// Check
	check1 := checkValue{id: 9, tsa: ts, tsb: ts}
	check2 := checkValue{id: 10, tsa: ts, tsb: ts}
	if err := checkMultipleRowTS(session, check1, check2); err != nil {
		t.Fatal(err)
	}
}

func TestGOCQL_Batch_Logged_UsingTimestamp_Columns(t *testing.T) {

	// Connect to Cassandra
	session, err := connectToCassandra()
	if err != nil {
		t.Fatal(err)
	}

	// Execute test
	ts1 := time.Now().UnixNano() / 1000
	ts2 := ts1 + 1000000
	batch := session.NewBatch(gocql.LoggedBatch)
	batch.Query(insertColAUsing, 11, "foo", ts1)
	batch.Query(insertColBUsing, 11, "bar", ts2)
	batch.Query(insertColAUsing, 12, "foo", ts2)
	batch.Query(insertColBUsing, 12, "bar", ts1)

	if err := session.ExecuteBatch(batch); err != nil {
		t.Fatal(err)
	}

	// Check
	check1 := checkValue{id: 11, tsa: ts1, tsb: ts2}
	check2 := checkValue{id: 12, tsa: ts2, tsb: ts1}
	if err := checkMultipleRowTS(session, check1, check2); err != nil {
		t.Fatal(err)
	}
}
