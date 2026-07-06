package rpc

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/ucsits/Luce/fsmgr"
)

func (s *Server) ListBlocks(c echo.Context) error {
	s.mu.RLock()
	height := s.chain.Height()
	blocks := make([]BlockResponse, 0, height)
	for i := uint64(0); i < height; i++ {
		blocks = append(blocks, NewBlockResponse(s.chain.GetBlock(i)))
	}
	s.mu.RUnlock()
	return c.JSON(http.StatusOK, blocks)
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
	block := s.chain.AppendBlock(req.Author, req.Data)
	if err := fsmgr.PersistBlock(s.config.DataDir, block); err != nil {
		s.chain.TruncateLast()
		s.mu.Unlock()
		c.Logger().Errorf("persisting block: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to persist block")
	}
	s.mu.Unlock()
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
