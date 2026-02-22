package stats

import (
	"fmt"
	"testing"

	"github.com/wcharczuk/go-chart/v2"
)

func TestService_renderActivityChart(t *testing.T) {
	s := &Service{}

	tests := []struct {
		name    string
		count   int
		maximum float64
	}{
		{
			name:    "Small dataset (7 days)",
			count:   7,
			maximum: 100,
		},
		{
			name:    "Medium dataset (30 days)",
			count:   30,
			maximum: 500,
		},
		{
			name:    "Large dataset (60 days) - Line Chart",
			count:   60,
			maximum: 1000,
		},
		{
			name:    "Empty dataset",
			count:   0,
			maximum: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := make([]chart.Value, tt.count)
			for i := 0; i < tt.count; i++ {
				values[i] = chart.Value{
					Value: float64(i * 10),
					Label: fmt.Sprintf("Day %d", i),
				}
			}

			buf, err := s.renderActivityChart("Test Chart", values, tt.maximum)
			if err != nil {
				t.Errorf("renderActivityChart() error = %v", err)
				return
			}

			if tt.count == 0 {
				if buf != nil {
					t.Error("renderActivityChart() should return nil for empty dataset")
				}
				return
			}

			if buf == nil {
				t.Error("renderActivityChart() returned nil buffer")
				return
			}

			if buf.Len() == 0 {
				t.Error("renderActivityChart() returned empty buffer")
			}
		})
	}
}
