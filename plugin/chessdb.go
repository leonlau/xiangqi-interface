package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	xq "github.com/leonlau/xiangqi-interface"
	"github.com/leonlau/xiangqi/chessapi"
)

var (
	chessdbMu      sync.Mutex
	chessdbCounter int64
	chessdbClients = map[xq.ChessDBClientHandle]*chessapi.Client{}
)

// chessdbTimeout is the per-request HTTP deadline used by all ChessDB queries.
const chessdbTimeout = 30 * time.Second

func getChessDBClient(h xq.ChessDBClientHandle) (*chessapi.Client, error) {
	chessdbMu.Lock()
	defer chessdbMu.Unlock()
	c, ok := chessdbClients[h]
	if !ok {
		return nil, fmt.Errorf("invalid chessdb client handle %d", h)
	}
	return c, nil
}

func (XiangqiEngineImpl) NewChessDBClient(baseURL string) (xq.ChessDBClientHandle, error) {
	c := chessapi.New()
	if baseURL != "" {
		c.BaseURL = baseURL
	}
	chessdbMu.Lock()
	chessdbCounter++
	h := xq.ChessDBClientHandle(chessdbCounter)
	chessdbClients[h] = c
	chessdbMu.Unlock()
	return h, nil
}

func (XiangqiEngineImpl) CloseChessDBClient(h xq.ChessDBClientHandle) error {
	chessdbMu.Lock()
	_, ok := chessdbClients[h]
	if ok {
		delete(chessdbClients, h)
	}
	chessdbMu.Unlock()
	if !ok {
		return fmt.Errorf("invalid chessdb client handle %d", h)
	}
	return nil
}

func chessdbOptsCommon(o xq.ChessDBCommonOptions) chessapi.CommonOptions {
	return chessapi.CommonOptions{
		Ban:        o.Ban,
		EGTBMetric: o.EGTBMetric,
		Learn:      o.Learn,
	}
}

func chessdbOpts(o xq.ChessDBQueryOptions) chessapi.QueryOptions {
	return chessapi.QueryOptions{
		CommonOptions: chessdbOptsCommon(o.ChessDBCommonOptions),
		Endgame:       o.Endgame,
	}
}

func chessdbOptsAll(o xq.ChessDBQueryAllOptions) chessapi.QueryAllOptions {
	return chessapi.QueryAllOptions{
		CommonOptions: chessdbOptsCommon(o.ChessDBCommonOptions),
		ShowAll:       o.ShowAll,
	}
}

func chessdbOptsRule(o xq.ChessDBQueryRuleOptions) chessapi.QueryRuleOptions {
	return chessapi.QueryRuleOptions{Reptimes: o.Reptimes}
}

func chessdbConvertMoves(in []chessapi.SuggestedMove) []xq.ChessDBSuggestedMove {
	out := make([]xq.ChessDBSuggestedMove, len(in))
	for i, m := range in {
		out[i] = xq.ChessDBSuggestedMove{Kind: m.Kind, Move: m.Move}
	}
	return out
}

func chessdbConvertInfos(in []chessapi.MoveInfo) []xq.ChessDBMoveInfo {
	out := make([]xq.ChessDBMoveInfo, len(in))
	for i, m := range in {
		out[i] = xq.ChessDBMoveInfo{
			Move:    m.Move,
			Score:   m.Score,
			Rank:    m.Rank,
			WinRate: m.WinRate,
			Note:    m.Note,
		}
	}
	return out
}

func chessdbCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), chessdbTimeout)
}

func (XiangqiEngineImpl) ChessDBQuery(h xq.ChessDBClientHandle, board string, opts xq.ChessDBQueryOptions) ([]xq.ChessDBSuggestedMove, error) {
	c, err := getChessDBClient(h)
	if err != nil {
		return nil, err
	}
	ctx, cancel := chessdbCtx()
	defer cancel()
	res, err := c.Query(ctx, board, chessdbOpts(opts))
	if err != nil {
		return nil, fmt.Errorf("chessdb query: %w", err)
	}
	return chessdbConvertMoves(res), nil
}

func (XiangqiEngineImpl) ChessDBQueryBest(h xq.ChessDBClientHandle, board string, opts xq.ChessDBQueryOptions) ([]xq.ChessDBSuggestedMove, error) {
	c, err := getChessDBClient(h)
	if err != nil {
		return nil, err
	}
	ctx, cancel := chessdbCtx()
	defer cancel()
	res, err := c.QueryBest(ctx, board, chessdbOpts(opts))
	if err != nil {
		return nil, fmt.Errorf("chessdb querybest: %w", err)
	}
	return chessdbConvertMoves(res), nil
}

