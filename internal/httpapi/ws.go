package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/ModernSelfHost/vertexos/internal/jobqueue"
	"nhooyr.io/websocket"
)

func registerWS(mux *http.ServeMux, deps RouterDeps) {
	if deps.JobManager == nil {
		return
	}
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close(websocket.StatusNormalClosure, "")

		ch := deps.JobManager.Subscribe()
		defer deps.JobManager.Unsubscribe(ch)

		// Send initial job list.
		for _, j := range deps.JobManager.List() {
			b, _ := json.Marshal(jobqueue.Event{Type: "job", Job: j})
			_ = c.Write(r.Context(), websocket.MessageText, b)
		}

		for ev := range ch {
			b, err := json.Marshal(ev)
			if err != nil {
				continue
			}
			if err := c.Write(r.Context(), websocket.MessageText, b); err != nil {
				return
			}
		}
	})
}
