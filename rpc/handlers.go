package rpc

import (
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/ucsits/Luce/fsmgr"
)

func (s *Server) ListBlocks(c echo.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Parse pagination query params
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	desc, err := strconv.ParseBool(c.QueryParam("desc"))
	if err != nil {
		desc = false
	}

	height := s.chain.Height()
	total := height

	start := uint64((page - 1) * limit)
	if start >= height {
		return c.JSON(http.StatusOK, PaginatedLightweightBlocksResponse{
			Data: []LightweightBlockResponse{},
			Pagination: PaginationMeta{
				Page:       page,
				Limit:      limit,
				Total:      total,
				TotalPages: int((total + uint64(limit) - 1) / uint64(limit)),
			},
		})
	}

	end := start + uint64(limit)
	if end > height {
		end = height
	}

	blocks := make([]LightweightBlockResponse, 0, end-start)
	if desc {
		// Reverse chronological order (newest first) — suitable for block explorers
		for j := uint64(0); j < end-start; j++ {
			idx := height - 1 - start - j
			blocks = append(blocks, NewLightweightBlockResponse(s.chain.GetBlock(idx)))
		}
	} else {
		// Chronological order (oldest first) — default
		for i := start; i < end; i++ {
			blocks = append(blocks, NewLightweightBlockResponse(s.chain.GetBlock(i)))
		}
	}

	return c.JSON(http.StatusOK, PaginatedLightweightBlocksResponse{
		Data: blocks,
		Pagination: PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: int((total + uint64(limit) - 1) / uint64(limit)),
		},
	})
}

func (s *Server) GetBlockByHash(c echo.Context) error {
	hashStr := c.Param("hash")
	hashBytes, err := hex.DecodeString(hashStr)
	if err != nil || len(hashBytes) != 32 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid hash")
	}
	var hash [32]byte
	copy(hash[:], hashBytes)

	s.mu.RLock()
	defer s.mu.RUnlock()

	block, found := s.chain.GetBlockByHash(hash)
	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "block not found")
	}
	return c.JSON(http.StatusOK, NewBlockResponse(*block))
}

func (s *Server) ChainSummary(c echo.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	height := s.chain.Height()
	resp := ChainSummaryResponse{
		Height: height,
		Blocks: height,
	}

	if last, ok := s.chain.LastBlock(); ok {
		blockResp := NewBlockResponse(*last)
		resp.BestBlockHash = blockResp.Hash
		resp.LastBlock = &blockResp
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) GetBlock(c echo.Context) error {
	heightStr := c.Param("height")
	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid height")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if height >= s.chain.Height() {
		return echo.NewHTTPError(http.StatusNotFound, "block not found")
	}
	return c.JSON(http.StatusOK, NewBlockResponse(s.chain.GetBlock(height)))
}

func (s *Server) AppendBlock(c echo.Context) error {
	var req AppendBlockRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Data == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "data must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	block := s.chain.AppendBlock(req.Author, req.Data)
	if err := fsmgr.PersistBlock(s.config.DataDir, block); err != nil {
		s.chain.TruncateLast()
		c.Logger().Errorf("persisting block: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to persist block")
	}
	return c.JSON(http.StatusCreated, NewBlockResponse(block))
}

func (s *Server) ValidateChain(c echo.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return c.JSON(http.StatusOK, map[string]bool{"valid": s.chain.Validate()})
}

func (s *Server) GetHeight(c echo.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return c.JSON(http.StatusOK, map[string]uint64{"height": s.chain.Height()})
}
