package benchmark

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/android-lewis/dbsmith/internal/tui/components"
)

// generateQueryResult creates a mock QueryResult with the specified dimensions.
func generateQueryResult(rows, cols int) *models.QueryResult {
	columns := make([]string, cols)
	columnTypes := make([]string, cols)
	for i := 0; i < cols; i++ {
		columns[i] = fmt.Sprintf("column_%d", i)
		columnTypes[i] = "VARCHAR"
	}

	data := make([][]interface{}, rows)
	for i := 0; i < rows; i++ {
		row := make([]interface{}, cols)
		for j := 0; j < cols; j++ {
			// Mix of different data types
			switch j % 4 {
			case 0:
				row[j] = fmt.Sprintf("text_value_%d_%d", i, j)
			case 1:
				row[j] = i*j + j
			case 2:
				row[j] = float64(i) * 1.23
			case 3:
				row[j] = nil // NULL values
			}
		}
		data[i] = row
	}

	return &models.QueryResult{
		Columns:     columns,
		ColumnTypes: columnTypes,
		Rows:        data,
		RowCount:    int64(rows),
		ExecutionMs: 100,
	}
}

// BenchmarkNewQueryResultContent measures the initialization time for different result sizes.
func BenchmarkNewQueryResultContent(b *testing.B) {
	testCases := []struct {
		name string
		rows int
		cols int
	}{
		{"Small_100x10", 100, 10},
		{"Medium_1000x20", 1000, 20},
		{"Large_10000x20", 10000, 20},
		{"Wide_1000x50", 1000, 50},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			result := generateQueryResult(tc.rows, tc.cols)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = components.NewQueryResultContent(result)
			}
		})
	}
}

// BenchmarkGetCellSequential measures sequential cell access performance.
func BenchmarkGetCellSequential(b *testing.B) {
	testCases := []struct {
		name string
		rows int
		cols int
	}{
		{"Small_100x10", 100, 10},
		{"Medium_1000x20", 1000, 20},
		{"Large_10000x20", 10000, 20},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			result := generateQueryResult(tc.rows, tc.cols)
			content := components.NewQueryResultContent(result)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Access cells sequentially (simulating scrolling)
				row := i % tc.rows
				col := i % tc.cols
				_ = content.GetCell(row, col)
			}
		})
	}
}

// BenchmarkGetCellRandom measures random cell access performance (worst case for cache).
func BenchmarkGetCellRandom(b *testing.B) {
	testCases := []struct {
		name string
		rows int
		cols int
	}{
		{"Small_100x10", 100, 10},
		{"Medium_1000x20", 1000, 20},
		{"Large_10000x20", 10000, 20},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			result := generateQueryResult(tc.rows, tc.cols)
			content := components.NewQueryResultContent(result)

			// Pre-generate random positions to avoid measuring RNG overhead
			positions := make([][2]int, b.N)
			for i := 0; i < b.N; i++ {
				positions[i] = [2]int{
					rand.Intn(tc.rows),
					rand.Intn(tc.cols),
				}
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				pos := positions[i]
				_ = content.GetCell(pos[0], pos[1])
			}
		})
	}
}

// BenchmarkGetCellCacheHits measures performance when cache hits are high.
func BenchmarkGetCellCacheHits(b *testing.B) {
	testCases := []struct {
		name    string
		rows    int
		cols    int
		hotRows int // Number of rows to repeatedly access
		hotCols int // Number of columns to repeatedly access
	}{
		{"HotSpot_10x5_in_1000x20", 1000, 20, 10, 5},
		{"HotSpot_50x10_in_10000x20", 10000, 20, 50, 10},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			result := generateQueryResult(tc.rows, tc.cols)
			content := components.NewQueryResultContent(result)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Access only a small subset of cells repeatedly
				row := i % tc.hotRows
				col := i % tc.hotCols
				_ = content.GetCell(row, col)
			}
		})
	}
}

// BenchmarkLRUCacheOperations measures raw LRU cache performance.
func BenchmarkLRUCacheOperations(b *testing.B) {
	testCases := []struct {
		name      string
		cacheSize int
		ops       int // Number of unique operations
	}{
		{"Cache_1000_Ops_100", 1000, 100},
		{"Cache_4000_Ops_500", 4000, 500},
		{"Cache_4000_Ops_5000", 4000, 5000}, // Heavy eviction
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			cache := components.NewLRUCellCache(tc.cacheSize)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				row := i % tc.ops
				col := (i / tc.ops) % 10
				cache.Set(row, col, nil) // Use nil to minimize allocation overhead
				_ = cache.Get(row, col)
			}
		})
	}
}

// BenchmarkApplyAlternatingRowColors measures the overhead of alternating row colors.
func BenchmarkApplyAlternatingRowColors(b *testing.B) {
	result := generateQueryResult(1000, 20)

	b.Run("WithoutAlternating", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			content := components.NewQueryResultContent(result)
			for row := 0; row < 100; row++ {
				for col := 0; col < 20; col++ {
					_ = content.GetCell(row, col)
				}
			}
		}
	})

	b.Run("WithAlternating", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			content := components.NewQueryResultContent(result)
			content.ApplyAlternatingRowColors()
			for row := 0; row < 100; row++ {
				for col := 0; col < 20; col++ {
					_ = content.GetCell(row, col)
				}
			}
		}
	})
}

// BenchmarkCellValueFormatting measures the overhead of different value types.
func BenchmarkCellValueFormatting(b *testing.B) {
	b.Run("ShortStrings", func(b *testing.B) {
		result := generateQueryResult(1000, 10)
		content := components.NewQueryResultContent(result)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = content.GetCell(i%1000, i%10)
		}
	})

	b.Run("LongStrings_Truncated", func(b *testing.B) {
		result := generateQueryResult(1000, 10)
		// Create long strings that will be truncated
		for i := 0; i < len(result.Rows); i++ {
			for j := 0; j < len(result.Rows[i]); j++ {
				result.Rows[i][j] = string(make([]byte, 200)) // Exceeds MaxCellDisplayLen
			}
		}
		content := components.NewQueryResultContent(result)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = content.GetCell(i%1000, i%10)
		}
	})

	b.Run("MixedTypes", func(b *testing.B) {
		result := generateQueryResult(1000, 10)
		content := components.NewQueryResultContent(result)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = content.GetCell(i%1000, i%10)
		}
	})
}

// BenchmarkFullTableRender simulates rendering a visible portion of a large table.
func BenchmarkFullTableRender(b *testing.B) {
	testCases := []struct {
		name        string
		totalRows   int
		totalCols   int
		visibleRows int
		visibleCols int
	}{
		{"Viewport_10x10_in_1000x20", 1000, 20, 10, 10},
		{"Viewport_50x20_in_10000x50", 10000, 50, 50, 20},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			result := generateQueryResult(tc.totalRows, tc.totalCols)
			content := components.NewQueryResultContent(result)
			content.ApplyAlternatingRowColors()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Simulate scrolling through the table
				startRow := (i * 10) % (tc.totalRows - tc.visibleRows)
				for row := startRow; row < startRow+tc.visibleRows; row++ {
					for col := 0; col < tc.visibleCols; col++ {
						_ = content.GetCell(row, col)
					}
				}
			}
		})
	}
}
