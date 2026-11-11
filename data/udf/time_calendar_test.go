package udf

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/viant/toolbox/data"
)

func TestFormatTimeCalendarOffset(t *testing.T) {
	for _, tc := range []struct {
		name, input, format, zone, truncate string
		offset                              interface{}
		want                                interface{}
	}{
		{"thirty days", "2026-10-31", "yyyy-MM-dd", "UTC", "", data.Map{"days": -30}, "2026-10-01"},
		{"sixty days", "2026-11-08", "yyyy-MM-dd", "UTC", "", data.Map{"days": -60}, "2026-09-09"},
		{"ninety days", "2026-11-08", "yyyy-MM-dd", "UTC", "", data.Map{"days": -90}, "2026-08-10"},
		{"year boundary", "2026-01-01", "yyyy-MM-dd", "UTC", "", data.Map{"days": -1}, "2025-12-31"},
		{"leap day", "2024-03-01", "yyyy-MM-dd", "UTC", "", data.Map{"days": -1}, "2024-02-29"},
		{"leap year normalization", "2024-02-29", "yyyy-MM-dd", "UTC", "", data.Map{"years": 1}, "2025-03-01"},
		{"anchored previous month", "2026-11-01", "yyyy-MM-dd", "UTC", "", data.Map{"months": -1}, "2026-10-01"},
		{"month normalization", "2026-10-31", "yyyy-MM-dd", "UTC", "", data.Map{"months": -1}, "2026-10-01"},
		{"spring DST", "2026-03-07T12:00:00-05:00", "yyyy-MM-dd HH:mm:ss ZZ", "America/New_York", "", data.Map{"days": 1}, "2026-03-08 12:00:00 -0400"},
		{"fall DST", "2026-10-31T12:00:00-04:00", "yyyy-MM-dd HH:mm:ss ZZ", "America/New_York", "", data.Map{"days": 1}, "2026-11-01 12:00:00 -0500"},
		{"timezone before offset", "2026-11-01T01:00:00Z", "yyyy-MM-dd", "America/New_York", "", data.Map{"days": 1}, "2026-11-01"},
		{"weekday after offset", "2026-10-31", "yyyy-MM-dd", "UTC", "weekday", data.Map{"days": 1}, time.Sunday},
		{"seconds of day", "2026-10-31T12:34:56Z", "secondsOfDay", "UTC", "", data.Map{"days": 1}, 45296},
		{"truncate after offset", "2026-10-31", "yyyy-MM-dd", "UTC", "yyyy-MM", data.Map{"days": 1}, "2026-11-01"},
		{"empty offset", "2026-10-31", "yyyy-MM-dd", "UTC", "", data.Map{}, "2026-10-31"},
		{"JSON integer", "2026-10-31", "yyyy-MM-dd", "UTC", "", data.Map{"days": float64(-30)}, "2026-10-01"},
		{"JSON number", "2026-10-31", "yyyy-MM-dd", "UTC", "", data.Map{"days": json.Number("-30")}, "2026-10-01"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := FormatTime([]interface{}{tc.input, tc.format, tc.zone, tc.truncate, tc.offset}, nil)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestFormatTimeLegacyEmptyTruncation(t *testing.T) {
	got, err := FormatTime([]interface{}{"2026-10-31", "yyyy-MM-dd", "UTC", ""}, nil)
	require.NoError(t, err)
	require.Equal(t, "0000-01-01", got)
	got, err = FormatTime([]interface{}{"2026-10-31", "yyyy-MM-dd", "UTC", "", data.Map{}}, nil)
	require.NoError(t, err)
	require.Equal(t, "2026-10-31", got)
}

func TestFormatTimeCalendarOffsetInvalid(t *testing.T) {
	for _, offset := range []interface{}{
		nil, "1 day", []int{1}, data.Map{"hours": 1}, data.Map{"days": "1"},
		data.Map{"days": 1.5}, data.Map{"days": math.Inf(1)}, data.Map{"days": math.NaN()},
		data.Map{"days": uint64(math.MaxUint64)}, data.Map{"days": float64(math.MaxInt64)},
		data.Map{"days": json.Number("9223372036854775808")}, data.Map{"days": true},
	} {
		_, err := FormatTime([]interface{}{"2026-10-31", "yyyy-MM-dd", "UTC", "", offset}, nil)
		require.Error(t, err, "%#v", offset)
	}
	_, err := FormatTime([]interface{}{"2026-10-31", "yyyy-MM-dd", "No/Such_Zone", "", data.Map{"days": 1}}, nil)
	var parseErr *time.ParseError
	require.Error(t, err)
	_, err = FormatTime([]interface{}{"invalid", "yyyy-MM-dd", "UTC", "", data.Map{"days": 1}}, nil)
	require.True(t, errors.As(err, &parseErr))
}
