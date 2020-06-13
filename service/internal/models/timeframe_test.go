package models_test

import (
	"prediction-league/service/internal/models"
	"testing"
	"time"
)

func TestTimeFrame_Valid(t *testing.T) {
	var now = time.Now()
	var oneNanosecondAgo = now.Add(-time.Nanosecond)

	t.Run("empty timeframe must not be valid", func(t *testing.T) {
		var tf models.TimeFrame

		if tf.Valid() {
			t.Fatalf("expected timeframe %+v to be invalid, but it was valid", tf)
		}
	})

	t.Run("timeframe with only a from timestamp must be valid", func(t *testing.T) {
		var tf = models.TimeFrame{
			From: &now,
		}

		if !tf.Valid() {
			t.Fatalf("expected timeframe %+v to be valid, but it was not valid", tf)
		}
	})

	t.Run("timeframe with only an until timestamp must be valid", func(t *testing.T) {
		var tf = models.TimeFrame{
			Until: &now,
		}

		if !tf.Valid() {
			t.Fatalf("expected timeframe %+v to be valid, but it was not valid", tf)
		}
	})

	t.Run("timeframe with an until timestamp occurring after a from timestamp must be valid", func(t *testing.T) {
		var tf = models.TimeFrame{
			From:  &oneNanosecondAgo,
			Until: &now,
		}

		if !tf.Valid() {
			t.Fatalf("expected timeframe %+v to be valid, but it was not valid", tf)
		}
	})

	t.Run("timeframe with a from timestamp occurring after an until timestamp must not be valid", func(t *testing.T) {
		var tf = models.TimeFrame{
			From:  &now,
			Until: &oneNanosecondAgo,
		}

		if tf.Valid() {
			t.Fatalf("expected timeframe %+v to not be valid, but it was valid", tf)
		}
	})

	t.Run("timeframe with a from timestamp equal to an until timestamp must not be valid", func(t *testing.T) {
		var tf = models.TimeFrame{
			From:  &now,
			Until: &now,
		}

		if tf.Valid() {
			t.Fatalf("expected timeframe %+v to not be valid, but it was valid", tf)
		}
	})
}

func TestTimeFrame_HasBegunBy(t *testing.T) {
	var now = time.Now()
	var oneNanosecondAgo = now.Add(-time.Nanosecond)

	t.Run("timeframe with empty from timestamp must return true", func(t *testing.T) {
		var tf = models.TimeFrame{
			Until: &now,
		}

		if !tf.HasBegunBy(now) {
			t.Fatalf("expected timeframe %+v to have begun by %+v, but it had not", tf, now)
		}
	})

	t.Run("timeframe with from timestamp in the past must return true", func(t *testing.T) {
		var tf = models.TimeFrame{
			From: &oneNanosecondAgo,
		}

		if !tf.HasBegunBy(now) {
			t.Fatalf("expected timeframe %+v to have begun by %+v, but it had not", tf, now)
		}
	})

	t.Run("timeframe with from timestamp in the future must return false", func(t *testing.T) {
		var tf = models.TimeFrame{
			From: &now,
		}

		if tf.HasBegunBy(oneNanosecondAgo) {
			t.Fatalf("expected timeframe %+v to have not begun by %+v, but it had", tf, now)
		}
	})

	t.Run("timeframe with from timestamp that matches current timestamp must return true", func(t *testing.T) {
		var tf = models.TimeFrame{
			From: &now,
		}

		if !tf.HasBegunBy(now) {
			t.Fatalf("expected timeframe %+v to have begun by %+v, but it had not", tf, now)
		}
	})
}

func TestTimeFrame_HasElapsedBy(t *testing.T) {
	var now = time.Now()
	var oneNanosecondAgo = now.Add(-time.Nanosecond)

	t.Run("timeframe with empty until timestamp must return false", func(t *testing.T) {
		var tf = models.TimeFrame{
			From: &now,
		}

		if tf.HasElapsedBy(now) {
			t.Fatalf("expected timeframe %+v to have not elapsed by %+v, but it had", tf, now)
		}
	})

	t.Run("timeframe with until timestamp in the past must return true", func(t *testing.T) {
		var tf = models.TimeFrame{
			Until: &oneNanosecondAgo,
		}

		if !tf.HasElapsedBy(now) {
			t.Fatalf("expected timeframe %+v to have elapsed by %+v, but it had not", tf, now)
		}
	})

	t.Run("timeframe with until timestamp in the future must return false", func(t *testing.T) {
		var tf = models.TimeFrame{
			Until: &now,
		}

		if tf.HasElapsedBy(oneNanosecondAgo) {
			t.Fatalf("expected timeframe %+v to have not elapsed by %+v, but it had", tf, now)
		}
	})

	t.Run("timeframe with until timestamp that matches current timestamp must return false", func(t *testing.T) {
		var tf = models.TimeFrame{
			Until: &now,
		}

		if tf.HasElapsedBy(now) {
			t.Fatalf("expected timeframe %+v to have not elapsed by %+v, but it had", tf, now)
		}
	})
}
