package risk

import (
	"solmeme-trader/models"
)

// FromModelPosition converts a models.Position to a risk.Position
func FromModelPosition(p *models.Position) *Position {
	if p == nil {
		return nil
	}

	var stopLoss, takeProfit *float64
	if p.StopLoss != 0 {
		sl := p.StopLoss
		stopLoss = &sl
	}
	if p.TakeProfit != 0 {
		tp := p.TakeProfit
		takeProfit = &tp
	}

	return &Position{
		ID:           p.ID,
		TokenAddress: p.TokenAddress,
		Side:         p.Side,
		EntryPrice:   p.EntryPrice,
		CurrentPrice: p.CurrentPrice,
		Size:         p.Size,
		Leverage:     p.Leverage,
		PnL:          p.RealizedPnL,
		OpenTime:     p.OpenTime,
		UpdateTime:   p.UpdateTime,
		StopLoss:     stopLoss,
		TakeProfit:   takeProfit,
	}
}

// ToModelPosition converts a risk.Position to a models.Position
func ToModelPosition(p *Position) *models.Position {
	if p == nil {
		return nil
	}

	pos := &models.Position{
		ID:           p.ID,
		TokenAddress: p.TokenAddress,
		Side:         p.Side,
		EntryPrice:   p.EntryPrice,
		CurrentPrice: p.CurrentPrice,
		Size:         p.Size,
		Leverage:     p.Leverage,
		RealizedPnL:  p.PnL,
		OpenTime:     p.OpenTime,
		UpdateTime:   p.UpdateTime,
	}

	if p.StopLoss != nil {
		pos.StopLoss = *p.StopLoss
	}
	if p.TakeProfit != nil {
		pos.TakeProfit = *p.TakeProfit
	}

	return pos
}

// FromModelPositions converts a slice of models.Position to a slice of risk.Position
func FromModelPositions(positions []*models.Position) []*Position {
	if positions == nil {
		return nil
	}

	result := make([]*Position, len(positions))
	for i, p := range positions {
		result[i] = FromModelPosition(p)
	}
	return result
}

// ToModelPositions converts a slice of risk.Position to a slice of models.Position
func ToModelPositions(positions []*Position) []*models.Position {
	if positions == nil {
		return nil
	}

	result := make([]*models.Position, len(positions))
	for i, p := range positions {
		result[i] = ToModelPosition(p)
	}
	return result
}
