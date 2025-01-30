import streamlit as st
import pandas as pd
import plotly.graph_objects as go
from plotly.subplots import make_subplots
import numpy as np
from datetime import datetime, timedelta
from typing import Dict, List
from pathlib import Path
import json
import sqlite3
from backtest_system import BacktestSystem, BacktestConfig, BacktestResult
from reporting_system import ReportingSystem
from market_data_service import MarketDataService
from agent_system import AgentSystem
from performance_monitor import PerformanceMonitor

# 配置页面
st.set_page_config(
    page_title="Trading System Dashboard",
    page_icon="📈",
    layout="wide"
)

# 初始化系统
reporting = ReportingSystem()

def main():
    """主函数"""
    st.title("Trading System Dashboard")
    
    # 侧边栏
    st.sidebar.title("Navigation")
    page = st.sidebar.radio(
        "Select Page",
        ["System Status", "Backtest", "Performance Analysis", "Agent Analysis", "Performance Monitor"]
    )
    
    if page == "System Status":
        show_system_status()
    elif page == "Backtest":
        show_backtest_page()
    elif page == "Performance Analysis":
        show_performance_analysis()
    elif page == "Agent Analysis":
        show_agent_analysis()
    else:
        show_performance_monitor()

def show_system_status():
    """显示系统状态"""
    st.header("System Status")
    
    # 获取最新的性能报告
    with sqlite3.connect(reporting.db_path) as conn:
        latest_perf = pd.read_sql_query("""
            SELECT * FROM performance_reports
            ORDER BY timestamp DESC LIMIT 1
        """, conn)
    
    if not latest_perf.empty:
        # 创建三列布局
        col1, col2, col3 = st.columns(3)
        
        with col1:
            st.metric(
                "Total P&L",
                f"${latest_perf['total_pnl'].iloc[0]:,.2f}",
                f"{latest_perf['daily_pnl'].iloc[0]:,.2f}"
            )
        
        with col2:
            st.metric(
                "Win Rate",
                f"{latest_perf['win_rate'].iloc[0]:.1%}"
            )
        
        with col3:
            st.metric(
                "Sharpe Ratio",
                f"{latest_perf['sharpe_ratio'].iloc[0]:.2f}"
            )
        
        # 获取最近的交易
        recent_trades = pd.read_sql_query("""
            SELECT * FROM trades
            ORDER BY close_time DESC LIMIT 10
        """, conn)
        
        if not recent_trades.empty:
            st.subheader("Recent Trades")
            st.dataframe(recent_trades)
        
        # 显示权益曲线
        equity_data = pd.read_sql_query("""
            SELECT timestamp, total_pnl
            FROM performance_reports
            ORDER BY timestamp
        """, conn)
        
        if not equity_data.empty:
            fig = go.Figure()
            fig.add_trace(go.Scatter(
                x=equity_data['timestamp'],
                y=equity_data['total_pnl'],
                mode='lines',
                name='Equity'
            ))
            fig.update_layout(
                title="Equity Curve",
                xaxis_title="Time",
                yaxis_title="Equity ($)",
                height=400
            )
            st.plotly_chart(fig, use_container_width=True)

def show_backtest_page():
    """显示回测页面"""
    st.header("Backtest")
    
    # 回测配置
    st.subheader("Backtest Configuration")
    
    col1, col2 = st.columns(2)
    
    with col1:
        start_date = st.date_input(
            "Start Date",
            datetime.now() - timedelta(days=30)
        )
        
        symbols = st.multiselect(
            "Symbols",
            ["BTC/USDT", "ETH/USDT", "BNB/USDT"],
            ["BTC/USDT"]
        )
    
    with col2:
        end_date = st.date_input(
            "End Date",
            datetime.now()
        )
        
        timeframes = st.multiselect(
            "Timeframes",
            ["1m", "5m", "15m", "1h", "4h", "1d"],
            ["1h"]
        )
    
    initial_capital = st.number_input(
        "Initial Capital ($)",
        min_value=1000,
        value=100000
    )
    
    if st.button("Run Backtest"):
        # 创建回测配置
        config = BacktestConfig(
            start_date=datetime.combine(start_date, datetime.min.time()),
            end_date=datetime.combine(end_date, datetime.max.time()),
            initial_capital=initial_capital,
            symbols=symbols,
            timeframes=timeframes
        )
        
        # 运行回测
        backtest = BacktestSystem(config)
        result = backtest.run()
        
        # 显示回测结果
        show_backtest_result(result)

