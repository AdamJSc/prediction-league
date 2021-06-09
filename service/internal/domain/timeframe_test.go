package domain_test

import (
	"prediction-league/service/internal/domain"
	"testing"
	"time"
)

func TestTimeFrame_Valid(t *testing.T) {
	var now = time.Now()
	var oneNanosecondAgo = now.Add(-time.Nanosecond)

	t.Run("empty timeframe must not be valid", func(t *testing.T) {
		var tf domain.TimeFrame

		if tf.Valid() {
			t.Fatalf("expected timeframe %+v to be invalid, but it was valid", tf)
		}
	})

	t.Run("timeframe with only a from timestamp must not be valid", func(t *testing.T) {
		var tf = domain.TimeFrame{
			From: now,
		}

		if tf.Valid() {
			t.Fatalf("expected timeframe %+v to be valid, but it was not valid", tf)
		}
	})

	t.Run("timeframe with only an until timestamp must not be valid", func(t *testing.T) {
		var tf = domain.TimeFrame{
			Until: now,
		}

		if tf.Valid() {
			t.Fatalf("expected timeframe %+v to be valid, but it was not valid", tf)
		}
	})

	t.Run("timeframe with an until timestamp occurring after a from timestamp must be valid", func(t *testing.T) {
		var tf = domain.TimeFrame{
			From:  oneNanosecondAgo,
			Until: now,
		}

		if !tf.Valid() {
			t.Fatalf("expected timeframe %+v to be valid, but it was not valid", tf)
		}
	})

	t.Run("timeframe with a from timestamp occurring after an until timestamp must not be valid", func(t *testing.T) {
		var tf = domain.TimeFrame{
			From:  now,
			Until: oneNanosecondAgo,
		}

		if tf.Valid() {
			t.Fatalf("expected timeframe %+v to not be valid, but it was valid", tf)
		}
	})

	t.Run("timeframe with a from timestamp equal to an until timestamp must not be valid", func(t *testing.T) {
		var tf = domain.TimeFrame{
			From:  now,
			Until: now,
		}

		if tf.Valid() {
			t.Fatalf("expected timeframe %+v to not be valid, but it was valid", tf)
		}
	})
}

func TestTimeFrame_HasBegunBy(t *testing.T) {
	var now = time.Now()
	var oneNanosecondAgo = now.Add(-time.Nanosecond)

	t.Run("timeframe with from timestamp in the past must return true", func(t *testing.T) {
		var tf = domain.TimeFrame{
			From: oneNanosecondAgo,
		}

		if !tf.HasBegunBy(now) {
			t.Fatalf("expected timeframe %+v to have begun by %+v, but it had not", tf, now)
		}
	})

	t.Run("timeframe with from timestamp that matches current timestamp must return true", func(t *testing.T) {
		var tf = domain.TimeFrame{
			From: now,
		}

		if !tf.HasBegunBy(now) {
			t.Fatalf("expected timeframe %+v to have begun by %+v, but it had not", tf, now)
		}
	})

	t.Run("timeframe with from timestamp in the future must return false", func(t *testing.T) {
		var tf = domain.TimeFrame{
			From: now,
		}

		if tf.HasBegunBy(oneNanosecondAgo) {
			t.Fatalf("expected timeframe %+v to have not begun by %+v, but it had", tf, now)
		}
	})
}

func TestTimeFrame_HasElapsedBy(t *testing.T) {
	var now = time.Now()
	var oneNanosecondAgo = now.Add(-time.Nanosecond)

	t.Run("timeframe with until timestamp in the past must return true", func(t *testing.T) {
		var tf = domain.TimeFrame{
			Until: oneNanosecondAgo,
		}

		if !tf.HasElapsedBy(now) {
			t.Fatalf("expected timeframe %+v to have elapsed by %+v, but it had not", tf, now)
		}
	})

	t.Run("timeframe with until timestamp that matches current timestamp must return true", func(t *testing.T) {
		var tf = domain.TimeFrame{
			Until: now,
		}

		if !tf.HasElapsedBy(now) {
			t.Fatalf("expected timeframe %+v to have elapsed by %+v, but it had not", tf, now)
		}
	})

	t.Run("timeframe with until timestamp in the future must return false", func(t *testing.T) {
		var tf = domain.TimeFrame{
			Until: now,
		}

		if tf.HasElapsedBy(oneNanosecondAgo) {
			t.Fatalf("expected timeframe %+v to have not elapsed by %+v, but it had", tf, now)
		}
	})
}

