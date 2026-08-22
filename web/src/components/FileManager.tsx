import React, { useState, useEffect } from 'react';
import { FileItem } from '../types';

export const FileManager: React.FC = () => {
  const [currentPath, setCurrentPath] = useState<string>('/var/www');
  const [items, setItems] = useState<FileItem[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [editingFile, setEditingFile] = useState<{ path: string; name: string; content: string } | null>(null);
  const [isSaving, setIsSaving] = useState<boolean>(false);
  const [showNewModal, setShowNewModal] = useState<boolean>(false);
  const [newItemName, setNewItemName] = useState<string>('');
  const [isNewDir, setIsNewDir] = useState<boolean>(false);
  const [searchQuery, setSearchQuery] = useState<string>('');

  const fetchDirectory = async (targetPath: string) => {
    setLoading(true);
    try {
      const res = await fetch(`/api/v1/ops/files/browse?path=${encodeURIComponent(targetPath)}`);
      if (res.ok) {
        const data: FileItem[] = await res.json();
        setItems(data || []);
        setCurrentPath(targetPath);
      }
    } catch (e) {
      console.error('Failed to browse files:', e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDirectory(currentPath);
  }, []);

  const handleOpenItem = (item: FileItem) => {
    if (item.is_dir) {
      fetchDirectory(item.path);
    } else {
      handleOpenFile(item.path, item.name);
    }
  };

  const handleOpenFile = async (path: string, name: string) => {
    try {
      const res = await fetch(`/api/v1/ops/files/read?path=${encodeURIComponent(path)}`);
      if (res.ok) {
        const text = await res.text();
        setEditingFile({ path, name, content: text });
      } else {
        alert('Could not view file. Binary or exceeds 5MB size limit.');
      }
    } catch (e) {
      alert('Error reading file: ' + e);
    }
  };

  const handleSaveFile = async () => {
    if (!editingFile) return;
    setIsSaving(true);
    try {
      const res = await fetch('/api/v1/ops/files/write', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: editingFile.path, content: editingFile.content }),
      });
      const data = await res.json();
      if (res.ok && data.success) {
        alert('File saved successfully.');
        setEditingFile(null);
        fetchDirectory(currentPath);
      } else {
        alert('Failed to save file: ' + (data.error || 'Unknown error'));
      }
    } catch (e) {
      alert('Error saving file: ' + e);
    } finally {
      setIsSaving(false);
    }
  };

  const handleCreateNew = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newItemName.trim()) return;

    const fullPath = currentPath === '/' ? `/${newItemName.trim()}` : `${currentPath}/${newItemName.trim()}`;
    try {
      const res = await fetch('/api/v1/ops/files/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: fullPath, is_dir: isNewDir }),
      });
      const data = await res.json();
      if (res.ok && data.success) {
        setShowNewModal(false);
        setNewItemName('');
        fetchDirectory(currentPath);
      } else {
        alert('Failed to create: ' + (data.error || 'Unknown error'));
      }
    } catch (e) {
      alert('Error: ' + e);
    }
  };

  const handleDeleteItem = async (item: FileItem) => {
    if (!window.confirm(`Are you sure you want to permanently delete "${item.name}"?`)) return;
    try {
      const res = await fetch('/api/v1/ops/files/delete', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: item.path }),
      });
      const data = await res.json();
      if (res.ok && data.success) {
        fetchDirectory(currentPath);
      } else {
        alert('Failed to delete: ' + (data.error || 'Unknown error'));
      }
    } catch (e) {
      alert('Error deleting: ' + e);
    }
  };

  const navigateUp = () => {
    if (currentPath === '/' || currentPath === '') return;
    const parts = currentPath.split('/').filter(Boolean);
    parts.pop();
    const parent = parts.length === 0 ? '/' : '/' + parts.join('/');
    fetchDirectory(parent);
  };

  const quickPaths = ['/var/www', '/etc/nginx', '/etc', '/var/log', '/opt', '/home', '/root'];

  const filteredItems = items.filter((it) =>
    it.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Top Header & Breadcrumb Toolbar */}
      <div className="panel-card">
        <div className="panel-header" style={{ marginBottom: '1rem' }}>
          <div className="panel-title">
            <span>📁</span> File Manager & Code Editor
          </div>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <button className="btn" onClick={() => fetchDirectory(currentPath)}>
              🔄 Refresh
            </button>
            <button className="btn btn-primary" onClick={() => { setIsNewDir(false); setShowNewModal(true); }}>
              📄 New File
            </button>
            <button className="btn" onClick={() => { setIsNewDir(true); setShowNewModal(true); }}>
              📁 New Directory
            </button>
          </div>
        </div>

        {/* Quick bookmarks */}
        <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap', marginBottom: '1rem', alignItems: 'center' }}>
          <span style={{ fontSize: '0.8rem', color: 'var(--text-subtle)' }}>Quick Jump:</span>
          {quickPaths.map((qp) => (
            <button
              key={qp}
              className="quick-cmd-badge"
              onClick={() => fetchDirectory(qp)}
              style={{ background: currentPath === qp ? 'var(--primary)' : undefined, color: currentPath === qp ? '#fff' : undefined }}
            >
              {qp}
            </button>
          ))}
        </div>

        {/* Breadcrumb Path & Search Bar */}
        <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
          <button className="btn" onClick={navigateUp} disabled={currentPath === '/'}>
            ⬆️ Up
          </button>
          <div className="file-breadcrumbs" style={{ flex: 1 }}>
            <span
              className="breadcrumb-crumb"
              onClick={() => fetchDirectory('/')}
              style={{ fontWeight: 700, color: 'var(--accent-cyan)' }}
            >
              root
            </span>
            {currentPath.split('/').filter(Boolean).map((part, idx, arr) => {
              const subPath = '/' + arr.slice(0, idx + 1).join('/');
              return (
                <React.Fragment key={subPath}>
                  <span style={{ color: 'var(--border)' }}>/</span>
                  <span className="breadcrumb-crumb" onClick={() => fetchDirectory(subPath)}>
                    {part}
                  </span>
                </React.Fragment>
              );
            })}
          </div>

          <input
            type="text"
            placeholder="Search items..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={{
              padding: '0.55rem 0.85rem',
              background: 'var(--bg-base)',
              border: '1px solid var(--border)',
              borderRadius: '8px',
              color: '#fff',
              fontSize: '0.82rem',
              width: '200px',
            }}
          />
        </div>
      </div>

      {/* Directory Content Table */}
      <div className="panel-card">
        {loading ? (
          <div style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-subtle)' }}>
            Reading filesystem contents...
          </div>
        ) : filteredItems.length === 0 ? (
          <div style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-subtle)' }}>
            This directory is empty.
          </div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table className="custom-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Size</th>
                  <th>Permissions</th>
                  <th>Modified Date</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredItems.map((item) => (
                  <tr key={item.path} style={{ cursor: 'pointer' }}>
                    <td onClick={() => handleOpenItem(item)}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.65rem' }}>
                        <span style={{ fontSize: '1.2rem' }}>{item.is_dir ? '📁' : '📄'}</span>
                        <span style={{ fontWeight: item.is_dir ? 700 : 500, color: item.is_dir ? 'var(--accent-cyan)' : '#fff' }}>
                          {item.name}
                        </span>
                      </div>
                    </td>
                    <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.82rem', color: 'var(--text-muted)' }}>
                      {item.is_dir ? '--' : item.size_human}
                    </td>
                    <td style={{ fontFamily: 'var(--font-mono)', fontSize: '0.8rem', color: 'var(--text-subtle)' }}>
                      {item.permissions}
                    </td>
                    <td style={{ fontSize: '0.8rem', color: 'var(--text-subtle)' }}>
                      {new Date(item.modified_time).toLocaleString()}
                    </td>
                    <td>
                      <div style={{ display: 'flex', gap: '0.4rem' }}>
                        {!item.is_dir && (
                          <button
                            className="btn"
                            onClick={(e) => { e.stopPropagation(); handleOpenFile(item.path, item.name); }}
                          >
                            ✏️ Edit
                          </button>
                        )}
                        <button
                          className="btn btn-danger"
                          onClick={(e) => { e.stopPropagation(); handleDeleteItem(item); }}
                        >
                          🗑️
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

      {/* Code Editor Modal */}
      {editingFile && (
        <div className="modal-backdrop" onClick={() => setEditingFile(null)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '900px' }}>
            <div className="modal-header">
              <div>
                <h4 style={{ fontWeight: 700, fontSize: '1.1rem' }}>Editing: {editingFile.name}</h4>
                <p style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                  {editingFile.path}
                </p>
              </div>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <button className="btn" onClick={() => setEditingFile(null)}>
                  Cancel
                </button>
                <button className="btn btn-primary" onClick={handleSaveFile} disabled={isSaving}>
                  {isSaving ? 'Saving...' : '💾 Save Changes'}
                </button>
              </div>
            </div>
            <div className="modal-body">
              <textarea
                className="code-editor"
                value={editingFile.content}
                onChange={(e) => setEditingFile({ ...editingFile, content: e.target.value })}
                spellCheck={false}
              />
            </div>
          </div>
        </div>
      )}

      {/* New File / Folder Modal */}
      {showNewModal && (
        <div className="modal-backdrop" onClick={() => setShowNewModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '480px' }}>
            <div className="modal-header">
              <h4 style={{ fontWeight: 700 }}>
                {isNewDir ? '📁 Create New Directory' : '📄 Create New File'}
              </h4>
              <button className="btn" onClick={() => setShowNewModal(false)}>
                ✕
              </button>
            </div>
            <form onSubmit={handleCreateNew} style={{ padding: '1.5rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.4rem' }}>
                  Target Location
                </label>
                <div style={{ fontFamily: 'var(--font-mono)', fontSize: '0.8rem', color: 'var(--accent-cyan)' }}>
                  {currentPath}
                </div>
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.4rem' }}>
                  {isNewDir ? 'Directory Name' : 'File Name'}
                </label>
                <input
                  type="text"
                  required
                  placeholder={isNewDir ? 'e.g. static' : 'e.g. config.json'}
                  value={newItemName}
                  onChange={(e) => setNewItemName(e.target.value)}
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
                <button type="button" className="btn" onClick={() => setShowNewModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary">
                  Create
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
