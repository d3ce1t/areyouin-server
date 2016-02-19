package images_server

import (
	"fmt"
	"github.com/gocql/gocql"
	"log"
	"math"
	"net/http"
	"peeple/areyouin/dao"
	"strconv"
)

const (
	IMAGE_MDPI    = 160              // 160dpi
	IMAGE_HDPI    = 1.5 * IMAGE_MDPI // 240dpi
	IMAGE_XHDPI   = 2 * IMAGE_MDPI   // 320dpi
	IMAGE_XXHDPI  = 3 * IMAGE_MDPI   // 480dpi
	IMAGE_XXXHDPI = 4 * IMAGE_MDPI   // 640dpi
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

func (s *ImageServer) handler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "Invalid request received", http.StatusBadRequest)
		log.Printf("< (?) GET THUMBNAIL ERROR: Invalid Request\n")
		return
	}

	defer func() {
		r := recover()
		if r != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Printf("< (?) GET THUMBNAIL ERROR: %v\n", r)
		}
	}()

	values := r.URL.Query()
	user_id_str, ok1 := values["userid"]
	token, ok2 := values["token"]
	thumbnail_id_str, ok3 := values["thumbnail"]
	dpi_str, ok4 := values["dpi"]

	if !ok1 || !ok2 || !ok3 || len(user_id_str) == 0 || len(token) == 0 || len(thumbnail_id_str) == 0 {
		http.Error(w, "Invalid request received", http.StatusBadRequest)
		log.Printf("< (?) GET THUMBNAIL ERROR: Invalid Request\n")
		return
	}

	user_id, err := strconv.ParseUint(user_id_str[0], 10, 64)
	manageError(err)

	thumbnail_id, err := strconv.ParseUint(thumbnail_id_str[0], 10, 64)
	manageError(err)

	var dpi int64
	if ok4 {
		dpi, err = strconv.ParseInt(dpi_str[0], 10, 32)
		manageError(err)
	} else {
		dpi = 160
	}

	log.Printf("> (%v) GET THUMBNAIL (UserID: %v, ScreenDensity: %v)\n", user_id, thumbnail_id, dpi)

	user_dao := dao.NewUserDAO(s.db_session)
	ok, err := user_dao.CheckAuthToken(user_id, token[0])
	manageError(err)

	if !ok {
		http.Error(w, "Authentication required", http.StatusForbidden)
		log.Printf("< (%v) GET THUMBNAIL ERROR: ACCESS DENIED", user_id)
	}

	data, err := s.loadThumbnail(thumbnail_id, int32(dpi))
	manageError(err)

	n, err := w.Write(data)
	manageError(err)

	log.Printf("< (%v) SEND THUMBNAIL (%v/%v bytes)\n", user_id, n, len(data))
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

		http.HandleFunc("/", s.handler)

		err := http.ListenAndServe(fmt.Sprintf(":%v", s.listen_port), nil)

		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()
}
