import sqlite3
import pandas as pd
from datetime import datetime, timedelta
from config import Config
import time
import os
import logging
import json
import psutil

logging.basicConfig(**Config.get_log_config())
logger = logging.getLogger(__name__)

class SystemMonitor:
    """系统监控"""
    
    def __init__(self):
        self.db_config = Config.get_db_config()
    
    def get_portfolio_status(self):
        """获取投资组合状态"""
        with sqlite3.connect(self.db_config['risk_management']) as conn:
            query = """
                SELECT * FROM portfolio_states
                ORDER BY timestamp DESC
                LIMIT 1
            """
            return pd.read_sql_query(query, conn)
    
    def get_recent_trades(self, hours=24):
        """获取最近交易"""
        with sqlite3.connect(self.db_config['market_data']) as conn:
            query = """
                SELECT * FROM trades
                WHERE timestamp >= ?
                ORDER BY timestamp DESC
            """
            start_time = datetime.now() - timedelta(hours=hours)
            return pd.read_sql_query(query, conn, params=(start_time,))
    
    def get_agent_performance(self):
        """获取Agent性能"""
        with sqlite3.connect(self.db_config['agent_system']) as conn:
            query = """
                SELECT * FROM performance
                ORDER BY timestamp DESC
            """
            return pd.read_sql_query(query, conn)
    
    def get_risk_events(self, hours=24):
        """获取风险事件"""
        with sqlite3.connect(self.db_config['risk_management']) as conn:
            query = """
                SELECT * FROM risk_events
                WHERE timestamp >= ?
                ORDER BY timestamp DESC
            """
            start_time = datetime.now() - timedelta(hours=hours)
            return pd.read_sql_query(query, conn, params=(start_time,))
    
    def display_status(self):
        """显示系统状态"""
        try:
            # 获取投资组合状态
            portfolio = self.get_portfolio_status()
            if not portfolio.empty:
                print("\n=== 投资组合状态 ===")
                print(f"总权益: {portfolio['total_equity'].iloc[0]:.2f}")
                print(f"已用保证金: {portfolio['used_margin'].iloc[0]:.2f}")
                print(f"可用保证金: {portfolio['free_margin'].iloc[0]:.2f}")
                print(f"保证金率: {portfolio['margin_level'].iloc[0]:.2f}")
                print(f"当日盈亏: {portfolio['daily_pnl'].iloc[0]:.2f}")
                print(f"回撤: {portfolio['drawdown'].iloc[0]:.2%}")
                
                # 添加合约交易信息
                exposure = json.loads(portfolio['exposure'].iloc[0])
                print("\n=== 合约交易信息 ===")
                print(f"总杠杆率: {exposure.get('total_leverage', 0):.2f}x")
                print(f"多头暴露: {exposure.get('long_exposure', 0):.2f}")
                print(f"空头暴露: {exposure.get('short_exposure', 0):.2f}")
                print(f"净暴露: {exposure.get('net_exposure', 0):.2f}")
                
                # 显示资金费率信息
                funding_info = json.loads(portfolio.get('funding_info', '{}'))
                if funding_info:
                    print("\n=== 资金费率信息 ===")
                    for symbol, info in funding_info.items():
                        print(f"{symbol}:")
                        print(f"  当前费率: {info['current_rate']:.4%}")
                        print(f"  预测费率: {info['predicted_rate']:.4%}")
                        print(f"  下次收取时间: {info['next_time']}")
                
                # 显示强平风险信息
                liquidation_risks = json.loads(portfolio.get('liquidation_risks', '{}'))
                if liquidation_risks:
                    print("\n=== 强平风险警告 ===")
                    for symbol, risk in liquidation_risks.items():
                        print(f"{symbol}:")
                        print(f"  当前价格: {risk['current_price']:.2f}")
                        print(f"  强平价格: {risk['liquidation_price']:.2f}")
                        print(f"  价格距离: {risk['price_distance']:.2%}")
            
            # 获取最近交易
            trades = self.get_recent_trades()
            if not trades.empty:
                print("\n=== 最近24小时交易 ===")
                # 分别显示现货和合约交易
                spot_trades = trades[trades['type'] == 'spot']
                contract_trades = trades[trades['type'] == 'contract']
                
                print("\n现货交易:")
                if not spot_trades.empty:
                    print(f"交易数量: {len(spot_trades)}")
                    print(f"交易量: {spot_trades['amount'].sum():.2f}")
                    print(f"平均价格: {spot_trades['price'].mean():.2f}")
                
                print("\n合约交易:")
                if not contract_trades.empty:
                    print(f"交易数量: {len(contract_trades)}")
                    print(f"名义价值: {contract_trades['notional_value'].sum():.2f}")
                    print(f"平均杠杆: {contract_trades['leverage'].mean():.2f}x")
                    print(f"已实现盈亏: {contract_trades['realized_pnl'].sum():.2f}")
            
            # 获取风险事件
            risk_events = self.get_risk_events()
            if not risk_events.empty:
                print("\n=== 风险事件 ===")
                for _, event in risk_events.iterrows():
                    print(f"[{event['severity']}] {event['event_type']}: {event['description']}")
            
            # 显示系统指标
            self._display_system_metrics()
            
        except Exception as e:
            print(f"Error displaying status: {e}")
    
    def _display_system_metrics(self):
        """显示系统性能指标"""
        try:
            # CPU使用率
            cpu_percent = psutil.cpu_percent(interval=1)
            print(f"CPU使用率: {cpu_percent}%")
            
            # 内存使用
            memory = psutil.Process().memory_info()
            print(f"内存使用: {memory.rss / 1024 / 1024:.1f}MB")
            
            # 数据库大小
            db_sizes = {}
            for name, path in self.db_config.items():
                if os.path.exists(path):
                    size = os.path.getsize(path) / 1024 / 1024  # MB
                    db_sizes[name] = size
            
            print("\n数据库大小:")
            for name, size in db_sizes.items():
                print(f"{name}: {size:.1f}MB")
            
            # 网络连接状态
            with sqlite3.connect(self.db_config['market_data']) as conn:
                last_update = pd.read_sql_query("""
                    SELECT MAX(timestamp) as last_update
                    FROM trades
                """, conn).iloc[0]['last_update']
                
                if last_update:
                    last_update = pd.to_datetime(last_update)
                    delay = (datetime.now() - last_update).total_seconds()
                    print(f"\n数据延迟: {delay:.1f}秒")
            
            # 错误日志统计
            log_file = Config.get_log_config()['file']
            if os.path.exists(log_file):
                with open(log_file, 'r') as f:
                    logs = f.readlines()
                    recent_logs = logs[-100:]  # 最近100条日志
                    error_count = sum(1 for log in recent_logs if 'ERROR' in log)
                    warning_count = sum(1 for log in recent_logs if 'WARNING' in log)
                    print(f"\n最近错误数: {error_count}")
                    print(f"最近警告数: {warning_count}")
            
        except Exception as e:
            logger.error(f"Error displaying system metrics: {str(e)}")

    def get_detailed_performance(self, days: int = 30) -> pd.DataFrame:
        """获取详细的性能报告"""
        start_time = datetime.now() - timedelta(days=days)
        
        with sqlite3.connect(self.db_config['risk_management']) as conn:
            # 获取每日性能数据
            daily_stats = pd.read_sql_query("""
                SELECT 
                    DATE(timestamp) as date,
                    COUNT(*) as trades_count,
                    SUM(CASE WHEN pnl > 0 THEN 1 ELSE 0 END) as winning_trades,
                    SUM(CASE WHEN pnl < 0 THEN 1 ELSE 0 END) as losing_trades,
                    SUM(pnl) as total_pnl,
                    AVG(CASE WHEN pnl > 0 THEN pnl END) as avg_win,
                    AVG(CASE WHEN pnl < 0 THEN pnl END) as avg_loss
                FROM trades
                WHERE timestamp >= ?
                GROUP BY DATE(timestamp)
                ORDER BY date DESC
            """, conn, params=(start_time,))
            
            # 计算累计指标
            daily_stats['cumulative_pnl'] = daily_stats['total_pnl'].cumsum()
            daily_stats['win_rate'] = daily_stats['winning_trades'] / daily_stats['trades_count']
            daily_stats['profit_factor'] = abs(
                daily_stats['avg_win'] * daily_stats['winning_trades'] /
                (daily_stats['avg_loss'] * daily_stats['losing_trades'])
            )
            
            return daily_stats

def main():
    """主函数"""
    monitor = SystemMonitor()
    
    try:
        while True:
            os.system('clear')  # 清屏
            print("\n=== 交易系统监控 ===")
            print(f"当前时间: {datetime.now()}")
            monitor.display_status()
            time.sleep(5)  # 每5秒更新一次
            
    except KeyboardInterrupt:
        print("\n监控已停止")

if __name__ == "__main__":
    main() 