package api

import (
	"encoding/json"
	"io"
	"net/http"

	"Raft3D/internal/models"
	"Raft3D/internal/raftnode"

	"github.com/gorilla/mux"
)

func RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/printers", addPrinter).Methods("POST")
	router.HandleFunc("/printers", getPrinters).Methods("GET")

	router.HandleFunc("/filaments", addFilament).Methods("POST")
	router.HandleFunc("/filaments", getFilaments).Methods("GET")

	router.HandleFunc("/printjobs", addPrintJob).Methods("POST")
	router.HandleFunc("/printjobs", getPrintJobs).Methods("GET")

	router.HandleFunc("/printjobs/{id}/status", updatePrintJobStatus).Methods("POST")
	
}

func addPrinter(w http.ResponseWriter, r *http.Request) {
	var p models.Printer
	body, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(body, &p)
	if !raftnode.IsLeader() {
		http.Error(w, "not leader", http.StatusForbidden)
		return
	}
	err := raftnode.ApplyCommand("printer", p)
	if err != nil {
		http.Error(w, "apply failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func getPrinters(w http.ResponseWriter, r *http.Request) {
	list := raftnode.GetFSM().Printers
	json.NewEncoder(w).Encode(list)
}

func addFilament(w http.ResponseWriter, r *http.Request) {
	var f models.Filament
	body, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(body, &f)
	if !raftnode.IsLeader() {
		http.Error(w, "not leader", http.StatusForbidden)
		return
	}
	err := raftnode.ApplyCommand("filament", f)
	if err != nil {
		http.Error(w, "apply failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func getFilaments(w http.ResponseWriter, r *http.Request) {
	list := raftnode.GetFSM().Filaments
	json.NewEncoder(w).Encode(list)
}

func addPrintJob(w http.ResponseWriter, r *http.Request) {
	var j models.PrintJob
	body, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(body, &j)
	if !raftnode.IsLeader() {
		http.Error(w, "not leader", http.StatusForbidden)
		return
	}
	err := raftnode.ApplyCommand("printjob", j)
	if err != nil {
		http.Error(w, "apply failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func getPrintJobs(w http.ResponseWriter, r *http.Request) {
	list := raftnode.GetFSM().PrintJobs
	json.NewEncoder(w).Encode(list)
}

func updatePrintJobStatus(w http.ResponseWriter, r *http.Request) {
	if !raftnode.IsLeader() {
		http.Error(w, "not leader", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	jobID := vars["id"]
	newStatus := r.URL.Query().Get("status")

	if newStatus == "" {
		http.Error(w, "Missing 'status' query param", http.StatusBadRequest)
		return
	}

	// Validate allowed statuses
	validStatuses := map[string]bool{
		"queued":   true,
		"running":  true,
		"done":     true,
		"canceled": true,
	}
	if !validStatuses[newStatus] {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	// Get current FSM state
	fsm := raftnode.GetFSM()

	// Find the job
	job, ok := fsm.PrintJobs[jobID]
	if !ok {
		http.Error(w, "Print job not found", http.StatusNotFound)
		return
	}

	// Transition rules
	if newStatus == "running" && job.Status != "queued" {
		http.Error(w, "Can only move to 'running' from 'queued'", http.StatusBadRequest)
		return
	}
	if newStatus == "done" && job.Status != "running" {
		http.Error(w, "Can only move to 'done' from 'running'", http.StatusBadRequest)
		return
	}
	if newStatus == "canceled" && !(job.Status == "queued" || job.Status == "running") {
		http.Error(w, "Can only cancel from 'queued' or 'running'", http.StatusBadRequest)
		return
	}

	// If moving to done, validate & deduct weight
	if newStatus == "done" {
		filament, ok := fsm.Filaments[job.FilamentID]
		if !ok {
			http.Error(w, "Associated filament not found", http.StatusInternalServerError)
			return
		}
		if job.PrintWeight > filament.RemainingWeight {
			http.Error(w, "Insufficient filament weight", http.StatusBadRequest)
			return
		}
		filament.RemainingWeight -= job.PrintWeight
		fsm.Filaments[job.FilamentID] = filament // Save updated filament
	}

	// Update job status
	job.Status = newStatus

	// Send the updated job to Raft
	err := raftnode.ApplyCommand("update_print_job_status", job)
	if err != nil {
		http.Error(w, "Apply failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Print job status updated successfully"))
}
