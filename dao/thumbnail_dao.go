package dao

import (
	"github.com/gocql/gocql"
	"log"
	core "peeple/areyouin/common"
)

type ThumbnailDAO struct {
	session *gocql.Session
}

func NewThumbnailDAO(session *gocql.Session) core.ThumbnailDAO {
	return &ThumbnailDAO{session: session}
}

func (dao *ThumbnailDAO) insertOne(id int64, digest []byte, dpi int32, thumbnail []byte, timestamp int64) error {
	dao.checkSession()
	stmt := `INSERT INTO thumbnails (id, digest, dpi, thumbnail, created_date)
            VALUES (?, ?, ?, ?, ?)`
	q := dao.session.Query(stmt, id, digest, dpi, thumbnail, timestamp)
	return q.Exec()
}

// Insert thumbnails into database. A batch implementation was used before, however
// it reaches Cassandra's batch_size limit of 50 Kb (despite being inserted to the
// same partition key). So a one-by-one insert implementation is preferred. If
// one of the inserts fails, this implementation keeps the first error but continues
// inserting the remainder ones. All errors are logged.
func (dao *ThumbnailDAO) Insert(id int64, digest []byte, thumbnails map[int32][]byte) error {
	dao.checkSession()
	var first_error error
	created_date := core.GetCurrentTimeMillis()
	for size, thumbnail := range thumbnails {
		tmp_err := dao.insertOne(id, digest, size, thumbnail, created_date)
		if tmp_err != nil {
			if first_error == nil {
				first_error = tmp_err
			}
			log.Printf("ThumbnailDAO.Insert Error: %v\n", tmp_err)
		}
	}

	return first_error
}

func (dao *ThumbnailDAO) Load(id int64, dpi int32) ([]byte, error) {

	dao.checkSession()

	stmt := `SELECT thumbnail FROM thumbnails WHERE id = ? AND dpi = ?`
	q := dao.session.Query(stmt, id, dpi)

	var thumbnail []byte

	err := q.Scan(&thumbnail)

	if err != nil {
		return nil, err
	}

	return thumbnail, nil
}

func (dao *ThumbnailDAO) Remove(id int64) error {
	dao.checkSession()
	return dao.session.Query(`DELETE FROM thumbnails WHERE id = ?`, id).Exec()
}

func (dao *ThumbnailDAO) checkSession() {
	if dao.session == nil {
		panic(ErrNoSession)
	}
}
