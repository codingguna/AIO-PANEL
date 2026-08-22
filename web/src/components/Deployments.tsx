import React, { useState } from 'react';

export const Deployments: React.FC = () => {
  const [appName, setAppName] = useState<string>('MemoTrack Production');
  const [appPath, setAppPath] = useState<string>('/var/www/memotrack');
  const [gitBranch, setGitBranch] = useState<string>('main');
  const [serviceName, setServiceName] = useState<string>('memotrack.service');
  const [buildCmd, setBuildCmd] = useState<string>('npm run build || pip install -r requirements.txt');
  const [restartService, setRestartService] = useState<boolean>(true);

  const [deploying, setDeploying] = useState<boolean>(false);
  const [deployLogs, setDeployLogs] = useState<string | null>(null);

  const handleTriggerDeploy = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!window.confirm(`Deploy ${appName}? This will run git pull, build commands, and restart ${serviceName}.`)) return;

    setDeploying(true);
    setDeployLogs('Triggering deployment sequence on host...');

    try {
      const res = await fetch('/api/v1/ops/deployments/run', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          app_name: appName,
          path: appPath,
          git_branch: gitBranch,
          service: serviceName,
          build_cmd: buildCmd,
          restart_service: restartService,
        }),
      });
      const data = await res.json();
      if (res.ok && data.success) {
        setDeployLogs(data.logs || 'Deployment completed successfully.');
      } else {
        setDeployLogs('Deployment failed: ' + (data.error || 'Unknown error'));
      }
    } catch (e: any) {
      setDeployLogs('Network / Execution Error: ' + e.message);
    } finally {
      setDeploying(false);
    }
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Overview Banner */}
      <div className="panel-card">
        <div className="panel-header">
          <div>
            <div className="panel-title">
              <span>🔄</span> 1-Click Deployment Manager
            </div>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.82rem', marginTop: '0.2rem' }}>
              Safely update production applications with automated Git pull, dependency resolution, build commands, and service reload
            </p>
          </div>
        </div>

        {/* Deployment Pipeline Diagram */}
        <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap', marginTop: '1rem', alignItems: 'center' }}>
          <div className="status-pill" style={{ background: 'rgba(99, 102, 241, 0.15)', borderColor: 'var(--primary)', color: '#a5b4fc' }}>
            1. Git Pull ({gitBranch})
          </div>
          <span style={{ color: 'var(--text-subtle)' }}>➔</span>
          <div className="status-pill" style={{ background: 'rgba(6, 182, 212, 0.15)', borderColor: 'var(--accent-cyan)', color: 'var(--accent-cyan)' }}>
            2. Build / Dependencies
          </div>
          <span style={{ color: 'var(--text-subtle)' }}>➔</span>
          <div className="status-pill" style={{ background: 'rgba(16, 185, 129, 0.15)', borderColor: 'var(--accent-emerald)', color: 'var(--accent-emerald)' }}>
            3. Systemd Service Restart
          </div>
          <span style={{ color: 'var(--text-subtle)' }}>➔</span>
          <div className="status-pill" style={{ background: 'rgba(245, 158, 11, 0.15)', borderColor: 'var(--accent-amber)', color: 'var(--accent-amber)' }}>
            4. Health Check
          </div>
        </div>
      </div>

      {/* Deployment Trigger Form */}
      <div className="panel-card">
        <div className="panel-header">
          <div className="panel-title">
            <span>🚀</span> Configure & Execute Deployment
          </div>
        </div>

        <form onSubmit={handleTriggerDeploy} style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '1.25rem', marginTop: '1rem' }}>
          <div>
            <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
              Application Name
            </label>
            <input
              type="text"
              required
              value={appName}
              onChange={(e) => setAppName(e.target.value)}
              style={{ width: '100%', padding: '0.6rem 0.85rem', background: 'var(--bg-base)', border: '1px solid var(--border)', borderRadius: '8px', color: '#fff' }}
            />
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
              Repository / Directory Path
            </label>
            <input
              type="text"
              required
              value={appPath}
              onChange={(e) => setAppPath(e.target.value)}
              style={{ width: '100%', padding: '0.6rem 0.85rem', background: 'var(--bg-base)', border: '1px solid var(--border)', borderRadius: '8px', color: '#fff', fontFamily: 'var(--font-mono)' }}
            />
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
              Target Git Branch
            </label>
            <input
              type="text"
              required
              value={gitBranch}
              onChange={(e) => setGitBranch(e.target.value)}
              style={{ width: '100%', padding: '0.6rem 0.85rem', background: 'var(--bg-base)', border: '1px solid var(--border)', borderRadius: '8px', color: '#fff', fontFamily: 'var(--font-mono)' }}
            />
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
              Linked systemd Service
            </label>
            <input
              type="text"
              value={serviceName}
              onChange={(e) => setServiceName(e.target.value)}
              style={{ width: '100%', padding: '0.6rem 0.85rem', background: 'var(--bg-base)', border: '1px solid var(--border)', borderRadius: '8px', color: '#fff', fontFamily: 'var(--font-mono)' }}
            />
          </div>

          <div style={{ gridColumn: '1 / -1' }}>
            <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.35rem' }}>
              Custom Build / Post-Pull Command (Optional)
            </label>
            <input
              type="text"
              value={buildCmd}
              onChange={(e) => setBuildCmd(e.target.value)}
              placeholder="e.g. npm install && npm run build, or python manage.py migrate"
              style={{ width: '100%', padding: '0.6rem 0.85rem', background: 'var(--bg-base)', border: '1px solid var(--border)', borderRadius: '8px', color: '#fff', fontFamily: 'var(--font-mono)' }}
            />
          </div>

          <div style={{ gridColumn: '1 / -1', display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '0.5rem' }}>
            <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.85rem', color: '#fff', cursor: 'pointer' }}>
              <input
                type="checkbox"
                checked={restartService}
                onChange={(e) => setRestartService(e.target.checked)}
              />
              Automatically reload/restart service on completion
            </label>

            <button
              type="submit"
              className="btn btn-primary"
              disabled={deploying}
              style={{ padding: '0.65rem 1.5rem', fontWeight: 700 }}
            >
              {deploying ? 'Deploying...' : '🚀 Start Deployment'}
            </button>
          </div>
        </form>
      </div>

      {/* Deployment Log Stream Output */}
      {deployLogs && (
        <div className="panel-card">
          <div className="panel-header">
            <div className="panel-title">
              <span>📜</span> Deployment Output & Logs
            </div>
            <button className="btn" onClick={() => setDeployLogs(null)}>
              Clear
            </button>
          </div>
          <div className="terminal-box" style={{ maxHeight: '350px', marginTop: '0.75rem' }}>
            {deployLogs}
          </div>
        </div>
      )}
    </div>
  );
};
