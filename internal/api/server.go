package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/codingguna/aio-panel/internal/api/handlers"
	"github.com/codingguna/aio-panel/internal/api/middleware"
	"github.com/codingguna/aio-panel/internal/auth"
	"github.com/codingguna/aio-panel/internal/config"
	"github.com/codingguna/aio-panel/internal/db"
	"github.com/codingguna/aio-panel/internal/logger"
	"github.com/codingguna/aio-panel/web"
)

// Server represents the AIO HTTP/HTTPS server daemon
type Server struct {
	cfg        *config.Config
	store      *db.Store
	httpServer *http.Server
}

// NewServer builds and configures a new Server instance
func NewServer(cfg *config.Config, store *db.Store) *Server {
	mux := http.NewServeMux()

	sessionMgr := auth.NewSessionManager()
	authHandler := handlers.NewAuthHandler(store, sessionMgr)
	storeHandler := handlers.NewStoreHandler(store)
	sysHandler := handlers.NewSystemHandler(store)
	auditHandler := handlers.NewAuditHandler(store)
	servicesHandler := handlers.NewServicesHandler(store)
	discoveryHandler := handlers.NewDiscoveryHandler(cfg)
	securityHandler := handlers.NewSecurityHandler(cfg, store)
	webHandler := handlers.NewWebServerHandler(cfg, store)
	dbHandler := handlers.NewDatabaseHandler(cfg, store)
	opsHandler := handlers.NewOpsHandler(cfg, store)

	// Auth Routes (Login, Logout, Initial Setup, Session Status)
	mux.HandleFunc("GET /api/v1/auth/status", authHandler.Status)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)
	mux.HandleFunc("POST /api/v1/auth/setup", authHandler.Setup)
	mux.HandleFunc("GET /api/v1/auth/users", authHandler.ListUsers)
	mux.HandleFunc("DELETE /api/v1/auth/users/{id}", authHandler.DeleteUser)

	// Store Routes (1-Click Package Installer)
	mux.HandleFunc("GET /api/v1/store/packages", storeHandler.ListPackages)
	mux.HandleFunc("POST /api/v1/store/install", storeHandler.InstallPackage)
	mux.HandleFunc("GET /api/v1/store/jobs/{id}", storeHandler.GetJobStatus)

	// API Telemetry & Audit Routes
	mux.HandleFunc("GET /health", handlers.HealthCheck)
	mux.HandleFunc("GET /api/v1/health", handlers.HealthCheck)
	mux.HandleFunc("GET /api/v1/system/info", sysHandler.GetSystemInfo)
	mux.HandleFunc("GET /api/v1/system/metrics", sysHandler.GetSystemMetrics)
	mux.HandleFunc("GET /api/v1/system/history", sysHandler.GetMetricsHistory)
	mux.HandleFunc("GET /api/v1/audit/events", auditHandler.GetAuditEvents)

	// Services Management & CRUD Routes
	mux.HandleFunc("GET /api/v1/services", servicesHandler.ListServices)
	mux.HandleFunc("POST /api/v1/services/create", servicesHandler.CreateService)
	mux.HandleFunc("GET /api/v1/services/{name}", servicesHandler.InspectService)
	mux.HandleFunc("POST /api/v1/services/{name}/action", servicesHandler.ControlService)
	mux.HandleFunc("GET /api/v1/services/{name}/logs", servicesHandler.GetServiceLogs)
	mux.HandleFunc("DELETE /api/v1/services/{name}", servicesHandler.DeleteService)

	// Discovery & Applications Routes
	mux.HandleFunc("GET /api/v1/discovery", discoveryHandler.GetPreflightReport)
	mux.HandleFunc("GET /api/v1/applications", discoveryHandler.GetApplications)
	mux.HandleFunc("GET /api/v1/nginx/sites", discoveryHandler.GetNginxSites)

	// Security Management Routes (SSH, Firewall, Linux Users)
	mux.HandleFunc("GET /api/v1/security/ssh", securityHandler.GetSSHConfig)
	mux.HandleFunc("POST /api/v1/security/ssh/config", securityHandler.UpdateSSHConfig)
	mux.HandleFunc("GET /api/v1/security/ssh/keys", securityHandler.GetAuthorizedKeys)
	mux.HandleFunc("POST /api/v1/security/ssh/keys", securityHandler.AddAuthorizedKey)
	mux.HandleFunc("GET /api/v1/security/ssh/sessions", securityHandler.GetSSHSessions)
	mux.HandleFunc("GET /api/v1/security/firewall", securityHandler.GetFirewallStatus)
	mux.HandleFunc("POST /api/v1/security/firewall/rules", securityHandler.AddFirewallRule)
	mux.HandleFunc("DELETE /api/v1/security/firewall/rules/{id}", securityHandler.DeleteFirewallRule)
	mux.HandleFunc("POST /api/v1/security/firewall/toggle", securityHandler.ToggleFirewall)
	mux.HandleFunc("GET /api/v1/security/users", securityHandler.ListUsers)
	mux.HandleFunc("POST /api/v1/security/users", securityHandler.CreateUser)
	mux.HandleFunc("DELETE /api/v1/security/users/{username}", securityHandler.DeleteUser)

	// WebServer & SSL Routes
	mux.HandleFunc("GET /api/v1/web/vhosts", webHandler.ListVHosts)
	mux.HandleFunc("POST /api/v1/web/vhosts", webHandler.CreateVHost)
	mux.HandleFunc("DELETE /api/v1/web/vhosts/{domain}", webHandler.DeleteVHost)
	mux.HandleFunc("GET /api/v1/web/ssl/certificates", webHandler.ListCertificates)
	mux.HandleFunc("POST /api/v1/web/ssl/issue", webHandler.IssueCertificate)

	// Databases Routes & CRUD (PostgreSQL, MySQL)
	mux.HandleFunc("GET /api/v1/databases/postgres", dbHandler.ListPostgresDBs)
	mux.HandleFunc("GET /api/v1/databases/postgres/users", dbHandler.ListPostgresUsers)
	mux.HandleFunc("POST /api/v1/databases/postgres/create", dbHandler.CreatePostgresDB)
	mux.HandleFunc("DELETE /api/v1/databases/postgres/{name}", dbHandler.DeletePostgresDB)
	mux.HandleFunc("POST /api/v1/databases/postgres/backup", dbHandler.BackupPostgres)
	mux.HandleFunc("GET /api/v1/databases/mysql", dbHandler.ListMySQLDBs)
	mux.HandleFunc("POST /api/v1/databases/mysql/create", dbHandler.CreateMySQLDB)
	mux.HandleFunc("DELETE /api/v1/databases/mysql/{name}", dbHandler.DeleteMySQLDB)
	mux.HandleFunc("POST /api/v1/databases/mysql/backup", dbHandler.BackupMySQL)

	// Operations Routes (File Manager, Terminal, Logs, Cron, Docker, Deployments, Backups)
	mux.HandleFunc("GET /api/v1/ops/files/browse", opsHandler.BrowseFiles)
	mux.HandleFunc("GET /api/v1/ops/files/read", opsHandler.ReadFile)
	mux.HandleFunc("POST /api/v1/ops/files/write", opsHandler.WriteFile)
	mux.HandleFunc("POST /api/v1/ops/files/create", opsHandler.CreateFile)
	mux.HandleFunc("POST /api/v1/ops/files/delete", opsHandler.DeleteFile)
	mux.HandleFunc("POST /api/v1/ops/terminal/exec", opsHandler.ExecuteTerminalCommand)
	mux.HandleFunc("GET /api/v1/ops/logs", opsHandler.GetSystemLogs)
	mux.HandleFunc("GET /api/v1/ops/backups", opsHandler.ListBackups)
	mux.HandleFunc("GET /api/v1/ops/cron", opsHandler.ListCronJobs)
	mux.HandleFunc("POST /api/v1/ops/cron/create", opsHandler.CreateCronJob)
	mux.HandleFunc("DELETE /api/v1/ops/cron/{id}", opsHandler.DeleteCronJob)
	mux.HandleFunc("GET /api/v1/ops/docker/containers", opsHandler.ListDockerContainers)
	mux.HandleFunc("POST /api/v1/ops/docker/containers/{id}/action", opsHandler.ControlDockerContainer)
	mux.HandleFunc("DELETE /api/v1/ops/docker/containers/{id}", opsHandler.DeleteDockerContainer)
	mux.HandleFunc("GET /api/v1/ops/docker/images", opsHandler.ListDockerImages)
	mux.HandleFunc("GET /api/v1/ops/docker/containers/{id}/logs", opsHandler.GetDockerContainerLogs)
	mux.HandleFunc("POST /api/v1/ops/deployments/run", opsHandler.RunDeployment)

	// Web UI handler (Serves embedded React UI with SPA fallback)
	webFS := web.GetFileSystem()
	if webFS != nil {
		fileServer := http.FileServer(webFS)
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if path == "" {
				path = "/"
			}
			f, err := webFS.Open(path)
			if err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
			// Fallback to index.html for client-side routing
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
		})
	} else {
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(DashboardHTML))
		})
	}

	// Wrap router with middleware chain: AuthEnforcer -> SecurityHeaders -> RequestLogger -> Recoverer -> CORS
	var handler http.Handler = mux
	handler = middleware.AuthEnforcer(store, sessionMgr)(handler)
	handler = middleware.CORS(handler)
	handler = middleware.SecurityHeaders(handler)
	handler = middleware.RequestLogger(handler)
	handler = middleware.Recoverer(handler)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		cfg:        cfg,
		store:      store,
		httpServer: srv,
	}
}

