import React, { useState, useEffect } from 'react';
import { Globe, ShieldCheck, Plus, RefreshCw, Lock } from 'lucide-react';
import { NginxSite, SSLCertificate } from '../types';

export const WebDomains: React.FC = () => {
  const [sites, setSites] = useState<NginxSite[]>([]);
  const [certs, setCerts] = useState<SSLCertificate[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [showNewModal, setShowNewModal] = useState<boolean>(false);
  const [domain, setDomain] = useState<string>('');
  const [proxyPass, setProxyPass] = useState<string>('http://127.0.0.1:8000');
  const [vhostType, setVhostType] = useState<string>('reverse_proxy');

  const fetchData = async () => {
    setLoading(true);
    try {
      const [sitesRes, certsRes] = await Promise.all([
        fetch('/api/v1/nginx/sites'),
        fetch('/api/v1/web/ssl/certificates'),
      ]);
      if (sitesRes.ok) setSites((await sitesRes.json()) || []);
      if (certsRes.ok) setCerts((await certsRes.json()) || []);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleCreateVHost = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const res = await fetch('/api/v1/web/vhosts', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          domain,
          type: vhostType,
          proxy_pass: vhostType === 'reverse_proxy' ? proxyPass : undefined,
          document_root: vhostType === 'static' ? `/var/www/${domain}` : undefined,
        }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed');
      alert(data.message);
      setShowNewModal(false);
      setDomain('');
      fetchData();
    } catch (e: any) {
      alert('Error creating vhost: ' + e.message);
    }
  };

  const handleIssueSSL = async (siteDomain: string) => {
    const email = window.prompt(`Enter admin email for ${siteDomain} Let's Encrypt certificate:`);
    if (email === null) return;

    try {
      const res = await fetch('/api/v1/web/ssl/issue', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ domain: siteDomain, email }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed');
      alert(data.message);
      fetchData();
    } catch (e: any) {
      alert('Error issuing SSL: ' + e.message);
    }
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Nginx Virtual Hosts */}
      <div className="panel-card">
        <div className="panel-header">
          <div>
            <div className="panel-title">
              <Globe size={18} /> Virtual Hosts & Reverse Proxies ({sites.length})
            </div>
            <div className="panel-subtitle">
              Nginx server blocks discovered in /etc/nginx/sites-enabled
            </div>
          </div>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <button className="btn" onClick={fetchData}>
              <RefreshCw size={14} /> Refresh
            </button>
            <button className="btn btn-primary" onClick={() => setShowNewModal(true)}>
              <Plus size={14} /> New Virtual Host
            </button>
          </div>
        </div>

        {loading ? (
          <div style={{ padding: '2.5rem', textAlign: 'center', color: 'var(--text-muted)' }}>
            Discovering Nginx virtual hosts...
          </div>
        ) : sites.length === 0 ? (
          <div style={{ padding: '2.5rem', textAlign: 'center', color: 'var(--text-muted)' }}>
            No Nginx virtual hosts discovered.
          </div>
        ) : (
          <div className="table-responsive">
            <table className="custom-table">
              <thead>
                <tr>
                  <th>Domain Name</th>
                  <th>Target / Document Root</th>
                  <th>SSL Status</th>
                  <th>Owner</th>
                  <th>Config Path</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {sites.map((s) => (
                  <tr key={s.domain + s.config_file}>
                    <td>
                      <strong style={{ color: 'var(--text-main)' }}>{s.domain}</strong>
                    </td>
                    <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.78rem', color: 'var(--primary)' }}>
                      {s.proxy_pass || s.document_root || '-'}
                    </td>
                    <td>
                      <span className={`badge ${s.ssl ? 'badge-emerald' : 'badge-amber'}`}>
                        {s.ssl ? 'HTTPS (Active)' : 'HTTP Only'}
                      </span>
                    </td>
                    <td>
                      <span className="badge-owner">{s.owner_type}</span>
                    </td>
                    <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                      {s.config_file}
                    </td>
                    <td>
                      {!s.ssl && (
                        <button className="btn btn-sm" onClick={() => handleIssueSSL(s.domain)}>
                          <Lock size={12} /> Issue SSL
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* SSL Certificates */}
      <div className="panel-card">
        <div className="panel-header">
          <div className="panel-title">
            <ShieldCheck size={18} /> TLS / SSL Certificates ({certs.length})
          </div>
        </div>

        {certs.length === 0 ? (
          <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--text-muted)' }}>
            No active Let's Encrypt certificates detected in /etc/letsencrypt/live.
          </div>
        ) : (
          <div className="table-responsive">
            <table className="custom-table">
              <thead>
                <tr>
                  <th>Domain</th>
                  <th>Issuer</th>
                  <th>Days Remaining</th>
                  <th>Auto Renew</th>
                  <th>Expiry Date</th>
                </tr>
              </thead>
              <tbody>
                {certs.map((c) => (
                  <tr key={c.domain}>
                    <td style={{ fontWeight: 600, color: 'var(--text-main)' }}>{c.domain}</td>
                    <td>{c.issuer}</td>
                    <td>
                      <span className="badge badge-emerald">{c.days_remaining} days</span>
                    </td>
                    <td>{c.auto_renew ? 'Active' : 'Manual'}</td>
                    <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.8rem' }}>
                      {new Date(c.valid_to).toLocaleDateString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Create Modal */}
      {showNewModal && (
        <div className="modal-backdrop" onClick={() => setShowNewModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '520px' }}>
            <div className="modal-header">
              <h4>Add New Virtual Host</h4>
              <button className="btn btn-sm" onClick={() => setShowNewModal(false)}>
                ✕
              </button>
            </div>
            <form onSubmit={handleCreateVHost} style={{ padding: '1.25rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
                  Domain Name
                </label>
                <input
                  type="text"
                  required
                  placeholder="e.g. app.example.com"
                  value={domain}
                  onChange={(e) => setDomain(e.target.value)}
                  style={{ width: '100%', fontFamily: 'var(--font-mono)' }}
                />
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
                  Virtual Host Type
                </label>
                <select
                  value={vhostType}
                  onChange={(e) => setVhostType(e.target.value)}
                  style={{ width: '100%' }}
                >
                  <option value="reverse_proxy">Reverse Proxy (Node/Django/FastAPI)</option>
                  <option value="static">Static SPA / HTML (React/Vue dist)</option>
                  <option value="php">PHP FastCGI (Laravel / WordPress)</option>
                </select>
              </div>

              {vhostType === 'reverse_proxy' && (
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
                    Proxy Pass Target
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="http://127.0.0.1:8000"
                    value={proxyPass}
                    onChange={(e) => setProxyPass(e.target.value)}
                    style={{ width: '100%', fontFamily: 'var(--font-mono)' }}
                  />
                </div>
              )}

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.5rem', marginTop: '0.5rem' }}>
                <button type="button" className="btn" onClick={() => setShowNewModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary">
                  Create & Reload Nginx
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
