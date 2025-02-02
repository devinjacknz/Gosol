module github.com/kwanRoshi/Gosol/backend/trading/analysis/storage

go 1.23.5

replace github.com/kwanRoshi/Gosol/backend/trading/analysis/streaming => ../streaming

require github.com/kwanRoshi/Gosol/backend/trading/analysis/streaming v0.0.0-00010101000000-000000000000

require github.com/lib/pq v1.10.9
