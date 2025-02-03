import React from 'react';

interface Position {
  symbol: string;
  size: number;
  entryPrice: number;
  markPrice?: number;
  pnl?: number;
  status?: 'open' | 'closed';
  lastUpdated?: number;
}

interface PositionTableProps {
  positions: Position[];
  onClosePosition?: (symbol: string) => void;
  loading?: boolean;
  'data-testid'?: string;
}

const formatNumber = (num: number | undefined, decimals: number = 2): string => {
  return num?.toFixed(decimals) || '-';
};

const PositionTable: React.FC<PositionTableProps> = ({ positions, onClosePosition, loading = false, 'data-testid': testId = 'position-table' }) => {
  if (loading) {
    return (
      <div className="position-table" data-testid="position-table">
        <h3>Positions</h3>
        <div className="loading">Loading positions...</div>
      </div>
    );
  }
  return (
    <div className="position-table" data-testid={testId}>
      <h3>Positions</h3>
      <table>
        <thead>
          <tr>
            <th>Symbol</th>
            <th>Size</th>
            <th>Entry Price</th>
            <th>Mark Price</th>
            <th>PnL</th>
            <th>Status</th>
            {onClosePosition && <th>Action</th>}
          </tr>
        </thead>
        <tbody>
          {positions.map((position, index) => (
            <tr key={position.symbol + index} data-testid={`position-row-${index}`}>
              <td>{position.symbol}</td>
              <td>{formatNumber(position.size, 4)}</td>
              <td>{formatNumber(position.entryPrice)}</td>
              <td>{formatNumber(position.markPrice)}</td>
              <td className={position.pnl && position.pnl > 0 ? 'profit' : 'loss'}>
                {formatNumber(position.pnl)}
              </td>
              <td>{position.status || 'open'}</td>
              {onClosePosition && position.status === 'open' && (
                <td>
                  <button 
                    onClick={() => onClosePosition(position.symbol)}
                    disabled={loading}
                    data-testid={`close-position-${position.symbol}`}
                  >
                    Close
                  </button>
                </td>
              )}
            </tr>
          ))}
          {positions.length === 0 && (
            <tr>
              <td colSpan={onClosePosition ? 7 : 6} className="no-data">
                No open positions
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
};

export default PositionTable;
