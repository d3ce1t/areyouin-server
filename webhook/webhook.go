package webhook

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"peeple/areyouin/api"
)

const VERIFY_TOKEN = "HK377WB7LPBWN87HMZ7X"

type FacebookUpdate struct {
	Object  string
	Entries []struct {
		Uid           string
		Id            string
		Time          uint64
		ChangedFields []string `json:"changed_fields"`
	} `json:"entry"`
}

type WebHookServer struct {
	callback   func(*FacebookUpdate)
	app_secret string
	config     api.Config
}

func New(secret string, config api.Config) *WebHookServer {
	return &WebHookServer{app_secret: secret, config: config}
}

func computeSignature(payload []byte, key string) string {
	key_bytes := []byte(key)
	mac := hmac.New(sha1.New, key_bytes)
	mac.Write(payload)
	return "sha1=" + fmt.Sprintf("%x", mac.Sum(nil))
}

func (wh *WebHookServer) RegisterCallback(f func(*FacebookUpdate)) {
	wh.callback = f
}

func (wh *WebHookServer) checkRequest(values url.Values) bool {
	_, ok1 := values["hub.mode"]
	_, ok2 := values["hub.challenge"]
	_, ok3 := values["hub.verify_token"]
	return ok1 && ok2 && ok3
}

func (wh *WebHookServer) verifyRequest(w http.ResponseWriter, r *http.Request) {

	values := r.URL.Query()

	if !wh.checkRequest(values) {
		log.Println("Webhook: Facebook Invalid request received")
		http.Error(w, "Invalid request received", http.StatusBadRequest)
		return
	}

	log.Println("Webhook: Facebook Verify Endpoint Request received")
	token := values["hub.verify_token"][0]
	mode := values["hub.mode"][0]

	if token != VERIFY_TOKEN {
		log.Println("Webhook: Token mismatch")
		http.Error(w, "Invalid request received", http.StatusBadRequest)
		return
	}

	if mode != "subscribe" {
		log.Println("Webhook: Invalide mode")
		http.Error(w, "Invalid request received", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "%s", values["hub.challenge"][0])
	log.Println("Webhook: Endpoint Verified")
}

func (wh *WebHookServer) handler(w http.ResponseWriter, r *http.Request) {

	// Verify Endpoint
	if r.Method == "GET" {
		wh.verifyRequest(w, r)
		return
	}

	if r.Method != "POST" {
		log.Println("Facebook Invalid request received")
		//http.Error(w, "Invalid request received", http.StatusBadRequest)
		return
	}

	// Check request errors
	remote_host := r.Host
	if val := r.Header.Get("X-Real-Ip"); val != "" {
		remote_host = val
	}

	x_hub_signature := r.Header.Get("X-Hub-Signature")
	if x_hub_signature == "" {
		log.Println("Facebook Invalid request received")
		//http.Error(w, "Invalid request received", http.StatusBadRequest)
	}

	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Println("Facebook Invalid request received from", remote_host)
		//http.Error(w, "Invalid request received", http.StatusBadRequest)
	}

	computed_signature := computeSignature(data, wh.app_secret)
	if x_hub_signature != computed_signature {
		log.Println("Webhook: Signature mismatch")
		return
	}

	// Manage update
	log.Printf("Webhook: Received Update from %v: %s\n", remote_host, data)

	// Decode JSON message
	v := &FacebookUpdate{}

	if err := json.Unmarshal(data, v); err != nil {
		log.Println("Webhook error:", err)
		return
	}

	wh.callback(v)
}

func (wh *WebHookServer) Run() {

	go func() {

		http.HandleFunc("/fbwebhook/", wh.handler)

		//err := http.ListenAndServeTLS(":443", "cert.pem", "key.pem", nil)
		addr := fmt.Sprintf("%v:%v", wh.config.ListenAddress(), wh.config.FBWebHookListenPort())
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()
}
