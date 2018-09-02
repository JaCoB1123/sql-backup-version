package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"fmt"
)

type service struct {
	Configuration *configuration
}

func (srv service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	head, tail := ShiftPath(r.URL.Path)
	switch head {
	case "api":
		srv.ServeAPI(w, r, tail)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Not found")
	}
}

func (srv service) ServeAPI(w http.ResponseWriter, r *http.Request, tail string) {
	head, tail := ShiftPath(tail)
	switch head {
	case "servers":
		srv.ServeServers(w, r, tail)
	case "files":
		srv.ServeFiles(w, r, tail)
	default:
		fmt.Fprintf(w, "hello, you've hit %s\n", r.URL.Path)
	}
}

func (srv service) ServeServers(w http.ResponseWriter, r *http.Request, tail string) {
	head, tail := ShiftPath(tail)
	switch head {
	case "list":
		result, err := json.Marshal(srv.Configuration.Servers)
		if HasError(w, err) {
			return
		}

		fmt.Fprintln(w, string(result))
	default:
		i, err := strconv.Atoi(head)
		if HasError(w, err) {
			return
		}

		server := srv.Configuration.Servers[i]

		head, _ := ShiftPath(tail)
		switch head {
		case "":
			result, err := json.Marshal(server)
			if HasError(w, err) {
				return
			}

			fmt.Fprintln(w, string(result))
		case "databases":
			db, err := getConnection(server)
			if HasError(w, err) {
				return
			}

			databases, err := getDatabases(db)
			if err != nil {
				panic(err)
			}

			result, err := json.Marshal(databases)
			if HasError(w, err) {
				return
			}

			fmt.Fprintln(w, string(result))
		}
	}
}

func (srv service) ServeFiles(w http.ResponseWriter, r *http.Request, tail string) {
	head, _ := ShiftPath(tail)
	switch head {
	case "list":
		result, err := json.Marshal(srv.Configuration.Files)
		if HasError(w, err) {
			return
		}

		fmt.Fprintln(w, string(result))
	default:
		i, err := strconv.Atoi(head)
		if HasError(w, err) {
			return
		}

		server := srv.Configuration.Files[i]
		result, err := json.Marshal(server)
		if HasError(w, err) {
			return
		}

		fmt.Fprintln(w, string(result))
	}
}
