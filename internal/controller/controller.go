package controller

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type Controller struct {
	SlackAPI *slack.Client
	SlackSocket *socketmode.Client
}

type Challenge struct {
	Type      string `json:"type"`
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
}

func (c *Controller) Register(r *mux.Router) {
	r.Handle("/", c.HandleChallenge()).Methods(http.MethodPost)
	r.Handle("/events-endpoint", c.HandleEvents()).Methods(http.MethodPost)
}

func (c *Controller) HandleChallenge() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		requestBytes, _ := ioutil.ReadAll(r.Body)

		var challenge Challenge
		json.Unmarshal(requestBytes, &challenge)
		responseBytes, _ := json.Marshal(challenge)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(responseBytes)
	}
}
