import React, { useState, useRef, useEffect } from 'react';
import { TerminalExecResult } from '../types';

export const Terminal: React.FC = () => {
  const [command, setCommand] = useState<string>('');
  const [history, setHistory] = useState<TerminalExecResult[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [cwd, setCwd] = useState<string>('/home/ubuntu');
  const [user, setUser] = useState<string>('ubuntu');
  const [hostname, setHostname] = useState<string>('server');
  const [cmdHistoryList, setCmdHistoryList] = useState<string[]>([]);
  const [historyIndex, setHistoryIndex] = useState<number>(-1);

  const screenRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (screenRef.current) {
      screenRef.current.scrollTop = screenRef.current.scrollHeight;
    }
  }, [history, loading]);

  const executeCommand = async (cmdToRun: string) => {
    const trimmed = cmdToRun.trim();
    if (!trimmed || loading) return;

    if (trimmed === 'clear' || trimmed === 'cls') {
      setHistory([]);
      setCommand('');
      return;
    }

    setLoading(true);
    setCmdHistoryList((prev) => [trimmed, ...prev]);
    setHistoryIndex(-1);

    const executingDir = cwd;
    const executingUser = user;
    const executingHost = hostname;

    try {
      const res = await fetch('/api/v1/ops/terminal/exec', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: trimmed, cwd: executingDir }),
      });
      const data: TerminalExecResult = await res.json();
      
      // Update persistent working directory and system identity
      if (data.cwd) setCwd(data.cwd);
      if (data.user) setUser(data.user);
      if (data.hostname) setHostname(data.hostname);

      setHistory((prev) => [...prev, {
        ...data,
        cwd: executingDir,
        user: executingUser,
        hostname: executingHost,
      }]);
      setCommand('');
    } catch (e: any) {
      setHistory((prev) => [
        ...prev,
        {
          command: trimmed,
          stdout: '',
          stderr: 'Connection error: ' + e.message,
          exit_code: 1,
          duration: '0ms',
          timestamp: new Date().toLocaleTimeString(),
          cwd: executingDir,
          user: executingUser,
          hostname: executingHost,
        },
      ]);
    } finally {
      setLoading(false);
      setTimeout(() => inputRef.current?.focus(), 50);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      executeCommand(command);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (cmdHistoryList.length > 0 && historyIndex < cmdHistoryList.length - 1) {
        const nextIdx = historyIndex + 1;
        setHistoryIndex(nextIdx);
        setCommand(cmdHistoryList[nextIdx]);
      }
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (historyIndex > 0) {
        const nextIdx = historyIndex - 1;
        setHistoryIndex(nextIdx);
        setCommand(cmdHistoryList[nextIdx]);
      } else if (historyIndex === 0) {
        setHistoryIndex(-1);
        setCommand('');
      }
    }
  };

  const quickShortcuts = [
    { label: 'Uptime', cmd: 'uptime' },
    { label: 'Memory (free -m)', cmd: 'free -m' },
    { label: 'Disk Space (df -h)', cmd: 'df -h' },
    { label: 'Listening Ports', cmd: 'ss -tuln' },
    { label: 'Active Services', cmd: 'systemctl list-units --type=service --state=running --no-pager' },
    { label: 'Docker Containers', cmd: 'docker ps -a' },
    { label: 'Nginx Config Test', cmd: 'nginx -t' },
    { label: 'Who Am I', cmd: 'whoami && pwd' },
  ];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
      {/* Top Banner & Quick Shortcuts */}
      <div className="panel-card">
        <div className="panel-header" style={{ marginBottom: '0.75rem' }}>
          <div>
            <div className="panel-title">
              <span>💻</span> Web Terminal & Interactive Shell
            </div>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.82rem', marginTop: '0.2rem' }}>
              Interactive Linux shell with persistent directory navigation (`cd`), real-time execution, and audit logging
            </p>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <span style={{ fontSize: '0.8rem', color: 'var(--text-subtle)' }}>Directory:</span>
            <input
              type="text"
              value={cwd}
              onChange={(e) => setCwd(e.target.value)}
              style={{
                background: 'var(--bg-base)',
                border: '1px solid var(--border)',
                borderRadius: '6px',
                padding: '0.35rem 0.65rem',
                color: '#fff',
                fontFamily: 'var(--font-mono)',
                fontSize: '0.8rem',
                width: '200px',
              }}
            />
            <button className="btn" onClick={() => setHistory([])}>
              🧹 Clear Console
            </button>
          </div>
        </div>

        {/* Quick Command Buttons */}
        <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap', alignItems: 'center' }}>
          <span style={{ fontSize: '0.78rem', color: 'var(--text-subtle)' }}>Quick Diagnostics:</span>
          {quickShortcuts.map((s) => (
            <button
              key={s.cmd}
              className="quick-cmd-badge"
              onClick={() => executeCommand(s.cmd)}
              disabled={loading}
            >
              {s.label}
            </button>
          ))}
        </div>
      </div>

      {/* Terminal Window */}
      <div className="terminal-window">
        <div className="terminal-titlebar">
          <div className="terminal-dots">
            <div className="dot dot-red" />
            <div className="dot dot-yellow" />
            <div className="dot dot-green" />
          </div>
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: '0.78rem', color: 'var(--text-subtle)' }}>
            {user}@{hostname}:{cwd}
          </span>
          <span style={{ fontSize: '0.75rem', color: 'var(--accent-emerald)' }}>● SSH Shell Session</span>
        </div>

        {/* Terminal Screen Output */}
        <div className="terminal-screen" ref={screenRef}>
          {history.length === 0 ? (
            <div style={{ color: 'var(--text-subtle)', lineHeight: '1.6' }}>
              Welcome to the AIO-PANEL Linux Web Shell.
              <br />
              Directory navigation with <code style={{ color: 'var(--accent-emerald)' }}>cd &lt;dir&gt;</code> persists across commands. Type <code style={{ color: 'var(--accent-emerald)' }}>clear</code> to reset console.
            </div>
          ) : (
            history.map((h, i) => (
              <div key={i} style={{ marginBottom: '1.25rem' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--accent-cyan)' }}>
                  <span style={{ color: 'var(--accent-emerald)', fontWeight: 600 }}>
                    {h.user || user}@{h.hostname || hostname}:{h.cwd || cwd}{h.user === 'root' ? '#' : '$'}
                  </span>
                  <strong style={{ color: '#ffffff' }}>{h.command}</strong>
                  <span style={{ fontSize: '0.72rem', color: 'var(--text-subtle)', marginLeft: 'auto' }}>
                    [{h.timestamp} • {h.duration} • Exit: {h.exit_code}]
                  </span>
                </div>
                {h.stdout && (
                  <pre style={{ margin: '0.4rem 0 0 0', whiteSpace: 'pre-wrap', color: '#cbd5e1' }}>
                    {h.stdout}
                  </pre>
                )}
                {h.stderr && (
                  <pre style={{ margin: '0.4rem 0 0 0', whiteSpace: 'pre-wrap', color: 'var(--accent-rose)' }}>
                    {h.stderr}
                  </pre>
                )}
              </div>
            ))
          )}
          {loading && (
            <div style={{ color: 'var(--accent-amber)', animation: 'pulse 1.5s infinite' }}>
              ⏳ Executing command on host...
            </div>
          )}
        </div>

        {/* Command Input Row */}
        <div className="terminal-input-row">
          <span style={{ color: 'var(--accent-emerald)', fontFamily: 'var(--font-mono)', fontWeight: 600, fontSize: '0.82rem', whiteSpace: 'nowrap' }}>
            {user}@{hostname}:{cwd}{user === 'root' ? '#' : '$'}
          </span>
          <input
            ref={inputRef}
            type="text"
            className="terminal-input"
            placeholder="Type command here (e.g. ls -la, cd /home/ubuntu, whoami)..."
            value={command}
            onChange={(e) => setCommand(e.target.value)}
            onKeyDown={handleKeyDown}
            disabled={loading}
            autoFocus
          />
          <button
            className="btn btn-primary"
            onClick={() => executeCommand(command)}
            disabled={loading || !command.trim()}
            style={{ padding: '0.4rem 0.85rem', fontSize: '0.8rem' }}
          >
            Run
          </button>
        </div>
      </div>
    </div>
  );
};
