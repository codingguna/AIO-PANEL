import React, { useState } from 'react';

export const Settings: React.FC = () => {
  const [port, setPort] = useState<number>(5555);
  const [bindHost, setBindHost] = useState<string>('0.0.0.0');
  const [tlsEnabled, setTlsEnabled] = useState<boolean>(false);
  const [apiKey, setApiKey] = useState<string>('aio_sec_99a8b1c4e207d5f9921b764a');

  const generateNewApiKey = () => {
    const chars = '0123456789abcdef';
    let key = 'aio_sec_';
    for (let i = 0; i < 24; i++) {
      key += chars[Math.floor(Math.random() * chars.length)];
    }
    setApiKey(key);
    alert('Generated new AIO Core API Token.\nMake sure to save it safely.');
  };

  const handleSaveSettings = (e: React.FormEvent) => {
    e.preventDefault();
    alert('Settings saved to /etc/aio/aio.conf.\nRestart aio-panel.service to apply changes.');
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Panel Identity Card */}
      <div className="panel-card">
        <div className="panel-header">
          <div>
            <div className="panel-title">
              <span>⚙️</span> AIO-PANEL Configuration & Security Settings
            </div>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.82rem', marginTop: '0.2rem' }}>
              Primary daemon configuration located at /etc/aio/aio.conf
            </p>
          </div>
        </div>

        <form onSubmit={handleSaveSettings} style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))', gap: '1.25rem', marginTop: '1rem' }}>
          <div>
            <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
              HTTP / HTTPS Port
            </label>
            <input
              type="number"
              value={port}
              onChange={(e) => setPort(Number(e.target.value))}
              style={{ width: '100%', padding: '0.6rem 0.85rem', background: 'var(--bg-base)', border: '1px solid var(--border)', borderRadius: '8px', color: '#fff', fontFamily: 'var(--font-mono)' }}
            />
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
              Host Binding Interface
            </label>
            <select
              value={bindHost}
              onChange={(e) => setBindHost(e.target.value)}
              style={{ width: '100%', padding: '0.6rem 0.85rem', background: 'var(--bg-base)', border: '1px solid var(--border)', borderRadius: '8px', color: '#fff' }}
            >
              <option value="0.0.0.0">0.0.0.0 (All Network Interfaces / Public)</option>
              <option value="127.0.0.1">127.0.0.1 (Localhost Only / High Security)</option>
            </select>
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
              Direct TLS / HTTPS Encryption
            </label>
            <select
              value={tlsEnabled ? 'true' : 'false'}
              onChange={(e) => setTlsEnabled(e.target.value === 'true')}
              style={{ width: '100%', padding: '0.6rem 0.85rem', background: 'var(--bg-base)', border: '1px solid var(--border)', borderRadius: '8px', color: '#fff' }}
            >
              <option value="false">Disabled (Plain HTTP / Proxied by Nginx)</option>
              <option value="true">Enabled (Direct ListenTLS with cert)</option>
            </select>
          </div>

          <div style={{ gridColumn: '1 / -1', display: 'flex', justifyContent: 'flex-end', marginTop: '0.5rem' }}>
            <button type="submit" className="btn btn-primary">
              💾 Save Configuration
            </button>
          </div>
        </form>
      </div>

      {/* API Token Card */}
      <div className="panel-card">
        <div className="panel-header">
          <div>
            <div className="panel-title">
              <span>🔑</span> REST API Authentication Token
            </div>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.82rem', marginTop: '0.2rem' }}>
              Use this Bearer token for CLI automation, programmatic webhooks, and remote monitoring
            </p>
          </div>
          <button className="btn" onClick={generateNewApiKey}>
            🔄 Regenerate Token
          </button>
        </div>

        <div style={{ marginTop: '1rem', display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <input
            type="text"
            readOnly
            value={apiKey}
            style={{
              flex: 1,
              padding: '0.65rem 0.85rem',
              background: 'var(--bg-base)',
              border: '1px solid var(--border)',
              borderRadius: '8px',
              color: 'var(--accent-cyan)',
              fontFamily: 'var(--font-mono)',
              fontSize: '0.85rem',
            }}
          />
          <button
            className="btn"
            onClick={() => {
              navigator.clipboard.writeText(apiKey);
              alert('API token copied to clipboard.');
            }}
          >
            📋 Copy
          </button>
        </div>
      </div>

      {/* Service Control */}
      <div className="panel-card">
        <div className="panel-header">
          <div className="panel-title">
            <span>🛡️</span> AIO Core Daemon State
          </div>
        </div>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '1rem' }}>
          <div>
            <strong style={{ color: '#fff' }}>aio-panel.service</strong>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.82rem' }}>
              Self-contained Linux systemd service running on port {port}
            </p>
          </div>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <button
              className="btn"
              onClick={() => alert('Command dispatched:\nsudo systemctl restart aio-panel')}
            >
              🔄 Restart Daemon
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
