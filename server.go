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
		Header http.Header
		Method string
		URL    string
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

				requestData := map[string]interface{}{
					"method": r.Method,
					"url":    r.URL.String(),
					"query":  copyMap(map[string][]string(r.URL.Query())),
					"vars":   vars(r),
					"header": copyMap(map[string][]string(r.Header)),
					// form?
				}

				if r.Body != nil && r.ContentLength > 0 {
					buf, err := ioutil.ReadAll(r.Body)
					if err == nil {
						var jsonBody map[string]interface{}
						err := json.Unmarshal(buf, &jsonBody)
						if err == nil {
							requestData["body"] = jsonBody
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
					err = t.Execute(buf, requestData)
					if err != nil {
						log.Printf("execute failed: %v", err)
						break
					}
					w.Header()[k] = []string{buf.String()}
				}

				w.WriteHeader(code)
				if body != nil {
					t, err := template.New("test").Parse(string(body))
					if err != nil {
						log.Printf("parse failed: %v", err)
					}

					buf := &bytes.Buffer{}
					err = t.Execute(buf, requestData)
					if err != nil {
						log.Printf("execute failed: %v", err)
					}
					w.Write(buf.Bytes())
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
		walkMap(t, result)
	}
	return json.Marshal(result)
}

type valueWriter func(string, interface{})

func walkMap(m map[string]interface{}, result map[string]interface{}) {
	for k, v := range m {
		switch i := v.(type) {

		case string:
			result[k] = i

		case map[string]interface{}:
			sub := make(map[string]interface{})
			result[k] = sub
			walkMap(i, sub)

		// case []interface{}:
		// 	for _, x := range i {
		// 		walkMap(x)
		// 	}

		default:
			fmt.Printf("unknown: %s = %v\n", k, reflect.TypeOf(v))
			result[k] = v
		}
	}
}
