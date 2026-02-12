#!/usr/bin/env python3
"""
performance_dashboard.py - Simulation Speed Metrics Dashboard

Creates performance tracking dashboard for M2Sim with:
- Historical performance trends across commits
- Cross-mode performance comparison (emulation vs timing vs fast-timing)
- Performance regression detection and alerting
- Benchmark-specific performance analysis

Usage:
    python3 performance_dashboard.py --data-dir <perf_data> --output-dir <dashboard>
"""

import argparse
import json
import sqlite3
from datetime import datetime, timedelta
from dataclasses import dataclass, asdict
from pathlib import Path
from typing import Dict, List, Optional, Tuple
import subprocess


@dataclass
class PerformanceMetric:
    """Single performance measurement point."""
    timestamp: str
    commit_hash: str
    commit_message: str
    mode: str  # emulation, timing, fast-timing
    benchmark: str
    suite: str  # microbench, polybench
    instructions_per_sec: float
    cpi: float
    memory_usage_mb: float
    elapsed_time_sec: float
    success: bool


class PerformanceDatabase:
    """SQLite database for storing performance metrics."""

    def __init__(self, db_path: Path):
        self.db_path = db_path
        self._init_db()

    def _init_db(self):
        """Initialize database schema."""
        conn = sqlite3.connect(self.db_path)
        conn.execute('''
            CREATE TABLE IF NOT EXISTS performance_metrics (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                timestamp TEXT NOT NULL,
                commit_hash TEXT NOT NULL,
                commit_message TEXT,
                mode TEXT NOT NULL,
                benchmark TEXT NOT NULL,
                suite TEXT NOT NULL,
                instructions_per_sec REAL,
                cpi REAL,
                memory_usage_mb REAL,
                elapsed_time_sec REAL,
                success INTEGER NOT NULL,
                UNIQUE(timestamp, commit_hash, mode, benchmark, suite)
            )
        ''')
        conn.execute('CREATE INDEX IF NOT EXISTS idx_commit_hash ON performance_metrics(commit_hash)')
        conn.execute('CREATE INDEX IF NOT EXISTS idx_timestamp ON performance_metrics(timestamp)')
        conn.execute('CREATE INDEX IF NOT EXISTS idx_benchmark ON performance_metrics(benchmark)')
        conn.commit()
        conn.close()

    def store_metrics(self, metrics: List[PerformanceMetric]):
        """Store performance metrics in database."""
        conn = sqlite3.connect(self.db_path)

        for metric in metrics:
            conn.execute('''
                INSERT OR REPLACE INTO performance_metrics
                (timestamp, commit_hash, commit_message, mode, benchmark, suite,
                 instructions_per_sec, cpi, memory_usage_mb, elapsed_time_sec, success)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            ''', (
                metric.timestamp, metric.commit_hash, metric.commit_message,
                metric.mode, metric.benchmark, metric.suite,
                metric.instructions_per_sec, metric.cpi, metric.memory_usage_mb,
                metric.elapsed_time_sec, 1 if metric.success else 0
            ))

        conn.commit()
        conn.close()

    def get_metrics(self,
                   benchmark: Optional[str] = None,
                   mode: Optional[str] = None,
                   days_back: int = 30) -> List[PerformanceMetric]:
        """Retrieve performance metrics with filtering."""
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row

        query = '''
            SELECT * FROM performance_metrics
            WHERE timestamp >= datetime('now', '-{} days')
        '''.format(days_back)

        params = []
        if benchmark:
            query += ' AND benchmark = ?'
            params.append(benchmark)
        if mode:
            query += ' AND mode = ?'
            params.append(mode)

        query += ' ORDER BY timestamp DESC'

        cursor = conn.execute(query, params)
        rows = cursor.fetchall()
        conn.close()

        return [PerformanceMetric(
            timestamp=row['timestamp'],
            commit_hash=row['commit_hash'],
            commit_message=row['commit_message'],
            mode=row['mode'],
            benchmark=row['benchmark'],
            suite=row['suite'],
            instructions_per_sec=row['instructions_per_sec'],
            cpi=row['cpi'],
            memory_usage_mb=row['memory_usage_mb'],
            elapsed_time_sec=row['elapsed_time_sec'],
            success=bool(row['success'])
        ) for row in rows]

    def detect_regressions(self, threshold_percent: float = 10.0) -> List[Dict]:
        """Detect performance regressions compared to recent baseline."""
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row

        # Get latest metrics for each benchmark/mode combination
        query = '''
            WITH latest_metrics AS (
                SELECT *, ROW_NUMBER() OVER (
                    PARTITION BY benchmark, mode
                    ORDER BY timestamp DESC
                ) as rn
                FROM performance_metrics
                WHERE success = 1
            ),
            baseline_metrics AS (
                SELECT *, ROW_NUMBER() OVER (
                    PARTITION BY benchmark, mode
                    ORDER BY timestamp DESC
                ) as rn
                FROM performance_metrics
                WHERE success = 1 AND timestamp <= datetime('now', '-7 days')
            )
            SELECT
                l.benchmark, l.mode, l.suite,
                l.instructions_per_sec as current_perf,
                b.instructions_per_sec as baseline_perf,
                l.timestamp as current_timestamp,
                b.timestamp as baseline_timestamp,
                l.commit_hash as current_commit,
                b.commit_hash as baseline_commit,
                ((l.instructions_per_sec - b.instructions_per_sec) / b.instructions_per_sec * 100) as change_percent
            FROM latest_metrics l
            JOIN baseline_metrics b ON l.benchmark = b.benchmark AND l.mode = b.mode
            WHERE l.rn = 1 AND b.rn = 1
            AND l.instructions_per_sec < b.instructions_per_sec * (1 - ? / 100)
        '''

        cursor = conn.execute(query, (threshold_percent,))
        regressions = [dict(row) for row in cursor.fetchall()]
        conn.close()

        return regressions


