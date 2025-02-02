package risk

import (
	"context"
	"sync"
	"time"

	"github.com/leonzhao/gosol/backend/trading/analysis/monitoring"
)

// RiskLevel represents the risk level
type RiskLevel int

const (
	Low RiskLevel = iota + 1
	Medium
	High
	Critical
)

// RiskType represents the type of risk
type RiskType int

const (
	PositionRisk RiskType = iota + 1
	ExposureRisk
	DrawdownRisk
	VolatilityRisk
	LiquidityRisk
)

// RiskStatus represents the status of a risk check
type RiskStatus int

const (
	Pass RiskStatus = iota + 1
	Warning
	Violation
)

// RiskCheck represents a risk check result
type RiskCheck struct {
	ID          string
	Type        RiskType
	Level       RiskLevel
	Status      RiskStatus
	Value       float64
	Threshold   float64
	Symbol      string
	CreatedAt   time.Time
	Description string
}

// RiskManager interface defines the contract for risk management
type RiskManager interface {
	// Position limits
	CheckPositionLimit(ctx context.Context, params PositionLimitParams) (*RiskCheck, error)
	UpdatePositionLimit(ctx context.Context, symbol string, limit float64) error

	// Exposure limits
	CheckExposureLimit(ctx context.Context, params ExposureLimitParams) (*RiskCheck, error)
	UpdateExposureLimit(ctx context.Context, limit float64) error

	// Drawdown protection
	CheckDrawdown(ctx context.Context, params DrawdownParams) (*RiskCheck, error)
	UpdateDrawdownLimit(ctx context.Context, limit float64) error

	// Volatility adjustments
	CheckVolatility(ctx context.Context, params VolatilityParams) (*RiskCheck, error)
	UpdateVolatilityThresholds(ctx context.Context, params VolatilityThresholds) error

	// Risk metrics
	GetRiskMetrics(ctx context.Context) (*RiskMetrics, error)
	GetRiskHistory(ctx context.Context, filter RiskHistoryFilter) ([]*RiskCheck, error)
}

// PositionLimitParams contains parameters for position limit check
type PositionLimitParams struct {
	Symbol        string
	Size          float64
	CurrentPrice  float64
	TotalPosition float64
}

// ExposureLimitParams contains parameters for exposure limit check
type ExposureLimitParams struct {
	TotalExposure     float64
	AdditionalAmount  float64
	CollateralBalance float64
}

// DrawdownParams contains parameters for drawdown check
type DrawdownParams struct {
	CurrentEquity float64
	PeakEquity    float64
	TimeWindow    time.Duration
}

// VolatilityParams contains parameters for volatility check
type VolatilityParams struct {
	Symbol            string
	CurrentVolatility float64
	HistoricalPrices  []float64
	TimeWindow        time.Duration
}

// VolatilityThresholds contains thresholds for volatility levels
type VolatilityThresholds struct {
	LowThreshold      float64
	MediumThreshold   float64
	HighThreshold     float64
	CriticalThreshold float64
}

// RiskMetrics contains current risk metrics
type RiskMetrics struct {
	TotalExposure       float64
	LargestPosition     float64
	CurrentDrawdown     float64
	PortfolioVolatility float64
	RiskLevel           RiskLevel
	UpdatedAt           time.Time
}

// RiskHistoryFilter contains filters for risk history
type RiskHistoryFilter struct {
	Type      *RiskType
	Level     *RiskLevel
	Status    *RiskStatus
	Symbol    string
	StartTime *time.Time
	EndTime   *time.Time
}

// DefaultRiskManager implements RiskManager interface
type DefaultRiskManager struct {
	positionLimits       map[string]float64
	exposureLimit        float64
	drawdownLimit        float64
	volatilityThresholds VolatilityThresholds
	riskChecks           []*RiskCheck
	mu                   sync.RWMutex
}

// NewRiskManager creates a new risk manager instance
func NewRiskManager() RiskManager {
	return &DefaultRiskManager{
		positionLimits: make(map[string]float64),
		exposureLimit:  1000000.0, // Default 1M
		drawdownLimit:  0.25,      // Default 25%
		volatilityThresholds: VolatilityThresholds{
			LowThreshold:      0.15,
			MediumThreshold:   0.30,
			HighThreshold:     0.50,
			CriticalThreshold: 0.75,
		},
		riskChecks: make([]*RiskCheck, 0),
	}
}

