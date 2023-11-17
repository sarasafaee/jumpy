package http

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/libp2p/go-libp2p/core/host"
	"log"
	"myBlockchain/chain"
	"net/http"
	"strconv"
	"time"
)

type HttpServer struct {
	Host host.Host
	RW   *bufio.ReadWriter
}

func (h HttpServer) RunHttpServer(port int) error {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/blocks", h.WriteBlock).Methods("POST")

	log.Println("Listening on ", port)
	s := &http.Server{
		Addr:           ":" + strconv.Itoa(port),
		Handler:        muxRouter,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func (h HttpServer) WriteBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var transaction chain.Transaction

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&transaction); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()

	peers := h.Host.Peerstore().Peers()
	for _, p := range peers {
		fmt.Println(p.String())
	}

	_, err := h.RW.WriteString(fmt.Sprintf("%s\n", chain.PULL_BLOCK))
	if err != nil {
		respondWithJSON(w, r, http.StatusNotFound, err)
		return
	}
	if err := h.RW.Flush(); err != nil {
		respondWithJSON(w, r, http.StatusNotFound, err)
		return
	}

	/*
		randomPeerIndex := rand.Intn(peers.Len())
			targetPeerID := peers[randomPeerIndex]

			resp, err := h.Host.Peerstore().Get(targetPeerID, chain.PULL_BLOCK)
		if err != nil {
			respondWithJSON(w, r, http.StatusNotFound, err)
			return
		}

		targetBlock, ok := resp.(chain.Block)
		if !ok {
			err := errors.New("block assertion failed")
			respondWithJSON(w, r, http.StatusInternalServerError, err)
			return
		}

		oldBlock := chain.Blockchain[len(chain.Blockchain)-1]
		newBlock := chain.Node.GenerateBlock(oldBlock, targetBlock, transaction)
		chain.Blockchain = append(chain.Blockchain, newBlock)
	*/

	respondWithJSON(w, r, http.StatusCreated, nil)
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}
