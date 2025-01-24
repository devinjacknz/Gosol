package risk

import (
	"math"
	"time"

	"solmeme-trader/models"
)

// NewDailyStats creates a new daily stats instance
func NewDailyStats(date time.Time, startBalance float64) *models.DailyStats {
	return &models.DailyStats{
		Date:         date,
		StartBalance: startBalance,
		EndBalance:   startBalance,
	}
}

// UpdateStats updates daily statistics with a new trade
func UpdateStats(stats *models.DailyStats, pnl float64, commission float64, volume float64) {
	stats.TotalTrades++
	stats.RealizedPnL += pnl
	stats.Commissions += commission
	stats.Volume += volume
	stats.EndBalance = stats.StartBalance + stats.RealizedPnL - stats.Commissions

	if pnl > 0 {
		stats.WinningTrades++
		stats.CurrentConsecWins++
		stats.CurrentConsecLosses = 0
		if pnl > stats.LargestWin {
			stats.LargestWin = pnl
		}
		if stats.CurrentConsecWins > stats.MaxConsecWins {
			stats.MaxConsecWins = stats.CurrentConsecWins
		}
	} else if pnl < 0 {
		stats.LosingTrades++
		stats.CurrentConsecLosses++
		stats.CurrentConsecWins = 0
		if pnl < stats.LargestLoss {
			stats.LargestLoss = pnl
		}
		if stats.CurrentConsecLosses > stats.MaxConsecLosses {
			stats.MaxConsecLosses = stats.CurrentConsecLosses
		}
	}

	updateAverages(stats)
	updateRatios(stats)
}

// updateAverages updates average win/loss statistics
func updateAverages(stats *models.DailyStats) {
	if stats.WinningTrades > 0 {
		totalWins := stats.RealizedPnL - stats.LargestLoss*float64(stats.LosingTrades)
		stats.AverageWin = totalWins / float64(stats.WinningTrades)
	}

	if stats.LosingTrades > 0 {
		totalLosses := stats.RealizedPnL - stats.LargestWin*float64(stats.WinningTrades)
		stats.AverageLoss = totalLosses / float64(stats.LosingTrades)
	}
}

// updateRatios updates trading ratios
func updateRatios(stats *models.DailyStats) {
	if stats.TotalTrades > 0 {
		stats.WinRate = float64(stats.WinningTrades) / float64(stats.TotalTrades) * 100
	}

	if stats.LosingTrades > 0 && stats.AverageLoss != 0 {
		stats.ProfitFactor = (stats.AverageWin * float64(stats.WinningTrades)) / 
			(math.Abs(stats.AverageLoss) * float64(stats.LosingTrades))
	}

	// Simple Sharpe Ratio calculation (assuming risk-free rate = 0)
	if stats.TotalTrades > 0 {
		returns := stats.RealizedPnL / stats.StartBalance
		if returns > 0 {
			stats.SharpeRatio = returns / stats.MaxDrawdown
		}
	}
}

// UpdateDrawdown updates the maximum drawdown
func UpdateDrawdown(stats *models.DailyStats, currentDrawdown float64) {
	if currentDrawdown > stats.MaxDrawdown {
		stats.MaxDrawdown = currentDrawdown
	}
}

// GetROI returns the return on investment percentage
func GetROI(stats *models.DailyStats) float64 {
	if stats.StartBalance == 0 {
		return 0
	}
	return (stats.EndBalance - stats.StartBalance) / stats.StartBalance * 100
}

// GetNetPnL returns the net profit/loss after commissions
func GetNetPnL(stats *models.DailyStats) float64 {
	return stats.RealizedPnL - stats.Commissions
}