def show_backtest_result(result: BacktestResult):
    """显示回测结果"""
    st.subheader("Backtest Results")
    
    # 创建四列布局显示主要指标
    col1, col2, col3, col4 = st.columns(4)
    
    with col1:
        st.metric("Total P&L", f"${result.total_pnl:,.2f}")
        st.metric("Win Rate", f"{result.win_rate:.1%}")
    
    with col2:
        st.metric("Total Trades", str(result.total_trades))
        st.metric("Profit Factor", f"{result.profit_factor:.2f}")
    
    with col3:
        st.metric("Sharpe Ratio", f"{result.sharpe_ratio:.2f}")
        st.metric("Sortino Ratio", f"{result.sortino_ratio:.2f}")
    
    with col4:
        st.metric("Max Drawdown", f"{result.max_drawdown:.1%}")
        st.metric("Recovery Factor", f"{result.recovery_factor:.2f}")
    
    # 创建图表
    fig = make_subplots(
        rows=2, cols=1,
        shared_xaxes=True,
        vertical_spacing=0.03,
        subplot_titles=("Equity Curve", "Drawdown"),
        row_heights=[0.7, 0.3]
    )
    
    # 添加权益曲线
    fig.add_trace(
        go.Scatter(
            x=result.equity_curve.index,
            y=result.equity_curve.values,
            mode='lines',
            name='Equity'
        ),
        row=1, col=1
    )
    
    # 添加回撤曲线
    fig.add_trace(
        go.Scatter(
            x=result.drawdown_curve.index,
            y=result.drawdown_curve.values * 100,
            mode='lines',
            name='Drawdown',
            line=dict(color='red')
        ),
        row=2, col=1
    )
    
    fig.update_layout(
        height=800,
        showlegend=True,
        title_text="Backtest Performance"
    )
    
    fig.update_yaxes(title_text="Equity ($)", row=1, col=1)
    fig.update_yaxes(title_text="Drawdown (%)", row=2, col=1)
    
    st.plotly_chart(fig, use_container_width=True)
    
    # 显示月度收益
    st.subheader("Monthly Returns")
    monthly_returns_pct = result.monthly_returns * 100
    
    fig = go.Figure(data=[
        go.Bar(
            x=monthly_returns_pct.index,
            y=monthly_returns_pct.values,
            name='Monthly Returns'
        )
    ])
    
    fig.update_layout(
        title="Monthly Returns (%)",
        xaxis_title="Month",
        yaxis_title="Return (%)",
        height=400
    )
    
    st.plotly_chart(fig, use_container_width=True)
    
    # 显示交易记录
    st.subheader("Trade History")
    trades_df = pd.DataFrame(result.trades)
    st.dataframe(trades_df)
    
    # 显示Agent性能
    st.subheader("Agent Performance")
    for agent_name, metrics in result.agent_performance.items():
        st.write(f"**{agent_name}**")
        metrics_df = pd.DataFrame([metrics])
        st.dataframe(metrics_df)

