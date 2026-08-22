import React, { useState, useEffect } from 'react';
import { List, RefreshCw } from 'lucide-react';
import { AuditEvent } from '../types';

export const AuditLog: React.FC = () => {
  const [events, setEvents] = useState<AuditEvent[]>([]);
  const [loading, setLoading] = useState<boolean>(true);

  const fetchAudit = async () => {
    setLoading(true);
    try {
      const res = await fetch('/api/v1/audit/events?limit=100');
      if (res.ok) {
        const data = await res.json();
        setEvents(data || []);
      }
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAudit();
  }, []);

  return (
    <div className="panel-card">
      <div className="panel-header">
        <div>
          <div className="panel-title">
            <List size={18} /> System Audit Trail ({events.length})
          </div>
          <div className="panel-subtitle">
            Every administrative action is permanently recorded in the local SQLite database
          </div>
        </div>
        <button className="btn" onClick={fetchAudit}>
          <RefreshCw size={14} /> Refresh
        </button>
      </div>

      {loading ? (
        <div style={{ padding: '2.5rem', textAlign: 'center', color: 'var(--text-muted)' }}>
          Loading audit records...
        </div>
      ) : events.length === 0 ? (
        <div style={{ padding: '2.5rem', textAlign: 'center', color: 'var(--text-muted)' }}>
          No audit events recorded yet.
        </div>
      ) : (
        <div className="table-responsive">
          <table className="custom-table">
            <thead>
              <tr>
                <th>Timestamp</th>
                <th>Action</th>
                <th>Target Resource</th>
                <th>User</th>
                <th>Result</th>
                <th>Remote IP</th>
              </tr>
            </thead>
            <tbody>
              {events.map((ev) => (
                <tr key={ev.id}>
                  <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                    {new Date(ev.timestamp).toLocaleString()}
                  </td>
                  <td>
                    <strong style={{ color: 'var(--text-main)' }}>{ev.action}</strong>
                  </td>
                  <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.78rem', color: 'var(--primary)' }}>
                    {ev.target}
                  </td>
                  <td>{ev.user}</td>
                  <td>
                    <span className={`badge ${ev.result === 'SUCCESS' ? 'badge-emerald' : 'badge-rose'}`}>
                      {ev.result}
                    </span>
                  </td>
                  <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem' }}>{ev.ip_address || '-'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};
