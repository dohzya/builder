package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
)

type ExternalMessage struct {
	Id string `json:"id"`
}

type ExternalResponse struct {
	Id string `json:"id"`
}

type InternalMessage struct {
	Id  string
	Out chan InternalResponse
}

type InternalResponse struct {
	Id string
}

func mainLoop(in chan InternalMessage) {
	for {
		msg := <-in
		resp := InternalResponse{}

		// TODO actually do stuff :-)
		resp.Id = msg.Id

		msg.Out <- resp
	}
}

type RequestHandler struct {
	In chan InternalMessage
}

func (h *RequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var extMsg ExternalMessage
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&extMsg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"error\":\"Can't parse body\"}\n")
		return
	}

	intMsg := InternalMessage{extMsg.Id, make(chan InternalResponse)}

	h.In <- intMsg

	intResp := <-intMsg.Out
	extResp := ExternalResponse{intResp.Id}

	enc := json.NewEncoder(w)
	enc.Encode(extResp)
}

func main() {
	httpHost := flag.String("h", "127.0.0.1", "The host to bind")
	httpPort := flag.Int("p", 8080, "The port to listen")
	flag.Parse()

	c := make(chan InternalMessage)

	go mainLoop(c)

	http.Handle("/", &RequestHandler{c})

	hostPort := fmt.Sprintf("%s:%d", *httpHost, *httpPort)
	log.Fatal(http.ListenAndServe(hostPort, nil))
}
