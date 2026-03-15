package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	xq "github.com/leonlau/xiangqi-interface"
)

const testFEN = "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"

func newMockChessDBServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		action := r.URL.Query().Get("action")
		switch action {
		case "query":
			w.Write([]byte("move:h2e2|egtb:c3c5|search:b0c2"))
		case "querybest":
			w.Write([]byte("move:h2e2"))
		case "queryscore":
			w.Write([]byte("eval:120"))
		case "querypv":
			w.Write([]byte("score:80,depth:12,pv:h2e2 e7e6 b0c2"))
		case "querysearch":
			w.Write([]byte("search:b0c2"))
		case "queryall":
			w.Write([]byte("move:h2e2,score:50,rank:1,winrate:55%,note:*"))
		case "queryrule":
			w.Write([]byte("move:h2e2,rule:none|move:c3c5,rule:ban"))
		case "queue":
			w.Write([]byte("ok"))
		case "store":
			w.Write([]byte("ok"))
		default:
			http.Error(w, "unknown action", http.StatusBadRequest)
		}
	})
	return httptest.NewServer(mux)
}

func TestChessDBQuery(t *testing.T) {
	srv := newMockChessDBServer(t)
	defer srv.Close()

	impl := &XiangqiEngineImpl{}
	h, err := impl.NewChessDBClient(srv.URL)
	if err != nil {
		t.Fatalf("NewChessDBClient: %v", err)
	}
	defer impl.CloseChessDBClient(h)

	got, err := impl.ChessDBQuery(h, testFEN, xq.ChessDBQueryOptions{})
	if err != nil {
		t.Fatalf("ChessDBQuery: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 suggestions, got %d", len(got))
	}
	if got[0].Kind != "move" || got[0].Move != "h2e2" {
		t.Fatalf("suggestion[0]: %+v", got[0])
	}
}

func TestChessDBQueryBest(t *testing.T) {
	srv := newMockChessDBServer(t)
	defer srv.Close()
	impl := &XiangqiEngineImpl{}
	h, _ := impl.NewChessDBClient(srv.URL)
	defer impl.CloseChessDBClient(h)

	got, err := impl.ChessDBQueryBest(h, testFEN, xq.ChessDBQueryOptions{})
	if err != nil {
		t.Fatalf("ChessDBQueryBest: %v", err)
	}
	if len(got) != 1 || got[0].Move != "h2e2" {
		t.Fatalf("expected [{move h2e2}], got %+v", got)
	}
}

func TestChessDBQueryPV(t *testing.T) {
	srv := newMockChessDBServer(t)
	defer srv.Close()
	impl := &XiangqiEngineImpl{}
	h, _ := impl.NewChessDBClient(srv.URL)
	defer impl.CloseChessDBClient(h)

	pv, err := impl.ChessDBQueryPV(h, testFEN, xq.ChessDBQueryOptions{})
	if err != nil {
		t.Fatalf("ChessDBQueryPV: %v", err)
	}
	if pv.Depth != 12 || len(pv.Moves) != 3 {
		t.Fatalf("PV: %+v", pv)
	}
}

func TestChessDBQueue(t *testing.T) {
	srv := newMockChessDBServer(t)
	defer srv.Close()
	impl := &XiangqiEngineImpl{}
	h, _ := impl.NewChessDBClient(srv.URL)
	defer impl.CloseChessDBClient(h)

	if err := impl.ChessDBQueue(h, testFEN); err != nil {
		t.Fatalf("ChessDBQueue: %v", err)
	}
}

func TestChessDBStore(t *testing.T) {
	srv := newMockChessDBServer(t)
	defer srv.Close()
	impl := &XiangqiEngineImpl{}
	h, _ := impl.NewChessDBClient(srv.URL)
	defer impl.CloseChessDBClient(h)

	if err := impl.ChessDBStore(h, testFEN, "h2e2"); err != nil {
		t.Fatalf("ChessDBStore: %v", err)
	}
}

func TestChessDBClientBaseURL(t *testing.T) {
	srv := newMockChessDBServer(t)
	defer srv.Close()

	impl := &XiangqiEngineImpl{}
	h, _ := impl.NewChessDBClient(srv.URL)
	defer impl.CloseChessDBClient(h)

	// 走通一次请求确认客户端用了 srv.URL
	if _, err := impl.ChessDBQueryScore(h, testFEN, xq.ChessDBQueryOptions{}); err != nil {
		t.Fatalf("QueryScore: %v", err)
	}
}

func TestChessDBInvalidHandle(t *testing.T) {
	impl := &XiangqiEngineImpl{}
	if _, err := impl.ChessDBQuery(xq.InvalidChessDBClientHandle, testFEN, xq.ChessDBQueryOptions{}); err == nil {
		t.Fatal("expected error for invalid handle")
	}
	if err := impl.CloseChessDBClient(xq.ChessDBClientHandle(99999)); err == nil {
		t.Fatal("expected error closing nonexistent handle")
	}
}

func TestChessDBInvalidBoard(t *testing.T) {
	// 走一次完整 client 流程,验证 board 字段传到 URL
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Query().Get("board"), "rnbakabnr") {
			http.Error(w, "bad board", http.StatusBadRequest)
			return
		}
		w.Write([]byte("move:h2e2"))
	}))
	defer srv.Close()

	impl := &XiangqiEngineImpl{}
	h, _ := impl.NewChessDBClient(srv.URL)
	defer impl.CloseChessDBClient(h)

	if _, err := impl.ChessDBQuery(h, testFEN, xq.ChessDBQueryOptions{}); err != nil {
		t.Fatalf("Query: %v", err)
	}
}
