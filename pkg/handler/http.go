package handler

import (
	"encoding/json"
	"io"
	"net/http"

	awsevent "github.com/aws/aws-lambda-go/events"
)

func (h *Handler) HttpMount(w http.ResponseWriter, r *http.Request) {
	h.log.Debug().Msg("http handler called")

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		WriteResponse(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	if !json.Valid(bodyBytes) {
		WriteResponse(w, http.StatusBadRequest, "request body is not valid json")
		return
	}

	out, err := h.Event(r.Context(), json.RawMessage(bodyBytes))
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteResponse(w, http.StatusOK, string(out))
}

func WriteResponse(w http.ResponseWriter, status int, body string) {
	resp := awsevent.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       body,
	}

	encoded, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to marshal response"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(encoded)
}
