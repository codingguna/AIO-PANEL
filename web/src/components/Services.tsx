import React, { useState, useEffect } from 'react';
import { Server, Play, Square, RotateCw, FileText, RefreshCw, Plus, Trash2 } from 'lucide-react';
import { SystemService } from '../types';

export const Services: React.FC = () => {
  const [services, setServices] = useState<SystemService[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [selectedLogs, setSelectedLogs] = useState<{ name: string; content: string } | null>(null);

  // New Service Modal State
  const [showCreateModal, setShowCreateModal] = useState<boolean>(false);
  const [newSvcName, setNewSvcName] = useState<string>('');
  const [newSvcDesc, setNewSvcDesc] = useState<string>('');
  const [newSvcExec, setNewSvcExec] = useState<string>('');
  const [newSvcDir, setNewSvcDir] = useState<string>('');
  const [newSvcUser, setNewSvcUser] = useState<string>('root');
  const [newSvcRestart, setNewSvcRestart] = useState<string>('always');
  const [newSvcEnable, setNewSvcEnable] = useState<boolean>(true);

  const fetchServices = async () => {
    setLoading(true);
    try {
      const res = await fetch('/api/v1/services');
      if (res.ok) {
        const data = await res.json();
        setServices(data || []);
      }
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchServices();
  }, []);

  const handleAction = async (name: string, action: string) => {
    if (!window.confirm(`Are you sure you want to ${action.toUpperCase()} ${name}?`)) return;
    try {
      const res = await fetch(`/api/v1/services/${name}/action`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ action }),
      });
      const data = await res.json();
      alert(data.message);
      fetchServices();
    } catch (e) {
      alert('Error executing action: ' + e);
    }
  };

  const handleViewLogs = async (name: string) => {
    setSelectedLogs({ name, content: 'Fetching journalctl logs...' });
    try {
      const res = await fetch(`/api/v1/services/${name}/logs?lines=60`);
      const text = await res.text();
      setSelectedLogs({ name, content: text || 'No logs recorded.' });
    } catch (e) {
      setSelectedLogs({ name, content: 'Error loading logs: ' + e });
    }
  };

  const handleCreateService = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const res = await fetch('/api/v1/services/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: newSvcName,
          description: newSvcDesc,
          exec_start: newSvcExec,
          work_dir: newSvcDir,
          user: newSvcUser,
          restart: newSvcRestart,
          enable: newSvcEnable,
        }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed');
      alert(data.message);
      setShowCreateModal(false);
      setNewSvcName('');
      setNewSvcExec('');
      setNewSvcDir('');
      fetchServices();
    } catch (e: any) {
      alert('Error creating service: ' + e.message);
    }
  };

  const handleDeleteService = async (name: string) => {
    if (!window.confirm(`⚠️ Are you sure you want to stop, disable, and DELETE ${name}.service from systemd?`)) return;
    try {
      const res = await fetch(`/api/v1/services/${name}`, { method: 'DELETE' });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed');
      alert(data.message);
      fetchServices();
    } catch (e: any) {
      alert('Error deleting service: ' + e.message);
    }
  };

  const formatBytes = (bytes: number) => {
    if (!bytes || bytes === 0) return '-';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i];
  };

  return (
    <div className="panel-card">
      <div className="panel-header">
        <div>
          <div className="panel-title">
            <Server size={18} /> Native System Services ({services.length})
          </div>
          <div className="panel-subtitle">
            Discovered from systemd service manager with full lifecycle management & creation
          </div>
        </div>
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button className="btn" onClick={fetchServices}>
            <RefreshCw size={14} /> Refresh
          </button>
          <button className="btn btn-primary" onClick={() => setShowCreateModal(true)}>
            <Plus size={14} /> Create Service
          </button>
        </div>
      </div>

      {loading ? (
        <div style={{ padding: '2.5rem', textAlign: 'center', color: 'var(--text-muted)' }}>
          Discovering systemd services...
        </div>
      ) : services.length === 0 ? (
        <div style={{ padding: '2.5rem', textAlign: 'center', color: 'var(--text-muted)' }}>
          No active systemd services discovered on this host.
        </div>
      ) : (
        <div className="table-responsive">
          <table className="custom-table">
            <thead>
              <tr>
                <th>Service Name</th>
                <th>Description</th>
                <th>Status</th>
                <th>Owner</th>
                <th>PID / Memory</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {services.map((s) => {
                const isActive = s.active_state === 'active';
                return (
                  <tr key={s.name}>
                    <td>
                      <div style={{ fontWeight: 600, color: 'var(--text-main)' }}>{s.display_name}</div>
                      <div style={{ fontFamily: 'var(--font-mono)', fontSize: '0.72rem', color: 'var(--text-subtle)' }}>
                        {s.unit_file}
                      </div>
                    </td>
                    <td style={{ color: 'var(--text-muted)', fontSize: '0.8rem' }}>{s.description || '-'}</td>
                    <td>
                      <span className={`badge ${isActive ? 'badge-emerald' : 'badge-rose'}`}>
                        {s.active_state} ({s.sub_state})
                      </span>
                    </td>
                    <td>
                      <span className="badge-owner">{s.owner_type}</span>
                    </td>
                    <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.78rem' }}>
                      PID: {s.pid > 0 ? s.pid : '-'}
                      <br />
                      RAM: {formatBytes(s.memory_bytes)}
                    </td>
                    <td>
                      <div style={{ display: 'flex', gap: '0.35rem' }}>
                        {isActive ? (
                          <>
                            <button className="btn btn-sm" onClick={() => handleAction(s.name, 'restart')}>
                              <RotateCw size={12} /> Restart
                            </button>
                            <button className="btn btn-danger btn-sm" onClick={() => handleAction(s.name, 'stop')}>
                              <Square size={12} /> Stop
                            </button>
                          </>
                        ) : (
                          <button className="btn btn-sm" onClick={() => handleAction(s.name, 'start')}>
                            <Play size={12} /> Start
                          </button>
                        )}
                        <button className="btn btn-sm" onClick={() => handleViewLogs(s.name)}>
                          <FileText size={12} /> Logs
                        </button>
                        <button className="btn btn-danger btn-sm" onClick={() => handleDeleteService(s.name)} title="Delete unit file">
                          <Trash2 size={12} />
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* Logs Modal */}
      {selectedLogs && (
        <div className="modal-backdrop" onClick={() => setSelectedLogs(null)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '800px' }}>
            <div className="modal-header">
              <h4>Journal Logs: {selectedLogs.name}.service</h4>
              <button className="btn btn-sm" onClick={() => setSelectedLogs(null)}>
                ✕
              </button>
            </div>
            <div className="modal-body">
              <div className="terminal-box" style={{ maxHeight: '420px' }}>{selectedLogs.content}</div>
            </div>
          </div>
        </div>
      )}

      {/* Create Service Modal */}
      {showCreateModal && (
        <div className="modal-backdrop" onClick={() => setShowCreateModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '560px' }}>
            <div className="modal-header">
              <h4>Create New Systemd Service Unit</h4>
              <button className="btn btn-sm" onClick={() => setShowCreateModal(false)}>
                ✕
              </button>
            </div>
            <form onSubmit={handleCreateService} style={{ padding: '1.25rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
                  Service Name
                </label>
                <input
                  type="text"
                  required
                  placeholder="e.g. backend-api"
                  value={newSvcName}
                  onChange={(e) => setNewSvcName(e.target.value)}
                  style={{ width: '100%', fontFamily: 'var(--font-mono)' }}
                />
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
                  Description
                </label>
                <input
                  type="text"
                  placeholder="e.g. Production Node.js Backend Service"
                  value={newSvcDesc}
                  onChange={(e) => setNewSvcDesc(e.target.value)}
                  style={{ width: '100%' }}
                />
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
                  ExecStart (Command to Run)
                </label>
                <input
                  type="text"
                  required
                  placeholder="e.g. /usr/bin/node /var/www/my-app/dist/index.js"
                  value={newSvcExec}
                  onChange={(e) => setNewSvcExec(e.target.value)}
                  style={{ width: '100%', fontFamily: 'var(--font-mono)' }}
                />
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
                    Working Directory
                  </label>
                  <input
                    type="text"
                    placeholder="e.g. /var/www/my-app"
                    value={newSvcDir}
                    onChange={(e) => setNewSvcDir(e.target.value)}
                    style={{ width: '100%', fontFamily: 'var(--font-mono)' }}
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
                    Run as User
                  </label>
                  <input
                    type="text"
                    placeholder="root or www-data"
                    value={newSvcUser}
                    onChange={(e) => setNewSvcUser(e.target.value)}
                    style={{ width: '100%' }}
                  />
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
                    Restart Policy
                  </label>
                  <select
                    value={newSvcRestart}
                    onChange={(e) => setNewSvcRestart(e.target.value)}
                    style={{ width: '100%' }}
                  >
                    <option value="always">always (Auto-restart on crash)</option>
                    <option value="on-failure">on-failure (Restart on non-zero exit)</option>
                    <option value="no">no</option>
                  </select>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', marginTop: '1.25rem' }}>
                  <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.82rem', cursor: 'pointer' }}>
                    <input
                      type="checkbox"
                      checked={newSvcEnable}
                      onChange={(e) => setNewSvcEnable(e.target.checked)}
                    />
                    Enable & Start immediately
                  </label>
                </div>
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.5rem', marginTop: '0.5rem' }}>
                <button type="button" className="btn" onClick={() => setShowCreateModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary">
                  Create & Register Unit
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
