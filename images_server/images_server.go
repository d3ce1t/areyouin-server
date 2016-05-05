package images_server

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	core "peeple/areyouin/common"
	"peeple/areyouin/dao"
	"strconv"

	"github.com/gocql/gocql"
)

const (
	IMAGE_MDPI    = 160              // 160dpi
	IMAGE_HDPI    = 1.5 * IMAGE_MDPI // 240dpi
	IMAGE_XHDPI   = 2 * IMAGE_MDPI   // 320dpi
	IMAGE_XXHDPI  = 3 * IMAGE_MDPI   // 480dpi
	IMAGE_XXXHDPI = 4 * IMAGE_MDPI   // 640dpi
)

var (
	ErrInvalidRequest = errors.New("invalid request")
)

func NewServer(db_address string, keyspace string) *ImageServer {
	server := &ImageServer{
		listen_port: 40187,
		db_address:  db_address,
		keyspace:    keyspace,
	}
	server.init()
	return server
}

type ImageServer struct {
	listen_port  int
	cluster      *gocql.ClusterConfig
	db_session   *gocql.Session
	keyspace     string
	db_address   string
	supportedDpi []int32
}

func (s *ImageServer) init() {
	// Cassandra
	s.cluster = gocql.NewCluster(s.db_address)
	s.cluster.Keyspace = s.keyspace
	s.cluster.Consistency = gocql.LocalQuorum

	// Supported Screen densities
	s.supportedDpi = []int32{IMAGE_MDPI, IMAGE_HDPI, IMAGE_XHDPI,
		IMAGE_XXHDPI, IMAGE_XXXHDPI}
}

func (s *ImageServer) getClosestDpi(reqDpi int32) int32 {

	if reqDpi <= IMAGE_MDPI {
		return IMAGE_MDPI
	} else if reqDpi >= IMAGE_XXXHDPI {
		return IMAGE_XXXHDPI
	}

	min_dist := math.MaxFloat32
	dpi_index := 0

	for i, dpi := range s.supportedDpi {
		dist := math.Abs(float64(reqDpi - dpi))
		if dist < min_dist {
			min_dist = dist
			dpi_index = i
		}
	}

	if s.supportedDpi[dpi_index] < reqDpi {
		dpi_index++
	}

	return s.supportedDpi[dpi_index]
}

func (s *ImageServer) loadThumbnail(id uint64, reqDpi int32) ([]byte, error) {
	thumbnail_dao := dao.NewThumbnailDAO(s.db_session)
	dpi := s.getClosestDpi(reqDpi)
	return thumbnail_dao.Load(id, dpi)
}

func (s *ImageServer) loadEventImage(id uint64) ([]byte, error) {
	event_dao := dao.NewEventDAO(s.db_session)
	return event_dao.LoadEventPicture(id)
}

func (s *ImageServer) loadUserImage(id uint64) ([]byte, error) {
	user_dao := dao.NewUserDAO(s.db_session)
	return user_dao.LoadUserPicture(id)
}

// Check access and returns user_id if access is granted or
// 0 otherwise.
func (s *ImageServer) checkAccess(header http.Header) (uint64, error) {

	user_id_str := header.Get("userid")
	token := header.Get("token")

	if user_id_str == "" || token == "" {
		return 0, nil
	}

	user_id, err := strconv.ParseUint(user_id_str, 10, 64)
	if err != nil {
		return 0, err
	}

	access_dao := dao.NewAccessTokenDAO(s.db_session)
	ok, err := access_dao.CheckAccessToken(user_id, token)
	if err != nil || !ok {
		return 0, err
	}

	access_dao.SetLastUsed(user_id, core.GetCurrentTimeMillis()) // ignore possible errors
	return user_id, nil
}

func (s *ImageServer) parseImageParams(id *uint64, values url.Values) error {

	id_str := values.Get("id")

	if id_str == "" {
		return ErrInvalidRequest
	}

	var err error
	*id, err = strconv.ParseUint(id_str, 10, 64)
	if err != nil {
		return err
	}

	return nil
}

func (s *ImageServer) parseThumbnailsParams(thumbnail_id *uint64, dpi *int32, values url.Values) error {

	thumbnail_id_str := values.Get("thumbnail")
	dpi_str := values.Get("dpi")

	if thumbnail_id_str == "" {
		return ErrInvalidRequest
	}

	var err error

	*thumbnail_id, err = strconv.ParseUint(thumbnail_id_str, 10, 64)
	if err != nil {
		return err
	}

	var dpi64 int64

	if dpi_str != "" {
		dpi64, err = strconv.ParseInt(dpi_str, 10, 32)
		if err != nil {
			return err
		}
	} else {
		dpi64 = 160
	}

	*dpi = int32(dpi64)

	return nil
}

