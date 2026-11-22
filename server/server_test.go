package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testXML = `<dataroot>
	<My>
		<ID>1</ID><DAT>2026-09-10T00:00:00</DAT><UR>1</UR>
		<FAMIO>Иванов И.И.</FAMIO><SPPRED.NAIM>Предмет А</SPPRED.NAIM>
		<SPGRUP.NAIM>1234</SPGRUP.NAIM><AUD>101</AUD><CAMPUS>Э</CAMPUS>
	</My>
	<My>
		<ID>2</ID><DAT>2026-09-10T00:00:00</DAT><UR>2</UR>
		<FAMIO>Петров П.П.</FAMIO><SPPRED.NAIM>Предмет Б</SPPRED.NAIM>
		<SPGRUP.NAIM>5678</SPGRUP.NAIM><AUD>202</AUD><CAMPUS>Э</CAMPUS>
	</My>
</dataroot>`

func newSource(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func get(s *Server, method, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	return rec
}

func TestServeSchedule_NoGroup(t *testing.T) {
	s := &Server{Source: "http://unused", Loc: time.UTC}

	rec := get(s, http.MethodGet, "/schedule.ics")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("получен код %d, ожидался %d", rec.Code, http.StatusBadRequest)
	}
}

func TestServeSchedule_WrongMethod(t *testing.T) {
	s := &Server{Source: "http://unused", Loc: time.UTC}

	rec := get(s, http.MethodPost, "/schedule.ics?group=1234")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("получен код %d, ожидался %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestServeSchedule_OK(t *testing.T) {
	src := newSource(t, http.StatusOK, testXML)
	s := &Server{Source: src.URL, Loc: time.UTC}

	rec := get(s, http.MethodGet, "/schedule.ics?group=1234")

	if rec.Code != http.StatusOK {
		t.Fatalf("получен код %d, ожидался 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/calendar") {
		t.Errorf("Content-Type = %q, ожидался text/calendar", ct)
	}

	body := rec.Body.String()
	if n := strings.Count(body, "BEGIN:VEVENT"); n != 1 {
		t.Errorf("ожидалось 1 событие, получено %d\n%s", n, body)
	}
	if !strings.Contains(body, "SUMMARY:Предмет А\r\n") {
		t.Errorf("нет занятия своей группы:\n%s", body)
	}
	if strings.Contains(body, "Предмет Б") {
		t.Errorf("в календарь попало занятие чужой группы:\n%s", body)
	}
}

func TestServeSchedule_UnknownGroup(t *testing.T) {
	src := newSource(t, http.StatusOK, testXML)
	s := &Server{Source: src.URL, Loc: time.UTC}

	rec := get(s, http.MethodGet, "/schedule.ics?group=0000")

	if rec.Code != http.StatusOK {
		t.Fatalf("получен код %d, ожидался 200", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "BEGIN:VEVENT") {
		t.Error("для несуществующей группы событий быть не должно")
	}
}

func TestServeSchedule_SourceDown(t *testing.T) {
	src := newSource(t, http.StatusInternalServerError, "boom")
	s := &Server{Source: src.URL, Loc: time.UTC}

	rec := get(s, http.MethodGet, "/schedule.ics?group=1234")
	if rec.Code != http.StatusBadGateway {
		t.Errorf("получен код %d, ожидался %d", rec.Code, http.StatusBadGateway)
	}
}

func TestServeSchedule_BrokenXML(t *testing.T) {
	src := newSource(t, http.StatusOK, `<dataroot"`)
	s := &Server{Source: src.URL, Loc: time.UTC}

	rec := get(s, http.MethodGet, "/schedule.ics?group=1234")
	if rec.Code != http.StatusBadGateway {
		t.Errorf("получен код %d, ожидался %d", rec.Code, http.StatusBadGateway)
	}
}
