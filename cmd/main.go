package main

import (
	_ "embed"
	"fmt"
	"sort"
	"strings"
	"sync"

	"flag"
	"html/template"
	"log"
	"net/http"
	"time"
)

//go:embed index.html
var indexTplStr string

var indexTpl *template.Template

func init() {
	var err error
	indexTpl, err = template.New("index.html").Parse(indexTplStr)
	if err != nil {
		panic(err)
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

var Option struct {
	Addr string
}

func init() {
	flag.StringVar(&Option.Addr, "addr", ":9280", "http server listen address")
}

func main() {
	flag.Parse()
	var s Service
	if err := s.Serve(Option.Addr); err != nil {
		log.Fatal(err)
	}
}

type Job struct {
	URL   string
	Start time.Time
	Spend time.Duration
}

type Data struct {
	Alerts []string
	Jobs   []*Job
}

type sortJob []*Job

func (s sortJob) Len() int {
	return len(s)
}

func (s sortJob) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s sortJob) Less(i, j int) bool {
	return s[i].Start.Before(s[j].Start)
}

func sortJobs(jobs []*Job) {
	sort.Sort(sortJob(jobs))
}

type Service struct {
	mu       sync.Mutex
	mux      *http.ServeMux
	Jobs     map[string]*Job
	alertsmu sync.Mutex
	Alerts   []string
}

func (s *Service) Index(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL.String())
	var d Data
	d.Jobs = s.jobs()
	for _, j := range d.Jobs {
		j.Spend = time.Since(j.Start)
	}
	d.Alerts = s.alerts()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := indexTpl.Execute(w, d); err != nil {
		log.Println(err)
	}
}

func (s *Service) Submit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	url := strings.TrimSpace(r.Form.Get("url"))
	if url != "" {
		log.Println("Add new job", url)
		s.submitJob(url)
	}
	w.Header().Add("Location", "/")
	w.WriteHeader(303)
}

func (s *Service) addAlerts(t string) {
	s.alertsmu.Lock()
	s.Alerts = append(s.Alerts, t)
	s.alertsmu.Unlock()
}

func (s *Service) alerts() []string {
	s.alertsmu.Lock()
	a := s.Alerts
	s.Alerts = nil
	s.alertsmu.Unlock()
	return a
}

func (s *Service) submitJob(url string) *Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	if j, ok := s.Jobs[url]; ok {
		s.addAlerts(fmt.Sprintf("url exists %s", url))
		return j
	}
	j := &Job{
		URL:   url,
		Start: time.Now(),
	}
	if s.Jobs == nil {
		s.Jobs = map[string]*Job{}
	}
	s.Jobs[url] = j
	return j
}

func (s *Service) jobs() []*Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	var i int
	result := make([]*Job, len(s.Jobs))
	for _, v := range s.Jobs {
		result[i] = v
		i++
	}
	sortJobs(result)
	return result
}

func (s *Service) Handle(method, path string, h func(w http.ResponseWriter, r *http.Request)) {
	s.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if r.Method != method {
			w.WriteHeader(405)
			return
		}
		h(w, r)
	})
}

func (s *Service) Serve(addr string) error {
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}
	s.Handle("GET", "/", s.Index)
	s.Handle("POST", "/jobs/submit", s.Submit)

	return http.ListenAndServe(addr, s.mux)
}
