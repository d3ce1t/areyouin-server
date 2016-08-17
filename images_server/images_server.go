package images_server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"peeple/areyouin/api"
	"peeple/areyouin/cqldao"
	"peeple/areyouin/model"
	"peeple/areyouin/utils"
	"strconv"
)

var (
	ErrInvalidRequest = errors.New("invalid request")
)

func NewServer(session api.DbSession, model *model.AyiModel, config api.Config) *ImageServer {
	server := &ImageServer{
		DbSession: session,
		Model:     model,
		Config:    config,
	}
	server.init()
	return server
}

type ImageServer struct {
	DbSession api.DbSession
	Model     *model.AyiModel
	Config    api.Config
}

func (s *ImageServer) init() {
	// Initiliase whatever is needed
}

func (s *ImageServer) loadThumbnail(id int64, reqDpi int32) ([]byte, error) {
	thumbnail_dao := cqldao.NewThumbnailDAO(s.DbSession)
	dpi := s.Model.GetClosestDpi(reqDpi)
	return thumbnail_dao.Load(id, dpi)
}

func (s *ImageServer) loadEventImage(id int64) (*api.PictureDTO, error) {
	event_dao := cqldao.NewEventDAO(s.DbSession)
	return event_dao.LoadEventPicture(id)
}

func (s *ImageServer) loadUserImage(id int64) (*api.PictureDTO, error) {
	user_dao := cqldao.NewUserDAO(s.DbSession)
	return user_dao.LoadProfilePicture(id)
}

// Check access and returns user_id if access is granted or
// 0 otherwise.
func (s *ImageServer) checkAccess(header http.Header) (int64, error) {

	user_id_str := header.Get("userid")
	token := header.Get("token")

	if user_id_str == "" || token == "" {
		return 0, nil
	}

	user_id, err := strconv.ParseInt(user_id_str, 10, 64)
	if err != nil {
		return 0, err
	}

	access_dao := cqldao.NewAccessTokenDAO(s.DbSession)

	accessToken, err := access_dao.Load(user_id)
	if err != nil {
		return 0, err
	}

	if accessToken.Token != token {
		return 0, nil
	}

	access_dao.SetLastUsed(user_id, utils.GetCurrentTimeMillis()) // ignore possible errors
	return user_id, nil
}

func (s *ImageServer) parseImageParams(id *int64, values url.Values) error {

	id_str := values.Get("id")

	if id_str == "" {
		return ErrInvalidRequest
	}

	var err error
	*id, err = strconv.ParseInt(id_str, 10, 64)
	if err != nil {
		return err
	}

	return nil
}

func (s *ImageServer) parseThumbnailsParams(thumbnail_id *int64, dpi *int32, values url.Values) error {

	thumbnail_id_str := values.Get("thumbnail")
	dpi_str := values.Get("dpi")

	if thumbnail_id_str == "" {
		return ErrInvalidRequest
	}

	var err error

	*thumbnail_id, err = strconv.ParseInt(thumbnail_id_str, 10, 64)
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

	var user_id int64

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

	var thumbnail_id int64
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

	var user_id int64

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

	var image_id int64

	err = s.parseImageParams(&image_id, r.URL.Query())
	if err != nil {
		http.Error(w, "Invalid request received", http.StatusBadRequest)
		log.Printf("< (%v) GET EVENT IMAGE ERROR: %v\n", user_id, err)
		return
	}

	// Everything OK

	image, err := s.loadEventImage(image_id)
	manageError(err)

	n, err := w.Write(image.RawData)
	manageError(err)
	log.Printf("< (%v) SEND EVENT IMAGE (%v/%v bytes)\n", user_id, n, len(image.RawData))
}

func (s *ImageServer) handleUserImageRequest(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "Invalid request received", http.StatusBadRequest)
		log.Printf("< (?) GET USER IMAGE ERROR: Invalid Request\n")
		return
	}

	var user_id int64

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

	var image_id int64

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
	image, err := s.loadUserImage(image_id)
	manageError(err)

	n, err := w.Write(image.RawData)
	manageError(err)
	log.Printf("< (%v) SEND USER IMAGE (%v/%v bytes)\n", user_id, n, len(image.RawData))
}

func manageError(err error) {
	if err != nil {
		panic(err)
	}
}

func (s *ImageServer) Run() {

	http.HandleFunc("/api/", s.handleThumbnailRequest)
	http.HandleFunc("/api/img/thumbnail/", s.handleThumbnailRequest)
	http.HandleFunc("/api/img/original/", s.handleEventImageRequest)
	http.HandleFunc("/api/img/original/event/", s.handleEventImageRequest)
	http.HandleFunc("/api/img/original/user/", s.handleUserImageRequest)

	addr := fmt.Sprintf("%v:%v", s.Config.ListenAddress(), s.Config.ImageListenPort())

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