def generate_dashboard_html(metrics: List[PerformanceMetric],
                          regressions: List[Dict],
                          output_dir: Path):
    """Generate HTML performance dashboard."""

    # Group metrics by benchmark and mode for plotting
    benchmark_data = {}
    for metric in metrics:
        key = f"{metric.benchmark}_{metric.mode}"
        if key not in benchmark_data:
            benchmark_data[key] = []
        benchmark_data[key].append(metric)

    html_content = f'''
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>M2Sim Performance Dashboard</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        body {{ font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }}
        .container {{ max-width: 1200px; margin: 0 auto; }}
        .card {{ background: white; padding: 20px; margin: 20px 0; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }}
        .chart-container {{ height: 400px; margin: 20px 0; }}
        .regression-alert {{ background-color: #ffebee; border-left: 4px solid #f44336; padding: 10px; margin: 10px 0; }}
        .metrics-grid {{ display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }}
        .metric-card {{ background: #e3f2fd; padding: 15px; border-radius: 4px; }}
        .trend-up {{ color: #4caf50; }}
        .trend-down {{ color: #f44336; }}
        table {{ width: 100%; border-collapse: collapse; }}
        th, td {{ padding: 8px 12px; text-align: left; border-bottom: 1px solid #ddd; }}
        th {{ background-color: #f5f5f5; }}
    </style>
</head>
<body>
    <div class="container">
        <h1>M2Sim Performance Dashboard</h1>
        <p><strong>Last Updated:</strong> {datetime.now().strftime("%Y-%m-%d %H:%M:%S UTC")}</p>

        <!-- Performance Regression Alerts -->
        <div class="card">
            <h2>🚨 Performance Regression Alerts</h2>
            {_generate_regression_alerts_html(regressions)}
        </div>

        <!-- Performance Overview -->
        <div class="card">
            <h2>📊 Performance Overview</h2>
            <div class="metrics-grid">
                {_generate_overview_cards_html(metrics)}
            </div>
        </div>

        <!-- Performance Trends Charts -->
        <div class="card">
            <h2>📈 Performance Trends</h2>
            {_generate_charts_html(benchmark_data)}
        </div>

        <!-- Detailed Metrics Table -->
        <div class="card">
            <h2>📋 Detailed Performance Metrics</h2>
            {_generate_metrics_table_html(metrics[:50])}  <!-- Show last 50 entries -->
        </div>
    </div>

    <script>
        {_generate_chart_js(benchmark_data)}
    </script>
</body>
</html>
'''

    # Write HTML file
    (output_dir / "performance_dashboard.html").write_text(html_content)


def _generate_regression_alerts_html(regressions: List[Dict]) -> str:
    """Generate HTML for performance regression alerts."""
    if not regressions:
        return "<p>✅ No performance regressions detected in recent commits.</p>"

    html = ""
    for regression in regressions:
        html += f'''
        <div class="regression-alert">
            <strong>⚠️ Performance Regression Detected</strong><br>
            <strong>Benchmark:</strong> {regression['benchmark']} ({regression['mode']})<br>
            <strong>Performance Drop:</strong> {regression['change_percent']:.1f}%<br>
            <strong>Current:</strong> {regression['current_perf']:.0f} inst/s
            (commit: {regression['current_commit'][:8]})<br>
            <strong>Baseline:</strong> {regression['baseline_perf']:.0f} inst/s
            (commit: {regression['baseline_commit'][:8]})<br>
        </div>
        '''
    return html


def _generate_overview_cards_html(metrics: List[PerformanceMetric]) -> str:
    """Generate overview metric cards."""
    if not metrics:
        return "<p>No performance data available</p>"

    # Calculate summary statistics
    latest_by_mode = {}
    for metric in metrics:
        if metric.mode not in latest_by_mode or metric.timestamp > latest_by_mode[metric.mode].timestamp:
            latest_by_mode[metric.mode] = metric

    html = ""
    for mode, metric in latest_by_mode.items():
        html += f'''
        <div class="metric-card">
            <h3>{mode.title()} Mode</h3>
            <p><strong>Performance:</strong> {metric.instructions_per_sec:.0f} inst/s</p>
            <p><strong>CPI:</strong> {metric.cpi:.3f}</p>
            <p><strong>Memory:</strong> {metric.memory_usage_mb:.1f} MB</p>
            <p><strong>Last Update:</strong> {metric.timestamp[:10]}</p>
        </div>
        '''

    return html


