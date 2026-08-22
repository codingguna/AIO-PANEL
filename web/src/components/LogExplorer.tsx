import React, { useState, useEffect } from 'react';
import { LogResponse } from '../types';

export const LogExplorer: React.FC = () => {
  const [source, setSource] = useState<string>('journalctl');
  const [lines, setLines] = useState<number>(100);
  const [logContent, setLogContent] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(true);
  const [searchFilter, setSearchFilter] = useState<string>('');
  const [autoRefresh, setAutoRefresh] = useState<boolean>(false);

  const fetchLogs = async () => {
    setLoading(true);
    try {
      const res = await fetch(`/api/v1/ops/logs?source=${source}&lines=${lines}`);
      if (res.ok) {
        const data: LogResponse = await res.json();
        setLogContent(data.content || 'No log data received.');
      } else {
        setLogContent('Failed to fetch logs from server.');
      }
    } catch (e) {
      setLogContent('Error loading logs: ' + e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchLogs();
  }, [source, lines]);

  useEffect(() => {
    if (!autoRefresh) return;
    const interval = setInterval(fetchLogs, 3000);
    return () => clearInterval(interval);
  }, [autoRefresh, source, lines]);

  const sources = [
    { id: 'journalctl', label: 'systemd Journal', icon: '⚙️' },
    { id: 'nginx-access', label: 'Nginx Access Log', icon: '🌐' },
    { id: 'nginx-error', label: 'Nginx Error Log', icon: '⚠️' },
    { id: 'auth', label: 'Auth Log (SSH/Sudo)', icon: '🔒' },
    { id: 'syslog', label: 'System Syslog', icon: '📜' },
    { id: 'aio', label: 'AIO-PANEL Daemon Log', icon: '🛡️' },
  ];

  const filteredContent = logContent
    .split('\n')
    .filter((line) => !searchFilter || line.toLowerCase().includes(searchFilter.toLowerCase()))
    .join('\n');

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Log Sources Toolbar */}
      <div className="panel-card">
        <div className="panel-header" style={{ marginBottom: '1rem' }}>
          <div>
            <div className="panel-title">
              <span>📑</span> Centralized Log Explorer
            </div>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.82rem', marginTop: '0.2rem' }}>
              Real-time multi-source server log inspection directly from native Linux journald & /var/log
            </p>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.65rem' }}>
            <label style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontSize: '0.82rem', color: 'var(--text-muted)', cursor: 'pointer' }}>
              <input
                type="checkbox"
                checked={autoRefresh}
                onChange={(e) => setAutoRefresh(e.target.checked)}
              />
              Live Tail (3s)
            </label>
            <button className="btn" onClick={fetchLogs} disabled={loading}>
              🔄 Refresh
            </button>
          </div>
        </div>

        {/* Source Switcher Pills */}
        <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap', marginBottom: '1rem' }}>
          {sources.map((s) => (
            <button
              key={s.id}
              className="btn"
              onClick={() => setSource(s.id)}
              style={{
                background: source === s.id ? 'var(--primary)' : 'var(--bg-base)',
                borderColor: source === s.id ? 'var(--primary)' : 'var(--border)',
                color: source === s.id ? '#fff' : 'var(--text-muted)',
              }}
            >
              <span>{s.icon}</span>
              <span>{s.label}</span>
            </button>
          ))}
        </div>

        {/* Filter Bar */}
        <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
          <input
            type="text"
            placeholder="Filter log lines by keyword (e.g. error, 404, root, GET)..."
            value={searchFilter}
            onChange={(e) => setSearchFilter(e.target.value)}
            style={{
              flex: 1,
              padding: '0.6rem 0.85rem',
              background: 'var(--bg-base)',
              border: '1px solid var(--border)',
              borderRadius: '8px',
              color: '#fff',
              fontFamily: 'var(--font-mono)',
              fontSize: '0.82rem',
            }}
          />
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <span style={{ fontSize: '0.8rem', color: 'var(--text-subtle)' }}>Lines:</span>
            <select
              value={lines}
              onChange={(e) => setLines(Number(e.target.value))}
              style={{
                padding: '0.55rem 0.75rem',
                background: 'var(--bg-base)',
                border: '1px solid var(--border)',
                borderRadius: '8px',
                color: '#fff',
                fontSize: '0.82rem',
              }}
            >
              <option value={50}>50 lines</option>
              <option value={100}>100 lines</option>
              <option value={250}>250 lines</option>
              <option value={500}>500 lines</option>
            </select>
          </div>
        </div>
      </div>

      {/* Log Output Screen */}
      <div className="panel-card">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: '0.78rem', color: 'var(--accent-cyan)' }}>
            Viewing: {source} ({filteredContent.split('\n').filter(Boolean).length} lines shown)
          </span>
          {loading && <span style={{ fontSize: '0.8rem', color: 'var(--accent-amber)' }}>Fetching fresh logs...</span>}
        </div>

        <div className="log-viewer-screen">
          {filteredContent || 'No matching log records found.'}
        </div>
      </div>
    </div>
  );
};