// Start begins listening on the configured host & port
func (s *Server) Start() error {
	logger.Log.Info("AIO-PANEL starting",
		slog.String("address", s.httpServer.Addr),
		slog.Bool("tls", s.cfg.Server.TLS),
		slog.String("db_path", s.cfg.Database.Path),
	)

	// Log an audit event for server start
	if s.store != nil {
		_ = s.store.LogAudit(context.Background(), "system", "SERVER_START", s.httpServer.Addr, "SUCCESS", "AIO-PANEL daemon started", "127.0.0.1")
	}

	if s.cfg.Server.TLS && s.cfg.Server.CertFile != "" && s.cfg.Server.KeyFile != "" {
		return s.httpServer.ListenAndServeTLS(s.cfg.Server.CertFile, s.cfg.Server.KeyFile)
	}

	return s.httpServer.ListenAndServe()
}

// Shutdown stops the server gracefully
func (s *Server) Shutdown(ctx context.Context) error {
	logger.Log.Info("shutting down AIO-PANEL server...")
	if s.store != nil {
		_ = s.store.LogAudit(context.Background(), "system", "SERVER_STOP", s.httpServer.Addr, "SUCCESS", "AIO-PANEL daemon stopped gracefully", "127.0.0.1")
	}
	return s.httpServer.Shutdown(ctx)
}

