export interface Job {
  id: string
  name: string
  description: string
  scm_type: string
  scm_url: string
  scm_branch: string
  build_config: Record<string, any>
  environment_vars: Record<string, any>
  triggers: any[]
  enabled: boolean
  worker_labels: Record<string, any>
  plugins: any[]
  pipeline_stages: any[]
  timeout_minutes: number
  max_retries: number
  created_at: string
  updated_at: string
  created_by: string
}

export interface Build {
  id: string
  job_id: string
  job_name?: string
  build_number: number
  status: 'queued' | 'running' | 'success' | 'failed' | 'cancelled' | 'timeout'
  queued_at: string
  started_at?: string
  completed_at?: string
  duration_seconds?: number
  worker_id?: string
  scm_commit_sha: string
  scm_commit_message: string
  scm_author: string
  branch: string
  parameters: Record<string, any>
  environment_vars: Record<string, any>
  triggered_by: string
  trigger_metadata: Record<string, any>
  exit_code?: number
  error_message?: string
  log_url?: string
  artifact_count: number
}

export interface Worker {
  id: string
  name: string
  hostname: string
  ip_address: string
  max_concurrent_builds: number
  current_builds: number
  cpu_cores: number
  memory_mb: number
  labels: Record<string, any>
  capabilities: Record<string, any>
  status: 'online' | 'offline' | 'draining' | 'maintenance'
  last_heartbeat: string
  health_status: string
  agent_version: string
  registered_at: string
  updated_at: string
}

export interface Deployment {
  id: string
  build_id: string
  artifact_id?: string
  environment: string
  status: 'pending' | 'in_progress' | 'success' | 'failed' | 'rolled_back'
  target_type: string
  target_url?: string
  target_metadata: Record<string, any>
  started_at: string
  completed_at?: string
  duration_seconds?: number
  deployment_plugin: string
  exit_code?: number
  error_message?: string
  deployment_url?: string
  rollback_from_deployment_id?: string
  deployed_by: string
  deployment_notes?: string
  created_at: string
}

export interface Plugin {
  id: string
  name: string
  type: string
  version: string
  description: string
  author: string
  homepage_url: string
  config_schema: Record<string, any>
  enabled: boolean
  installed_at: string
  updated_at: string
}

export interface BuildLog {
  id: string
  build_id: string
  sequence_number: number
  timestamp: string
  log_line: string
  stream: 'stdout' | 'stderr'
}

export interface Artifact {
  id: string
  build_id: string
  name: string
  path: string
  size_bytes: number
  checksum_sha256: string
  content_type: string
  storage_plugin: string
  storage_url: string
  storage_metadata: Record<string, any>
  promotion_status: string
  promoted_at?: string
  promoted_by?: string
  created_at: string
  metadata: Record<string, any>
}
