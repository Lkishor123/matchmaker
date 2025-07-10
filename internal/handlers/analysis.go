package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"matchmaker/internal/logging"
)

// AnalysisRequest represents the payload for compatibility analysis.
type AnalysisRequest struct {
	PersonA BirthDetails `json:"personA"`
	PersonB BirthDetails `json:"personB"`
}

// CreateAnalysis handles POST /api/v1/analysis.
func CreateAnalysis(c *gin.Context) {
	start := time.Now()
	logging.Log.Info("analysis request started")

	var req AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Log.WithError(err).Warn("invalid analysis payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	reportURL := os.Getenv("REPORT_SERVICE_URL")
	if reportURL == "" {
		reportURL = "http://localhost:8082"
	}
	endpoint := reportURL + "/internal/v1/reports"

	var wg sync.WaitGroup
	wg.Add(2)
	reports := make([][]byte, 2)
	errs := make([]error, 2)

	fetch := func(idx int, bd BirthDetails) {
		defer wg.Done()
		body, err := json.Marshal(bd)
		if err != nil {
			errs[idx] = err
			return
		}
		resp, err := http.Post(endpoint, "application/json", bytes.NewReader(body))
		if err != nil {
			errs[idx] = err
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			errs[idx] = fmt.Errorf("report service status %d: %s", resp.StatusCode, string(b))
			return
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			errs[idx] = err
			return
		}
		reports[idx] = data
	}

	go fetch(0, req.PersonA)
	go fetch(1, req.PersonB)
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			logging.Log.WithError(err).Error("failed to fetch report")
			c.JSON(http.StatusBadGateway, gin.H{"error": "report service error"})
			logging.Log.WithField("latency", time.Since(start)).Info("analysis request finished")
			return
		}
	}

	result := calculateCompatibility(reports[0], reports[1])
	c.JSON(http.StatusOK, result)

	logging.Log.WithField("latency", time.Since(start)).Info("analysis request finished")
}

// calculateCompatibility computes a simple compatibility score from two reports.
func calculateCompatibility(repA, repB []byte) map[string]interface{} {
	diff := len(repA) - len(repB)
	if diff < 0 {
		diff = -diff
	}
	score := 100 - diff%100
	if score < 0 {
		score = 0
	}
	return map[string]interface{}{"score": score}
}