func (XiangqiEngineImpl) ChessDBQueryPV(h xq.ChessDBClientHandle, board string, opts xq.ChessDBQueryOptions) (xq.ChessDBPV, error) {
	c, err := getChessDBClient(h)
	if err != nil {
		return xq.ChessDBPV{}, err
	}
	ctx, cancel := chessdbCtx()
	defer cancel()
	res, err := c.QueryPV(ctx, board, chessdbOpts(opts))
	if err != nil {
		return xq.ChessDBPV{}, fmt.Errorf("chessdb querypv: %w", err)
	}
	return xq.ChessDBPV{Score: res.Score, Depth: res.Depth, Moves: res.Moves}, nil
}

func (XiangqiEngineImpl) ChessDBQueryScore(h xq.ChessDBClientHandle, board string, opts xq.ChessDBQueryOptions) (xq.ChessDBScore, error) {
	c, err := getChessDBClient(h)
	if err != nil {
		return xq.ChessDBScore{}, err
	}
	ctx, cancel := chessdbCtx()
	defer cancel()
	res, err := c.QueryScore(ctx, board, chessdbOpts(opts))
	if err != nil {
		return xq.ChessDBScore{}, fmt.Errorf("chessdb queryscore: %w", err)
	}
	return xq.ChessDBScore{Value: res.Value}, nil
}

func (XiangqiEngineImpl) ChessDBQuerySearch(h xq.ChessDBClientHandle, board string, opts xq.ChessDBQueryOptions) ([]xq.ChessDBSuggestedMove, error) {
	c, err := getChessDBClient(h)
	if err != nil {
		return nil, err
	}
	ctx, cancel := chessdbCtx()
	defer cancel()
	res, err := c.QuerySearch(ctx, board, chessdbOpts(opts))
	if err != nil {
		return nil, fmt.Errorf("chessdb querysearch: %w", err)
	}
	return chessdbConvertMoves(res), nil
}

func (XiangqiEngineImpl) ChessDBQueryAll(h xq.ChessDBClientHandle, board string, opts xq.ChessDBQueryAllOptions) ([]xq.ChessDBMoveInfo, error) {
	c, err := getChessDBClient(h)
	if err != nil {
		return nil, err
	}
	ctx, cancel := chessdbCtx()
	defer cancel()
	res, err := c.QueryAll(ctx, board, chessdbOptsAll(opts))
	if err != nil {
		return nil, fmt.Errorf("chessdb queryall: %w", err)
	}
	return chessdbConvertInfos(res), nil
}

func (XiangqiEngineImpl) ChessDBQueryRule(h xq.ChessDBClientHandle, board string, movelist []string, opts xq.ChessDBQueryRuleOptions) ([]xq.ChessDBRuleResult, error) {
	c, err := getChessDBClient(h)
	if err != nil {
		return nil, err
	}
	ctx, cancel := chessdbCtx()
	defer cancel()
	res, err := c.QueryRule(ctx, board, movelist, chessdbOptsRule(opts))
	if err != nil {
		return nil, fmt.Errorf("chessdb queryrule: %w", err)
	}
	out := make([]xq.ChessDBRuleResult, len(res))
	for i, r := range res {
		out[i] = xq.ChessDBRuleResult{Move: r.Move, Rule: r.Rule}
	}
	return out, nil
}

func (XiangqiEngineImpl) ChessDBQueue(h xq.ChessDBClientHandle, board string) error {
	c, err := getChessDBClient(h)
	if err != nil {
		return err
	}
	ctx, cancel := chessdbCtx()
	defer cancel()
	if err := c.Queue(ctx, board); err != nil {
		return fmt.Errorf("chessdb queue: %w", err)
	}
	return nil
}

func (XiangqiEngineImpl) ChessDBStore(h xq.ChessDBClientHandle, board, move string) error {
	c, err := getChessDBClient(h)
	if err != nil {
		return err
	}
	ctx, cancel := chessdbCtx()
	defer cancel()
	if err := c.Store(ctx, board, move); err != nil {
		return fmt.Errorf("chessdb store: %w", err)
	}
	return nil
}