func TestTimeFrame_OverlapsWith(t *testing.T) {
	var now = time.Now()
	var oneNanosecondAgo = now.Add(-time.Nanosecond)
	var twoNanosecondsAgo = now.Add(-2 * time.Nanosecond)
	var threeNanosecondsAgo = now.Add(-3 * time.Nanosecond)

	t.Run("timeframes that do not overlap at all must return false", func(t *testing.T) {
		tf1 := domain.TimeFrame{
			From:  threeNanosecondsAgo,
			Until: twoNanosecondsAgo,
		}
		tf2 := domain.TimeFrame{
			From:  oneNanosecondAgo,
			Until: now,
		}

		if tf1.OverlapsWith(tf2) {
			t.Fatalf("expected timeframe %+v to not overlap with %+v, but it did", tf1, tf2)
		}

		if tf2.OverlapsWith(tf1) {
			t.Fatalf("expected timeframe %+v to not overlap with %+v, but it did", tf2, tf1)
		}
	})

	t.Run("timeframes that are identical must return true", func(t *testing.T) {
		tf1 := domain.TimeFrame{
			From:  oneNanosecondAgo,
			Until: now,
		}
		tf2 := domain.TimeFrame{
			From:  oneNanosecondAgo,
			Until: now,
		}

		if !tf1.OverlapsWith(tf2) {
			t.Fatalf("expected timeframe %+v to overlap with %+v, but it did not", tf1, tf2)
		}

		if !tf2.OverlapsWith(tf1) {
			t.Fatalf("expected timeframe %+v to overlap with %+v, but it did not", tf2, tf1)
		}
	})

	t.Run("timeframes that overlap completely must return true", func(t *testing.T) {
		tf1 := domain.TimeFrame{
			From:  threeNanosecondsAgo,
			Until: now,
		}
		tf2 := domain.TimeFrame{
			From:  twoNanosecondsAgo,
			Until: oneNanosecondAgo,
		}

		if !tf1.OverlapsWith(tf2) {
			t.Fatalf("expected timeframe %+v to overlap with %+v, but it did not", tf1, tf2)
		}

		if !tf2.OverlapsWith(tf1) {
			t.Fatalf("expected timeframe %+v to overlap with %+v, but it did not", tf2, tf1)
		}
	})

	t.Run("timeframes that overlap partially must return true", func(t *testing.T) {
		tf1 := domain.TimeFrame{
			From:  threeNanosecondsAgo,
			Until: oneNanosecondAgo,
		}
		tf2 := domain.TimeFrame{
			From:  twoNanosecondsAgo,
			Until: now,
		}

		if !tf1.OverlapsWith(tf2) {
			t.Fatalf("expected timeframe %+v to overlap with %+v, but it did not", tf1, tf2)
		}

		if !tf2.OverlapsWith(tf1) {
			t.Fatalf("expected timeframe %+v to overlap with %+v, but it did not", tf2, tf1)
		}
	})

	t.Run("timeframes that are consecutive must return false", func(t *testing.T) {
		tf1 := domain.TimeFrame{
			From:  threeNanosecondsAgo,
			Until: twoNanosecondsAgo,
		}
		tf2 := domain.TimeFrame{
			From:  twoNanosecondsAgo,
			Until: oneNanosecondAgo,
		}

		if tf1.OverlapsWith(tf2) {
			t.Fatalf("expected timeframe %+v to not overlap with %+v, but it did", tf1, tf2)
		}

		if tf2.OverlapsWith(tf1) {
			t.Fatalf("expected timeframe %+v to not overlap with %+v, but it did", tf2, tf1)
		}
	})
}