// DashboardHTML provides an immediate rich dark-mode UI with live system telemetry, services & apps control
const DashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>AIO-PANEL • All-In-One Linux Server Control</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet">
  <style>
    :root {
      --bg-base: #090d16;
      --bg-surface: #111827;
      --bg-surface-elevated: #1a2234;
      --border: rgba(255, 255, 255, 0.08);
      --border-focus: rgba(99, 102, 241, 0.5);
      --primary: #6366f1;
      --primary-gradient: linear-gradient(135deg, #6366f1 0%, #a855f7 100%);
      --accent-cyan: #06b6d4;
      --accent-emerald: #10b981;
      --accent-amber: #f59e0b;
      --accent-rose: #f43f5e;
      --text-main: #f8fafc;
      --text-muted: #94a3b8;
      --text-subtle: #64748b;
      --font-main: 'Outfit', -apple-system, BlinkMacSystemFont, sans-serif;
      --font-mono: 'JetBrains Mono', monospace;
    }

    * { box-sizing: border-box; margin: 0; padding: 0; }

    body {
      background-color: var(--bg-base);
      background-image: 
        radial-gradient(at 0% 0%, rgba(99, 102, 241, 0.12) 0px, transparent 50%),
        radial-gradient(at 100% 100%, rgba(168, 85, 247, 0.08) 0px, transparent 50%);
      color: var(--text-main);
      font-family: var(--font-main);
      min-height: 100vh;
      display: flex;
      flex-direction: column;
      overflow-x: hidden;
    }

    header {
      border-bottom: 1px solid var(--border);
      background: rgba(17, 24, 39, 0.7);
      backdrop-filter: blur(12px);
      position: sticky;
      top: 0;
      z-index: 50;
      padding: 0.85rem 2rem;
      display: flex;
      align-items: center;
      justify-content: space-between;
    }

    .brand {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      text-decoration: none;
      color: inherit;
    }

    .logo-badge {
      width: 36px;
      height: 36px;
      border-radius: 10px;
      background: var(--primary-gradient);
      display: flex;
      align-items: center;
      justify-content: center;
      font-weight: 800;
      font-size: 1.1rem;
      color: #fff;
      box-shadow: 0 4px 15px rgba(99, 102, 241, 0.4);
    }

    .brand-text h1 {
      font-size: 1.2rem;
      font-weight: 700;
      letter-spacing: -0.02em;
      background: linear-gradient(180deg, #fff 0%, #cbd5e1 100%);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
    }

    .brand-text p {
      font-size: 0.72rem;
      color: var(--text-subtle);
      font-weight: 500;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }

    .header-status {
      display: flex;
      align-items: center;
      gap: 1.25rem;
    }

    .status-pill {
      display: inline-flex;
      align-items: center;
      gap: 0.45rem;
      padding: 0.35rem 0.85rem;
      border-radius: 9999px;
      background: rgba(16, 185, 129, 0.12);
      border: 1px solid rgba(16, 185, 129, 0.3);
      color: var(--accent-emerald);
      font-size: 0.8rem;
      font-weight: 600;
    }

    .status-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background: var(--accent-emerald);
      box-shadow: 0 0 10px var(--accent-emerald);
      animation: pulse 2s infinite ease-in-out;
    }

    @keyframes pulse {
      0%, 100% { opacity: 1; transform: scale(1); }
      50% { opacity: 0.4; transform: scale(0.85); }
    }

    main {
      flex: 1;
      padding: 2rem;
      max-width: 1400px;
      margin: 0 auto;
      width: 100%;
      display: flex;
      flex-direction: column;
      gap: 2rem;
    }

    .hero-banner {
      background: var(--bg-surface);
      border: 1px solid var(--border);
      border-radius: 16px;
      padding: 1.75rem 2rem;
      display: flex;
      justify-content: space-between;
      align-items: center;
      flex-wrap: wrap;
      gap: 1.5rem;
      box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
    }

    .hero-title h2 {
      font-size: 1.6rem;
      font-weight: 700;
      margin-bottom: 0.3rem;
    }

    .hero-title p {
      color: var(--text-muted);
      font-size: 0.95rem;
    }

    .tag-container {
      display: flex;
      gap: 0.5rem;
      flex-wrap: wrap;
    }

    .tag {
      background: var(--bg-surface-elevated);
      border: 1px solid var(--border);
      color: var(--text-muted);
      font-size: 0.78rem;
      padding: 0.35rem 0.75rem;
      border-radius: 8px;
      font-family: var(--font-mono);
    }

    .grid-metrics {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
      gap: 1.25rem;
    }

    .card {
      background: var(--bg-surface);
      border: 1px solid var(--border);
      border-radius: 16px;
      padding: 1.5rem;
      display: flex;
      flex-direction: column;
      justify-content: space-between;
      transition: transform 0.2s ease, border-color 0.2s ease;
      position: relative;
      overflow: hidden;
    }

    .card:hover {
      transform: translateY(-2px);
      border-color: rgba(99, 102, 241, 0.4);
    }

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 1rem;
    }

    .card-title {
      font-size: 0.85rem;
      font-weight: 600;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      color: var(--text-muted);
    }

    .card-icon { font-size: 1.25rem; }

    .metric-value {
      font-size: 2.2rem;
      font-weight: 800;
      font-family: var(--font-main);
      letter-spacing: -0.03em;
      margin-bottom: 0.5rem;
      color: #fff;
    }

    .progress-bar-bg {
      width: 100%;
      height: 8px;
      background: var(--bg-surface-elevated);
      border-radius: 9999px;
      overflow: hidden;
      margin-bottom: 0.75rem;
    }

    .progress-bar-fill {
      height: 100%;
      border-radius: 9999px;
      transition: width 0.5s ease;
    }

    .cpu-fill { background: linear-gradient(90deg, #6366f1, #a855f7); }
    .mem-fill { background: linear-gradient(90deg, #06b6d4, #3b82f6); }
    .disk-fill { background: linear-gradient(90deg, #10b981, #06b6d4); }
    .load-fill { background: linear-gradient(90deg, #f59e0b, #f43f5e); }

    .card-subtext {
      font-size: 0.8rem;
      color: var(--text-subtle);
      display: flex;
      justify-content: space-between;
    }

    .panel-card {
      background: var(--bg-surface);
      border: 1px solid var(--border);
      border-radius: 16px;
      padding: 1.5rem;
    }

    .panel-card h3 {
      font-size: 1.15rem;
      font-weight: 700;
      margin-bottom: 1.25rem;
      display: flex;
      align-items: center;
      justify-content: space-between;
    }

    .services-table {
      width: 100%;
      border-collapse: collapse;
      font-size: 0.88rem;
    }

    .services-table th {
      text-align: left;
      padding: 0.75rem 1rem;
      color: var(--text-subtle);
      font-size: 0.75rem;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      border-bottom: 1px solid var(--border);
    }

    .services-table td {
      padding: 0.85rem 1rem;
      border-bottom: 1px solid rgba(255, 255, 255, 0.04);
      vertical-align: middle;
    }

    .services-table tr:hover td {
      background: rgba(255, 255, 255, 0.02);
    }

    .badge-status {
      display: inline-flex;
      align-items: center;
      gap: 0.35rem;
      padding: 0.2rem 0.6rem;
      border-radius: 6px;
      font-size: 0.75rem;
      font-weight: 600;
    }

    .badge-active {
      background: rgba(16, 185, 129, 0.15);
      color: var(--accent-emerald);
      border: 1px solid rgba(16, 185, 129, 0.3);
    }

    .badge-inactive {
      background: rgba(244, 63, 94, 0.15);
      color: var(--accent-rose);
      border: 1px solid rgba(244, 63, 94, 0.3);
    }

    .badge-owner {
      background: var(--bg-surface-elevated);
      color: var(--text-muted);
      border: 1px solid var(--border);
      padding: 0.2rem 0.5rem;
      border-radius: 6px;
      font-size: 0.7rem;
      font-family: var(--font-mono);
    }

    .btn-action {
      background: var(--bg-surface-elevated);
      border: 1px solid var(--border);
      color: var(--text-main);
      padding: 0.35rem 0.65rem;
      border-radius: 6px;
      font-size: 0.75rem;
      cursor: pointer;
      font-weight: 500;
      transition: background 0.2s, border-color 0.2s;
      display: inline-flex;
      align-items: center;
      gap: 0.3rem;
    }

    .btn-action:hover {
      background: rgba(99, 102, 241, 0.2);
      border-color: var(--primary);
    }

    .actions-group {
      display: flex;
      gap: 0.4rem;
    }

    .details-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 1.5rem;
    }

    @media (max-width: 900px) {
      .details-grid {
        grid-template-columns: 1fr;
      }
    }

    .info-table {
      width: 100%;
      border-collapse: collapse;
      font-size: 0.88rem;
    }

    .info-table tr {
      border-bottom: 1px solid rgba(255, 255, 255, 0.04);
    }

    .info-table td {
      padding: 0.75rem 0;
    }

    .info-table td:first-child {
      color: var(--text-muted);
      width: 35%;
    }

    .info-table td:last-child {
      font-family: var(--font-mono);
      color: #fff;
      text-align: right;
    }

    /* Modal */
    .modal-overlay {
      position: fixed;
      top: 0; left: 0; right: 0; bottom: 0;
      background: rgba(0, 0, 0, 0.8);
      backdrop-filter: blur(8px);
      z-index: 100;
      display: none;
      align-items: center;
      justify-content: center;
      padding: 2rem;
    }

    .modal-box {
      background: var(--bg-surface);
      border: 1px solid var(--border);
      border-radius: 16px;
      width: 100%;
      max-width: 800px;
      max-height: 80vh;
      display: flex;
      flex-direction: column;
      box-shadow: 0 20px 50px rgba(0, 0, 0, 0.5);
    }

    .modal-header {
      padding: 1.25rem 1.5rem;
      border-bottom: 1px solid var(--border);
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .modal-body {
      padding: 1.5rem;
      overflow-y: auto;
      flex: 1;
      font-family: var(--font-mono);
      font-size: 0.82rem;
      color: #cbd5e1;
      background: #060911;
      white-space: pre-wrap;
      line-height: 1.5;
    }

    .modal-close {
      background: none;
      border: none;
      color: var(--text-muted);
      font-size: 1.5rem;
      cursor: pointer;
    }

    footer {
      border-top: 1px solid var(--border);
      padding: 1.25rem 2rem;
      text-align: center;
      font-size: 0.8rem;
      color: var(--text-subtle);
      background: rgba(17, 24, 39, 0.4);
    }
  </style>
</head>
<body>
  <header>
    <a href="/" class="brand">
      <div class="logo-badge">A</div>
      <div class="brand-text">
        <h1>AIO-PANEL</h1>
        <p>All-In-One Linux Server Control</p>
      </div>
    </a>
    <div class="header-status">
      <div class="status-pill">
        <span class="status-dot"></span>
        <span>DAEMON RUNNING</span>
      </div>
    </div>
  </header>

  <main>
    <div class="hero-banner">
      <div class="hero-title">
        <h2>Server Overview & Health</h2>
        <p>Non-invasive Linux management running natively on single Go executable</p>
      </div>
      <div class="tag-container">
        <span class="tag" id="tag-host">Host: loading...</span>
        <span class="tag" id="tag-os">OS: loading...</span>
        <span class="tag" id="tag-uptime">Uptime: loading...</span>
      </div>
    </div>

    <!-- Live Metrics Grid -->
    <div class="grid-metrics">
      <div class="card">
        <div class="card-header">
          <span class="card-title">CPU Utilization</span>
          <span class="card-icon">⚡</span>
        </div>
        <div class="metric-value" id="val-cpu">--%</div>
        <div class="progress-bar-bg">
          <div class="progress-bar-fill cpu-fill" id="bar-cpu" style="width: 0%"></div>
        </div>
        <div class="card-subtext">
          <span id="sub-cpu-cores">Cores: --</span>
          <span id="sub-cpu-load">Load: --</span>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <span class="card-title">Memory (RAM)</span>
          <span class="card-icon">🧠</span>
        </div>
        <div class="metric-value" id="val-mem">--%</div>
        <div class="progress-bar-bg">
          <div class="progress-bar-fill mem-fill" id="bar-mem" style="width: 0%"></div>
        </div>
        <div class="card-subtext">
          <span id="sub-mem-used">Used: --</span>
          <span id="sub-mem-total">Total: --</span>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <span class="card-title">Disk Storage</span>
          <span class="card-icon">💾</span>
        </div>
        <div class="metric-value" id="val-disk">--%</div>
        <div class="progress-bar-bg">
          <div class="progress-bar-fill disk-fill" id="bar-disk" style="width: 0%"></div>
        </div>
        <div class="card-subtext">
          <span id="sub-disk-path">Mount: /</span>
          <span id="sub-disk-free">Free: --</span>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <span class="card-title">System Load (1/5/15m)</span>
          <span class="card-icon">📊</span>
        </div>
        <div class="metric-value" id="val-load">0.00</div>
        <div class="progress-bar-bg">
          <div class="progress-bar-fill load-fill" id="bar-load" style="width: 10%"></div>
        </div>
        <div class="card-subtext">
          <span id="sub-procs">Active Processes: --</span>
          <span>Health: Optimal</span>
        </div>
      </div>
    </div>

    <!-- Discovered Applications Table -->
    <div class="panel-card">
      <h3>
        <span>🚀 Discovered Web Applications (e.g. MemoTrack)</span>
        <button class="btn-action" onclick="fetchApplications()">🔄 Scan Apps</button>
      </h3>
      <table class="services-table">
        <thead>
          <tr>
            <th>Application Name</th>
            <th>Framework / Type</th>
            <th>Path</th>
            <th>Runtime</th>
            <th>Linked Service</th>
            <th>Domain</th>
            <th>Ownership</th>
          </tr>
        </thead>
        <tbody id="apps-tbody">
          <tr><td colspan="7" style="text-align: center; color: var(--text-subtle);">Discovering existing applications...</td></tr>
        </tbody>
      </table>
    </div>

    <!-- Services & Daemons Table -->
    <div class="panel-card">
      <h3>
        <span>⚙️ Native System Services (Discovery & Control)</span>
        <button class="btn-action" onclick="fetchServices()">🔄 Refresh</button>
      </h3>
      <table class="services-table">
        <thead>
          <tr>
            <th>Service Name</th>
            <th>Description</th>
            <th>Status</th>
            <th>Ownership</th>
            <th>PID / Memory</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody id="services-tbody">
          <tr><td colspan="6" style="text-align: center; color: var(--text-subtle);">Discovering system services...</td></tr>
        </tbody>
      </table>
    </div>

    <!-- Specifications & Architecture Cards -->
    <div class="details-grid">
      <div class="panel-card">
        <h3><span>🖥️</span> Host & Runtime Specifications</h3>
        <table class="info-table">
          <tbody>
            <tr><td>Hostname</td><td id="info-hostname">--</td></tr>
            <tr><td>Operating System</td><td id="info-os">--</td></tr>
            <tr><td>Kernel & Arch</td><td id="info-kernel">--</td></tr>
            <tr><td>CPU Model</td><td id="info-cpu-model">--</td></tr>
            <tr><td>Go Runtime</td><td id="info-go-version">--</td></tr>
            <tr><td>AIO Version</td><td id="info-aio-version">--</td></tr>
          </tbody>
        </table>
      </div>

      <div class="panel-card">
        <h3><span>🛡️</span> Non-Invasive Safety Guarantees</h3>
        <div class="feature-list" style="display: flex; flex-direction: column; gap: 0.75rem;">
          <div style="background: var(--bg-surface-elevated); border: 1px solid var(--border); border-radius: 10px; padding: 0.85rem 1rem; display: flex; justify-content: space-between; align-items: center;">
            <span style="display: flex; align-items: center; gap: 0.5rem; font-size: 0.85rem;"><span>👁️</span> Observer-First Discovery</span>
            <span class="badge-status badge-active">Read-Only</span>
          </div>
          <div style="background: var(--bg-surface-elevated); border: 1px solid var(--border); border-radius: 10px; padding: 0.85rem 1rem; display: flex; justify-content: space-between; align-items: center;">
            <span style="display: flex; align-items: center; gap: 0.5rem; font-size: 0.85rem;"><span>📦</span> Failure Isolation</span>
            <span class="badge-status badge-active">Guaranteed</span>
          </div>
          <div style="background: var(--bg-surface-elevated); border: 1px solid var(--border); border-radius: 10px; padding: 0.85rem 1rem; display: flex; justify-content: space-between; align-items: center;">
            <span style="display: flex; align-items: center; gap: 0.5rem; font-size: 0.85rem;"><span>⚡</span> Port 5555 Binding</span>
            <span class="badge-status badge-active">Healthy</span>
          </div>
        </div>
      </div>
    </div>
  </main>

  <!-- Logs Modal -->
  <div class="modal-overlay" id="logs-modal" onclick="closeLogs(event)">
    <div class="modal-box" onclick="event.stopPropagation()">
      <div class="modal-header">
        <h4 id="modal-title">Service Journal Logs</h4>
        <button class="modal-close" onclick="closeLogs()">×</button>
      </div>
      <div class="modal-body" id="modal-content">Loading journal logs...</div>
    </div>
  </div>

  <footer>
    AIO-PANEL • Lightweight, Powerful, Security-First Linux Server Management
  </footer>

  <script>
    function formatBytes(bytes) {
      if (!bytes || bytes === 0) return '0 B';
      const k = 1024;
      const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i];
    }

    function formatUptime(seconds) {
      if (!seconds) return '0m';
      const days = Math.floor(seconds / (3600*24));
      const hours = Math.floor((seconds % (3600*24)) / 3600);
      const mins = Math.floor((seconds % 3600) / 60);
      var res = '';
      if (days > 0) res += days + 'd ';
      if (hours > 0) res += hours + 'h ';
      res += mins + 'm';
      return res;
    }

    async function fetchInfo() {
      try {
        const res = await fetch('/api/v1/system/info');
        if (!res.ok) return;
        const data = await res.json();
        document.getElementById('tag-host').innerText = 'Host: ' + data.hostname;
        document.getElementById('tag-os').innerText = 'OS: ' + data.os;
        document.getElementById('tag-uptime').innerText = 'Uptime: ' + formatUptime(data.uptime_seconds);

        document.getElementById('info-hostname').innerText = data.hostname;
        document.getElementById('info-os').innerText = data.os;
        document.getElementById('info-kernel').innerText = data.kernel + ' (' + data.architecture + ')';
        document.getElementById('info-cpu-model').innerText = data.cpu_model + ' (' + data.cpu_cores + ' cores)';
        document.getElementById('info-go-version').innerText = data.go_version;
        document.getElementById('info-aio-version').innerText = 'v' + data.panel_version;
      } catch (e) {
        console.error("Info error:", e);
      }
    }

    async function fetchMetrics() {
      try {
        const res = await fetch('/api/v1/system/metrics');
        if (!res.ok) return;
        const m = await res.json();

        // CPU
        const cpuPct = m.cpu.usage_percent.toFixed(1);
        document.getElementById('val-cpu').innerText = cpuPct + '%';
        document.getElementById('bar-cpu').style.width = Math.min(100, Math.max(0, cpuPct)) + '%';
        document.getElementById('sub-cpu-cores').innerText = 'Cores: ' + m.cpu.cores;
        document.getElementById('sub-cpu-load').innerText = 'Load: ' + m.load_average[0].toFixed(2);

        // Memory
        const memPct = m.memory.usage_percent.toFixed(1);
        document.getElementById('val-mem').innerText = memPct + '%';
        document.getElementById('bar-mem').style.width = Math.min(100, Math.max(0, memPct)) + '%';
        document.getElementById('sub-mem-used').innerText = 'Used: ' + formatBytes(m.memory.used_bytes);
        document.getElementById('sub-mem-total').innerText = 'Total: ' + formatBytes(m.memory.total_bytes);

        // Disk
        const diskPct = m.disk.usage_percent.toFixed(1);
        document.getElementById('val-disk').innerText = diskPct + '%';
        document.getElementById('bar-disk').style.width = Math.min(100, Math.max(0, diskPct)) + '%';
        document.getElementById('sub-disk-free').innerText = 'Free: ' + formatBytes(m.disk.free_bytes);

        // Load
        const load1 = m.load_average[0].toFixed(2);
        document.getElementById('val-load').innerText = load1;
        document.getElementById('sub-procs').innerText = 'Goroutines/Procs: ' + m.processes;
        document.getElementById('bar-load').style.width = Math.min(100, Math.max(5, load1 * 20)) + '%';
      } catch (e) {
        console.error("Metrics error:", e);
      }
    }

    async function fetchApplications() {
      try {
        const res = await fetch('/api/v1/applications');
        if (!res.ok) return;
        const apps = await res.json();

        const tbody = document.getElementById('apps-tbody');
        tbody.innerHTML = '';

        if (!apps || apps.length === 0) {
          tbody.innerHTML = '<tr><td colspan="7" style="text-align: center; color: var(--text-subtle);">No web applications found in /var/www or /srv</td></tr>';
          return;
        }

        apps.forEach(function(app) {
          const tr = document.createElement('tr');
          tr.innerHTML = '<td><strong style="color: #fff;">' + app.name + '</strong></td>' +
            '<td><span class="tag" style="background: rgba(99, 102, 241, 0.15); color: #818cf8;">' + app.type + '</span></td>' +
            '<td style="font-family: var(--font-mono); font-size: 0.78rem; color: var(--text-muted);">' + app.path + '</td>' +
            '<td style="font-size: 0.82rem;">' + app.runtime + '</td>' +
            '<td style="font-family: var(--font-mono); font-size: 0.78rem; color: var(--accent-cyan);">' + (app.service || '-') + '</td>' +
            '<td style="font-family: var(--font-mono); font-size: 0.78rem; color: #fff;">' + (app.nginx_domain || '-') + '</td>' +
            '<td><span class="badge-owner">' + app.owner_type.toUpperCase() + '</span></td>';
          tbody.appendChild(tr);
        });
      } catch (e) {
        console.error("Apps fetch error:", e);
      }
    }

    async function fetchServices() {
      try {
        const res = await fetch('/api/v1/services');
        if (!res.ok) return;
        const services = await res.json();

        const tbody = document.getElementById('services-tbody');
        tbody.innerHTML = '';

        services.forEach(function(s) {
          const tr = document.createElement('tr');
          const isActive = s.active_state === 'active';
          const badgeClass = isActive ? 'badge-active' : 'badge-inactive';
          const mem = s.memory_bytes > 0 ? formatBytes(s.memory_bytes) : '-';
          const pid = s.pid > 0 ? s.pid : '-';

          var actionsHtml = '';
          if (isActive) {
            actionsHtml = '<button class="btn-action" onclick="controlService(\'' + s.name + '\', \'restart\')">🔄 Restart</button> ' +
                          '<button class="btn-action" onclick="controlService(\'' + s.name + '\', \'stop\')">⏹️ Stop</button> ';
          } else {
            actionsHtml = '<button class="btn-action" onclick="controlService(\'' + s.name + '\', \'start\')">▶️ Start</button> ';
          }
          actionsHtml += '<button class="btn-action" onclick="viewLogs(\'' + s.name + '\')">📜 Logs</button>';

          tr.innerHTML = '<td>' +
            '<div style="font-weight: 600; color: #fff;">' + s.display_name + '</div>' +
            '<div style="font-family: var(--font-mono); font-size: 0.75rem; color: var(--text-subtle);">' + s.unit_file + '</div>' +
            '</td>' +
            '<td style="color: var(--text-muted); font-size: 0.82rem;">' + (s.description || '-') + '</td>' +
            '<td><span class="badge-status ' + badgeClass + '">' + s.active_state + ' (' + s.sub_state + ')</span></td>' +
            '<td><span class="badge-owner">' + s.owner_type.toUpperCase() + '</span></td>' +
            '<td style="font-family: var(--font-mono); font-size: 0.8rem;">PID: ' + pid + '<br>RAM: ' + mem + '</td>' +
            '<td><div class="actions-group">' + actionsHtml + '</div></td>';

          tbody.appendChild(tr);
        });
      } catch (e) {
        console.error("Services fetch error:", e);
      }
    }

    async function controlService(name, action) {
      if (!confirm('Are you sure you want to ' + action.toUpperCase() + ' ' + name + '?')) return;
      try {
        const res = await fetch('/api/v1/services/' + name + '/action', {
          method: 'POST',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify({action: action})
        });
        const data = await res.json();
        alert(data.message);
        fetchServices();
      } catch (e) {
        alert('Failed to execute action: ' + e);
      }
    }

    async function viewLogs(name) {
      document.getElementById('modal-title').innerText = 'Journal Logs: ' + name + '.service';
      document.getElementById('modal-content').innerText = 'Fetching logs from journalctl...';
      document.getElementById('logs-modal').style.display = 'flex';

      try {
        const res = await fetch('/api/v1/services/' + name + '/logs?lines=60');
        const text = await res.text();
        document.getElementById('modal-content').innerText = text || 'No recent log records found.';
      } catch (e) {
        document.getElementById('modal-content').innerText = 'Error fetching logs: ' + e;
      }
    }

    function closeLogs() {
      document.getElementById('logs-modal').style.display = 'none';
    }

    fetchInfo();
    fetchMetrics();
    fetchServices();
    fetchApplications();
    setInterval(fetchMetrics, 2000);
  </script>
</body>
</html>
`
