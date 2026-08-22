import React, { useState, useEffect } from 'react';
import { Database, Plus, RefreshCw, HardDrive, Shield, Trash2 } from 'lucide-react';
import { PostgresDB, MySQLDB, PostgresUser } from '../types';

export const Databases: React.FC = () => {
  const [pgDbs, setPgDbs] = useState<PostgresDB[]>([]);
  const [pgUsers, setPgUsers] = useState<PostgresUser[]>([]);
  const [mysqlDbs, setMysqlDbs] = useState<MySQLDB[]>([]);
  const [loading, setLoading] = useState<boolean>(true);

  // New DB Modal state
  const [showModal, setShowModal] = useState<boolean>(false);
  const [dbEngine, setDbEngine] = useState<string>('postgres');
  const [newDbName, setNewDbName] = useState<string>('');
  const [newDbOwner, setNewDbOwner] = useState<string>('');

  const fetchData = async () => {
    setLoading(true);
    try {
      const [pgRes, usersRes, mysqlRes] = await Promise.all([
        fetch('/api/v1/databases/postgres'),
        fetch('/api/v1/databases/postgres/users'),
        fetch('/api/v1/databases/mysql'),
      ]);
      if (pgRes.ok) setPgDbs((await pgRes.json()) || []);
      if (usersRes.ok) setPgUsers((await usersRes.json()) || []);
      if (mysqlRes.ok) setMysqlDbs((await mysqlRes.json()) || []);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleBackup = async (engine: string, dbName: string) => {
    if (!window.confirm(`Trigger instant SQL dump backup for ${dbName}?`)) return;
    try {
      const url = engine === 'postgres' ? '/api/v1/databases/postgres/backup' : '/api/v1/databases/mysql/backup';
      const res = await fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ database: dbName }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed');
      alert(`✅ Backup saved to:\n${data.backup_file}`);
    } catch (e: any) {
      alert('Error creating backup: ' + e.message);
    }
  };

  const handleCreateDB = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const url = dbEngine === 'postgres' ? '/api/v1/databases/postgres/create' : '/api/v1/databases/mysql/create';
      const res = await fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: newDbName,
          owner: newDbOwner || undefined,
        }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed');
      alert(data.message);
      setShowModal(false);
      setNewDbName('');
      setNewDbOwner('');
      fetchData();
    } catch (e: any) {
      alert('Error creating database: ' + e.message);
    }
  };

  const handleDeleteDB = async (engine: 'postgres' | 'mysql', name: string) => {
    if (!window.confirm(`⚠️ Are you sure you want to permanently DROP ${engine.toUpperCase()} database '${name}'? This cannot be undone!`)) return;
    try {
      const url = engine === 'postgres' ? `/api/v1/databases/postgres/${name}` : `/api/v1/databases/mysql/${name}`;
      const res = await fetch(url, { method: 'DELETE' });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed');
      alert(data.message);
      fetchData();
    } catch (e: any) {
      alert('Error deleting database: ' + e.message);
    }
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* PostgreSQL Section */}
      <div className="panel-card">
        <div className="panel-header">
          <div>
            <div className="panel-title">
              <Database size={18} /> PostgreSQL Database Server ({pgDbs.length} DBs)
            </div>
            <div className="panel-subtitle">
              Discovered from native PostgreSQL cluster without modifying existing databases
            </div>
          </div>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <button className="btn" onClick={fetchData}>
              <RefreshCw size={14} /> Refresh
            </button>
            <button
              className="btn btn-primary"
              onClick={() => {
                setDbEngine('postgres');
                setShowModal(true);
              }}
            >
              <Plus size={14} /> Create PostgreSQL DB
            </button>
          </div>
        </div>

        {loading ? (
          <div style={{ padding: '2.5rem', textAlign: 'center', color: 'var(--text-muted)' }}>
            Discovering database instances...
          </div>
        ) : pgDbs.length === 0 ? (
          <div style={{ padding: '2.5rem', textAlign: 'center', color: 'var(--text-muted)' }}>
            No PostgreSQL databases found (or PostgreSQL service is inactive).
          </div>
        ) : (
          <div className="table-responsive">
            <table className="custom-table">
              <thead>
                <tr>
                  <th>Database Name</th>
                  <th>Owner Role</th>
                  <th>Encoding</th>
                  <th>Disk Size</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {pgDbs.map((d) => (
                  <tr key={d.name}>
                    <td><strong style={{ color: 'var(--text-main)' }}>{d.name}</strong></td>
                    <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.8rem' }}>{d.owner}</td>
                    <td style={{ fontSize: '0.8rem' }}>{d.encoding}</td>
                    <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.8rem', color: 'var(--primary)' }}>
                      {d.size_human}
                    </td>
                    <td>
                      <div style={{ display: 'flex', gap: '0.35rem' }}>
                        <button className="btn btn-sm" onClick={() => handleBackup('postgres', d.name)}>
                          <HardDrive size={12} /> Dump Backup
                        </button>
                        <button className="btn btn-danger btn-sm" onClick={() => handleDeleteDB('postgres', d.name)} title="Drop database">
                          <Trash2 size={12} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* PostgreSQL Roles */}
        {pgUsers.length > 0 && (
          <div style={{ marginTop: '1.25rem', borderTop: '1px solid var(--border)', paddingTop: '1rem' }}>
            <h5 style={{ fontWeight: 600, fontSize: '0.82rem', marginBottom: '0.6rem', color: 'var(--text-muted)' }}>
              Configured Roles / Users ({pgUsers.length})
            </h5>
            <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
              {pgUsers.map((u) => (
                <div
                  key={u.role_name}
                  className="badge badge-gray"
                  style={{ display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}
                >
                  <Shield size={12} />
                  <span>{u.role_name}</span>
                  {u.is_superuser && <span style={{ color: 'var(--accent-amber)', fontWeight: 700 }}>(superuser)</span>}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* MySQL / MariaDB Section */}
      <div className="panel-card">
        <div className="panel-header">
          <div>
            <div className="panel-title">
              <Database size={18} /> MySQL / MariaDB Database Server ({mysqlDbs.length} DBs)
            </div>
            <div className="panel-subtitle">
              Discovered from native MySQL engine socket
            </div>
          </div>
          <button
            className="btn btn-primary"
            onClick={() => {
              setDbEngine('mysql');
              setShowModal(true);
            }}
          >
            <Plus size={14} /> Create MySQL DB
          </button>
        </div>

        {mysqlDbs.length === 0 ? (
          <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--text-muted)' }}>
            No MySQL databases found (or MySQL service is inactive).
          </div>
        ) : (
          <div className="table-responsive">
            <table className="custom-table">
              <thead>
                <tr>
                  <th>Database Name</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {mysqlDbs.map((d) => (
                  <tr key={d.name}>
                    <td><strong style={{ color: 'var(--text-main)' }}>{d.name}</strong></td>
                    <td>
                      <div style={{ display: 'flex', gap: '0.35rem' }}>
                        <button className="btn btn-sm" onClick={() => handleBackup('mysql', d.name)}>
                          <HardDrive size={12} /> Dump Backup
                        </button>
                        <button className="btn btn-danger btn-sm" onClick={() => handleDeleteDB('mysql', d.name)} title="Drop database">
                          <Trash2 size={12} />
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

      {/* Create Database Modal */}
      {showModal && (
        <div className="modal-backdrop" onClick={() => setShowModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '480px' }}>
            <div className="modal-header">
              <h4>Create New {dbEngine === 'postgres' ? 'PostgreSQL' : 'MySQL'} Database</h4>
              <button className="btn btn-sm" onClick={() => setShowModal(false)}>
                ✕
              </button>
            </div>
            <form onSubmit={handleCreateDB} style={{ padding: '1.25rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
                  Database Engine
                </label>
                <select
                  value={dbEngine}
                  onChange={(e) => setDbEngine(e.target.value)}
                  style={{ width: '100%' }}
                >
                  <option value="postgres">PostgreSQL</option>
                  <option value="mysql">MySQL / MariaDB</option>
                </select>
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
                  Database Name
                </label>
                <input
                  type="text"
                  required
                  placeholder="e.g. app_production"
                  value={newDbName}
                  onChange={(e) => setNewDbName(e.target.value)}
                  style={{ width: '100%', fontFamily: 'var(--font-mono)' }}
                />
              </div>

              {dbEngine === 'postgres' && (
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
                    Owner Role (Optional)
                  </label>
                  <input
                    type="text"
                    placeholder="e.g. postgres or app_user"
                    value={newDbOwner}
                    onChange={(e) => setNewDbOwner(e.target.value)}
                    style={{ width: '100%', fontFamily: 'var(--font-mono)' }}
                  />
                </div>
              )}

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.5rem', marginTop: '0.5rem' }}>
                <button type="button" className="btn" onClick={() => setShowModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary">
                  Create Database
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