func TestTimeFrame_BeginsWithin(t *testing.T) {
	var now = time.Now()
	var oneNanosecondAgo = now.Add(-time.Nanosecond)
	var twoNanosecondsAgo = now.Add(-2 * time.Nanosecond)
	var threeNanosecondsAgo = now.Add(-3 * time.Nanosecond)

	t.Run("base timeframe that begins within provided timeframe must return true", func(t *testing.T) {
		base := domain.TimeFrame{
			From:  twoNanosecondsAgo,
			Until: oneNanosecondAgo,
		}
		provided := domain.TimeFrame{
			From:  threeNanosecondsAgo,
			Until: now,
		}

		if !base.BeginsWithin(provided) {
			t.Fatalf("expected timeframe %+v to begin within %+v, but it doesn't", base, provided)
		}
	})

	t.Run("base timeframe that does not begin within provided timeframe must return false", func(t *testing.T) {
		base := domain.TimeFrame{
			From:  threeNanosecondsAgo,
			Until: oneNanosecondAgo,
		}
		provided := domain.TimeFrame{
			From:  twoNanosecondsAgo,
			Until: now,
		}

		if base.BeginsWithin(provided) {
			t.Fatalf("expected timeframe %+v to not begin within %+v, but it does", base, provided)
		}
	})

	t.Run("base timeframe that begins at same time as provided timeframe begins must return true", func(t *testing.T) {
		base := domain.TimeFrame{
			From:  threeNanosecondsAgo,
			Until: now,
		}
		provided := domain.TimeFrame{
			From:  threeNanosecondsAgo,
			Until: oneNanosecondAgo,
		}

		if !base.BeginsWithin(provided) {
			t.Fatalf("expected timeframe %+v to begin within %+v, but it doesn't", base, provided)
		}
	})

	t.Run("base timeframe that begins at same time as provided timeframe ends must return true", func(t *testing.T) {
		base := domain.TimeFrame{
			From:  twoNanosecondsAgo,
			Until: now,
		}
		provided := domain.TimeFrame{
			From:  threeNanosecondsAgo,
			Until: twoNanosecondsAgo,
		}

		if !base.BeginsWithin(provided) {
			t.Fatalf("expected timeframe %+v to begin within %+v, but it doesn't", base, provided)
		}
	})
}

func TestTimeFrame_EndsWithin(t *testing.T) {
	var now = time.Now()
	var oneNanosecondAgo = now.Add(-time.Nanosecond)
	var twoNanosecondsAgo = now.Add(-2 * time.Nanosecond)
	var threeNanosecondsAgo = now.Add(-3 * time.Nanosecond)

	t.Run("base timeframe that ends within provided timeframe must return true", func(t *testing.T) {
		base := domain.TimeFrame{
			From:  threeNanosecondsAgo,
			Until: oneNanosecondAgo,
		}
		provided := domain.TimeFrame{
			From:  twoNanosecondsAgo,
			Until: now,
		}

		if !base.EndsWithin(provided) {
			t.Fatalf("expected timeframe %+v to end within %+v, but it doesn't", base, provided)
		}
	})

	t.Run("base timeframe that does not end within provided timeframe must return false", func(t *testing.T) {
		base := domain.TimeFrame{
			From:  oneNanosecondAgo,
			Until: now,
		}
		provided := domain.TimeFrame{
			From:  threeNanosecondsAgo,
			Until: twoNanosecondsAgo,
		}

		if base.EndsWithin(provided) {
			t.Fatalf("expected timeframe %+v to not end within %+v, but it does", base, provided)
		}
	})

	t.Run("base timeframe that ends at same time as provided timeframe begins must return true", func(t *testing.T) {
		base := domain.TimeFrame{
			From:  threeNanosecondsAgo,
			Until: twoNanosecondsAgo,
		}
		provided := domain.TimeFrame{
			From:  twoNanosecondsAgo,
			Until: oneNanosecondAgo,
		}

		if !base.EndsWithin(provided) {
			t.Fatalf("expected timeframe %+v to end within %+v, but it doesn't", base, provided)
		}
	})

	t.Run("base timeframe that ends at same time as provided timeframe ends must return true", func(t *testing.T) {
		base := domain.TimeFrame{
			From:  twoNanosecondsAgo,
			Until: oneNanosecondAgo,
		}
		provided := domain.TimeFrame{
			From:  threeNanosecondsAgo,
			Until: oneNanosecondAgo,
		}

		if !base.EndsWithin(provided) {
			t.Fatalf("expected timeframe %+v to end within %+v, but it doesn't", base, provided)
		}
	})
}
