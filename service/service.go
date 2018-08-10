package service

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/andrecronje/lachesis/node"
	"github.com/sirupsen/logrus"
)

type Service struct {
	bindAddress string
	node        *node.Node
	logger      *logrus.Logger
}

func NewService(bindAddress string, node *node.Node, logger *logrus.Logger) *Service {
	service := Service{
		bindAddress: bindAddress,
		node:        node,
		logger:      logger,
	}

	return &service
}

func (s *Service) Serve() {
	s.logger.WithField("bind_address", s.bindAddress).Debug("Service serving")
	http.HandleFunc("/stats", s.GetStats)
	http.HandleFunc("/block/", s.GetBlock)
	http.HandleFunc("/events/", s.GetKnownEvents)
	http.HandleFunc("/consensusevents/", s.GetConsensusEvents)
	http.HandleFunc("/participants/", s.GetParticipants)
	http.HandleFunc("/event/", s.GetEvent)
	http.HandleFunc("/roundwitnesses/", s.GetRoundWitnesses)
	http.HandleFunc("/roundevents/", s.GetRoundEvents)
	err := http.ListenAndServe(s.bindAddress, nil)
	if err != nil {
		s.logger.WithField("error", err).Error("Service failed")
	}
}

func (s *Service) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := s.node.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
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

func (s *Service) GetParticipants(w http.ResponseWriter, r *http.Request) {
	participants := s.node.GetParticipants()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(participants)
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

func (s *Service) GetEvent(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Path[len("/event/"):]
	eventIndex, err := strconv.Atoi(param)
	if err != nil {
		s.logger.WithError(err).Errorf("Parsing eventIndex parameter %s", param)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	event, err := s.node.GetEvent(eventIndex)
	if err != nil {
		s.logger.WithError(err).Errorf("Retrieving event %d", event)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

func (s *Service) GetRoundWitnesses(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Path[len("/roundwitnesses/"):]
	roundWitnessesIndex, err := strconv.Atoi(param)
	if err != nil {
		s.logger.WithError(err).Errorf("Parsing roundWitnessesIndex parameter %s", param)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	roundWitnesses, err := s.node.GetRoundWitnesses(roundWitnessesIndex)
	if err != nil {
		s.logger.WithError(err).Errorf("Retrieving roundWitnesses %d", roundWitnesses)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	roundEvent, err := s.node.GetRoundEvents(roundEventsIndex)
	if err != nil {
		s.logger.WithError(err).Errorf("Retrieving roundEvent %d", roundEvent)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roundEvent)
}
