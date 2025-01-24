package models

type MarketStats struct {
    Symbol         string
    Volume         float64
    Volatility     float64
    Liquidity      float64
    Spread         float64
    PriceChange1h  float64
    PriceChange24h float64
    High24h        float64
    Low24h         float64
}
