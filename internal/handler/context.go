package handler

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/dto"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/internal/app"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/utils"
)

const (
	VERSION = "1.0.18"
)

var tn = time.Now()

func GetState(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("GetState, REQUEST\n")

	from := time.Since(tn)

	str := fmt.Sprintf("%dd, %d:%d:%d", int(from.Hours())/24, int(from.Hours()), int(from.Minutes()), int(from.Seconds()))

	body := fmt.Sprintf(`{"name":"Pcsc Daemon - %s", "version":%q, "upTime":%q}`, utils.OSName, VERSION, str)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", body)
}

func ListAllReaders(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("ListAllReaders, REQUEST\n")

	readers, err := app.Instance().ListReaders()
	if err != nil {
		fmt.Println(err)
		resp := make([]*dto.Reader, 0)
		body, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%s", err)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", body)
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

	fmt.Printf("ListAllReaders, response: %s\n", body)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", body)
}

func ReaderInformation(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("ReaderInformation, REQUEST\n")
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

	fmt.Printf("CardInReader, REQUEST\n")

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

	card, err := app.Instance().VerifyCardInReader(nameReader)
	if err != nil {
		fmt.Printf("error VerifyCardInReader: %s\n", err)
		resp := dto.NewSmartCardStatus(nil, dto.NotPresent)
		body, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%s", err)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", body)
		return
	}

	state, err := card.Status()
	if err != nil {
		fmt.Printf("error CardInReader: %s\n", err)
	}

	atr, _ := card.Atr()

	reader := dto.NewSmartCardStatus(atr, dto.StatusCode(state))

	body, err := json.Marshal(reader)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
	}

	fmt.Printf("CardInReader, response: %s\n", body)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", body)
}

func SendAPUs(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("SendAPUs, REQUEST\n")

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
		Requests  []string `json:"requests"`
		SessionId string   `json:"sessionId"`
		// Is7816       bool     `json:"is7816"`
		CloseSession bool `json:"closeSession"`
	}{}

	if err := json.Unmarshal(body, &keys); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", err)
		return
	}

	fmt.Printf("reader: %s, request: %+v\n", nameReader, keys)

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

	debug := utils.Debug
	ch, err := app.Instance().SendAPUs(nameReader, keys.SessionId, keys.CloseSession, debug, apdus...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
		return
	}

	responses := make([]*dto.APDUResponse, 0)
	idx := 0
	for v := range ch {
		apdu := apdus[idx]
		response := dto.NewAPDUResponse(apdu, v)
		responses = append(responses, response)
		idx++
	}

	bodyResponse, err := json.Marshal(responses)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", bodyResponse)
}

func SendAPU(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("SendAPU, REQUEST\n")

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
		Request   string `json:"request"`
		SessionId string `json:"sessionId"`
		// Is7816       bool   `json:"is7816"`
		CloseSession bool `json:"closeSession"`
	}{}

	if err := json.Unmarshal(body, &keys); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", err)
		return
	}

	fmt.Printf("reader: %s, request: %+v\n", nameReader, keys)

	apdu, err := hex.DecodeString(keys.Request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", err)
		return
	}

	if len(keys.SessionId) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "len session error: len = %d", len(keys.SessionId))
		return
	}

	debug := utils.Debug
	ch, err := app.Instance().SendAPUs(nameReader, keys.SessionId, keys.CloseSession, debug, apdu)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
		return
	}

	var response *dto.APDUResponse
	idx := 0
	for v := range ch {
		apdu := apdu
		response = dto.NewAPDUResponse(apdu, v)
		idx++
	}

	bodyResponse, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", bodyResponse)
}