def show_performance_analysis():
    """显示性能分析"""
    st.header("Performance Analysis")
    
    # 时间范围选择
    col1, col2 = st.columns(2)
    with col1:
        start_date = st.date_input(
            "Start Date",
            datetime.now() - timedelta(days=30),
            key="perf_start_date"
        )
    with col2:
        end_date = st.date_input(
            "End Date",
            datetime.now(),
            key="perf_end_date"
        )
    
    # 获取历史性能数据
    historical_perf = reporting.get_historical_performance(
        start_date=datetime.combine(start_date, datetime.min.time()),
        end_date=datetime.combine(end_date, datetime.max.time())
    )
    
    if not historical_perf.empty:
        # 创建性能指标图表
        fig = make_subplots(
            rows=2, cols=2,
            subplot_titles=(
                "Equity Curve",
                "Rolling Sharpe Ratio",
                "Rolling Win Rate",
                "Rolling Profit Factor"
            )
        )
        
        # 权益曲线
        fig.add_trace(
            go.Scatter(
                x=historical_perf['timestamp'],
                y=historical_perf['total_pnl'],
                mode='lines',
                name='Equity'
            ),
            row=1, col=1
        )
        
        # 滚动夏普比率
        fig.add_trace(
            go.Scatter(
                x=historical_perf['timestamp'],
                y=historical_perf['sharpe_ratio'].rolling(30).mean(),
                mode='lines',
                name='Rolling Sharpe'
            ),
            row=1, col=2
        )
        
        # 滚动胜率
        fig.add_trace(
            go.Scatter(
                x=historical_perf['timestamp'],
                y=historical_perf['win_rate'].rolling(30).mean(),
                mode='lines',
                name='Rolling Win Rate'
            ),
            row=2, col=1
        )
        
        # 计算滚动利润因子
        profit_factor = (
            historical_perf['avg_profit'] * historical_perf['winning_trades'] /
            abs(historical_perf['avg_loss'] * historical_perf['losing_trades'])
        ).rolling(30).mean()
        
        fig.add_trace(
            go.Scatter(
                x=historical_perf['timestamp'],
                y=profit_factor,
                mode='lines',
                name='Rolling Profit Factor'
            ),
            row=2, col=2
        )
        
        fig.update_layout(height=800, showlegend=True)
        st.plotly_chart(fig, use_container_width=True)
        
        # 显示统计摘要
        st.subheader("Performance Statistics")
        stats = {
            'Total P&L': f"${historical_perf['total_pnl'].iloc[-1]:,.2f}",
            'Win Rate': f"{historical_perf['win_rate'].mean():.1%}",
            'Avg Profit': f"${historical_perf['avg_profit'].mean():,.2f}",
            'Avg Loss': f"${historical_perf['avg_loss'].mean():,.2f}",
            'Sharpe Ratio': f"{historical_perf['sharpe_ratio'].mean():.2f}",
            'Max Drawdown': f"{historical_perf['max_drawdown'].max():.1%}"
        }
        
        st.json(stats)

def show_agent_analysis():
    """显示Agent分析"""
    st.header("Agent Analysis")
    
    # 获取所有Agent
    with sqlite3.connect(reporting.db_path) as conn:
        agents = pd.read_sql_query("""
            SELECT DISTINCT agent_name FROM trades
        """, conn)['agent_name'].tolist()
    
    if agents:
        # Agent选择
        selected_agent = st.selectbox(
            "Select Agent",
            agents
        )
        
        # 获取Agent性能数据
        agent_perf = reporting.get_agent_performance(selected_agent)
        
        if agent_perf:
            # 显示Agent统计
            st.subheader("Agent Statistics")
            st.json(agent_perf)
            
            # 获取Agent的交易记录
            with sqlite3.connect(reporting.db_path) as conn:
                agent_trades = pd.read_sql_query("""
                    SELECT * FROM trades
                    WHERE agent_name = ?
                    ORDER BY close_time DESC
                """, conn, params=(selected_agent,))
            
            if not agent_trades.empty:
                # 计算累计收益
                agent_trades['cumulative_pnl'] = agent_trades['pnl'].cumsum()
                
                # 创建收益曲线
                fig = go.Figure()
                fig.add_trace(go.Scatter(
                    x=agent_trades['close_time'],
                    y=agent_trades['cumulative_pnl'],
                    mode='lines',
                    name='Cumulative P&L'
                ))
                
                fig.update_layout(
                    title=f"{selected_agent} Performance",
                    xaxis_title="Time",
                    yaxis_title="Cumulative P&L ($)",
                    height=400
                )
                
                st.plotly_chart(fig, use_container_width=True)
                
                # 显示最近的交易
                st.subheader("Recent Trades")
                st.dataframe(agent_trades.head(10))