// CheckPositionLimit checks if a position exceeds the limit
func (m *DefaultRiskManager) CheckPositionLimit(ctx context.Context, params PositionLimitParams) (*RiskCheck, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		monitoring.RecordIndicatorCalculation("check_position_limit", duration)
	}()

	m.mu.RLock()
	limit, exists := m.positionLimits[params.Symbol]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrLimitNotSet
	}

	totalExposure := params.TotalPosition * params.CurrentPrice
	newExposure := params.Size * params.CurrentPrice

	check := &RiskCheck{
		ID:          generateCheckID(),
		Type:        PositionRisk,
		Value:       totalExposure + newExposure,
		Threshold:   limit,
		Symbol:      params.Symbol,
		CreatedAt:   time.Now(),
		Description: "Position limit check",
	}

	if check.Value >= limit*0.9 && check.Value < limit {
		check.Status = Warning
		check.Level = High
	} else if check.Value >= limit {
		check.Status = Violation
		check.Level = Critical
		monitoring.RecordIndicatorError("position_limit", "Position limit exceeded")
		return check, ErrPositionLimitExceeded
	} else {
		check.Status = Pass
		check.Level = Low
	}

	m.mu.Lock()
	m.riskChecks = append(m.riskChecks, check)
	m.mu.Unlock()

	monitoring.RecordIndicatorValue("position_utilization", check.Value/check.Threshold)
	return check, nil
}

// UpdatePositionLimit updates the position limit for a symbol
func (m *DefaultRiskManager) UpdatePositionLimit(ctx context.Context, symbol string, limit float64) error {
	if limit <= 0 {
		return ErrInvalidLimit
	}

	m.mu.Lock()
	m.positionLimits[symbol] = limit
	m.mu.Unlock()

	monitoring.RecordIndicatorValue("position_limit_"+symbol, limit)
	return nil
}

// CheckExposureLimit checks if total exposure exceeds the limit
func (m *DefaultRiskManager) CheckExposureLimit(ctx context.Context, params ExposureLimitParams) (*RiskCheck, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		monitoring.RecordIndicatorCalculation("check_exposure_limit", duration)
	}()

	m.mu.RLock()
	limit := m.exposureLimit
	m.mu.RUnlock()

	totalExposure := params.TotalExposure + params.AdditionalAmount
	maxExposure := params.CollateralBalance * limit

	check := &RiskCheck{
		ID:          generateCheckID(),
		Type:        ExposureRisk,
		Value:       totalExposure,
		Threshold:   maxExposure,
		CreatedAt:   time.Now(),
		Description: "Exposure limit check",
	}

	if totalExposure >= maxExposure*0.9 && totalExposure < maxExposure {
		check.Status = Warning
		check.Level = High
	} else if totalExposure >= maxExposure {
		check.Status = Violation
		check.Level = Critical
		monitoring.RecordIndicatorError("exposure_limit", "Exposure limit exceeded")
		return check, ErrExposureLimitExceeded
	} else {
		check.Status = Pass
		check.Level = Low
	}

	m.mu.Lock()
	m.riskChecks = append(m.riskChecks, check)
	m.mu.Unlock()

	monitoring.RecordIndicatorValue("exposure_utilization", check.Value/check.Threshold)
	return check, nil
}

// UpdateExposureLimit updates the exposure limit
func (m *DefaultRiskManager) UpdateExposureLimit(ctx context.Context, limit float64) error {
	if limit <= 0 {
		return ErrInvalidLimit
	}

	m.mu.Lock()
	m.exposureLimit = limit
	m.mu.Unlock()

	monitoring.RecordIndicatorValue("exposure_limit", limit)
	return nil
}

// CheckDrawdown checks if drawdown exceeds the limit
func (m *DefaultRiskManager) CheckDrawdown(ctx context.Context, params DrawdownParams) (*RiskCheck, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		monitoring.RecordIndicatorCalculation("check_drawdown", duration)
	}()

	m.mu.RLock()
	limit := m.drawdownLimit
	m.mu.RUnlock()

	drawdown := (params.PeakEquity - params.CurrentEquity) / params.PeakEquity

	check := &RiskCheck{
		ID:          generateCheckID(),
		Type:        DrawdownRisk,
		Value:       drawdown,
		Threshold:   limit,
		CreatedAt:   time.Now(),
		Description: "Drawdown check",
	}

	if drawdown >= limit*0.8 && drawdown < limit {
		check.Status = Warning
		check.Level = High
	} else if drawdown >= limit {
		check.Status = Violation
		check.Level = Critical
		monitoring.RecordIndicatorError("drawdown", "Maximum drawdown exceeded")
		return check, ErrDrawdownLimitExceeded
	} else {
		check.Status = Pass
		check.Level = Low
	}

	m.mu.Lock()
	m.riskChecks = append(m.riskChecks, check)
	m.mu.Unlock()

	monitoring.RecordIndicatorValue("drawdown", drawdown)
	return check, nil
}

