package dao

import (
	"github.com/gocql/gocql"
	core "peeple/areyouin/common"
)

type ThumbnailDAO struct {
	session *gocql.Session
}

func NewThumbnailDAO(session *gocql.Session) core.ThumbnailDAO {
	return &ThumbnailDAO{session: session}
}

func (dao *ThumbnailDAO) Insert(id uint64, digest []byte, thumbnails map[int32][]byte) error {

	dao.checkSession()

	stmt := `INSERT INTO thumbnails (id, digest, dpi, thumbnail, created_date)
            VALUES (?, ?, ?, ?, ?)`

	created_date := core.GetCurrentTimeMillis()

	batch := dao.session.NewBatch(gocql.UnloggedBatch)

	for size, thumbnail := range thumbnails {
		batch.Query(stmt, id, digest, size, thumbnail, created_date)
	}

	return dao.session.ExecuteBatch(batch)
}

func (dao *ThumbnailDAO) Load(id uint64, dpi int32) ([]byte, error) {

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

func (dao *ThumbnailDAO) checkSession() {
	if dao.session == nil {
		panic(ErrNoSession)
	}
}
