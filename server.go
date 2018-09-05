package atticus

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/gorilla/mux"
)

type (
	//Server for mocked responses
	Server struct {
		http    *http.Server
		control *http.Server
		mu      sync.Mutex
		router  *mux.Router
		canned  []Canned
	}
	// ResponseTemplate .
	ResponseTemplate struct {
		Body       interface{}
		Header     map[string]string
		StatusCode int
	}
	// RequestMatch .
	RequestMatch struct {
		URL    string
		Method string
		Header map[string]string
	}
	// Canned .
	Canned struct {
		Name     string
		Label    string
		Match    RequestMatch
		Template ResponseTemplate
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
func New(initial string) (*Server, error) {
	var canned []Canned
	if initial != "" {
		f, err := ioutil.ReadFile(initial)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(f, &canned)
		if err != nil {
			return nil, err
		}
	}

	log.Printf("%d canned results configured", len(canned))

	return &Server{
		canned: canned,
	}, nil
}

// Run the configured server
func (s *Server) Run(ctrlListener net.Listener, runListener net.Listener) error {
	log.Printf("listening ...")

	root := mux.NewRouter()
	s.router = root

	for _, c := range s.canned {
		s.addCanned(c)
	}

	s.http = &http.Server{Handler: s}
	s.control = &http.Server{Handler: controlPlane(s)}

	go func() {
		err := s.control.Serve(ctrlListener)
		if err != nil {
			log.Printf("Control plane serve : %v", err)
		}
	}()

	return s.http.Serve(runListener)
}

func (s *Server) addCanned(c Canned) error {
	var valid []string
	if c.Match.URL == "" {
		valid = append(valid, "missing match URL")
	}
	if c.Match.URL == "" {
		valid = append(valid, "missing match URL")
	}

	if len(valid) > 0 {
		return errors.New(strings.Join(valid, ","))
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	r := s.router.NewRoute()
	r.Path(c.Match.URL)

	if c.Match.Method != "" {
		r.Methods(c.Match.Method)
	}

	r.HandlerFunc(runtimeHandler(
		c.Template.Body,
		c.Template.Header,
		c.Template.StatusCode,
	))

	return nil
}

func (s *Server) Close() error {
	ctx := context.Background()

	if err := s.control.Shutdown(ctx); err != nil {
		log.Printf("Control plane shutdown: %v", err)
	}

	return s.http.Shutdown(ctx)
}

func copyMap(src map[string][]string) map[string]interface{} {
	hdr := make(map[string]interface{})
	for k, v := range src {
		k = strings.Replace(strings.ToLower(k), "-", "_", -1)
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
	if !s.router.Match(r, &match) {
		w.WriteHeader(599)
		fmt.Fprintf(w, "Atticus Failed to marshal response: %v", match.MatchErr)
		return
	}

	ctx := context.WithValue(r.Context(), "vars", match.Vars)
	match.Handler.ServeHTTP(w, r.WithContext(ctx))
}

func vars(r *http.Request) map[string]string {
	return r.Context().Value("vars").(map[string]string)
}
