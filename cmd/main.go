package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
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
	Addr   string
	BBDown string
}

func init() {
	flag.StringVar(&Option.Addr, "addr", ":9280", "http server listen address")
	flag.StringVar(&Option.BBDown, "bbdown", "./BBDown", "BBDown path")
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
	Cmd   *Cmd
	State string
}

type Cmd struct {
	Cmd    *exec.Cmd
	Output *os.File
}

func Exec(name string, args ...string) (*Cmd, error) {
	file, err := os.CreateTemp("", "bbdown-*")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(name, args...)
	cmd.Stdout = file
	cmd.Stderr = file
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &Cmd{Cmd: cmd, Output: file}, nil
}

func (c *Cmd) Tail() ([]byte, error) {
	offset, err := c.Output.Seek(0, os.SEEK_CUR)
	if err != nil {
		return nil, err
	}
	var resp = make([]byte, 4096)
	start := offset - int64(len(resp))
	if start <= 0 {
		start = 0
	}
	if offset == 0 {
		return nil, nil
	}
	if _, err := c.Output.ReadAt(resp, start); err != nil {
		return nil, err
	}
	return resp, nil
}

// Start a download job
func Start(url string) (*Job, error) {
	var j Job
	j.URL = url
	j.Start = time.Now()
	cmd, err := Exec(Option.BBDown,
		"--multi-thread",
		"--work-dir",
		"/downloads",
		"--encoding-priority",
		"hevc,av1,avc",
		"--delay-per-page",
		"5",
		url,
	)
	if err != nil {
		return nil, err
	}
	j.Cmd = cmd
	go func() {
		log.Println(j.URL, "started", j.Cmd.Output.Name())
		if err := j.Cmd.Cmd.Wait(); err != nil {
			log.Println(j.URL, "fails", err, j.Cmd.Output.Name())
		} else {
			log.Println(j.URL, "finish", j.Cmd.Output.Name())
		}
	}()
	return &j, nil
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

var (
	loginMu  sync.Mutex
	loginCmd *exec.Cmd
)

func (s *Service) Login(w http.ResponseWriter, r *http.Request) {
	loginMu.Lock()
	defer loginMu.Unlock()

	if loginCmd != nil {
		loginCmd.Process.Kill()
	}

	os.Remove("./qrcode.png")
	cmd := exec.Command(Option.BBDown, "login")
	if err := cmd.Start(); err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err)
		return
	}
	loginCmd = cmd
	defer func() {
		go func() {
			err := cmd.Wait()
			loginMu.Lock()
			if cmd == loginCmd {
				loginCmd = nil
			}
			loginMu.Unlock()
			log.Println("login return with", err)
		}()
	}()

	time.Sleep(time.Second)

	file, err := os.Open("./qrcode.png")
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(200)
	io.Copy(w, file)
}

func (s *Service) Status(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	url := strings.TrimSpace(r.Form.Get("job"))
	s.mu.Lock()
	j := s.Jobs[url]
	s.mu.Unlock()
	if j == nil {
		w.WriteHeader(404)
		return
	}
	resp, err := j.Cmd.Tail()
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err)
		return
	}
	w.Write(resp)
	return
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
	j, err := Start(url)
	if err != nil {
		s.addAlerts(fmt.Sprintf("url(%s) fails: %v", url, err))
		return nil
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
		if v.Cmd.Cmd.ProcessState != nil {
			v.State = v.Cmd.Cmd.ProcessState.String()
		} else {
			v.State = "running"
		}
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
		log.Println(r.Method, r.URL)
		h(w, r)
	})
}

func (s *Service) Serve(addr string) error {
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}
	s.Handle("GET", "/", s.Index)
	s.Handle("POST", "/jobs/submit", s.Submit)
	s.Handle("GET", "/jobs/status", s.Status)
	s.Handle("GET", "/login", s.Login)

	return http.ListenAndServe(addr, s.mux)
}