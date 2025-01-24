import React, { useEffect, useState } from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer
} from 'recharts';
import { MarketData, AnalysisResult } from '../types';
import styled from '@emotion/styled';

interface Props {
  tokenAddress: string;
}

const Value = styled.span`
  display: block;
  font-size: 1.2em;
  font-weight: 600;
  color: #212529;
`;

const MarketAnalysis: React.FC<Props> = ({ tokenAddress }) => {
  const [marketData, setMarketData] = useState<MarketData | null>(null);
  const [analysis, setAnalysis] = useState<AnalysisResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);

        // Fetch market data
        const marketResponse = await fetch(`http://localhost:8080/api/market-data/${tokenAddress}`);
        if (!marketResponse.ok) {
          throw new Error('Failed to fetch market data');
        }
        const marketData = await marketResponse.json();
        setMarketData(marketData);

        // Fetch analysis
        const analysisResponse = await fetch(`http://localhost:8080/api/analysis`);
        if (!analysisResponse.ok) {
          throw new Error('Failed to fetch analysis');
        }
        const analysis = await analysisResponse.json();
        setAnalysis(analysis);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
    const interval = setInterval(fetchData, 30000); // Update every 30 seconds

    return () => clearInterval(interval);
  }, [tokenAddress]);

  if (loading) {
    return <div className="loading">Loading market data...</div>;
  }

  if (error) {
    return <div className="error">Error: {error}</div>;
  }

  if (!marketData || !analysis) {
    return <div>No data available</div>;
  }

  const deepseekAnalysis = JSON.parse(analysis.deepseek_analysis);

  return (
    <Container>
      <Section>
        <h2>Market Analysis</h2>
        <MetricsGrid>
          <Metric>
            <Label>Market Sentiment</Label>
            <Value>{analysis.sentiment}</Value>
          </Metric>
          <Metric>
            <Label>Risk Level</Label>
            <Value>{analysis.risk_level}</Value>
          </Metric>
          <Metric>
            <Label>Recommendation</Label>
            <Value>{deepseekAnalysis.recommendation.action}</Value>
          </Metric>
          <Metric>
            <Label>Confidence</Label>
            <Value>{analysis.confidence}%</Value>
          </Metric>
        </MetricsGrid>
      </Section>

      <Section>
        <h2>Price Analysis</h2>
        <MetricsGrid>
          <Metric>
            <Label>Current Price</Label>
            <Value>{marketData.price.toFixed(6)} SOL</Value>
          </Metric>
          <Metric>
            <Label>24h Volume</Label>
            <Value>{marketData.volume_24h.toFixed(2)} SOL</Value>
          </Metric>
          <Metric>
            <Label>Market Cap</Label>
            <Value>{marketData.market_cap.toFixed(2)} SOL</Value>
          </Metric>
          <Metric>
            <Label>Liquidity</Label>
            <Value>{marketData.liquidity.toFixed(2)} SOL</Value>
          </Metric>
        </MetricsGrid>

        <ChartContainer>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={[{ time: 'now', price: marketData.price }]}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="time" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Line type="monotone" dataKey="price" stroke="#8884d8" />
            </LineChart>
          </ResponsiveContainer>
        </ChartContainer>
      </Section>
    </Container>
  );
};

const Container = styled.div`
  padding: 20px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
`;

const Section = styled.div`
  margin-bottom: 30px;
`;

const MetricsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin: 20px 0;
`;

const Metric = styled.div`
  padding: 15px;
  background: #f8f9fa;
  border-radius: 6px;
  border: 1px solid #e9ecef;
`;

const Label = styled.label`
  display: block;
  color: #6c757d;
  font-size: 0.9em;
  margin-bottom: 5px;
`;

const ChartContainer = styled.div`
  margin: 20px 0;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 6px;
  border: 1px solid #e9ecef;
`;

const SentimentValue = styled(Value)<{ sentiment: string }>`
  color: ${props => {
    switch (props.sentiment.toLowerCase()) {
      case 'bullish': return '#28a745';
      case 'bearish': return '#dc3545';
      default: return '#6c757d';
    }
  }};
`;

const ActionValue = styled(Value)<{ action: string }>`
  color: ${props => {
    switch (props.action.toLowerCase()) {
      case 'buy': return '#28a745';
      case 'sell': return '#dc3545';
      default: return '#6c757d';
    }
  }};
  font-weight: 600;
`;

const RiskDetails = styled.div`
  font-size: 0.9em;
  line-height: 1.5;
`;

const Recommendation = styled.div`
  font-size: 0.9em;
  line-height: 1.5;
`;

export default MarketAnalysis;
