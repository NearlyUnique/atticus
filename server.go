package atticus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"text/template"

	"github.com/gorilla/mux"
)

type (
	//Server for mocked responses
	Server struct {
		r *mux.Router
	}
	// CannedResponse .
	CannedResponse struct {
		Body       interface{}
		Header     map[string]string
		StatusCode int
	}
	// CannedRequest .
	CannedRequest struct {
		URL    string
		Method string
	}
	// Canned .
	Canned struct {
		Name     string
		Label    string
		Request  CannedRequest
		Response CannedResponse
	}
	// TemplateData based on request data
	TemplateData struct {
		Method string
		URL    string
		Vars   map[string]string
		Header map[string]interface{}
		Query  map[string]interface{}
		Body   map[string]interface{}
	}
)

// New server configured to start
func New() *Server {
	return &Server{}
}

// Run the configured server
func (s *Server) Run() error {
	log.Printf("listening ...")

	var canned []Canned
	f, err := ioutil.ReadFile("./canned.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(f, &canned)
	if err != nil {
		return err
	}

	root := mux.NewRouter()
	for _, c := range canned {
		r := root.NewRoute()
		body, _ := json.Marshal(c.Response.Body)
		code := c.Response.StatusCode
		header := c.Response.Header
		r.
			Path(c.Request.URL).
			Methods(c.Request.Method).
			HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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

				for k, v := range header {
					t, err := template.New("test").Parse(v)
					if err != nil {
						log.Printf("parse failed: %v", err)
						break
					}

					buf := &bytes.Buffer{}
					err = t.Execute(buf, data)
					if err != nil {
						log.Printf("execute failed: %v", err)
						break
					}
					w.Header()[k] = []string{buf.String()}
				}

				w.WriteHeader(code)
				if body != nil {

					// t, err := template.New("test").Parse(string(body))
					// if err != nil {
					// 	log.Printf("parse failed: %v", err)
					// }

					// buf := &bytes.Buffer{}
					// err = t.Execute(buf, data)
					// if err != nil {
					// 	log.Printf("execute failed: %v", err)
					// }

					// data has the data for this request

					// w.Write(buf.Bytes())
				}
			})
	}
	s.r = root

	return http.ListenAndServe(":10000", s)
}

func copyMap(src map[string][]string) map[string]interface{} {
	hdr := make(map[string]interface{})
	for k, v := range src {
		if len(v) == 1 {
			hdr[k] = v[0]
		} else {
			hdr[k] = v
		}
	}
	return hdr
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var match mux.RouteMatch
	if !s.r.Match(r, &match) {
		w.WriteHeader(418)
		fmt.Fprintf(w, "Atticus Failed to marshal response: %v", match.MatchErr)
		return
	}

	ctx := context.WithValue(r.Context(), "vars", match.Vars)
	match.Handler.ServeHTTP(w, r.WithContext(ctx))
}

func vars(r *http.Request) map[string]string {
	return r.Context().Value("vars").(map[string]string)
}

// ApplyTemplate as configured using the data
func ApplyTemplate(template interface{}, data TemplateData) ([]byte, error) {
	result := make(map[string]interface{})

	switch t := template.(type) {
	case map[string]interface{}:
		walkMap(t, mapWriter(result))
	}
	fmt.Printf("%v\n", reflect.TypeOf(template))
	return json.Marshal(result)
}

type valueWriter func(key string, value interface{})

func mapWriter(m map[string]interface{}) valueWriter {
	return func(key string, value interface{}) {
		m[key] = value
	}
}
func sliceWriter(s *[]interface{}) valueWriter {
	return func(k string, value interface{}) {
		fmt.Printf("sliceWriter: %v , %v", k, value)
		*s = append(*s, value)
	}
}

func walkMap(m map[string]interface{}, result valueWriter) {
	for k, v := range m {
		switch value := v.(type) {

		case string:
			result(k, value)

		case map[string]interface{}:
			sub := make(map[string]interface{})
			result(k, sub)
			walkMap(value, mapWriter(sub))

		case []interface{}:
			var sub []interface{}
			sw := sliceWriter(&sub)

			msub := make(map[string]interface{})

			for _, msub["_"] = range value {
				walkMap(msub, sw)
			}

			result(k, sub)

		default:
			result(k, value)

		}
	}
}
