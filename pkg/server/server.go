package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/SubhanAfz/scraper/pkg/browser"
	"github.com/SubhanAfz/scraper/pkg/conversion"
)

type Server struct {
	BrowserService browser.BrowserService
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeJsonError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
}

func (s *Server) GetPageHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	url := r.URL.Query().Get("url")
	if url == "" {
		writeJsonError(w, http.StatusBadRequest, fmt.Errorf("url parameter is required"))
		return
	}

	waitTimeStr := r.URL.Query().Get("wait_time")
	var waitTime uint64 = 1000 // default 1 second
	if waitTimeStr != "" {
		if parsedWaitTime, err := strconv.ParseUint(waitTimeStr, 10, 64); err != nil {
			writeJsonError(w, http.StatusBadRequest, fmt.Errorf("invalid wait_time parameter: %s", err.Error()))
			return
		} else {
			waitTime = parsedWaitTime
		}
	}

	format := r.URL.Query().Get("format")

	// Create the request
	pageReq := browser.GetPage{
		URL:      url,
		WaitTime: waitTime,
	}

	page, err := s.BrowserService.GetPage(pageReq)
	if err != nil {
		writeJsonError(w, http.StatusInternalServerError, err)
		return
	}

	if format != "" {
		if conversionService, exists := conversion.GetService(format); exists {
			page, err = conversionService.Convert(page)
			if err != nil {
				writeJsonError(w, http.StatusInternalServerError, err)
				return
			}
		} else {
			writeJsonError(w, http.StatusBadRequest, fmt.Errorf("unsupported format: %s", format))
			return
		}
	}

	resp := browser.Page{
		Title:   page.Title,
		Content: page.Content,
		URL:     page.URL,
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) ScreenShotHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	url := r.URL.Query().Get("url")
	if url == "" {
		writeJsonError(w, http.StatusBadRequest, fmt.Errorf("url parameter is required"))
		return
	}

	waitTimeStr := r.URL.Query().Get("wait_time")
	var waitTime uint64 = 1000 // default 1 second
	if waitTimeStr != "" {
		if parsedWaitTime, err := strconv.ParseUint(waitTimeStr, 10, 64); err != nil {
			writeJsonError(w, http.StatusBadRequest, fmt.Errorf("invalid wait_time parameter: %s", err.Error()))
			return
		} else {
			waitTime = parsedWaitTime
		}
	}

	// Create the request
	req := browser.GetScreenShotRequest{
		URL:      url,
		WaitTime: waitTime,
	}

	resp, err := s.BrowserService.ScreenShot(req)
	if err != nil {
		writeJsonError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
