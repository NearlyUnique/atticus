package atticus

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
)

func runtimeHandler(templateBody interface{}, templateHeader map[string]string, templateCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		data := TemplateData{
			Method: r.Method,
			URL:    r.URL.String(),
			Query:  copyMap(map[string][]string(r.URL.Query())),
			Vars:   vars(r),
			Header: copyMap(map[string][]string(r.Header)),
		}

		if r.Body != nil && r.ContentLength > 0 {
			buf, err := ioutil.ReadAll(r.Body)
			if err == nil {
				err := json.Unmarshal(buf, &data.Body)
				if err == nil {
					log.Printf("Bad Template: %v", err)
				}
			}
		}

		for k, v := range templateHeader {

			t, err := template.New("test").Parse(v)
			if err != nil {
				log.Printf("parse failed: %v", err)
				break
			}

			buf := &bytes.Buffer{}
			err = t.Execute(buf, data)
			if err != nil {
				log.Printf("execute failed: %v", err)
				w.Header()[k] = []string{v}
			} else {
				w.Header()[k] = []string{buf.String()}
			}
		}

		var buf []byte

		if templateBody != nil {
			buf, err = ApplyBodyTemplate(templateBody, &data)
			if err != nil {
				log.Printf("ERROR:%v", err)
			}
		}

		if err != nil {
			w.Header().Set("Atticus-Error", err.Error())
			w.WriteHeader(599)
		} else {
			w.WriteHeader(templateCode)
		}
		if len(buf) > 0 {
			w.Write(buf)
		}
	}
}
