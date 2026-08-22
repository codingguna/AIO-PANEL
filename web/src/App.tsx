import React, { useState, useEffect } from 'react';
import {
  Activity,
  Server,
  Layers,
  Globe,
  Folder,
  Terminal as TerminalIcon,
  FileText,
  GitBranch,
  Shield,
  Database,
  Archive,
  Box,
  List,
  Settings as SettingsIcon,
  ShoppingBag,
  LogOut,
  User,
  Menu,
  X,
} from 'lucide-react';

import { SystemInfo, LiveMetrics, AuthStatus } from './types';
import { Login } from './components/Login';
import { Dashboard } from './components/Dashboard';
import { AppStore } from './components/AppStore';
import { Services } from './components/Services';
import { Applications } from './components/Applications';
import { WebDomains } from './components/WebDomains';
import { Security } from './components/Security';
import { Databases } from './components/Databases';
import { FileManager } from './components/FileManager';
import { Terminal } from './components/Terminal';
import { LogExplorer } from './components/LogExplorer';
import { Deployments } from './components/Deployments';
import { Backups } from './components/Backups';
import { DockerCron } from './components/DockerCron';
import { AuditLog } from './components/AuditLog';
import { Settings } from './components/Settings';

export const App: React.FC = () => {
  const [authStatus, setAuthStatus] = useState<AuthStatus | null>(null);
  const [checkingAuth, setCheckingAuth] = useState<boolean>(true);
  const [mobileMenuOpen, setMobileMenuOpen] = useState<boolean>(false);
  const [activeTab, setActiveTab] = useState<string>('telemetry');
  const [info, setInfo] = useState<SystemInfo | null>(null);
  const [metrics, setMetrics] = useState<LiveMetrics | null>(null);
  const [connected, setConnected] = useState<boolean>(true);

  // 1. Initial Authentication Check
  const checkAuth = async () => {
    try {
      const res = await fetch('/api/v1/auth/status');
      if (res.ok) {
        const data: AuthStatus = await res.json();
        setAuthStatus(data);
      } else {
        setAuthStatus({ authenticated: false, setup_required: false });
      }
    } catch (err) {
      setAuthStatus({ authenticated: false, setup_required: false });
    } finally {
      setCheckingAuth(false);
    }
  };

  useEffect(() => {
    checkAuth();
  }, []);

  // 2. Initial system info fetch
  useEffect(() => {
    if (!authStatus?.authenticated) return;
    fetch('/api/v1/system/info')
      .then((res) => res.json())
      .then((data) => setInfo(data))
      .catch((err) => console.error('Failed to fetch system info', err));
  }, [authStatus?.authenticated]);

  // 3. Real-time metrics streaming: Polling every 2 seconds
  useEffect(() => {
    if (!authStatus?.authenticated) return;
    const fetchMetrics = async () => {
      try {
        const res = await fetch('/api/v1/system/metrics');
        if (res.ok) {
          const data = await res.json();
          setMetrics(data);
          setConnected(true);
        } else {
          setConnected(false);
        }
      } catch (err) {
        setConnected(false);
      }
    };

    fetchMetrics();
    const interval = setInterval(fetchMetrics, 2000);
    return () => clearInterval(interval);
  }, [authStatus?.authenticated]);

  const handleLogout = async () => {
    try {
      await fetch('/api/v1/auth/logout', { method: 'POST' });
      setAuthStatus({ authenticated: false, setup_required: false });
    } catch (e) {
      console.error('Logout error', e);
    }
  };

  if (checkingAuth) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh', backgroundColor: 'var(--bg-page)', color: 'var(--text-muted)' }}>
        Connecting to AIO-PANEL daemon...
      </div>
    );
  }

  if (!authStatus?.authenticated) {
    return (
      <Login
        authStatus={authStatus || { authenticated: false, setup_required: false }}
        onLoginSuccess={(username) => {
          setAuthStatus({
            authenticated: true,
            username,
            role: 'admin',
            setup_required: false,
          });
        }}
      />
    );
  }

  const navSections = [
    {
      title: 'Overview',
      items: [
        { id: 'telemetry', label: 'System Health', icon: Activity },
      ],
    },
    {
      title: 'Management',
      items: [
        { id: 'store', label: 'Software Store', icon: ShoppingBag },
        { id: 'services', label: 'Services (systemd)', icon: Server },
        { id: 'apps', label: 'Applications', icon: Layers },
        { id: 'web', label: 'Web & Domains', icon: Globe },
        { id: 'databases', label: 'Databases', icon: Database },
        { id: 'docker', label: 'Docker & Cron', icon: Box },
      ],
    },
    {
      title: 'Operations',
      items: [
        { id: 'files', label: 'File Manager', icon: Folder },
        { id: 'terminal', label: 'Web Terminal', icon: TerminalIcon },
        { id: 'logs', label: 'Log Explorer', icon: FileText },
        { id: 'deployments', label: 'Deployments', icon: GitBranch },
        { id: 'backups', label: 'Backups', icon: Archive },
      ],
    },
    {
      title: 'Security & System',
      items: [
        { id: 'security', label: 'Security & UFW', icon: Shield },
        { id: 'audit', label: 'Audit Trail', icon: List },
        { id: 'settings', label: 'Panel Settings', icon: SettingsIcon },
      ],
    },
  ];

  const getActiveTabTitle = () => {
    for (const section of navSections) {
      for (const item of section.items) {
        if (item.id === activeTab) return item.label;
      }
    }
    return 'Dashboard';
  };

  return (
    <div className="layout-wrapper">
      {/* Mobile Drawer Overlay */}
      {mobileMenuOpen && (
        <div
          className="sidebar-overlay"
          onClick={() => setMobileMenuOpen(false)}
        />
      )}

      {/* Left Sidebar */}
      <aside className={`sidebar ${mobileMenuOpen ? 'mobile-open' : ''}`}>
        <div className="sidebar-brand">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <div className="sidebar-logo">AIO</div>
            <div>
              <div className="sidebar-brand-title">AIO-PANEL</div>
              <div className="sidebar-brand-sub">Server Control</div>
            </div>
          </div>
          {mobileMenuOpen && (
            <button
              className="btn btn-sm mobile-close-btn"
              onClick={() => setMobileMenuOpen(false)}
              style={{ background: 'transparent', border: 'none', color: '#94a3b8' }}
            >
              <X size={18} />
            </button>
          )}
        </div>

        <div className="sidebar-nav">
          {navSections.map((section) => (
            <div key={section.title}>
              <div className="sidebar-section-header">{section.title}</div>
              {section.items.map((item) => {
                const IconComponent = item.icon;
                const isActive = activeTab === item.id;
                return (
                  <button
                    key={item.id}
                    className={`sidebar-btn ${isActive ? 'active' : ''}`}
                    onClick={() => {
                      setActiveTab(item.id);
                      setMobileMenuOpen(false);
                    }}
                  >
                    <IconComponent />
                    <span>{item.label}</span>
                  </button>
                );
              })}
            </div>
          ))}
        </div>

        {/* Sidebar Footer with Host & User Profile */}
        <div className="sidebar-footer">
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '0.65rem', paddingBottom: '0.65rem', borderBottom: '1px solid var(--border)' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.45rem' }}>
              <div style={{ width: '24px', height: '24px', borderRadius: '50%', backgroundColor: 'var(--primary)', color: '#fff', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '0.75rem' }}>
                <User size={13} />
              </div>
              <span style={{ fontWeight: 600, fontSize: '0.82rem', color: 'var(--text-main)' }}>
                {authStatus.username || 'admin'}
              </span>
            </div>
            <button
              onClick={handleLogout}
              title="Sign Out"
              style={{
                background: 'none',
                border: 'none',
                color: 'var(--text-subtle)',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                padding: '0.2rem',
              }}
            >
              <LogOut size={14} />
            </button>
          </div>
          <div>Host: <strong>{info?.hostname || 'localhost'}</strong></div>
          <div>OS: {info?.os || 'Linux'}</div>
        </div>
      </aside>

      {/* Main Content Area */}
      <div className="main-wrapper">
        {/* Top Navbar */}
        <header className="top-navbar">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <button
              className="hamburger-btn"
              onClick={() => setMobileMenuOpen(true)}
              aria-label="Open Navigation Menu"
            >
              <Menu size={20} />
            </button>
            <div className="top-navbar-title">
              <span>{getActiveTabTitle()}</span>
            </div>
          </div>

          <div className="top-navbar-actions">
            <div className={`status-pill ${connected ? '' : 'disconnected'}`}>
              <span className={`status-dot ${connected ? '' : 'disconnected'}`} />
              <span>{connected ? 'Live (2s)' : 'Disconnected'}</span>
            </div>
          </div>
        </header>

        {/* Dynamic Content View */}
        <main className="content-container">
          {activeTab === 'telemetry' && <Dashboard info={info} metrics={metrics} />}
          {activeTab === 'store' && <AppStore />}
          {activeTab === 'services' && <Services />}
          {activeTab === 'apps' && <Applications />}
          {activeTab === 'web' && <WebDomains />}
          {activeTab === 'files' && <FileManager />}
          {activeTab === 'terminal' && <Terminal />}
          {activeTab === 'logs' && <LogExplorer />}
          {activeTab === 'deployments' && <Deployments />}
          {activeTab === 'security' && <Security />}
          {activeTab === 'databases' && <Databases />}
          {activeTab === 'backups' && <Backups />}
          {activeTab === 'docker' && <DockerCron />}
          {activeTab === 'audit' && <AuditLog />}
          {activeTab === 'settings' && <Settings />}
        </main>
      </div>
    </div>
  );
};