func (s *ImageServer) handleThumbnailRequest(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "Invalid request received", http.StatusBadRequest)
		log.Printf("< (?) GET THUMBNAIL ERROR: Invalid Request\n")
		return
	}

	var user_id uint64

	defer func() {
		r := recover()
		if r != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Printf("< (%v) GET THUMBNAIL ERROR: %v\n", user_id, r)
		}
	}()

	log.Printf("> (%v) GET THUMBNAIL (ID: %v, ScreenDensity: %v)\n",
		r.Header.Get("userid"), r.URL.Query().Get("thumbnail"), r.URL.Query().Get("dpi"))

	user_id, err := s.checkAccess(r.Header)
	manageError(err)
	if user_id == 0 {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		log.Printf("< (%v) GET THUMBNAIL ERROR: ACCESS DENIED", r.Header.Get("userid"))
		return
	}

	var thumbnail_id uint64
	var dpi int32

	err = s.parseThumbnailsParams(&thumbnail_id, &dpi, r.URL.Query())
	if err != nil {
		http.Error(w, "Invalid request received", http.StatusBadRequest)
		log.Printf("< (%v) GET THUMBNAIL ERROR: %v\n", user_id, err)
		return
	}

	// Everything OK

	data, err := s.loadThumbnail(thumbnail_id, int32(dpi))
	manageError(err)

	n, err := w.Write(data)
	manageError(err)
	log.Printf("< (%v) SEND THUMBNAIL (%v/%v bytes, %v dpi)\n", user_id, n, len(data), dpi)
}

func (s *ImageServer) handleEventImageRequest(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "Invalid request received", http.StatusBadRequest)
		log.Printf("< (?) GET EVENT IMAGE ERROR: Invalid Request\n")
		return
	}

	var user_id uint64

	defer func() {
		r := recover()
		if r != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Printf("< (%v) GET EVENT IMAGE ERROR: %v\n", user_id, r)
		}
	}()

	log.Printf("> (%v) GET EVENT IMAGE (ID: %v)\n", r.Header.Get("userid"), r.URL.Query().Get("id"))

	user_id, err := s.checkAccess(r.Header)
	manageError(err)
	if user_id == 0 {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		log.Printf("< (%v) GET EVENT IMAGE ERROR: ACCESS DENIED", r.Header.Get("userid"))
		return
	}

	var image_id uint64

	err = s.parseImageParams(&image_id, r.URL.Query())
	if err != nil {
		http.Error(w, "Invalid request received", http.StatusBadRequest)
		log.Printf("< (%v) GET EVENT IMAGE ERROR: %v\n", user_id, err)
		return
	}

	// Everything OK

	data, err := s.loadEventImage(image_id)
	manageError(err)

	n, err := w.Write(data)
	manageError(err)
	log.Printf("< (%v) SEND EVENT IMAGE (%v/%v bytes)\n", user_id, n, len(data))

}

func (s *ImageServer) handleUserImageRequest(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "Invalid request received", http.StatusBadRequest)
		log.Printf("< (?) GET USER IMAGE ERROR: Invalid Request\n")
		return
	}

	var user_id uint64

	defer func() {
		r := recover()
		if r != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Printf("< (%v) GET USER IMAGE ERROR: %v\n", user_id, r)
		}
	}()

	log.Printf("> (%v) GET USER IMAGE (ID: %v)\n", r.Header.Get("userid"), r.URL.Query().Get("id"))

	user_id, err := s.checkAccess(r.Header)
	manageError(err)
	if user_id == 0 {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		log.Printf("< (%v) GET USER IMAGE ERROR: ACCESS DENIED", r.Header.Get("userid"))
		return
	}

	var image_id uint64

	err = s.parseImageParams(&image_id, r.URL.Query())
	if err != nil {
		http.Error(w, "Invalid request received", http.StatusBadRequest)
		log.Printf("< (%v) GET USER IMAGE ERROR: %v\n", user_id, err)
		return
	}

	// Only allow retrieve of own profile image
	if user_id != image_id {
		http.Error(w, "Access Forbidden", http.StatusUnauthorized)
		log.Printf("< (%v) GET USER IMAGE ERROR: %v\n", user_id, "Access Forbidden")
		return
	}

	// Everything OK
	data, err := s.loadUserImage(image_id)
	manageError(err)

	n, err := w.Write(data)
	manageError(err)
	log.Printf("< (%v) SEND USER IMAGE (%v/%v bytes)\n", user_id, n, len(data))
}

func (s *ImageServer) connectToDB() {
	if session, err := s.cluster.CreateSession(); err == nil {
		s.db_session = session
	} else {
		log.Println("Error connecting to cassandra:", err)
	}
}

func manageError(err error) {
	if err != nil {
		panic(err)
	}
}

func (s *ImageServer) Run() {

	// Connect to Cassandra
	s.connectToDB()

	go func() {
		http.HandleFunc("/api/", s.handleThumbnailRequest)
		http.HandleFunc("/api/img/thumbnail/", s.handleThumbnailRequest)
		http.HandleFunc("/api/img/original/", s.handleEventImageRequest)
		http.HandleFunc("/api/img/original/event/", s.handleEventImageRequest)
		http.HandleFunc("/api/img/original/user/", s.handleUserImageRequest)
		err := http.ListenAndServe(fmt.Sprintf(":%v", s.listen_port), nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()
}
