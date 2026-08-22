export interface SystemInfo {
  hostname: string;
  os: string;
  platform: string;
  kernel: string;
  architecture: string;
  cpu_model: string;
  cpu_cores: number;
  total_memory: number;
  total_disk: number;
  uptime_seconds: number;
  boot_time: string;
  go_version: string;
  panel_version: string;
}

export interface LiveMetrics {
  timestamp: string;
  cpu: {
    usage_percent: number;
    cores: number;
  };
  memory: {
    total_bytes: number;
    used_bytes: number;
    free_bytes: number;
    available_bytes: number;
    usage_percent: number;
  };
  swap: {
    total_bytes: number;
    used_bytes: number;
    free_bytes: number;
    usage_percent: number;
  };
  disk: {
    total_bytes: number;
    used_bytes: number;
    free_bytes: number;
    usage_percent: number;
    path: string;
  };
  load_average: [number, number, number];
  processes: number;
}

export interface SystemService {
  name: string;
  display_name: string;
  description: string;
  unit_file: string;
  active_state: string;
  sub_state: string;
  load_state: string;
  enabled: boolean;
  owner_type: 'aio' | 'external';
  pid: number;
  memory_bytes: number;
  uptime: string;
}

export interface DiscoveredApp {
  name: string;
  type: string;
  path: string;
  runtime: string;
  service: string;
  nginx_domain: string;
  owner_type: 'external' | 'aio';
  status: string;
}

export interface NginxSite {
  domain: string;
  aliases?: string[];
  config_file: string;
  document_root?: string;
  proxy_pass?: string;
  ssl: boolean;
  owner_type: 'external' | 'aio';
  enabled: boolean;
}

export interface SSLCertificate {
  domain: string;
  issuer: string;
  valid_from: string;
  valid_to: string;
  days_remaining: number;
  auto_renew: boolean;
}

export interface SSHConfig {
  port: number;
  permit_root_login: string;
  password_authentication: boolean;
  pubkey_authentication: boolean;
  config_path: string;
}

export interface SSHSession {
  user: string;
  terminal: string;
  host: string;
  login_time: string;
}

export interface FirewallRule {
  id: number;
  to_port: string;
  protocol: string;
  action: string;
  from_ip: string;
  comment?: string;
}

export interface FirewallStatus {
  active: boolean;
  default_incoming: string;
  default_outgoing: string;
  rules: FirewallRule[];
}

export interface LinuxUser {
  username: string;
  uid: number;
  gid: number;
  home_dir: string;
  shell: string;
  is_sudo: boolean;
  has_ssh_key: boolean;
  is_system: boolean;
}

export interface PostgresDB {
  name: string;
  owner: string;
  encoding: string;
  size_bytes: number;
  size_human: string;
}

export interface MySQLDB {
  name: string;
  size_human: string;
}

export interface DockerContainer {
  id: string;
  names: string;
  image: string;
  status: string;
  state: string;
  ports: string;
  created: string;
}

export interface CronJob {
  id: number;
  schedule: string;
  command: string;
  user: string;
}

export interface AuditEvent {
  id: number;
  timestamp: string;
  user: string;
  action: string;
  target: string;
  result: string;
  details?: string;
  ip_address?: string;
}

export interface FileItem {
  name: string;
  path: string;
  is_dir: boolean;
  size_bytes: number;
  size_human: string;
  permissions: string;
  modified_time: string;
  owner?: string;
}

export interface DockerImage {
  id: string;
  repository: string;
  tag: string;
  size: string;
  created: string;
}

export interface BackupItem {
  name: string;
  path: string;
  size_bytes: number;
  size_human: string;
  created_at: string;
  type: 'postgres' | 'mysql' | 'config' | 'app' | 'file';
}

export interface TerminalExecResult {
  command: string;
  stdout: string;
  stderr: string;
  exit_code: number;
  duration: string;
  timestamp: string;
}

export interface LogResponse {
  source: string;
  lines: number;
  content: string;
}

export interface PostgresUser {
  role_name: string;
  is_superuser: boolean;
  can_login: boolean;
}

export interface StorePackage {
  id: string;
  name: string;
  category: string;
  description: string;
  icon: string;
  installed: boolean;
  version: string;
  versions: string[];
  install_cmd: string;
}

export interface InstallJob {
  package_id: string;
  status: 'RUNNING' | 'SUCCESS' | 'FAILED';
  output: string;
  started_at: string;
  ended_at: string;
  error?: string;
}

export interface AuthStatus {
  authenticated: boolean;
  username?: string;
  role?: string;
  setup_required: boolean;
}

