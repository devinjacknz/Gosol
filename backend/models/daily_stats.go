package models

import "time"

type DailyStats struct {
    Date               time.Time `bson:"date"`
    StartBalance       float64   `bson:"start_balance"`
    EndBalance         float64   `bson:"end_balance"`
    TotalVolume        float64   `bson:"total_volume"`  // Deprecated: Use Volume instead
    Volume             float64   `bson:"volume"`
    RealizedPnL        float64   `bson:"realized_pnl"`
    Commissions        float64   `bson:"commissions"`
    TotalTrades        int       `bson:"total_trades"`
    WinningTrades      int       `bson:"winning_trades"`
    LosingTrades       int       `bson:"losing_trades"`
    MaxDrawdown        float64   `bson:"max_drawdown"`
    ProfitFactor       float64   `bson:"profit_factor"`
    WinRate            float64   `bson:"win_rate"`
    AverageWin         float64   `bson:"average_win"`
    AverageLoss        float64   `bson:"average_loss"`
    LargestWin         float64   `bson:"largest_win"`
    LargestLoss        float64   `bson:"largest_loss"`
    MaxConsecWins      int       `bson:"max_consec_wins"`
    MaxConsecLosses    int       `bson:"max_consec_losses"`
    CurrentConsecWins  int       `bson:"current_consec_wins"`
    CurrentConsecLosses int      `bson:"current_consec_losses"`
    SharpeRatio        float64   `bson:"sharpe_ratio"`
}
