package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewHandlerServesCalendarForAnyPath(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			body := `<?xml version="1.0"?><schedule generated="2026-09-25T12:00:00"><My><DAT>2026-09-25T09:00:00</DAT><UR>1</UR><SPGRUP><NAIM>group</NAIM></SPGRUP><SPPRED><NAIM>Математика</NAIM></SPPRED></My></schedule>`
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}),
	}

	handler := NewHandler(client, "group")
	for _, path := range []string{"/", "/bebebe", "/calendar.ics"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))

		if recorder.Code != http.StatusOK {
			t.Fatalf("path %q returned status %d", path, recorder.Code)
		}
		if !strings.HasPrefix(recorder.Header().Get("Content-Type"), "text/calendar") {
			t.Fatalf("path %q returned content type %q", path, recorder.Header().Get("Content-Type"))
		}
		if !strings.Contains(recorder.Body.String(), "BEGIN:VCALENDAR") {
			t.Fatalf("path %q did not return a calendar", path)
		}
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
