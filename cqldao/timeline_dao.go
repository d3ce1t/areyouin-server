package cqldao

import (
	"peeple/areyouin/api"
	"peeple/areyouin/utils"
	"sort"

	"github.com/gocql/gocql"
)

const (
	timeLinePartitionLimit = 10000000
)

// Beginning of time line has the event with earliest end date
type EventTimeLineDAO struct {
	session     *GocqlSession
	bucketIndex []int // This memory should be distributed if more than one node is gonna be user
}

func NewTimeLineDAO(session api.DbSession) api.EventTimeLineDAO {
	reconnectIfNeeded(session)
	dao := &EventTimeLineDAO{session: session.(*GocqlSession)}
	dao.loadBucketIndex()
	return dao
}

// Read events in timeline from date. Returned items are in range [date, infinite)
// and it includes events that have not finished yet (not started or in progress).
func (d *EventTimeLineDAO) FindAllFrom(date int64) ([]*api.TimeLineEntryDTO, error) {

	checkSession(d.session)

	stmt := `SELECT event_id, position
		FROM events_timeline
		WHERE bucket = ? and position >= ? LIMIT ?`

	fromDate := utils.MillisToTimeUTC(date)

	// Compute buckets to read
	var exploreBuckets []int
	_, pos := d.hasBucket(fromDate.Year())

	if pos < len(d.bucketIndex) {
		exploreBuckets = d.bucketIndex[pos:]
	}

	// Reserve initial memory to load results
	var err error
	results := make([]*api.TimeLineEntryDTO, 0, 1000)

	for _, bucket := range exploreBuckets {

		// Read buckets

		q := d.session.Query(stmt, bucket, date, timeLinePartitionLimit)
		if results, err = d.findAllAux(q, results); err != nil {
			return nil, err
		}
	}

	return results, nil
}

func (d *EventTimeLineDAO) Insert(item *api.TimeLineEntryDTO) error {

	checkSession(d.session)

	bucket := item.Position.Year()

	stmt := `INSERT INTO events_timeline (bucket, event_id, position) VALUES (?, ?, ?)`

	q := d.session.Query(stmt, bucket, item.EventID, utils.TimeToMillis(item.Position))

	if err := convErr(q.Exec()); err != nil {
		return err
	}

	d.updateBucketIndex(bucket)

	return nil
}

func (d *EventTimeLineDAO) Delete(item *api.TimeLineEntryDTO) error {

	checkSession(d.session)
	stmt := `DELETE FROM events_timeline WHERE bucket = ? AND position = ? AND event_id = ?`
	q := d.session.Query(stmt, item.Position.Year(), utils.TimeToMillis(item.Position), item.EventID)
	return convErr(q.Exec())
}

func (d *EventTimeLineDAO) DeleteAll() error {
	checkSession(d.session)
	return d.session.Query(`TRUNCATE events_timeline`).Exec()
}

func (d *EventTimeLineDAO) Replace(oldItem *api.TimeLineEntryDTO, newItem *api.TimeLineEntryDTO) error {

	checkSession(d.session)

	if oldItem.EventID == newItem.EventID && oldItem.Position == newItem.Position {
		return nil
	}

	deleteStmt := `DELETE FROM events_timeline WHERE bucket = ? AND position = ? AND event_id = ?`
	insertStmt := `INSERT INTO events_timeline (bucket, event_id, position) VALUES (?, ?, ?)`

	var batchType gocql.BatchType

	if oldItem.Position.Year() == newItem.Position.Year() {
		batchType = gocql.UnloggedBatch // updates to the same partition key
	} else {
		batchType = gocql.LoggedBatch
	}

	batch := d.session.NewBatch(batchType)
	batch.Query(deleteStmt, oldItem.Position.Year(), utils.TimeToMillis(oldItem.Position), oldItem.EventID)
	batch.Query(insertStmt, newItem.Position.Year(), newItem.EventID, utils.TimeToMillis(newItem.Position))

	if err := d.session.ExecuteBatch(batch); err != nil {
		return convErr(err)
	}

	d.updateBucketIndex(newItem.Position.Year())

	return nil
}

func (d *EventTimeLineDAO) loadBucketIndex() error {

	checkSession(d.session)

	stmt := `SELECT DISTINCT bucket FROM events_timeline`
	iter := d.session.Query(stmt).Iter()

	results := make([]int, 0, 10)

	var bucket int

	for iter.Scan(&bucket) {
		results = append(results, bucket)
	}

	if err := iter.Close(); err != nil {
		return convErr(err)
	}

	// sort slide from lower to higher
	sort.Ints(results)
	d.bucketIndex = results

	return nil
}

func (d *EventTimeLineDAO) updateBucketIndex(bucket int) {
	if exist, pos := d.hasBucket(bucket); !exist {
		// Insert new bucket into index and keep order
		d.bucketIndex = append(d.bucketIndex, 0)
		copy(d.bucketIndex[pos+1:], d.bucketIndex[pos:])
		d.bucketIndex[pos] = bucket
	}
}

func (d *EventTimeLineDAO) hasBucket(bucket int) (bool, int) {

	pos := sort.SearchInts(d.bucketIndex, bucket)

	if pos == len(d.bucketIndex) || d.bucketIndex[pos] != bucket {
		return false, pos
	}

	return true, pos
}

func (d *EventTimeLineDAO) findAllAux(query *gocql.Query, results []*api.TimeLineEntryDTO) ([]*api.TimeLineEntryDTO, error) {

	checkSession(d.session)

	iter := query.Iter()

	var eventID int64
	var position int64

	// Assume LIMIT is NOT reached so all items of this partition should
	// have been read
	for iter.Scan(&eventID, &position) {
		results = append(results, &api.TimeLineEntryDTO{
			EventID:  eventID,
			Position: utils.MillisToTimeUTC(position),
		})
	}

	if err := iter.Close(); err != nil {
		return nil, convErr(err)
	}

	return results, nil
}
