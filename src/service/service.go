package service

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/andrecronje/lachesis/src/node"
	"github.com/sirupsen/logrus"
)

type Service struct {
	bindAddress string
	node        *node.Node
	graph       *node.Graph
	logger      *logrus.Logger
}

func NewService(bindAddress string, n *node.Node, logger *logrus.Logger) *Service {
	service := Service{
		bindAddress: bindAddress,
		node:        n,
		graph:       node.NewGraph(n),
		logger:      logger,
	}

	return &service
}

func (s *Service) Serve() {
	s.logger.WithField("bind_address", s.bindAddress).Debug("Service serving")
	mux := http.NewServeMux()
	mux.Handle("/stats", corsHandler(s.GetStats))
	mux.Handle("/participants/", corsHandler(s.GetParticipants))
	mux.Handle("/event/", corsHandler(s.GetEvent))
	mux.Handle("/lasteventfrom/", corsHandler(s.GetLastEventFrom))
	mux.Handle("/events/", corsHandler(s.GetKnownEvents))
	mux.Handle("/consensusevents/", corsHandler(s.GetConsensusEvents))
	mux.Handle("/round/", corsHandler(s.GetRound))
	mux.Handle("/lastround/", corsHandler(s.GetLastRound))
	mux.Handle("/roundwitnesses/", corsHandler(s.GetRoundWitnesses))
	mux.Handle("/roundevents/", corsHandler(s.GetRoundEvents))
	mux.Handle("/root/", corsHandler(s.GetRoot))
	mux.Handle("/block/", corsHandler(s.GetBlock))
	mux.Handle("/graph", corsHandler(s.GetGraph))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("src/service/static/"))))
	err := http.ListenAndServe(s.bindAddress, mux)
	if err != nil {
		s.logger.WithField("error", err).Error("Service failed")
	}
}

func corsHandler(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
		if r.Method == "OPTIONS" {
			/*w.Header().Set("Access-Control-Allow-Origin", "*")
			    	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
						w.Header().Set("Access-Control-Allow-Headers",
			        "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")*/
		} else {
			/*w.Header().Set("Access-Control-Allow-Origin", "*")
			    	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
						w.Header().Set("Access-Control-Allow-Headers",
			        "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")*/
			h.ServeHTTP(w, r)
		}
	}
}

func (s *Service) GetStats(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("Stats")

	stats := s.node.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *Service) GetParticipants(w http.ResponseWriter, r *http.Request) {
	participants, err := s.node.GetParticipants()
	if err != nil {
		s.logger.WithError(err).Errorf("Parsing participants parameter")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(participants)
}

func (s *Service) GetEvent(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Path[len("/event/"):]
	event, err := s.node.GetEvent(param)
	if err != nil {
		s.logger.WithError(err).Errorf("Retrieving event %s", param)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

func (s *Service) GetLastEventFrom(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Path[len("/lasteventfrom/"):]
	event, _, err := s.node.GetLastEventFrom(param)
	if err != nil {
		s.logger.WithError(err).Errorf("Retrieving event %s", event)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

func (s *Service) GetGraph(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
 	encoder := json.NewEncoder(w)
 	res := s.graph.GetInfos()
 	encoder.Encode(res)
}

func (s *Service) GetKnownEvents(w http.ResponseWriter, r *http.Request) {
	knownEvents := s.node.GetKnownEvents()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(knownEvents)
}

func (s *Service) GetConsensusEvents(w http.ResponseWriter, r *http.Request) {
	consensusEvents := s.node.GetConsensusEvents()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(consensusEvents)
}

func (s *Service) GetRound(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Path[len("/round/"):]
	roundIndex, err := strconv.Atoi(param)
	if err != nil {
		s.logger.WithError(err).Errorf("Parsing roundIndex parameter %s", param)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	round, err := s.node.GetRound(roundIndex)
	if err != nil {
		s.logger.WithError(err).Errorf("Retrieving round %d", roundIndex)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(round)
}

func (s *Service) GetLastRound(w http.ResponseWriter, r *http.Request) {
	lastRound := s.node.GetLastRound()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lastRound)
}

func (s *Service) GetRoundWitnesses(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Path[len("/roundwitnesses/"):]
	roundWitnessesIndex, err := strconv.Atoi(param)
	if err != nil {
		s.logger.WithError(err).Errorf("Parsing roundWitnessesIndex parameter %s", param)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	roundWitnesses := s.node.GetRoundWitnesses(roundWitnessesIndex)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roundWitnesses)
}

func (s *Service) GetRoundEvents(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Path[len("/roundevents/"):]
	roundEventsIndex, err := strconv.Atoi(param)
	if err != nil {
		s.logger.WithError(err).Errorf("Parsing roundEventsIndex parameter %s", param)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	roundEvent := s.node.GetRoundEvents(roundEventsIndex)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roundEvent)
}

func (s *Service) GetRoot(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Path[len("/root/"):]
	root, err := s.node.GetRoot(param)
	if err != nil {
		s.logger.WithError(err).Errorf("Retrieving root %s", param)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(root)
}

func (s *Service) GetBlock(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Path[len("/block/"):]
	blockIndex, err := strconv.Atoi(param)
	if err != nil {
		s.logger.WithError(err).Errorf("Parsing block_index parameter %s", param)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	block, err := s.node.GetBlock(blockIndex)
	if err != nil {
		s.logger.WithError(err).Errorf("Retrieving block %d", blockIndex)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(block)
}
