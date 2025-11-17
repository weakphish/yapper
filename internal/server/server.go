package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jack/yapper/go-note/internal/core"
	"github.com/jack/yapper/go-note/internal/logging"
	"github.com/jack/yapper/go-note/internal/rpc"
)

// Run launches the blocking stdio JSON-RPC loop.
func Run(domain *core.Domain) error {
	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var request rpc.Request
		if err := json.Unmarshal([]byte(line), &request); err != nil {
			logging.Warnf("malformed JSON: %v", err)
			resp := rpc.ResponseError(rpc.NullID(), rpc.ParseError(err.Error()))
			if err := writeResponse(writer, resp); err != nil {
				return err
			}
			continue
		}

		resp, ok := handleRequest(domain, request)
		if ok {
			if err := writeResponse(writer, resp); err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stdin read error: %w", err)
	}
	logging.Infof("stdin closed, shutting down")
	return nil
}

func writeResponse(w *bufio.Writer, resp rpc.Response) error {
	payload, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	if _, err := w.Write(payload); err != nil {
		return err
	}
	if err := w.WriteByte('\n'); err != nil {
		return err
	}
	return w.Flush()
}

func handleRequest(domain *core.Domain, req rpc.Request) (rpc.Response, bool) {
	id := rpc.NullID()
	if req.ID != nil {
		id = *req.ID
	}

	if req.JSONRPC != "2.0" {
		return rpc.ResponseError(id, rpc.InvalidRequest("jsonrpc must be \"2.0\"")), true
	}

	result, err := dispatch(domain, req.Method, req.Params)
	if err.Code != 0 {
		if req.ID == nil {
			logging.Warnf("notification for method %q failed: %v", req.Method, err)
			return rpc.Response{}, false
		}
		return rpc.ResponseError(id, err), true
	}

	if req.ID == nil {
		return rpc.Response{}, false
	}
	return rpc.ResponseResult(id, result), true
}

func dispatch(domain *core.Domain, method string, params json.RawMessage) (interface{}, rpc.Error) {
	switch method {
	case "core.reindex":
		if err := domain.ReindexAll(); err != nil {
			return nil, rpc.ServerError(err.Error())
		}
		return map[string]string{"status": "ok"}, rpc.Error{}
	case "core.list_tasks":
		payload, err := rpc.ParseParams[rpc.ListTasksParams](params)
		if err.Code != 0 {
			return nil, err
		}
		filter := &core.TaskFilter{}
		if payload.Status != nil {
			filter.Status = payload.Status
		}
		if len(payload.Tags) > 0 {
			filter.Tags = payload.Tags
		}
		if payload.TextSearch != nil {
			filter.TextSearch = payload.TextSearch
		}
		if payload.TouchedSince != nil && *payload.TouchedSince != "" {
			date, err := rpc.ParseDate(*payload.TouchedSince)
			if err.Code != 0 {
				return nil, err
			}
			filter.TouchedSince = &date
		}
		return domain.ListTasks(filter), rpc.Error{}
	case "core.task_detail":
		payload, err := rpc.ParseParams[rpc.TaskDetailParams](params)
		if err.Code != 0 {
			return nil, err
		}
		taskID := core.TaskID(payload.TaskID)
		task, mentions, ok := domain.TaskDetail(taskID)
		if !ok {
			return nil, rpc.InvalidRequest("task not found")
		}
		logEntries := domain.LogEntriesForTask(taskID)
		return map[string]interface{}{
			"task":        task,
			"mentions":    mentions,
			"log_entries": logEntries,
		}, rpc.Error{}
	case "core.items_for_tag":
		payload, err := rpc.ParseParams[rpc.TagParams](params)
		if err.Code != 0 {
			return nil, err
		}
		return domain.ItemsForTag(payload.Tag), rpc.Error{}
	case "core.notes_in_range":
		payload, err := rpc.ParseParams[rpc.RangeParams](params)
		if err.Code != 0 {
			return nil, err
		}
		rangeSel, err2 := rpc.ParseRange(payload.Start, payload.End)
		if err2.Code != 0 {
			return nil, err2
		}
		return domain.NotesInRange(&rangeSel), rpc.Error{}
	case "core.weekly_summary":
		payload, err := rpc.ParseParams[rpc.RangeParams](params)
		if err.Code != 0 {
			return nil, err
		}
		rangeSel, err2 := rpc.ParseRange(payload.Start, payload.End)
		if err2.Code != 0 {
			return nil, err2
		}
		return domain.WeeklySummary(&rangeSel), rpc.Error{}
	case "core.open_daily":
		payload, err := rpc.ParseParams[rpc.OpenDailyParams](params)
		if err.Code != 0 {
			return nil, err
		}
		date, err2 := rpc.ParseDate(payload.Date)
		if err2.Code != 0 {
			return nil, err2
		}
		note, noteErr := domain.OpenDaily(date)
		if noteErr != nil {
			return nil, rpc.ServerError(noteErr.Error())
		}
		return note, rpc.Error{}
	case "core.read_note":
		payload, err := rpc.ParseParams[rpc.NoteParams](params)
		if err.Code != 0 {
			return nil, err
		}
		if note, ok := domain.ReadNote(core.NoteID(payload.NoteID)); ok {
			return note, rpc.Error{}
		}
		return nil, rpc.InvalidRequest("note not found")
	case "core.write_note":
		payload, err := rpc.ParseParams[rpc.WriteNoteParams](params)
		if err.Code != 0 {
			return nil, err
		}
		note, writeErr := domain.WriteNote(core.NoteID(payload.NoteID), payload.Content)
		if writeErr != nil {
			return nil, rpc.ServerError(writeErr.Error())
		}
		return note, rpc.Error{}
	default:
		return nil, rpc.MethodNotFound(method)
	}
}
