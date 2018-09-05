package atticus

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type cannedCollection interface {
	addCanned(Canned) error
}

func controlPlane(cc cannedCollection) http.Handler {

	r := mux.NewRouter()

	r.PathPrefix("/canned").
		Methods("POST").HandlerFunc(addCannedResponse(cc.addCanned))

	return r
}

func addCannedResponse(add func(canned Canned) error) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var c Canned
		buf, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "reading body: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		log.Printf("json:%s", string(buf))
		err = json.Unmarshal(buf, &c)
		if err != nil {
			http.Error(w, "parsing body: "+err.Error(), http.StatusBadRequest)
			return
		}

		err = add(c)
		if err != nil {
			return
		}
	}
}
