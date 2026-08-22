import React, { useState, useEffect } from 'react';
import { SSHConfig, SSHSession, FirewallStatus, LinuxUser } from '../types';

export const Security: React.FC = () => {
  const [sshCfg, setSshCfg] = useState<SSHConfig | null>(null);
  const [sessions, setSessions] = useState<SSHSession[]>([]);
  const [firewall, setFirewall] = useState<FirewallStatus | null>(null);
  const [users, setUsers] = useState<LinuxUser[]>([]);
  const [loading, setLoading] = useState(true);

  // Form states
  const [sshPort, setSshPort] = useState(22);
  const [permitRoot, setPermitRoot] = useState('yes');
  const [passwordAuth, setPasswordAuth] = useState(true);

  const [rulePort, setRulePort] = useState('');
  const [ruleAction, setRuleAction] = useState('allow');
  const [ruleProto, setRuleProto] = useState('tcp');

  // New user state
  const [showNewUserModal, setShowNewUserModal] = useState(false);
  const [newUsername, setNewUsername] = useState('');
  const [newUserShell, setNewUserShell] = useState('/bin/bash');
  const [newUserSudo, setNewUserSudo] = useState(false);

  const handleCreateUser = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const res = await fetch('/api/v1/security/users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          username: newUsername,
          shell: newUserShell,
          is_sudo: newUserSudo,
        }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed');
      alert(data.message);
      setShowNewUserModal(false);
      setNewUsername('');
      fetchData();
    } catch (e: any) {
      alert('Error creating user: ' + e.message);
    }
  };

  const fetchData = async () => {
    try {
      const [sshRes, sessRes, fwRes, usersRes] = await Promise.all([
        fetch('/api/v1/security/ssh'),
        fetch('/api/v1/security/ssh/sessions'),
        fetch('/api/v1/security/firewall'),
        fetch('/api/v1/security/users'),
      ]);

      if (sshRes.ok) {
        const d = await sshRes.json();
        setSshCfg(d);
        setSshPort(d.port);
        setPermitRoot(d.permit_root_login);
        setPasswordAuth(d.password_authentication);
      }
      if (sessRes.ok) setSessions((await sessRes.json()) || []);
      if (fwRes.ok) setFirewall(await fwRes.json());
      if (usersRes.ok) setUsers((await usersRes.json()) || []);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleUpdateSSH = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!window.confirm('Save SSH changes? AIO will validate syntax via sshd -t and auto-rollback if any error occurs.')) return;

    try {
      const res = await fetch('/api/v1/security/ssh/config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          port: Number(sshPort),
          permit_root_login: permitRoot,
          password_authentication: passwordAuth,
        }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed');
      alert(data.message);
      fetchData();
    } catch (e: any) {
      alert('Error updating SSH: ' + e.message);
    }
  };

  const handleAddRule = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const res = await fetch('/api/v1/security/firewall/rules', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          port: rulePort,
          action: ruleAction,
          protocol: ruleProto,
        }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed');
      setRulePort('');
      fetchData();
    } catch (e: any) {
      alert('Error adding rule: ' + e.message);
    }
  };

  const handleDeleteRule = async (id: number) => {
    if (!window.confirm(`Delete rule #${id}?`)) return;
    try {
      await fetch(`/api/v1/security/firewall/rules/${id}`, { method: 'DELETE' });
      fetchData();
    } catch (e) {
      alert('Error deleting rule: ' + e);
    }
  };

  const handleToggleFirewall = async (enable: boolean) => {
    const action = enable ? 'ENABLE' : 'DISABLE';
    if (!window.confirm(`Are you sure you want to ${action} UFW firewall?`)) return;
    try {
      const res = await fetch('/api/v1/security/firewall/toggle', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ enable, ssh_port: sshPort }),
      });
      const data = await res.json();
      alert(data.message);
      fetchData();
    } catch (e) {
      alert('Error toggling firewall: ' + e);
    }
  };

  if (loading) {
    return <div style={{ padding: '2rem', textAlign: 'center' }}>Loading security configuration...</div>;
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
      {/* OpenSSH Hardening */}
      <div className="panel-card">
        <div className="panel-header">
          <div>
            <div className="panel-title">
              <span>🛡️</span> OpenSSH Configuration (with Auto-Rollback)
            </div>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.82rem', marginTop: '0.2rem' }}>
              Config file: {sshCfg?.config_path || '/etc/ssh/sshd_config'}
            </p>
          </div>
        </div>

        <form onSubmit={handleUpdateSSH} style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1.25rem' }}>
          <div>
            <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
              SSH Port
            </label>
            <input
              type="number"
              value={sshPort}
              onChange={(e) => setSshPort(Number(e.target.value))}
              style={{ width: '100%', padding: '0.6rem', background: 'var(--bg-base)', border: '1px solid var(--border)', borderRadius: '8px', color: '#fff' }}
            />
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
              Permit Root Login
            </label>
            <select
              value={permitRoot}
              onChange={(e) => setPermitRoot(e.target.value)}
              style={{ width: '100%', padding: '0.6rem', background: 'var(--bg-base)', border: '1px solid var(--border)', borderRadius: '8px', color: '#fff' }}
            >
              <option value="yes">yes (Allow Root SSH)</option>
              <option value="prohibit-password">prohibit-password (Keys only)</option>
              <option value="no">no (Disable Root Login)</option>
            </select>
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
              Password Authentication
            </label>
            <select
              value={passwordAuth ? 'yes' : 'no'}
              onChange={(e) => setPasswordAuth(e.target.value === 'yes')}
              style={{ width: '100%', padding: '0.6rem', background: 'var(--bg-base)', border: '1px solid var(--border)', borderRadius: '8px', color: '#fff' }}
            >
              <option value="yes">yes (Allow Passwords)</option>
              <option value="no">no (Enforce SSH Keys Only)</option>
            </select>
          </div>

          <div style={{ display: 'flex', alignItems: 'flex-end' }}>
            <button type="submit" className="btn btn-primary" style={{ width: '100%', padding: '0.65rem' }}>
              💾 Safe Apply & Validate
            </button>
          </div>
        </form>

        {/* Active sessions */}
        {sessions.length > 0 && (
          <div style={{ marginTop: '1rem', borderTop: '1px solid var(--border)', paddingTop: '1rem' }}>
            <div style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-muted)', marginBottom: '0.5rem' }}>
              Active Connected SSH Sessions ({sessions.length})
            </div>
            <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
              {sessions.map((s, idx) => (
                <span key={idx} className="tag">
                  👤 {s.user} via {s.terminal} ({s.host || 'local'})
                </span>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* UFW Firewall */}
      <div className="panel-card">
        <div className="panel-header">
          <div className="panel-title">
            <span>🔥</span> UFW Firewall ({firewall?.active ? 'Active' : 'Disabled'})
          </div>
          <button
            className={`btn ${firewall?.active ? 'btn-danger' : 'btn-primary'}`}
            onClick={() => handleToggleFirewall(!firewall?.active)}
          >
            {firewall?.active ? '🔴 Disable Firewall' : '🟢 Enable Firewall'}
          </button>
        </div>

        {/* Add rule form */}
        <form onSubmit={handleAddRule} style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap' }}>
          <input
            type="text"
            required
            placeholder="Port (e.g. 80, 443, 3000)"
            value={rulePort}
            onChange={(e) => setRulePort(e.target.value)}
            style={{ flex: 1, minWidth: '160px', padding: '0.55rem', background: 'var(--bg-base)', border: '1px solid var(--border)', borderRadius: '8px', color: '#fff' }}
          />
          <select
            value={ruleProto}
            onChange={(e) => setRuleProto(e.target.value)}
            style={{ padding: '0.55rem', background: 'var(--bg-base)', border: '1px solid var(--border)', borderRadius: '8px', color: '#fff' }}
          >
            <option value="tcp">TCP</option>
            <option value="udp">UDP</option>
            <option value="any">Any Proto</option>
          </select>
          <select
            value={ruleAction}
            onChange={(e) => setRuleAction(e.target.value)}
            style={{ padding: '0.55rem', background: 'var(--bg-base)', border: '1px solid var(--border)', borderRadius: '8px', color: '#fff' }}
          >
            <option value="allow">ALLOW</option>
            <option value="deny">DENY</option>
          </select>
          <button type="submit" className="btn btn-primary">
            ➕ Add Port Rule
          </button>
        </form>

        {/* Rules table */}
        <table className="custom-table">
          <thead>
            <tr>
              <th>Rule #</th>
              <th>To Port</th>
              <th>Protocol</th>
              <th>Action</th>
              <th>From Source</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            {firewall?.rules.length === 0 ? (
              <tr>
                <td colSpan={6} style={{ textAlign: 'center', color: 'var(--text-subtle)', padding: '1.5rem' }}>
                  No active UFW port rules configured.
                </td>
              </tr>
            ) : (
              firewall?.rules.map((r) => (
                <tr key={r.id}>
                  <td>#{r.id}</td>
                  <td style={{ fontWeight: 600, color: '#fff' }}>{r.to_port}</td>
                  <td>{r.protocol.toUpperCase()}</td>
                  <td>
                    <span className={`badge ${r.action === 'ALLOW' ? 'badge-emerald' : 'badge-rose'}`}>
                      {r.action}
                    </span>
                  </td>
                  <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.8rem' }}>{r.from_ip}</td>
                  <td>
                    <button className="btn btn-danger" onClick={() => handleDeleteRule(r.id)}>
                      🗑️
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Linux Users */}
      <div className="panel-card">
        <div className="panel-header">
          <div className="panel-title">
            <span>👥</span> System User Accounts ({users.length})
          </div>
          <button className="btn btn-primary" onClick={() => setShowNewUserModal(true)}>
            ➕ Create User
          </button>
        </div>
        <table className="custom-table">
          <thead>
            <tr>
              <th>Username</th>
              <th>UID / GID</th>
              <th>Home Directory</th>
              <th>Shell</th>
              <th>Sudo Privileges</th>
            </tr>
          </thead>
          <tbody>
            {users.map((u) => (
              <tr key={u.username}>
                <td style={{ fontWeight: 700, color: '#fff' }}>{u.username}</td>
                <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.8rem' }}>{u.uid} : {u.gid}</td>
                <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.78rem', color: 'var(--text-muted)' }}>{u.home_dir}</td>
                <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.78rem' }}>{u.shell}</td>
                <td>
                  <span className={`badge ${u.is_sudo ? 'badge-emerald' : 'badge-owner'}`}>
                    {u.is_sudo ? '⭐ Sudo Admin' : 'Standard'}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* New User Modal */}
      {showNewUserModal && (
        <div className="modal-backdrop" onClick={() => setShowNewUserModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '480px' }}>
            <div className="modal-header">
              <h4 style={{ fontWeight: 700 }}>Create New Linux User</h4>
              <button className="btn" onClick={() => setShowNewUserModal(false)}>
                ✕
              </button>
            </div>
            <form onSubmit={handleCreateUser} style={{ padding: '1.5rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
                  Username
                </label>
                <input
                  type="text"
                  required
                  placeholder="e.g. deploy"
                  value={newUsername}
                  onChange={(e) => setNewUsername(e.target.value)}
                  style={{ width: '100%', padding: '0.65rem 0.85rem', background: 'var(--bg-base)', border: '1px solid var(--border)', borderRadius: '8px', color: '#fff', fontFamily: 'var(--font-mono)' }}
                />
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
                  Login Shell
                </label>
                <select
                  value={newUserShell}
                  onChange={(e) => setNewUserShell(e.target.value)}
                  style={{ width: '100%', padding: '0.65rem 0.85rem', background: 'var(--bg-base)', border: '1px solid var(--border)', borderRadius: '8px', color: '#fff' }}
                >
                  <option value="/bin/bash">/bin/bash</option>
                  <option value="/bin/sh">/bin/sh</option>
                  <option value="/usr/bin/zsh">/usr/bin/zsh</option>
                  <option value="/usr/sbin/nologin">/usr/sbin/nologin (No interactive login)</option>
                </select>
              </div>

              <div>
                <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.85rem', color: '#fff', cursor: 'pointer' }}>
                  <input
                    type="checkbox"
                    checked={newUserSudo}
                    onChange={(e) => setNewUserSudo(e.target.checked)}
                  />
                  Grant Sudo Administrative Privileges
                </label>
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.5rem', marginTop: '0.5rem' }}>
                <button type="button" className="btn" onClick={() => setShowNewUserModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary">
                  Create User
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
