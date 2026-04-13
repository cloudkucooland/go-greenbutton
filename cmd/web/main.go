package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"

	// "golang.org/x/crypto/acme/autocert"

	"github.com/cloudkucooland/go-greenbutton"
	"github.com/urfave/cli/v3"
)

const jsonType = "application/json"

type Server struct {
	staticdir string
	plans     []greenbutton.Plan
}

func main() {
	cmd := &cli.Command{
		Name:  "greenbutton",
		Usage: "SMT Solar Plan Simulator",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "port",
				Value: ":8090",
				Usage: "https port",
			},
			&cli.StringFlag{
				Name:  "dir",
				Value: "/home/gb",
				Usage: "Working directory",
			},
		},

		Action: startup,
	}

	// use signals
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("startup failed", "error", err)
		os.Exit(1)
	}
}

func startup(ctx context.Context, cmd *cli.Command) error {
	workdir := cmd.String("dir")

	plans, err := greenbutton.LoadPlans(path.Join(workdir, "plans.json"))
	if err != nil {
		return err
	}

	srv := &Server{
		plans:     plans,
		staticdir: path.Join(workdir, "static"),
	}

	fs := http.FileServer(http.Dir(srv.staticdir))

	mux := http.NewServeMux()
	mux.HandleFunc("POST /upload", srv.handleUpload)
	mux.HandleFunc("GET /plans", srv.handleGetPlans)
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	slog.Info("Listening", "port", cmd.String("port"))
	return http.ListenAndServe(cmd.String("port"), mux)

	// do simple HTTPS using Acme cert - later in dev cycle, after getting external host online
	/*
	   m := &autocert.Manager{
	       Cache:      autocert.DirCache(path.Join(workdir, "acme")),
	       Prompt:     autocert.AcceptTOS,
	       HostPolicy: autocert.HostWhitelist("julysun.store"),
	   }
	   s := &http.Server{
	   Addr:      ":https",
	       TLSConfig: m.TLSConfig(),
	   }
	   s.ListenAndServeTLS("", "")
	*/
}

func jsonError(e error) string {
	return fmt.Sprintf(`{"status":"error","error":"%s"}`, e.Error())
}

func (s *Server) headers(w http.ResponseWriter, r *http.Request) {
	ref := r.Header.Get("Origin")

	w.Header().Add("Access-Control-Allow-Origin", ref)
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Accept, If-Modified-Since, If-Match, If-None-Match, Authorization")

	w.Header().Add("Server", "Solar Plan Helper")
	w.Header().Add("Content-Type", jsonType)
}

// accepts incoming SmartMeterTexas CSV files and returns the processed data
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	s.headers(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 12MB max (2 years, max allowed export from SMT)
	r.ParseMultipartForm(12 << 20)

	file, header, err := r.FormFile("d")
	if err != nil {
		fmt.Printf("%+v\n", r.PostForm["d"])
		slog.Error("error retrieving data", "error", err)
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	slog.Info("DEBUG", "filename", header.Filename, "header", header.Header)

	isCSV := false
	if strings.HasSuffix(header.Filename, ".csv") {
		isCSV = true
	}

	monthlymap, err := greenbutton.Loader(file, isCSV)
	if err != nil {
		slog.Error("error parsing  data", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO if the client says we can save the file for analytics, save it to workdir/analyze
	// scrub the PII from the file

	// should we push this logic to the client?
	// make the client sort the months at display time, we just calculate them?
	months := make([]string, 0, len(monthlymap))
	for m := range monthlymap {
		months = append(months, m)
	}
	sort.Strings(months)

	type mm struct {
		Month string
		MBI   greenbutton.MonthlyBillInterval
		Cents float64
	}

	type pd struct {
		Name string
		Data []mm
	}

	out := make([]pd, 0, len(s.plans))

	for _, p := range s.plans {
		npd := pd{
			Name: p.Name,
			Data: make([]mm, 0, len(months)),
		}

		for _, month := range months {
			mbi, cents, err := monthlymap[month].Sum(r.Context(), p)
			if err != nil {
				slog.Error("error building results", "error", err)
				http.Error(w, jsonError(err), http.StatusInternalServerError)
				return
			}
			nmm := mm{
				Month: month,
				MBI:   mbi,
				Cents: cents,
			}
			npd.Data = append(npd.Data, nmm)
		}
		out = append(out, npd)
	}

	if err := json.NewEncoder(w).Encode(out); err != nil {
		slog.Error("error sending results", "error", err)
		http.Error(w, jsonError(err), http.StatusInternalServerError)
		return
	}

	// allow optional custom plan
	// allow selection of specific plans (by name?)
	// return JSON of the Sum
}

// sends plan data to UI
func (s *Server) handleGetPlans(w http.ResponseWriter, r *http.Request) {
	s.headers(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := json.NewEncoder(w).Encode(s.plans); err != nil {
		slog.Error("error sending plans", "error", err)
		http.Error(w, jsonError(err), http.StatusInternalServerError)
		return
	}
}
