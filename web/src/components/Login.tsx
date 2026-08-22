import React, { useState } from 'react';
import { Lock, User, Shield, Terminal, ArrowRight } from 'lucide-react';
import { AuthStatus } from '../types';

interface LoginProps {
  authStatus: AuthStatus;
  onLoginSuccess: (username: string) => void;
}

export const Login: React.FC<LoginProps> = ({ authStatus, onLoginSuccess }) => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const isSetup = authStatus.setup_required;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    if (isSetup && password !== confirmPassword) {
      setError('Passwords do not match');
      setLoading(false);
      return;
    }

    try {
      const endpoint = isSetup ? '/api/v1/auth/setup' : '/api/v1/auth/login';
      const res = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });

      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error || 'Authentication failed');
      }

      onLoginSuccess(data.username || username);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      minHeight: '100vh',
      width: '100vw',
      backgroundColor: 'var(--bg-page)',
      padding: '1.5rem',
    }}>
      <div style={{
        width: '100%',
        maxWidth: '440px',
        backgroundColor: '#ffffff',
        border: '1px solid var(--border)',
        borderRadius: '8px',
        boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.05), 0 2px 4px -1px rgba(0, 0, 0, 0.03)',
        overflow: 'hidden',
      }}>
        {/* Card Header */}
        <div style={{
          padding: '2rem 2rem 1.5rem 2rem',
          borderBottom: '1px solid var(--border)',
          textAlign: 'center',
          backgroundColor: '#f8fafc',
        }}>
          <div style={{
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            width: '48px',
            height: '48px',
            borderRadius: '8px',
            backgroundColor: 'var(--primary)',
            color: '#ffffff',
            fontWeight: 800,
            fontSize: '1.25rem',
            marginBottom: '0.75rem',
          }}>
            AIO
          </div>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 700, color: 'var(--text-main)' }}>
            {isSetup ? 'Initial Administrator Setup' : 'Sign in to AIO-PANEL'}
          </h2>
          <p style={{ fontSize: '0.82rem', color: 'var(--text-muted)', marginTop: '0.25rem' }}>
            {isSetup
              ? 'Create the primary superuser account for your server'
              : 'Enter your administrator credentials to access the panel'}
          </p>
        </div>

        {/* Error Alert */}
        {error && (
          <div style={{
            margin: '1.25rem 1.75rem 0 1.75rem',
            padding: '0.75rem 1rem',
            backgroundColor: 'var(--accent-rose-bg)',
            border: '1px solid #fecaca',
            borderRadius: '6px',
            color: 'var(--accent-rose)',
            fontSize: '0.82rem',
            fontWeight: 500,
          }}>
            {error}
          </div>
        )}

        {/* Form Body */}
        <form onSubmit={handleSubmit} style={{ padding: '1.75rem', display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
          <div>
            <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.4rem' }}>
              Username
            </label>
            <div style={{ position: 'relative' }}>
              <input
                type="text"
                required
                placeholder={isSetup ? 'e.g. admin' : 'Your username'}
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                style={{ width: '100%', paddingLeft: '2.4rem' }}
              />
              <User size={16} style={{ position: 'absolute', left: '0.8rem', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-subtle)' }} />
            </div>
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.4rem' }}>
              Password
            </label>
            <div style={{ position: 'relative' }}>
              <input
                type="password"
                required
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                style={{ width: '100%', paddingLeft: '2.4rem' }}
              />
              <Lock size={16} style={{ position: 'absolute', left: '0.8rem', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-subtle)' }} />
            </div>
          </div>

          {isSetup && (
            <div>
              <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.4rem' }}>
                Confirm Password
              </label>
              <div style={{ position: 'relative' }}>
                <input
                  type="password"
                  required
                  placeholder="••••••••"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  style={{ width: '100%', paddingLeft: '2.4rem' }}
                />
                <Shield size={16} style={{ position: 'absolute', left: '0.8rem', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-subtle)' }} />
              </div>
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="btn btn-primary"
            style={{
              justifyContent: 'center',
              padding: '0.65rem 1rem',
              fontSize: '0.88rem',
              fontWeight: 600,
              marginTop: '0.5rem',
            }}
          >
            {loading ? 'Authenticating...' : isSetup ? 'Complete Setup & Log In' : 'Sign In'}
            {!loading && <ArrowRight size={16} />}
          </button>
        </form>

        {/* CLI Helper Footer */}
        <div style={{
          padding: '1rem 1.75rem',
          backgroundColor: '#f8fafc',
          borderTop: '1px solid var(--border)',
          fontSize: '0.75rem',
          color: 'var(--text-muted)',
          display: 'flex',
          alignItems: 'center',
          gap: '0.5rem',
        }}>
          <Terminal size={14} style={{ color: 'var(--primary)', flexShrink: 0 }} />
          <span>
            Tip: You can also manage superusers via SSH with <code>sudo aio createsuperuser</code>
          </span>
        </div>
      </div>
    </div>
  );
};
