import React, { useState, useEffect } from 'react';
import {
  ShoppingBag,
  Search,
  CheckCircle,
  Download,
  RotateCw,
  Globe,
  Server,
  Zap,
  Layers,
  FileCode,
  Database,
  Activity,
  Box,
  Shield,
  GitBranch,
  X,
  Terminal,
} from 'lucide-react';
import { StorePackage, InstallJob } from '../types';

export const AppStore: React.FC = () => {
  const [packages, setPackages] = useState<StorePackage[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [search, setSearch] = useState<string>('');
  const [selectedCategory, setSelectedCategory] = useState<string>('All');
  const [selectedVersions, setSelectedVersions] = useState<Record<string, string>>({});

  // Active Installation State
  const [activeJob, setActiveJob] = useState<InstallJob | null>(null);
  const [pollingPkg, setPollingPkg] = useState<string | null>(null);

  const fetchPackages = async () => {
    setLoading(true);
    try {
      const res = await fetch('/api/v1/store/packages');
      if (res.ok) {
        const data: StorePackage[] = await res.json();
        setPackages(data || []);

        // Default selected versions
        const versions: Record<string, string> = {};
        data.forEach((p) => {
          if (p.versions && p.versions.length > 0) {
            versions[p.id] = p.versions[0];
          }
        });
        setSelectedVersions((prev) => ({ ...versions, ...prev }));
      }
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPackages();
  }, []);

  // Poll active installation job
  useEffect(() => {
    if (!pollingPkg) return;

    const interval = setInterval(async () => {
      try {
        const res = await fetch(`/api/v1/store/jobs/${pollingPkg}`);
        if (res.ok) {
          const job: InstallJob = await res.json();
          setActiveJob(job);
          if (job.status !== 'RUNNING') {
            setPollingPkg(null);
            fetchPackages(); // Refresh installed status
          }
        }
      } catch (e) {
        console.error(e);
      }
    }, 1500);

    return () => clearInterval(interval);
  }, [pollingPkg]);

  const handleInstall = async (pkg: StorePackage) => {
    const version = selectedVersions[pkg.id] || pkg.versions[0] || 'latest';
    if (!window.confirm(`Install ${pkg.name} (${version}) on this server?`)) return;

    try {
      const res = await fetch('/api/v1/store/install', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          package_id: pkg.id,
          version: version,
        }),
      });

      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Installation request failed');

      setActiveJob(data);
      setPollingPkg(pkg.id);
    } catch (e: any) {
      alert('Error starting installation: ' + e.message);
    }
  };

  const getIcon = (name: string) => {
    switch (name) {
      case 'Globe': return <Globe size={22} />;
      case 'Server': return <Server size={22} />;
      case 'Zap': return <Zap size={22} />;
      case 'Layers': return <Layers size={22} />;
      case 'FileCode': return <FileCode size={22} />;
      case 'Database': return <Database size={22} />;
      case 'Activity': return <Activity size={22} />;
      case 'Box': return <Box size={22} />;
      case 'Shield': return <Shield size={22} />;
      case 'GitBranch': return <GitBranch size={22} />;
      default: return <ShoppingBag size={22} />;
    }
  };

  const categories = ['All', 'Web Server', 'Runtime', 'Database', 'Containers', 'Security', 'DevOps'];

  const filteredPackages = packages.filter((p) => {
    const matchesSearch =
      p.name.toLowerCase().includes(search.toLowerCase()) ||
      p.description.toLowerCase().includes(search.toLowerCase()) ||
      p.category.toLowerCase().includes(search.toLowerCase());
    const matchesCat = selectedCategory === 'All' || p.category === selectedCategory;
    return matchesSearch && matchesCat;
  });

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Top Search & Filter Bar */}
      <div className="panel-card" style={{ padding: '1.25rem' }}>
        <div style={{ display: 'flex', flexWrap: 'wrap', alignItems: 'center', justifyContent: 'space-between', gap: '1rem' }}>
          <div>
            <div className="panel-title">
              <ShoppingBag size={20} /> AIO Software Store & 1-Click Installer
            </div>
            <div className="panel-subtitle">
              Easily install and manage web servers, runtimes, databases, and DevOps tools
            </div>
          </div>

          {/* Search Box */}
          <div style={{ position: 'relative', minWidth: '260px' }}>
            <input
              type="text"
              placeholder="Search packages (e.g. nginx, node, postgres)..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              style={{ width: '100%', paddingLeft: '2.2rem' }}
            />
            <Search size={15} style={{ position: 'absolute', left: '0.75rem', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-subtle)' }} />
          </div>
        </div>

        {/* Category Pills */}
        <div style={{ display: 'flex', gap: '0.4rem', flexWrap: 'wrap', marginTop: '1rem', borderTop: '1px solid var(--border)', paddingTop: '0.85rem' }}>
          {categories.map((cat) => (
            <button
              key={cat}
              className={`btn btn-sm ${selectedCategory === cat ? 'btn-primary' : ''}`}
              onClick={() => setSelectedCategory(cat)}
            >
              {cat}
            </button>
          ))}
        </div>
      </div>

      {/* Package Grid */}
      {loading ? (
        <div style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-muted)' }}>
          Loading software catalog and inspecting server stack...
        </div>
      ) : filteredPackages.length === 0 ? (
        <div style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-muted)' }}>
          No software packages found matching "{search}".
        </div>
      ) : (
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))',
          gap: '1.25rem',
        }}>
          {filteredPackages.map((pkg) => {
            return (
              <div
                key={pkg.id}
                className="panel-card"
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  justifyContent: 'space-between',
                  gap: '1rem',
                  padding: '1.25rem',
                }}
              >
                {/* Header */}
                <div>
                  <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: '0.75rem' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                      <div style={{
                        width: '42px',
                        height: '42px',
                        borderRadius: '8px',
                        backgroundColor: '#f1f5f9',
                        color: 'var(--primary)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        flexShrink: 0,
                      }}>
                        {getIcon(pkg.icon)}
                      </div>
                      <div>
                        <h4 style={{ fontSize: '0.95rem', fontWeight: 600, color: 'var(--text-main)' }}>{pkg.name}</h4>
                        <span className="badge badge-gray" style={{ fontSize: '0.7rem', marginTop: '0.15rem' }}>
                          {pkg.category}
                        </span>
                      </div>
                    </div>

                    {/* Status Badge */}
                    {pkg.installed ? (
                      <span className="badge badge-emerald" style={{ flexShrink: 0 }}>
                        <CheckCircle size={12} /> Installed
                      </span>
                    ) : (
                      <span className="badge badge-gray" style={{ flexShrink: 0 }}>
                        Not Installed
                      </span>
                    )}
                  </div>

                  {/* Description */}
                  <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)', marginTop: '0.75rem', lineHeight: '1.45' }}>
                    {pkg.description}
                  </p>

                  {/* Detected Version if installed */}
                  {pkg.installed && pkg.version && (
                    <div style={{
                      marginTop: '0.6rem',
                      padding: '0.35rem 0.55rem',
                      backgroundColor: '#f8fafc',
                      border: '1px solid var(--border)',
                      borderRadius: '4px',
                      fontFamily: 'var(--font-mono)',
                      fontSize: '0.72rem',
                      color: 'var(--text-muted)',
                      wordBreak: 'break-all',
                    }}>
                      Detected: {pkg.version}
                    </div>
                  )}
                </div>

                {/* Footer Controls */}
                <div style={{ borderTop: '1px solid var(--border)', paddingTop: '0.85rem', display: 'flex', flexDirection: 'column', gap: '0.65rem' }}>
                  {/* Version Picker */}
                  {pkg.versions && pkg.versions.length > 1 && (
                    <div>
                      <label style={{ display: 'block', fontSize: '0.72rem', color: 'var(--text-muted)', marginBottom: '0.25rem' }}>
                        Version Selection
                      </label>
                      <select
                        value={selectedVersions[pkg.id] || pkg.versions[0]}
                        onChange={(e) => setSelectedVersions({ ...selectedVersions, [pkg.id]: e.target.value })}
                        style={{ width: '100%', fontSize: '0.8rem', padding: '0.35rem 0.55rem' }}
                      >
                        {pkg.versions.map((ver) => (
                          <option key={ver} value={ver}>{ver}</option>
                        ))}
                      </select>
                    </div>
                  )}

                  {/* Install / Reinstall Button */}
                  <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.5rem' }}>
                    {pkg.installed ? (
                      <button className="btn btn-sm" onClick={() => handleInstall(pkg)}>
                        <RotateCw size={12} /> Reinstall
                      </button>
                    ) : (
                      <button className="btn btn-primary btn-sm" onClick={() => handleInstall(pkg)}>
                        <Download size={12} /> 1-Click Install
                      </button>
                    )}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Live Installation Progress Modal */}
      {activeJob && (
        <div className="modal-backdrop" onClick={() => { if (activeJob.status !== 'RUNNING') setActiveJob(null); }}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '750px' }}>
            <div className="modal-header">
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Terminal size={18} style={{ color: 'var(--primary)' }} />
                <h4>
                  {activeJob.status === 'RUNNING' ? 'Installing Software Package...' : 'Installation Summary'}
                </h4>
              </div>
              {activeJob.status !== 'RUNNING' && (
                <button className="btn btn-sm" onClick={() => setActiveJob(null)}>
                  <X size={14} />
                </button>
              )}
            </div>

            <div className="modal-body">
              <div style={{ marginBottom: '0.75rem', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <div>
                  Status:{' '}
                  <span className={`badge ${
                    activeJob.status === 'SUCCESS'
                      ? 'badge-emerald'
                      : activeJob.status === 'RUNNING'
                      ? 'badge-amber'
                      : 'badge-rose'
                  }`}>
                    {activeJob.status}
                  </span>
                </div>
                {activeJob.status === 'RUNNING' && (
                  <span style={{ fontSize: '0.78rem', color: 'var(--text-muted)' }}>Streaming terminal logs...</span>
                )}
              </div>

              <div className="terminal-box" style={{ maxHeight: '420px', minHeight: '220px' }}>
                {activeJob.output}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
