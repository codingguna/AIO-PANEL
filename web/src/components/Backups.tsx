import React, { useState, useEffect } from 'react';
import { BackupItem } from '../types';

export const Backups: React.FC = () => {
  const [backups, setBackups] = useState<BackupItem[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [showCreateModal, setShowCreateModal] = useState<boolean>(false);
  const [backupTarget, setBackupTarget] = useState<string>('postgres');
  const [targetName, setTargetName] = useState<string>('memotrack');

  const fetchBackups = async () => {
    setLoading(true);
    try {
      const res = await fetch('/api/v1/ops/backups');
      if (res.ok) {
        const data = await res.json();
        setBackups(data || []);
      }
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchBackups();
  }, []);

  const handleCreateSnapshot = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      let url = '/api/v1/databases/postgres/backup';
      let body: any = { database: targetName };

      if (backupTarget === 'mysql') {
        url = '/api/v1/databases/mysql/backup';
      }

      const res = await fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      const data = await res.json();
      if (res.ok && data.success) {
        alert('Backup snapshot generated successfully:\n' + data.backup_file);
        setShowCreateModal(false);
        fetchBackups();
      } else {
        alert('Backup failed: ' + (data.error || 'Unknown error'));
      }
    } catch (e: any) {
      alert('Error creating backup: ' + e.message);
    }
  };

  const handleDownload = (backup: BackupItem) => {
    alert(`Backup located at:\n${backup.path}\n\nYou can copy or download this archive securely via SFTP / SCP.`);
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Overview Banner */}
      <div className="panel-card">
        <div className="panel-header">
          <div>
            <div className="panel-title">
              <span>💾</span> Centralized Backup Center
            </div>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.82rem', marginTop: '0.2rem' }}>
              Instant SQL dumps, Nginx configurations, and application archives stored in /var/lib/aio/backups
            </p>
          </div>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <button className="btn" onClick={fetchBackups}>
              🔄 Refresh
            </button>
            <button className="btn btn-primary" onClick={() => setShowCreateModal(true)}>
              ➕ Create Snapshot
            </button>
          </div>
        </div>
      </div>

      {/* Backups Inventory Table */}
      <div className="panel-card">
        <div className="panel-header">
          <div className="panel-title">
            <span>📦</span> Backup Archive Inventory ({backups.length})
          </div>
        </div>

        {loading ? (
          <div style={{ padding: '2.5rem', textAlign: 'center', color: 'var(--text-subtle)' }}>
            Loading backup files...
          </div>
        ) : backups.length === 0 ? (
          <div style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-subtle)' }}>
            No backup archives found in /var/lib/aio/backups. Click "Create Snapshot" above to generate one.
          </div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table className="custom-table">
              <thead>
                <tr>
                  <th>Archive Name</th>
                  <th>Type</th>
                  <th>File Size</th>
                  <th>Timestamp</th>
                  <th>Full Location</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {backups.map((b) => (
                  <tr key={b.path}>
                    <td>
                      <strong style={{ color: '#fff' }}>{b.name}</strong>
                    </td>
                    <td>
                      <span className="tag" style={{ textTransform: 'uppercase' }}>
                        {b.type}
                      </span>
                    </td>
                    <td style={{ fontFamily: 'var(--font-mono)', color: 'var(--accent-cyan)', fontSize: '0.82rem' }}>
                      {b.size_human}
                    </td>
                    <td style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                      {new Date(b.created_at).toLocaleString()}
                    </td>
                    <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem', color: 'var(--text-subtle)' }}>
                      {b.path}
                    </td>
                    <td>
                      <div style={{ display: 'flex', gap: '0.4rem' }}>
                        <button className="btn" onClick={() => handleDownload(b)}>
                          ⬇️ Location
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Create Snapshot Modal */}
      {showCreateModal && (
        <div className="modal-backdrop" onClick={() => setShowCreateModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '480px' }}>
            <div className="modal-header">
              <h4 style={{ fontWeight: 700 }}>Generate Backup Snapshot</h4>
              <button className="btn" onClick={() => setShowCreateModal(false)}>
                ✕
              </button>
            </div>
            <form onSubmit={handleCreateSnapshot} style={{ padding: '1.5rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.4rem' }}>
                  Snapshot Target Type
                </label>
                <select
                  value={backupTarget}
                  onChange={(e) => setBackupTarget(e.target.value)}
                  style={{
                    width: '100%',
                    padding: '0.65rem 0.85rem',
                    background: 'var(--bg-base)',
                    border: '1px solid var(--border)',
                    borderRadius: '8px',
                    color: '#fff',
                  }}
                >
                  <option value="postgres">PostgreSQL Database (pg_dump)</option>
                  <option value="mysql">MySQL / MariaDB Database (mysqldump)</option>
                </select>
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.4rem' }}>
                  Database / Schema Name
                </label>
                <input
                  type="text"
                  required
                  value={targetName}
                  onChange={(e) => setTargetName(e.target.value)}
                  style={{
                    width: '100%',
                    padding: '0.65rem 0.85rem',
                    background: 'var(--bg-base)',
                    border: '1px solid var(--border)',
                    borderRadius: '8px',
                    color: '#fff',
                    fontFamily: 'var(--font-mono)',
                  }}
                />
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.5rem', marginTop: '0.5rem' }}>
                <button type="button" className="btn" onClick={() => setShowCreateModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary">
                  💾 Dump & Save
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
