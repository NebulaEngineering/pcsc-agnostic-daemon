package handler

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gitlab.com/nebulaeng/fleet/pcscrest/dto"
	"gitlab.com/nebulaeng/fleet/pcscrest/internal/app"
)

const (
	version = "1.0.0"
)

var tn = time.Now()

func GetState(w http.ResponseWriter, r *http.Request) {

	from := time.Since(tn)

	str := fmt.Sprintf("%dd, %d:%d:%d", int(from.Hours())/24, int(from.Hours()), int(from.Minutes()), int(from.Seconds()))

	body := fmt.Sprintf(`{"name":"Pcsc Daemon - %s", "version":%q, "upTime":%q}`, osName, version, str)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", body)
}

func ListAllReaders(w http.ResponseWriter, r *http.Request) {

	readers, err := app.Instance().ListReaders()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
		return
	}

	rds := make([]*dto.Reader, 0)
	for _, r := range readers {
		rd := dto.NewReader(r, r)
		rds = append(rds, rd)
	}

	body, err := json.Marshal(rds)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", body)
}

func ReaderInformation(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nameReader, err := app.Instance().ReaderInformation(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
		return
	}

	reader := dto.NewReader(nameReader, nameReader)

	body, err := json.Marshal(reader)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", body)
}

func CardInReader(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id, ok := vars["readerIdHex"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	card, err := app.Instance().VerifyCardInReader(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
		return
	}

	state, err := card.Status()
	if err != nil {
		log.Printf("errr CardInReader: %s", err)
	}

	atr, _ := card.Atr()

	reader := dto.NewSmartCardStatus(atr, dto.StatusCode(state))

	body, err := json.Marshal(reader)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", body)
}

func SendAPUs(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id, ok := vars["readerIdHex"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	idBytes, err := hex.DecodeString(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nameReader := fmt.Sprintf("%s", idBytes)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", err)
		return
	}
	defer r.Body.Close()

	keys := struct {
		Requests     []string `json:"requests"`
		SessionId    string   `json:"sessionId"`
		Is7816       bool     `json:"is7816"`
		CloseSession bool     `json:"closeSession"`
	}{}

	if err := json.Unmarshal(body, &keys); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", err)
		return
	}

	apdus := make([][]byte, 0)

	for _, request := range keys.Requests {
		apdu, err := hex.DecodeString(request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "%s", err)
			return
		}
		apdus = append(apdus, apdu)
	}

	if len(keys.SessionId) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "len session error: len = %d", len(keys.SessionId))
		return
	}

	ch, err := app.Instance().SendAPUs(nameReader, keys.SessionId, keys.CloseSession, apdus...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
	}

	responses := make([]*dto.APDUResponse, 0)
	idx := 0
	for v := range ch {
		apdu := apdus[idx]
		response := dto.NewAPDUResponse(apdu, v, keys.Is7816)
		responses = append(responses, response)
		idx++
	}

	bodyResponse, err := json.Marshal(responses)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", bodyResponse)
}
