import React from 'react';
import { Cpu, HardDrive, Zap, Activity, Server } from 'lucide-react';
import { SystemInfo, LiveMetrics } from '../types';

interface DashboardProps {
  info: SystemInfo | null;
  metrics: LiveMetrics | null;
}

export const Dashboard: React.FC<DashboardProps> = ({ info, metrics }) => {
  const formatBytes = (bytes?: number) => {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i];
  };

  const formatUptime = (seconds?: number) => {
    if (!seconds) return '0m';
    const days = Math.floor(seconds / (3600 * 24));
    const hours = Math.floor((seconds % (3600 * 24)) / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    let res = '';
    if (days > 0) res += `${days}d `;
    if (hours > 0) res += `${hours}h `;
    res += `${mins}m`;
    return res;
  };

  const cpuPct = metrics?.cpu.usage_percent.toFixed(1) || '0.0';
  const memPct = metrics?.memory.usage_percent.toFixed(1) || '0.0';
  const diskPct = metrics?.disk.usage_percent.toFixed(1) || '0.0';
  const load1 = metrics?.load_average[0].toFixed(2) || '0.00';

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* 4 Live Metrics Cards */}
      <div className="grid-metrics">
        {/* CPU */}
        <div className="metric-card">
          <div className="metric-card-header">
            <span>CPU Usage</span>
            <Cpu size={16} />
          </div>
          <div className="metric-card-value">{cpuPct}%</div>
          <div className="metric-progress-bg">
            <div
              className="metric-progress-fill fill-blue"
              style={{ width: `${Math.min(100, Math.max(0, Number(cpuPct)))}%` }}
            />
          </div>
          <div className="metric-card-sub">
            <span>{metrics?.cpu.cores || '--'} Cores</span>
            <span>Load 1m: {load1}</span>
          </div>
        </div>

        {/* RAM */}
        <div className="metric-card">
          <div className="metric-card-header">
            <span>Memory (RAM)</span>
            <Zap size={16} />
          </div>
          <div className="metric-card-value">{memPct}%</div>
          <div className="metric-progress-bg">
            <div
              className="metric-progress-fill fill-emerald"
              style={{ width: `${Math.min(100, Math.max(0, Number(memPct)))}%` }}
            />
          </div>
          <div className="metric-card-sub">
            <span>Used: {formatBytes(metrics?.memory.used_bytes)}</span>
            <span>Total: {formatBytes(metrics?.memory.total_bytes)}</span>
          </div>
        </div>

        {/* Disk */}
        <div className="metric-card">
          <div className="metric-card-header">
            <span>Disk Storage</span>
            <HardDrive size={16} />
          </div>
          <div className="metric-card-value">{diskPct}%</div>
          <div className="metric-progress-bg">
            <div
              className="metric-progress-fill fill-cyan"
              style={{ width: `${Math.min(100, Math.max(0, Number(diskPct)))}%` }}
            />
          </div>
          <div className="metric-card-sub">
            <span>Mount: {metrics?.disk.path || '/'}</span>
            <span>Free: {formatBytes(metrics?.disk.free_bytes)}</span>
          </div>
        </div>

        {/* System Load */}
        <div className="metric-card">
          <div className="metric-card-header">
            <span>Load Average</span>
            <Activity size={16} />
          </div>
          <div className="metric-card-value">{load1}</div>
          <div className="metric-progress-bg">
            <div
              className="metric-progress-fill fill-amber"
              style={{ width: `${Math.min(100, Math.max(5, Number(load1) * 20))}%` }}
            />
          </div>
          <div className="metric-card-sub">
            <span>Processes: {metrics?.processes || '--'}</span>
            <span>Uptime: {formatUptime(info?.uptime_seconds)}</span>
          </div>
        </div>
      </div>

      {/* Host Specifications Card */}
      <div className="panel-card">
        <div className="panel-header">
          <div className="panel-title">
            <Server size={18} /> Host & System Information
          </div>
          <span className="badge badge-emerald">Operational</span>
        </div>
        <table className="custom-table">
          <tbody>
            <tr>
              <td style={{ width: '30%', color: 'var(--text-muted)' }}>Hostname</td>
              <td style={{ fontFamily: 'var(--font-mono)', fontWeight: 600 }}>{info?.hostname || 'loading...'}</td>
            </tr>
            <tr>
              <td style={{ color: 'var(--text-muted)' }}>Operating System</td>
              <td>{info?.os || 'loading...'}</td>
            </tr>
            <tr>
              <td style={{ color: 'var(--text-muted)' }}>Kernel & Architecture</td>
              <td style={{ fontFamily: 'var(--font-mono)' }}>{info?.kernel} ({info?.architecture})</td>
            </tr>
            <tr>
              <td style={{ color: 'var(--text-muted)' }}>CPU Model</td>
              <td>{info?.cpu_model} ({info?.cpu_cores} Cores)</td>
            </tr>
            <tr>
              <td style={{ color: 'var(--text-muted)' }}>Go Runtime</td>
              <td style={{ fontFamily: 'var(--font-mono)', color: 'var(--primary)' }}>{info?.go_version}</td>
            </tr>
            <tr>
              <td style={{ color: 'var(--text-muted)' }}>AIO-PANEL Version</td>
              <td style={{ fontFamily: 'var(--font-mono)', color: 'var(--accent-emerald)', fontWeight: 600 }}>
                v{info?.panel_version}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  );
};
