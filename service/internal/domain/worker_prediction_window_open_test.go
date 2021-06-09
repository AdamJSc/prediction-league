package domain_test

import (
	"prediction-league/service/internal/domain"
	"testing"
	"time"
)

func TestGenerateTimeFrameForPredictionWindowOpenQuery(t *testing.T) {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("generated timeframe for prediction window open cron must match expected", func(t *testing.T) {
		baseTime := time.Date(2018, 5, 26, 14, 0, 0, 0, loc)

		expectedTimeframe := domain.TimeFrame{
			// 24 hours prior to base time
			From: time.Date(2018, 5, 25, 14, 0, 0, 0, loc),
			// one minute prior to base time
			Until: time.Date(2018, 5, 26, 13, 59, 0, 0, loc),
		}

		actualTimeFrame := domain.GenerateTimeFrameForPredictionWindowOpenQuery(baseTime)

		if expectedTimeframe != actualTimeFrame {
			t.Fatalf("expected %+v, got %+v", expectedTimeframe, actualTimeFrame)
		}
	})
}
