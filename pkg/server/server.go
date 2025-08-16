package server

import (
	"encoding/json"
	"fmt"
	"net/http"

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

type GetPageRequest struct {
	browser.GetPage `json:"get_page"`
	Format          string `json:"format"` // format to convert the page content to
}

func (s *Server) GetPageHandler(w http.ResponseWriter, r *http.Request) {
	var req GetPageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJsonError(w, http.StatusBadRequest, err)
		return
	}

	page, err := s.BrowserService.GetPage(req.GetPage)
	if err != nil {
		writeJsonError(w, http.StatusInternalServerError, err)
		return
	}

	if req.Format != "" {
		if conversionService, exists := conversion.GetService(req.Format); exists {
			page, err = conversionService.Convert(page)
			if err != nil {
				writeJsonError(w, http.StatusInternalServerError, err)
				return
			}
		} else {
			writeJsonError(w, http.StatusBadRequest, fmt.Errorf("unsupported format: %s", req.Format))
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
	var req browser.GetScreenShotRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJsonError(w, http.StatusBadRequest, err)
		return
	}

	resp, err := s.BrowserService.ScreenShot(req)
	if err != nil {
		writeJsonError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