def _generate_charts_html(benchmark_data: Dict) -> str:
    """Generate chart containers."""
    html = ""
    for key in benchmark_data.keys():
        chart_id = key.replace('-', '_').replace('.', '_')
        html += f'''
        <div class="chart-container">
            <h3>{key.replace('_', ' - ').title()}</h3>
            <canvas id="chart_{chart_id}"></canvas>
        </div>
        '''
    return html


def _generate_metrics_table_html(metrics: List[PerformanceMetric]) -> str:
    """Generate detailed metrics table."""
    if not metrics:
        return "<p>No metrics data available</p>"

    html = '''
    <table>
        <thead>
            <tr>
                <th>Timestamp</th>
                <th>Commit</th>
                <th>Mode</th>
                <th>Benchmark</th>
                <th>Instructions/sec</th>
                <th>CPI</th>
                <th>Memory (MB)</th>
            </tr>
        </thead>
        <tbody>
    '''

    for metric in metrics:
        html += f'''
        <tr>
            <td>{metric.timestamp[:16]}</td>
            <td>{metric.commit_hash[:8]}</td>
            <td>{metric.mode}</td>
            <td>{metric.benchmark}</td>
            <td>{metric.instructions_per_sec:.0f}</td>
            <td>{metric.cpi:.3f}</td>
            <td>{metric.memory_usage_mb:.1f}</td>
        </tr>
        '''

    html += '</tbody></table>'
    return html


def _generate_chart_js(benchmark_data: Dict) -> str:
    """Generate JavaScript for Chart.js visualizations."""
    js_code = ""

    for key, metrics in benchmark_data.items():
        chart_id = key.replace('-', '_').replace('.', '_')

        # Prepare data for Chart.js
        timestamps = [m.timestamp[:10] for m in sorted(metrics, key=lambda x: x.timestamp)]
        perf_data = [m.instructions_per_sec for m in sorted(metrics, key=lambda x: x.timestamp)]

        js_code += f'''
        new Chart(document.getElementById('chart_{chart_id}'), {{
            type: 'line',
            data: {{
                labels: {json.dumps(timestamps)},
                datasets: [{{
                    label: 'Instructions/sec',
                    data: {json.dumps(perf_data)},
                    borderColor: 'rgb(75, 192, 192)',
                    backgroundColor: 'rgba(75, 192, 192, 0.2)',
                    tension: 0.1
                }}]
            }},
            options: {{
                responsive: true,
                maintainAspectRatio: false,
                scales: {{
                    y: {{
                        beginAtZero: false,
                        title: {{
                            display: true,
                            text: 'Instructions per Second'
                        }}
                    }},
                    x: {{
                        title: {{
                            display: true,
                            text: 'Date'
                        }}
                    }}
                }}
            }}
        }});
        '''

    return js_code


def collect_git_commit_info() -> Tuple[str, str]:
    """Get current git commit hash and message."""
    try:
        commit_hash = subprocess.check_output(['git', 'rev-parse', 'HEAD']).decode().strip()
        commit_message = subprocess.check_output(['git', 'log', '-1', '--pretty=%s']).decode().strip()
        return commit_hash, commit_message
    except subprocess.SubprocessError:
        return "unknown", "unknown"


def main():
    parser = argparse.ArgumentParser(description='Generate M2Sim performance dashboard')
    parser.add_argument('--data-dir', type=Path, required=True,
                       help='Directory containing performance data files')
    parser.add_argument('--output-dir', type=Path, default='dashboard',
                       help='Output directory for dashboard files')
    parser.add_argument('--db-path', type=Path, default='performance_metrics.db',
                       help='SQLite database path for performance metrics')

    args = parser.parse_args()

    # Create output directory
    args.output_dir.mkdir(exist_ok=True)

    # Initialize database
    db = PerformanceDatabase(args.db_path)

    # TODO: Import performance data from profiling results
    # For now, create sample data structure
    sample_metrics = []

    print(f"Generating dashboard in {args.output_dir}")

    # Get recent metrics and detect regressions
    recent_metrics = db.get_metrics(days_back=30)
    regressions = db.detect_regressions(threshold_percent=10.0)

    # Generate dashboard
    generate_dashboard_html(recent_metrics, regressions, args.output_dir)

    print(f"Performance dashboard generated: {args.output_dir}/performance_dashboard.html")
    if regressions:
        print(f"⚠️  {len(regressions)} performance regressions detected!")
    else:
        print("✅ No performance regressions detected")


if __name__ == '__main__':
    exit(main())