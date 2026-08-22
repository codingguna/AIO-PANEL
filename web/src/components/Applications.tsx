import React, { useState, useEffect } from 'react';
import { Layers, RefreshCw } from 'lucide-react';
import { DiscoveredApp } from '../types';

export const Applications: React.FC = () => {
  const [apps, setApps] = useState<DiscoveredApp[]>([]);
  const [loading, setLoading] = useState<boolean>(true);

  const fetchApps = async () => {
    setLoading(true);
    try {
      const res = await fetch('/api/v1/applications');
      if (res.ok) {
        const data = await res.json();
        setApps(data || []);
      }
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchApps();
  }, []);

  return (
    <div className="panel-card">
      <div className="panel-header">
        <div>
          <div className="panel-title">
            <Layers size={18} /> Discovered Web Applications ({apps.length})
          </div>
          <div className="panel-subtitle">
            Django, Node.js, Next.js, React SPA, and PHP applications discovered in /var/www, /srv, /home
          </div>
        </div>
        <button className="btn" onClick={fetchApps}>
          <RefreshCw size={14} /> Scan Filesystem
        </button>
      </div>

      {loading ? (
        <div style={{ padding: '2.5rem', textAlign: 'center', color: 'var(--text-muted)' }}>
          Scanning directories for web applications...
        </div>
      ) : apps.length === 0 ? (
        <div style={{ padding: '2.5rem', textAlign: 'center', color: 'var(--text-muted)' }}>
          No web applications detected in standard directories (/var/www, /srv, /home).
        </div>
      ) : (
        <div className="table-responsive">
          <table className="custom-table">
            <thead>
              <tr>
                <th>App Name</th>
                <th>Framework</th>
                <th>Directory Path</th>
                <th>Runtime</th>
                <th>Linked Service</th>
                <th>Nginx Domain</th>
                <th>Owner</th>
              </tr>
            </thead>
            <tbody>
              {apps.map((app) => (
                <tr key={app.path}>
                  <td>
                    <strong style={{ color: 'var(--text-main)' }}>{app.name}</strong>
                  </td>
                  <td>
                    <span className="badge badge-blue">
                      {app.type}
                    </span>
                  </td>
                  <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                    {app.path}
                  </td>
                  <td style={{ fontSize: '0.8rem' }}>{app.runtime}</td>
                  <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem', color: 'var(--primary)' }}>
                    {app.service || '-'}
                  </td>
                  <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem' }}>
                    {app.nginx_domain || '-'}
                  </td>
                  <td>
                    <span className="badge-owner">{app.owner_type}</span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};