// UpdateDrawdownLimit updates the drawdown limit
func (m *DefaultRiskManager) UpdateDrawdownLimit(ctx context.Context, limit float64) error {
	if limit <= 0 || limit >= 1 {
		return ErrInvalidLimit
	}

	m.mu.Lock()
	m.drawdownLimit = limit
	m.mu.Unlock()

	monitoring.RecordIndicatorValue("drawdown_limit", limit)
	return nil
}

// CheckVolatility checks if volatility exceeds thresholds
func (m *DefaultRiskManager) CheckVolatility(ctx context.Context, params VolatilityParams) (*RiskCheck, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		monitoring.RecordIndicatorCalculation("check_volatility", duration)
	}()

	m.mu.RLock()
	thresholds := m.volatilityThresholds
	m.mu.RUnlock()

	check := &RiskCheck{
		ID:          generateCheckID(),
		Type:        VolatilityRisk,
		Value:       params.CurrentVolatility,
		Symbol:      params.Symbol,
		CreatedAt:   time.Now(),
		Description: "Volatility check",
	}

	switch {
	case params.CurrentVolatility >= thresholds.CriticalThreshold:
		check.Status = Violation
		check.Level = Critical
		check.Threshold = thresholds.CriticalThreshold
		monitoring.RecordIndicatorError("volatility", "Critical volatility level")
		return check, ErrVolatilityTooHigh
	case params.CurrentVolatility >= thresholds.HighThreshold:
		check.Status = Warning
		check.Level = High
		check.Threshold = thresholds.HighThreshold
	case params.CurrentVolatility >= thresholds.MediumThreshold:
		check.Status = Warning
		check.Level = Medium
		check.Threshold = thresholds.MediumThreshold
	default:
		check.Status = Pass
		check.Level = Low
		check.Threshold = thresholds.LowThreshold
	}

	m.mu.Lock()
	m.riskChecks = append(m.riskChecks, check)
	m.mu.Unlock()

	monitoring.RecordIndicatorValue("volatility_"+params.Symbol, params.CurrentVolatility)
	return check, nil
}

// UpdateVolatilityThresholds updates volatility thresholds
func (m *DefaultRiskManager) UpdateVolatilityThresholds(ctx context.Context, thresholds VolatilityThresholds) error {
	if !isValidThresholds(thresholds) {
		return ErrInvalidThresholds
	}

	m.mu.Lock()
	m.volatilityThresholds = thresholds
	m.mu.Unlock()

	return nil
}

// GetRiskMetrics returns current risk metrics
func (m *DefaultRiskManager) GetRiskMetrics(ctx context.Context) (*RiskMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Implementation would calculate metrics from current state
	// This is a placeholder implementation
	return &RiskMetrics{
		UpdatedAt: time.Now(),
	}, nil
}

// GetRiskHistory returns risk check history based on filter
func (m *DefaultRiskManager) GetRiskHistory(ctx context.Context, filter RiskHistoryFilter) ([]*RiskCheck, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	checks := make([]*RiskCheck, 0)
	for _, check := range m.riskChecks {
		if matchesRiskFilter(check, filter) {
			checks = append(checks, check)
		}
	}

	return checks, nil
}

func isValidThresholds(thresholds VolatilityThresholds) bool {
	return thresholds.LowThreshold > 0 &&
		thresholds.MediumThreshold > thresholds.LowThreshold &&
		thresholds.HighThreshold > thresholds.MediumThreshold &&
		thresholds.CriticalThreshold > thresholds.HighThreshold
}

func matchesRiskFilter(check *RiskCheck, filter RiskHistoryFilter) bool {
	if filter.Type != nil && check.Type != *filter.Type {
		return false
	}
	if filter.Level != nil && check.Level != *filter.Level {
		return false
	}
	if filter.Status != nil && check.Status != *filter.Status {
		return false
	}
	if filter.Symbol != "" && check.Symbol != filter.Symbol {
		return false
	}
	if filter.StartTime != nil && check.CreatedAt.Before(*filter.StartTime) {
		return false
	}
	if filter.EndTime != nil && check.CreatedAt.After(*filter.EndTime) {
		return false
	}
	return true
}

func generateCheckID() string {
	return time.Now().Format("20060102150405.000") + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