def show_performance_monitor():
    """显示性能监控页面"""
    st.header("Performance Monitor")
    
    # 时间范围选择
    col1, col2 = st.columns(2)
    with col1:
        start_time = st.date_input(
            "Start Time",
            datetime.now() - timedelta(days=1),
            key="monitor_start_time"
        )
    with col2:
        end_time = st.date_input(
            "End Time",
            datetime.now(),
            key="monitor_end_time"
        )
    
    # 获取性能数据
    monitor = PerformanceMonitor()
    system_metrics = monitor.get_system_metrics(
        datetime.combine(start_time, datetime.min.time()),
        datetime.combine(end_time, datetime.max.time())
    )
    
    trading_metrics = monitor.get_trading_metrics(
        datetime.combine(start_time, datetime.min.time()),
        datetime.combine(end_time, datetime.max.time())
    )
    
    if not system_metrics.empty:
        # 系统性能指标
        st.subheader("System Metrics")
        
        # 创建系统指标图表
        fig = make_subplots(
            rows=2, cols=2,
            subplot_titles=(
                "CPU Usage",
                "Memory Usage",
                "Disk Usage",
                "Network I/O"
            )
        )
        
        # CPU使用率
        fig.add_trace(
            go.Scatter(
                x=system_metrics['timestamp'],
                y=system_metrics['cpu_usage'],
                mode='lines',
                name='CPU Usage'
            ),
            row=1, col=1
        )
        
        # 内存使用率
        fig.add_trace(
            go.Scatter(
                x=system_metrics['timestamp'],
                y=system_metrics['memory_usage'],
                mode='lines',
                name='Memory Usage'
            ),
            row=1, col=2
        )
        
        # 磁盘使用率
        fig.add_trace(
            go.Scatter(
                x=system_metrics['timestamp'],
                y=system_metrics['disk_usage'],
                mode='lines',
                name='Disk Usage'
            ),
            row=2, col=1
        )
        
        # 网络IO
        network_io = pd.DataFrame([
            json.loads(io) for io in system_metrics['network_io']
        ])
        
        fig.add_trace(
            go.Scatter(
                x=system_metrics['timestamp'],
                y=network_io['bytes_sent'],
                mode='lines',
                name='Bytes Sent'
            ),
            row=2, col=2
        )
        
        fig.add_trace(
            go.Scatter(
                x=system_metrics['timestamp'],
                y=network_io['bytes_recv'],
                mode='lines',
                name='Bytes Received'
            ),
            row=2, col=2
        )
        
        fig.update_layout(height=800, showlegend=True)
        st.plotly_chart(fig, use_container_width=True)
    
    if not trading_metrics.empty:
        # 交易性能指标
        st.subheader("Trading Metrics")
        
        # 创建交易指标图表
        fig = make_subplots(
            rows=2, cols=2,
            subplot_titles=(
                "Execution Latency",
                "Signal Processing Time",
                "Order Success Rate",
                "Slippage"
            )
        )
        
        # 执行延迟
        fig.add_trace(
            go.Scatter(
                x=trading_metrics['timestamp'],
                y=trading_metrics['execution_latency'],
                mode='lines',
                name='Execution Latency'
            ),
            row=1, col=1
        )
        
        # 信号处理时间
        fig.add_trace(
            go.Scatter(
                x=trading_metrics['timestamp'],
                y=trading_metrics['signal_processing_time'],
                mode='lines',
                name='Processing Time'
            ),
            row=1, col=2
        )
        
        # 订单成功率
        fig.add_trace(
            go.Scatter(
                x=trading_metrics['timestamp'],
                y=trading_metrics['order_success_rate'],
                mode='lines',
                name='Success Rate'
            ),
            row=2, col=1
        )
        
        # 滑点
        fig.add_trace(
            go.Scatter(
                x=trading_metrics['timestamp'],
                y=trading_metrics['slippage'],
                mode='lines',
                name='Slippage'
            ),
            row=2, col=2
        )
        
        fig.update_layout(height=800, showlegend=True)
        st.plotly_chart(fig, use_container_width=True)
    
    # 性能分析
    st.subheader("Performance Analysis")
    analysis = monitor.analyze_performance(
        datetime.combine(start_time, datetime.min.time()),
        datetime.combine(end_time, datetime.max.time())
    )
    
    if analysis:
        # 显示系统性能
        st.write("System Performance")
        system_perf = pd.DataFrame([analysis['system_performance']])
        st.dataframe(system_perf)
        
        # 显示交易性能
        st.write("Trading Performance")
        trading_perf = pd.DataFrame([analysis['trading_performance']])
        st.dataframe(trading_perf)
        
        # 显示警告
        if analysis['warnings']:
            st.warning("Performance Warnings")
            for warning in analysis['warnings']:
                st.write(f"- {warning}")
        
        # 显示优化建议
        recommendations = monitor.optimize_performance(analysis)
        if recommendations:
            st.info("Optimization Recommendations")
            for recommendation in recommendations:
                st.write(f"- {recommendation}")
    
    # Agent性能监控
    st.subheader("Agent Performance")
    
    # 获取所有Agent
    with sqlite3.connect(monitor.db_path) as conn:
        agents = pd.read_sql_query("""
            SELECT DISTINCT agent_name FROM agent_metrics
        """, conn)['agent_name'].tolist()
    
    if agents:
        selected_agent = st.selectbox(
            "Select Agent",
            agents,
            key="monitor_agent"
        )
        
        agent_metrics = monitor.get_agent_metrics(
            selected_agent,
            datetime.combine(start_time, datetime.min.time()),
            datetime.combine(end_time, datetime.max.time())
        )
        
        if not agent_metrics.empty:
            # 创建Agent指标图表
            fig = make_subplots(
                rows=2, cols=2,
                subplot_titles=(
                    "Signal Count",
                    "Signal Quality",
                    "Response Time",
                    "Resource Usage"
                )
            )
            
            # 信号数量
            fig.add_trace(
                go.Scatter(
                    x=agent_metrics['timestamp'],
                    y=agent_metrics['signal_count'],
                    mode='lines',
                    name='Signal Count'
                ),
                row=1, col=1
            )
            
            # 信号质量
            fig.add_trace(
                go.Scatter(
                    x=agent_metrics['timestamp'],
                    y=agent_metrics['signal_quality'],
                    mode='lines',
                    name='Signal Quality'
                ),
                row=1, col=2
            )
            
            # 响应时间
            fig.add_trace(
                go.Scatter(
                    x=agent_metrics['timestamp'],
                    y=agent_metrics['response_time'],
                    mode='lines',
                    name='Response Time'
                ),
                row=2, col=1
            )
            
            # 资源使用
            fig.add_trace(
                go.Scatter(
                    x=agent_metrics['timestamp'],
                    y=agent_metrics['cpu_usage'],
                    mode='lines',
                    name='CPU Usage'
                ),
                row=2, col=2
            )
            
            fig.add_trace(
                go.Scatter(
                    x=agent_metrics['timestamp'],
                    y=agent_metrics['memory_usage'],
                    mode='lines',
                    name='Memory Usage'
                ),
                row=2, col=2
            )
            
            fig.update_layout(height=800, showlegend=True)
            st.plotly_chart(fig, use_container_width=True)

if __name__ == "__main__":
    main() 