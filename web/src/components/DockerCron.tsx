import React, { useState, useEffect } from 'react';
import { Box, Clock, Play, Square, RotateCw, FileText, Plus, Trash2, RefreshCw, X } from 'lucide-react';
import { DockerContainer, CronJob, DockerImage } from '../types';

export const DockerCron: React.FC = () => {
  const [containers, setContainers] = useState<DockerContainer[]>([]);
  const [images, setImages] = useState<DockerImage[]>([]);
  const [cronJobs, setCronJobs] = useState<CronJob[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [activeLogs, setActiveLogs] = useState<{ id: string; name: string; content: string } | null>(null);

  // New Cron Modal State
  const [showCronModal, setShowCronModal] = useState<boolean>(false);
  const [cronSchedule, setCronSchedule] = useState<string>('0 2 * * *');
  const [cronCommand, setCronCommand] = useState<string>('');

  const fetchData = async () => {
    setLoading(true);
    try {
      const [docRes, imgRes, cronRes] = await Promise.all([
        fetch('/api/v1/ops/docker/containers'),
        fetch('/api/v1/ops/docker/images'),
        fetch('/api/v1/ops/cron'),
      ]);
      if (docRes.ok) setContainers((await docRes.json()) || []);
      if (imgRes.ok) setImages((await imgRes.json()) || []);
      if (cronRes.ok) setCronJobs((await cronRes.json()) || []);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleContainerAction = async (id: string, action: string) => {
    try {
      const res = await fetch(`/api/v1/ops/docker/containers/${id}/action`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ action }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed');
      alert(data.message);
      fetchData();
    } catch (e: any) {
      alert('Error: ' + e.message);
    }
  };

  const handleDeleteContainer = async (id: string, name: string) => {
    if (!window.confirm(`⚠️ Are you sure you want to force REMOVE Docker container '${name}' (${id.substring(0, 12)})?`)) return;
    try {
      const res = await fetch(`/api/v1/ops/docker/containers/${id}`, { method: 'DELETE' });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed');
      alert(data.message);
      fetchData();
    } catch (e: any) {
      alert('Error removing container: ' + e.message);
    }
  };

  const handleViewContainerLogs = async (c: DockerContainer) => {
    setActiveLogs({ id: c.id, name: c.names, content: 'Fetching container logs...' });
    try {
      const res = await fetch(`/api/v1/ops/docker/containers/${c.id}/logs?lines=100`);
      const text = await res.text();
      setActiveLogs({ id: c.id, name: c.names, content: text || 'No logs recorded.' });
    } catch (e) {
      setActiveLogs({ id: c.id, name: c.names, content: 'Error loading logs: ' + e });
    }
  };

  const handleCreateCron = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const res = await fetch('/api/v1/ops/cron/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          schedule: cronSchedule,
          command: cronCommand,
        }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed');
      alert(data.message);
      setShowCronModal(false);
      setCronCommand('');
      fetchData();
    } catch (e: any) {
      alert('Error adding cron job: ' + e.message);
    }
  };

  const handleDeleteCron = async (id: number, cmd: string) => {
    if (!window.confirm(`⚠️ Remove scheduled cron job #${id} (${cmd})?`)) return;
    try {
      const res = await fetch(`/api/v1/ops/cron/${id}`, { method: 'DELETE' });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed');
      alert(data.message);
      fetchData();
    } catch (e: any) {
      alert('Error deleting cron job: ' + e.message);
    }
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Docker Containers */}
      <div className="panel-card">
        <div className="panel-header">
          <div>
            <div className="panel-title">
              <Box size={18} /> Docker Containers ({containers.length})
            </div>
            <div className="panel-subtitle">
              Isolated containers discovered via native Docker Engine socket
            </div>
          </div>
          <button className="btn" onClick={fetchData}>
            <RefreshCw size={14} /> Refresh
          </button>
        </div>

        {loading ? (
          <div style={{ padding: '2.5rem', textAlign: 'center', color: 'var(--text-muted)' }}>Querying Docker Engine...</div>
        ) : containers.length === 0 ? (
          <div style={{ padding: '2.5rem', textAlign: 'center', color: 'var(--text-muted)' }}>
            No Docker containers active (or Docker not installed on host).
          </div>
        ) : (
          <div className="table-responsive">
            <table className="custom-table">
              <thead>
                <tr>
                  <th>Container Name</th>
                  <th>Image</th>
                  <th>State</th>
                  <th>Port Mappings</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {containers.map((c) => {
                  const isRunning = c.state === 'running';
                  return (
                    <tr key={c.id}>
                      <td>
                        <strong style={{ color: 'var(--text-main)' }}>{c.names}</strong>
                        <div style={{ fontFamily: 'var(--font-mono)', fontSize: '0.72rem', color: 'var(--text-subtle)' }}>
                          {c.id.substring(0, 12)}
                        </div>
                      </td>
                      <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.8rem' }}>{c.image}</td>
                      <td>
                        <span className={`badge ${isRunning ? 'badge-emerald' : 'badge-rose'}`}>
                          {c.status}
                        </span>
                      </td>
                      <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.78rem' }}>{c.ports || '-'}</td>
                      <td>
                        <div style={{ display: 'flex', gap: '0.35rem' }}>
                          {isRunning ? (
                            <>
                              <button className="btn btn-sm" onClick={() => handleContainerAction(c.id, 'restart')} title="Restart container">
                                <RotateCw size={12} /> Restart
                              </button>
                              <button className="btn btn-danger btn-sm" onClick={() => handleContainerAction(c.id, 'stop')} title="Stop container">
                                <Square size={12} /> Stop
                              </button>
                            </>
                          ) : (
                            <button className="btn btn-sm" onClick={() => handleContainerAction(c.id, 'start')} title="Start container">
                              <Play size={12} /> Start
                            </button>
                          )}
                          <button className="btn btn-sm" onClick={() => handleViewContainerLogs(c)} title="View container logs">
                            <FileText size={12} /> Logs
                          </button>
                          <button className="btn btn-danger btn-sm" onClick={() => handleDeleteContainer(c.id, c.names)} title="Remove container">
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
      </div>

      {/* Docker Images */}
      {images.length > 0 && (
        <div className="panel-card">
          <div className="panel-header">
            <div className="panel-title">
              <Box size={18} /> Local Docker Images ({images.length})
            </div>
          </div>
          <div className="table-responsive">
            <table className="custom-table">
              <thead>
                <tr>
                  <th>Repository</th>
                  <th>Tag</th>
                  <th>Image ID</th>
                  <th>Virtual Size</th>
                </tr>
              </thead>
              <tbody>
                {images.map((img) => (
                  <tr key={img.id}>
                    <td style={{ fontWeight: 600, color: 'var(--text-main)' }}>{img.repository}</td>
                    <td><span className="badge badge-gray">{img.tag}</span></td>
                    <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem', color: 'var(--text-subtle)' }}>{img.id.substring(0, 12)}</td>
                    <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.8rem', color: 'var(--primary)' }}>{img.size}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Scheduled Cron Jobs */}
      <div className="panel-card">
        <div className="panel-header">
          <div>
            <div className="panel-title">
              <Clock size={18} /> Scheduled Cron Jobs ({cronJobs.length})
            </div>
            <div className="panel-subtitle">
              System cron tasks and user crontabs managed by native Linux cron daemon
            </div>
          </div>
          <button className="btn btn-primary" onClick={() => setShowCronModal(true)}>
            <Plus size={14} /> New Cron Job
          </button>
        </div>

        {cronJobs.length === 0 ? (
          <div style={{ padding: '2.5rem', textAlign: 'center', color: 'var(--text-muted)' }}>
            No user crontab tasks configured.
          </div>
        ) : (
          <div className="table-responsive">
            <table className="custom-table">
              <thead>
                <tr>
                  <th>Job #</th>
                  <th>Schedule Expression</th>
                  <th>Command</th>
                  <th>User</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {cronJobs.map((j) => (
                  <tr key={j.id}>
                    <td>#{j.id}</td>
                    <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.82rem', color: 'var(--primary)' }}>{j.schedule}</td>
                    <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.8rem', color: 'var(--text-main)' }}>{j.command}</td>
                    <td><span className="badge badge-gray">{j.user}</span></td>
                    <td>
                      <button className="btn btn-danger btn-sm" onClick={() => handleDeleteCron(j.id, j.command)} title="Delete cron job">
                        <Trash2 size={12} />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* New Cron Job Modal */}
      {showCronModal && (
        <div className="modal-backdrop" onClick={() => setShowCronModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '520px' }}>
            <div className="modal-header">
              <h4>Add Scheduled Cron Job</h4>
              <button className="btn btn-sm" onClick={() => setShowCronModal(false)}>
                <X size={14} />
              </button>
            </div>
            <form onSubmit={handleCreateCron}>
              <div className="modal-body" style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, marginBottom: '0.35rem' }}>
                    Schedule (Cron Expression)
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="0 2 * * * (e.g. Daily at 2 AM)"
                    value={cronSchedule}
                    onChange={(e) => setCronSchedule(e.target.value)}
                    style={{ width: '100%', fontFamily: 'var(--font-mono)' }}
                  />
                  <span style={{ fontSize: '0.72rem', color: 'var(--text-muted)', marginTop: '0.2rem', display: 'block' }}>
                    Format: minute hour day month day-of-week
                  </span>
                </div>

                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, marginBottom: '0.35rem' }}>
                    Shell Command
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="/usr/bin/python3 /opt/app/backup.py"
                    value={cronCommand}
                    onChange={(e) => setCronCommand(e.target.value)}
                    style={{ width: '100%', fontFamily: 'var(--font-mono)' }}
                  />
                </div>
              </div>
              <div className="modal-footer">
                <button type="button" className="btn" onClick={() => setShowCronModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary">
                  Save Cron Task
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Container Logs Modal */}
      {activeLogs && (
        <div className="modal-backdrop" onClick={() => setActiveLogs(null)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '850px' }}>
            <div className="modal-header">
              <h4>Docker Logs: {activeLogs.name}</h4>
              <button className="btn btn-sm" onClick={() => setActiveLogs(null)}>
                <X size={14} />
              </button>
            </div>
            <div className="modal-body">
              <div className="terminal-box" style={{ maxHeight: '420px' }}>{activeLogs.content}</div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
