package web

import (
	"database/sql"
	"encoding/json"
	"event-driven/internal/server/broker"
	genrepo "event-driven/internal/sqlc/generated/repository"
	"event-driven/types"
	"fmt"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
	"log"
	"net/http"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) (err error)
type Routes map[string]HandlerFunc

type Handler struct {
	repository *genrepo.Queries
	pb         *broker.PublisherServer
}

func NewHandler(repository *genrepo.Queries, pb *broker.PublisherServer) *Handler {
	return &Handler{repository: repository, pb: pb}
}

func (h Handler) GetRoutes() Routes {
	return Routes{
		"POST /task":             h.postMsg,
		"PATCH /task/{event_id}": h.patchMsg,

		"POST /saga":             h.postSaga,
		"PATCH /saga/{event_id}": h.patchSaga,
	}
}

func (h Handler) postMsg(w http.ResponseWriter, r *http.Request) (err error) {
	r.Header.Set("Content-Type", "application/json")

	var input types.PayloadType
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, fmt.Sprintf("could not decode input with error: %v", err), http.StatusBadRequest)
		return nil
	}

	opts, _ := json.Marshal(input.Opts)
	payload, _ := json.Marshal(input.Payload)
	info, _ := json.Marshal(input.Info)

	if err := h.repository.CreateTransaction(r.Context(), genrepo.CreateTransactionParams{
		EventID:   input.EventID,
		EventName: input.EventName,
		Opts:      opts,
		StartedAt: sql.NullTime{
			Time:  input.CreatedAt,
			Valid: true,
		},
		Info: pqtype.NullRawMessage{
			RawMessage: info,
			Valid:      true,
		},
		Payload: payload,
		Status:  "PENDING",
	}); err != nil {
		log.Printf("could not create transaction with error: %v\n", err)
	}

	if err := h.pb.Producer(r.Context(), input); err != nil {
		http.Error(w, fmt.Sprintf("could not send message with error: %v", err), http.StatusInternalServerError)
		return nil
	}

	w.WriteHeader(http.StatusAccepted)
	return
}

func (h Handler) patchMsg(w http.ResponseWriter, r *http.Request) (err error) {
	r.Header.Set("Content-Type", "application/json")

	eventID := r.PathValue("event_id")
	ID, err := uuid.Parse(eventID)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not decode input with error: %v", err), http.StatusBadRequest)
		return
	}

	fmt.Println("UPDATE INFOS", eventID)

	defer r.Body.Close()
	var input types.UpdatePayloadInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, fmt.Sprintf("could not decode input with error: %v", err), http.StatusBadRequest)
		return nil
	}

	if err := h.repository.UpdateTransaction(r.Context(), genrepo.UpdateTransactionParams{
		EventID: ID,
		Status:  input.Status,
		TotalRetry: sql.NullInt32{
			Int32: int32(input.TotalRetry),
			Valid: true,
		},
		EndedAt: sql.NullTime{
			Time:  input.FinishedAt,
			Valid: true,
		},
	}); err != nil {
		log.Printf("could not update transaction with error: %v\n", err)
	}

	w.WriteHeader(http.StatusAccepted)
	return
}

func (h Handler) postSaga(w http.ResponseWriter, r *http.Request) (err error) {
	r.Header.Set("Content-Type", "application/json")

	var input types.PayloadType
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, fmt.Sprintf("could not decode input with error: %v", err), http.StatusBadRequest)
		return nil
	}

	opts, _ := json.Marshal(input.Opts)
	payload, _ := json.Marshal(input.Payload)
	info, _ := json.Marshal(input.Info)

	transaction, err := h.repository.GetTransactionByEventID(r.Context(), input.TransactionEventID)
	if err != nil {
		log.Printf("could not get transaction with error: %v\n", err)
		http.Error(w, fmt.Sprintf("could not get transaction with error: %v", err), http.StatusInternalServerError)
		return nil
	}

	if err := h.repository.CreateTxSaga(r.Context(), genrepo.CreateTxSagaParams{
		TransactionID: transaction.ID,
		EventID:       input.EventID,
		EventName:     input.EventName,
		Opts:          opts,
		StartedAt: sql.NullTime{
			Time:  input.CreatedAt,
			Valid: true,
		},
		Info: pqtype.NullRawMessage{
			RawMessage: info,
			Valid:      true,
		},
		Payload: payload,
		Status:  "PENDING",
	}); err != nil {
		log.Printf("could not create tx_sagas with error: %v\n", err)
	}

	if err := h.pb.Producer(r.Context(), input); err != nil {
		log.Printf("could not send message with error: %v\n", err)
		http.Error(w, fmt.Sprintf("could not send message with error: %v", err), http.StatusInternalServerError)
		return nil
	}

	w.WriteHeader(http.StatusAccepted)
	return
}

func (h Handler) patchSaga(w http.ResponseWriter, r *http.Request) (err error) {
	r.Header.Set("Content-Type", "application/json")

	eventID := r.PathValue("event_id")
	ID, err := uuid.Parse(eventID)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not decode input with error: %v", err), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
	var input types.UpdatePayloadInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, fmt.Sprintf("could not decode input with error: %v", err), http.StatusBadRequest)
		return nil
	}

	fmt.Println("SAGA UPDATE INFOS", eventID, input.Status)
	if err := h.repository.UpdateTxSaga(r.Context(), genrepo.UpdateTxSagaParams{
		EventID: ID,
		Status:  input.Status,
		TotalRetry: sql.NullInt32{
			Int32: int32(input.TotalRetry),
			Valid: true,
		},
		EndedAt: sql.NullTime{
			Time:  input.FinishedAt,
			Valid: true,
		},
	}); err != nil {
		log.Printf("could not update transaction with error: %v\n", err)
	}

	w.WriteHeader(http.StatusAccepted)
	return
}
