package streaming

// DefaultIndicatorFactory implements the IndicatorFactory interface
type DefaultIndicatorFactory struct{}

// NewIndicatorFactory creates a new instance of DefaultIndicatorFactory
func NewIndicatorFactory() *DefaultIndicatorFactory {
	return &DefaultIndicatorFactory{}
}

// CreateRSI creates a new RSI indicator
func (f *DefaultIndicatorFactory) CreateRSI(period int) (WindowedIndicator, error) {
	return NewRSI(period)
}

// CreateEMA creates a new EMA indicator
func (f *DefaultIndicatorFactory) CreateEMA(period int) (WindowedIndicator, error) {
	return NewEMA(period)
}

// CreateMACD creates a new MACD indicator
func (f *DefaultIndicatorFactory) CreateMACD(fastPeriod, slowPeriod, signalPeriod int) (Indicator, error) {
	return NewMACD(fastPeriod, slowPeriod, signalPeriod)
}
