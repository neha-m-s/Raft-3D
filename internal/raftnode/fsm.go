package raftnode

import (
	"bytes"
	"encoding/gob"
	"io"
	"sync"

	"Raft3D/internal/models"

	"github.com/hashicorp/raft"
)

type Command struct {
	Type string
	Data []byte
}

type RaftFSM struct {
	mu        sync.Mutex
	Printers  map[string]models.Printer
	Filaments map[string]models.Filament
	PrintJobs map[string]models.PrintJob
}

func NewFSM() *RaftFSM {
	return &RaftFSM{
		Printers:  make(map[string]models.Printer),
		Filaments: make(map[string]models.Filament),
		PrintJobs: make(map[string]models.PrintJob),
	}
}

func (f *RaftFSM) Apply(log *raft.Log) interface{} {
	var cmd Command
	if err := gob.NewDecoder(bytes.NewReader(log.Data)).Decode(&cmd); err != nil {
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	switch cmd.Type {
	case "printer":
		var p models.Printer
		_ = gob.NewDecoder(bytes.NewReader(cmd.Data)).Decode(&p)
		f.Printers[p.ID] = p
	case "filament":
		var fl models.Filament
		_ = gob.NewDecoder(bytes.NewReader(cmd.Data)).Decode(&fl)
		f.Filaments[fl.ID] = fl
	case "printjob":
		var job models.PrintJob
		_ = gob.NewDecoder(bytes.NewReader(cmd.Data)).Decode(&job)
		f.PrintJobs[job.ID] = job
	case "update_print_job_status":
		var updated models.PrintJob
		_ = gob.NewDecoder(bytes.NewReader(cmd.Data)).Decode(&updated)
	
		// Fetch the existing job
		oldJob, exists := f.PrintJobs[updated.ID]
		if !exists {
			// Silently ignore if job doesn't exist (or log)
			return nil
		}
	
		// If transitioning to "done", deduct filament weight
		if oldJob.Status != "done" && updated.Status == "done" {
			filament, ok := f.Filaments[oldJob.FilamentID]
			if ok && filament.RemainingWeight >= oldJob.PrintWeight {
				filament.RemainingWeight -= oldJob.PrintWeight
				f.Filaments[oldJob.FilamentID] = filament
			}
		}
	
		// Preserve other job details (we're only updating status here)
		oldJob.Status = updated.Status
		f.PrintJobs[updated.ID] = oldJob
	}
	return nil
}

func (f *RaftFSM) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	buf := new(bytes.Buffer)
	gob.NewEncoder(buf).Encode(f)
	return &snapshot{state: buf.Bytes()}, nil
}

func (f *RaftFSM) Restore(rc io.ReadCloser) error {
	return gob.NewDecoder(rc).Decode(f)
}

type snapshot struct {
	state []byte
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	_, err := sink.Write(s.state)
	if err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *snapshot) Release() {}
