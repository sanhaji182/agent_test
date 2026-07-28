package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/schedule"
	"github.com/google/uuid"
)

// --- Schedule handlers ---

func (s *Server) handleCreateSchedule(w http.ResponseWriter, r *http.Request) {
	var sch schedule.Schedule
	if err := json.NewDecoder(r.Body).Decode(&sch); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if sch.TestListID != "" {
		list, err := s.planning.GetTestList(r.Context(), sch.TestListID)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "test list not found")
			return
		}
		if sch.ProjectID == "" {
			sch.ProjectID = list.ProjectID
		}
		if sch.Name == "" {
			sch.Name = list.Name + " schedule"
		}
	}
	if sch.Enabled {
		sch.NextRunAt = schedule.CalcNextRunInTZ(sch.Frequency, sch.CronExpr, sch.Timezone, time.Now())
	}
	result := s.schedules.Create(&sch)
	if result == nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleListSchedules(w http.ResponseWriter, r *http.Request) {
	list := s.schedules.List()
	if list == nil {
		list = []*schedule.Schedule{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (s *Server) handleGetSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sch, ok := s.schedules.Get(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sch)
}

func (s *Server) handleUpdateSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var patch map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	ok := s.schedules.Update(id, func(sch *schedule.Schedule) {
		if v, ok := safeBool(patch, "enabled"); ok {
			sch.Enabled = v
		}
		if v, ok := safeString(patch, "name"); ok {
			sch.Name = v
		}
		if v, ok := safeString(patch, "webhook_url"); ok {
			sch.WebhookURL = v
		}
		if v, ok := safeBool(patch, "notify_on_fail"); ok {
			sch.NotifyOnFail = v
		}
		if v, ok := safeString(patch, "environment"); ok {
			sch.Environment = v
		}
		if v, ok := safeString(patch, "base_url"); ok {
			sch.BaseURL = v
		}
		if v, ok := safeString(patch, "test_list_id"); ok {
			sch.TestListID = v
		}
		if v, ok := safeString(patch, "frequency"); ok {
			sch.Frequency = schedule.Frequency(v)
			sch.NextRunAt = schedule.CalcNextRunInTZ(sch.Frequency, sch.CronExpr, sch.Timezone, time.Now())
		}
		if v, ok := safeString(patch, "cron_expr"); ok {
			sch.CronExpr = v
			sch.NextRunAt = schedule.CalcNextRunInTZ(sch.Frequency, sch.CronExpr, sch.Timezone, time.Now())
		}
	})
	if !ok {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	sch, _ := s.schedules.Get(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sch)
}

func (s *Server) handleDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !s.schedules.Delete(id) {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleRunNow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sch, ok := s.schedules.Get(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	if sch.TestListID != "" {
		list, err := s.planning.GetTestList(r.Context(), sch.TestListID)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "test list not found")
			return
		}
		runIDs, err := s.startTestListRuns(r.Context(), list)
		if err != nil {
			slog.Error("start test list runs failed", "error", err, "test_list_id", sch.TestListID)
			writeJSONError(w, http.StatusInternalServerError, "failed to start test list runs")
			return
		}
		now := time.Now()
		lastRunID := ""
		if len(runIDs) > 0 {
			lastRunID = runIDs[0]
		}
		s.schedules.Update(id, func(sc *schedule.Schedule) {
			sc.LastRunAt = &now
			sc.LastRunID = lastRunID
			sc.LastRunStatus = string(agent.StateIdle)
			sc.NextRunAt = schedule.CalcNextRunInTZ(sc.Frequency, sc.CronExpr, sc.Timezone, now)
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{"run_id": lastRunID, "run_ids": runIDs, "state": string(agent.StateIdle), "test_list_id": list.ID})
		return
	}
	// Create a run from the schedule config
	now := time.Now()
	run, err := s.startScheduleRun(r.Context(), sch, id, now, "running", "Run created via schedule run-now")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	// Snapshot response fields BEFORE launching (run is mutated async after launch).
	resp := map[string]string{"run_id": run.ID, "state": string(run.State)}
	s.launchRun(run)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}

// startScheduleRun membuat TestRun dari konfigurasi schedule non-list, memperbarui
// metadata last-run schedule, dan meng-emit event run_started. Caller bertanggung
// jawab memanggil s.launchRun (agar handleRunNow bisa snapshot respons dulu).
// Dipakai bersama oleh handleRunNow dan ProcessDueSchedules (DL-4).
func (s *Server) startScheduleRun(ctx context.Context, sch *schedule.Schedule, scheduleID string, now time.Time, lastRunStatus, eventMsg string) (*agent.TestRun, error) {
	run := &agent.TestRun{
		ID:           uuid.New().String(),
		ProjectPath:  sch.ProjectPath,
		Requirements: sch.Requirements,
		Mode:         sch.Mode,
		State:        agent.StateIdle,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.store.CreateRun(ctx, run); err != nil {
		return nil, err
	}
	s.schedules.Update(scheduleID, func(sc *schedule.Schedule) {
		sc.LastRunAt = &now
		sc.LastRunID = run.ID
		sc.LastRunStatus = lastRunStatus
		sc.NextRunAt = schedule.CalcNextRunInTZ(sc.Frequency, sc.CronExpr, sc.Timezone, now)
	})
	// Trigger async execution — matches handleCreateRun and webhook paths
	s.events.Emit(run.ID, "run_started", "idle", eventMsg, map[string]string{"project": sch.ProjectPath, "mode": sch.Mode, "schedule_id": scheduleID})
	return run, nil
}

func (s *Server) ProcessDueSchedules(ctx context.Context, now time.Time) int {
	processed := 0
	for {
		sch := s.schedules.ClaimNextDue(now, "process-due")
		if sch == nil {
			break
		}
		if sch.TestListID != "" {
			list, err := s.planning.GetTestList(ctx, sch.TestListID)
			if err != nil {
				continue
			}
			runIDs, err := s.startTestListRuns(ctx, list)
			if err != nil {
				continue
			}
			lastRunID := ""
			if len(runIDs) > 0 {
				lastRunID = runIDs[0]
			}
			s.schedules.Update(sch.ID, func(sc *schedule.Schedule) {
				sc.LastRunAt = &now
				sc.LastRunID = lastRunID
				sc.LastRunStatus = string(agent.StateIdle)
				sc.NextRunAt = schedule.CalcNextRunInTZ(sc.Frequency, sc.CronExpr, sc.Timezone, now)
			})
			processed++
			continue
		}

		run, err := s.startScheduleRun(ctx, sch, sch.ID, now, string(agent.StateIdle), "Run created via due schedule")
		if err != nil {
			continue
		}
		s.launchRun(run)
		processed++
	}
	return processed
}
