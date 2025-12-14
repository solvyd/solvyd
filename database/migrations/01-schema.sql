-- Ritmo Database Schema
-- PostgreSQL 15+

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Jobs table: Stores job configurations
CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    
    -- SCM configuration
    scm_type VARCHAR(50), -- git, github, gitlab, etc.
    scm_url TEXT,
    scm_branch VARCHAR(255) DEFAULT 'main',
    scm_credentials_id UUID,
    
    -- Build configuration
    build_config JSONB NOT NULL, -- Flexible config for different build types
    environment_vars JSONB DEFAULT '{}'::jsonb,
    
    -- Scheduling
    triggers JSONB DEFAULT '[]'::jsonb, -- cron, webhook, manual
    enabled BOOLEAN DEFAULT true,
    
    -- Worker targeting
    worker_labels JSONB DEFAULT '{}'::jsonb,
    
    -- Plugin references
    plugins JSONB DEFAULT '[]'::jsonb,
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    
    -- Pipeline stages (for complex pipelines)
    pipeline_stages JSONB DEFAULT '[]'::jsonb,
    
    -- Timeout and retry
    timeout_minutes INTEGER DEFAULT 60,
    max_retries INTEGER DEFAULT 0
);

CREATE INDEX idx_jobs_name ON jobs(name);
CREATE INDEX idx_jobs_enabled ON jobs(enabled);
CREATE INDEX idx_jobs_created_at ON jobs(created_at DESC);

-- Builds table: Stores individual build executions
CREATE TABLE builds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    build_number SERIAL,
    
    -- Build status
    status VARCHAR(50) NOT NULL, -- queued, running, success, failed, cancelled, timeout
    
    -- Timing
    queued_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_seconds INTEGER,
    
    -- Worker assignment
    worker_id UUID,
    
    -- Build context
    scm_commit_sha VARCHAR(255),
    scm_commit_message TEXT,
    scm_author VARCHAR(255),
    branch VARCHAR(255),
    
    -- Build parameters
    parameters JSONB DEFAULT '{}'::jsonb,
    environment_vars JSONB DEFAULT '{}'::jsonb,
    
    -- Trigger info
    triggered_by VARCHAR(255), -- user, webhook, schedule, manual
    trigger_metadata JSONB DEFAULT '{}'::jsonb,
    
    -- Results
    exit_code INTEGER,
    error_message TEXT,
    
    -- Logs reference
    log_url TEXT,
    
    -- Artifacts
    artifact_count INTEGER DEFAULT 0,
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(job_id, build_number)
);

CREATE INDEX idx_builds_job_id ON builds(job_id);
CREATE INDEX idx_builds_status ON builds(status);
CREATE INDEX idx_builds_queued_at ON builds(queued_at DESC);
CREATE INDEX idx_builds_started_at ON builds(started_at DESC);
CREATE INDEX idx_builds_worker_id ON builds(worker_id);
CREATE INDEX idx_builds_scm_commit ON builds(scm_commit_sha);

-- Workers table: Stores worker node information
CREATE TABLE workers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    hostname VARCHAR(255),
    ip_address INET,
    
    -- Capacity
    max_concurrent_builds INTEGER DEFAULT 1,
    current_builds INTEGER DEFAULT 0,
    cpu_cores INTEGER,
    memory_mb INTEGER,
    
    -- Labels for targeting
    labels JSONB DEFAULT '{}'::jsonb,
    
    -- Status
    status VARCHAR(50) NOT NULL, -- online, offline, draining, maintenance
    
    -- Health
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    health_status VARCHAR(50) DEFAULT 'healthy', -- healthy, degraded, unhealthy
    
    -- Version
    agent_version VARCHAR(50),
    
    -- Metadata
    registered_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Capabilities
    capabilities JSONB DEFAULT '{}'::jsonb -- docker, kubernetes, vm, etc.
);

CREATE INDEX idx_workers_status ON workers(status);
CREATE INDEX idx_workers_last_heartbeat ON workers(last_heartbeat DESC);
CREATE INDEX idx_workers_labels ON workers USING gin(labels);

-- Artifacts table: Stores artifact metadata
CREATE TABLE artifacts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    build_id UUID NOT NULL REFERENCES builds(id) ON DELETE CASCADE,
    
    -- Artifact info
    name VARCHAR(255) NOT NULL,
    path TEXT NOT NULL,
    size_bytes BIGINT,
    checksum_sha256 VARCHAR(64),
    content_type VARCHAR(255),
    
    -- Storage
    storage_plugin VARCHAR(100), -- s3, gcs, artifactory, etc.
    storage_url TEXT NOT NULL,
    storage_metadata JSONB DEFAULT '{}'::jsonb,
    
    -- Promotion
    promotion_status VARCHAR(50) DEFAULT 'dev', -- dev, staging, prod
    promoted_at TIMESTAMP WITH TIME ZONE,
    promoted_by VARCHAR(255),
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_artifacts_build_id ON artifacts(build_id);
CREATE INDEX idx_artifacts_name ON artifacts(name);
CREATE INDEX idx_artifacts_promotion_status ON artifacts(promotion_status);

-- Deployments table: Stores CD deployment records
CREATE TABLE deployments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    build_id UUID NOT NULL REFERENCES builds(id),
    artifact_id UUID REFERENCES artifacts(id),
    
    -- Deployment info
    environment VARCHAR(100) NOT NULL, -- dev, staging, production
    status VARCHAR(50) NOT NULL, -- pending, in_progress, success, failed, rolled_back
    
    -- Deployment target
    target_type VARCHAR(100), -- kubernetes, docker, ssh, argocd, etc.
    target_url TEXT,
    target_metadata JSONB DEFAULT '{}'::jsonb,
    
    -- Timing
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_seconds INTEGER,
    
    -- Deployment plugin
    deployment_plugin VARCHAR(100),
    
    -- Results
    exit_code INTEGER,
    error_message TEXT,
    deployment_url TEXT, -- URL to deployed application
    
    -- Rollback info
    rollback_from_deployment_id UUID REFERENCES deployments(id),
    
    -- Metadata
    deployed_by VARCHAR(255),
    deployment_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_deployments_build_id ON deployments(build_id);
CREATE INDEX idx_deployments_environment ON deployments(environment);
CREATE INDEX idx_deployments_status ON deployments(status);
CREATE INDEX idx_deployments_started_at ON deployments(started_at DESC);

-- Build logs table: Stores build log chunks
CREATE TABLE build_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    build_id UUID NOT NULL REFERENCES builds(id) ON DELETE CASCADE,
    sequence_number INTEGER NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    log_line TEXT NOT NULL,
    stream VARCHAR(10) DEFAULT 'stdout', -- stdout, stderr
    
    UNIQUE(build_id, sequence_number)
);

CREATE INDEX idx_build_logs_build_id ON build_logs(build_id, sequence_number);

-- Credentials table: Stores encrypted credentials
CREATE TABLE credentials (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL, -- ssh_key, username_password, token, certificate
    
    -- Encrypted credential data
    encrypted_data BYTEA NOT NULL,
    encryption_key_id VARCHAR(255), -- Reference to key management system
    
    -- Metadata
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    
    -- Usage tracking
    last_used_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_credentials_name ON credentials(name);

-- Plugins table: Stores installed plugins
CREATE TABLE plugins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    type VARCHAR(100) NOT NULL, -- scm, build, artifact, notification, deployment
    version VARCHAR(50) NOT NULL,
    
    -- Plugin binary
    binary_path TEXT,
    binary_checksum VARCHAR(64),
    
    -- Plugin metadata
    description TEXT,
    author VARCHAR(255),
    homepage_url TEXT,
    
    -- Configuration schema
    config_schema JSONB DEFAULT '{}'::jsonb,
    
    -- Status
    enabled BOOLEAN DEFAULT true,
    
    -- Metadata
    installed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_plugins_type ON plugins(type);
CREATE INDEX idx_plugins_enabled ON plugins(enabled);

-- Pipeline stages table: For complex multi-stage pipelines
CREATE TABLE pipeline_stages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    build_id UUID NOT NULL REFERENCES builds(id) ON DELETE CASCADE,
    stage_name VARCHAR(255) NOT NULL,
    stage_order INTEGER NOT NULL,
    
    -- Status
    status VARCHAR(50) NOT NULL, -- pending, running, success, failed, skipped
    
    -- Timing
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_seconds INTEGER,
    
    -- Results
    exit_code INTEGER,
    error_message TEXT,
    
    -- Dependencies
    depends_on UUID[], -- Array of stage IDs
    
    UNIQUE(build_id, stage_name)
);

CREATE INDEX idx_pipeline_stages_build_id ON pipeline_stages(build_id, stage_order);

-- Webhooks table: Stores webhook configurations
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    
    -- Webhook config
    source VARCHAR(50) NOT NULL, -- github, gitlab, bitbucket, custom
    secret_token VARCHAR(255),
    
    -- Events
    events VARCHAR(50)[], -- push, pull_request, tag, etc.
    
    -- Filters
    branch_filter VARCHAR(255), -- regex for branch matching
    
    -- Status
    enabled BOOLEAN DEFAULT true,
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_triggered_at TIMESTAMP WITH TIME ZONE,
    trigger_count INTEGER DEFAULT 0
);

CREATE INDEX idx_webhooks_job_id ON webhooks(job_id);

-- Users table: For authentication and authorization
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    
    -- Authentication
    password_hash VARCHAR(255), -- For local auth
    oauth_provider VARCHAR(50), -- github, google, etc.
    oauth_id VARCHAR(255),
    
    -- Profile
    full_name VARCHAR(255),
    avatar_url TEXT,
    
    -- Status
    active BOOLEAN DEFAULT true,
    
    -- Roles
    roles VARCHAR(50)[], -- admin, developer, viewer
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);

-- Audit log table: Track all actions
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL, -- create_job, delete_build, trigger_deployment, etc.
    resource_type VARCHAR(50), -- job, build, worker, etc.
    resource_id UUID,
    
    -- Details
    details JSONB DEFAULT '{}'::jsonb,
    ip_address INET,
    user_agent TEXT,
    
    -- Metadata
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);

-- Metrics table: Store aggregated metrics
CREATE TABLE metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    metric_name VARCHAR(100) NOT NULL,
    metric_type VARCHAR(50) NOT NULL, -- counter, gauge, histogram
    
    -- Values
    value DOUBLE PRECISION,
    labels JSONB DEFAULT '{}'::jsonb,
    
    -- Timestamp
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_metrics_name_timestamp ON metrics(metric_name, timestamp DESC);
CREATE INDEX idx_metrics_labels ON metrics USING gin(labels);

-- Functions and triggers

-- Update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_jobs_updated_at BEFORE UPDATE ON jobs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_workers_updated_at BEFORE UPDATE ON workers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_credentials_updated_at BEFORE UPDATE ON credentials
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_plugins_updated_at BEFORE UPDATE ON plugins
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Calculate build duration on completion
CREATE OR REPLACE FUNCTION calculate_build_duration()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.completed_at IS NOT NULL AND NEW.started_at IS NOT NULL THEN
        NEW.duration_seconds = EXTRACT(EPOCH FROM (NEW.completed_at - NEW.started_at));
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER calculate_build_duration_trigger BEFORE UPDATE ON builds
    FOR EACH ROW EXECUTE FUNCTION calculate_build_duration();

-- Calculate deployment duration on completion
CREATE OR REPLACE FUNCTION calculate_deployment_duration()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.completed_at IS NOT NULL AND NEW.started_at IS NOT NULL THEN
        NEW.duration_seconds = EXTRACT(EPOCH FROM (NEW.completed_at - NEW.started_at));
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER calculate_deployment_duration_trigger BEFORE UPDATE ON deployments
    FOR EACH ROW EXECUTE FUNCTION calculate_deployment_duration();

-- Views for analytics

-- Build statistics by job
CREATE VIEW build_statistics AS
SELECT 
    j.id as job_id,
    j.name as job_name,
    COUNT(*) as total_builds,
    SUM(CASE WHEN b.status = 'success' THEN 1 ELSE 0 END) as successful_builds,
    SUM(CASE WHEN b.status = 'failed' THEN 1 ELSE 0 END) as failed_builds,
    ROUND(AVG(b.duration_seconds)) as avg_duration_seconds,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY b.duration_seconds) as p50_duration_seconds,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY b.duration_seconds) as p95_duration_seconds
FROM jobs j
LEFT JOIN builds b ON j.id = b.job_id
WHERE b.completed_at IS NOT NULL
GROUP BY j.id, j.name;

-- Worker utilization
CREATE VIEW worker_utilization AS
SELECT 
    w.id as worker_id,
    w.name as worker_name,
    w.max_concurrent_builds,
    w.current_builds,
    ROUND((w.current_builds::numeric / NULLIF(w.max_concurrent_builds, 0)) * 100, 2) as utilization_percent,
    w.status,
    w.last_heartbeat
FROM workers w;

-- Recent build activity
CREATE VIEW recent_builds AS
SELECT 
    b.id,
    b.build_number,
    j.name as job_name,
    b.status,
    b.started_at,
    b.duration_seconds,
    w.name as worker_name,
    b.scm_commit_sha,
    b.triggered_by
FROM builds b
JOIN jobs j ON b.job_id = j.id
LEFT JOIN workers w ON b.worker_id = w.id
ORDER BY b.queued_at DESC
LIMIT 100;

COMMENT ON TABLE jobs IS 'Stores CI/CD job configurations';
COMMENT ON TABLE builds IS 'Stores individual build execution records';
COMMENT ON TABLE workers IS 'Stores worker node registrations and status';
COMMENT ON TABLE artifacts IS 'Stores build artifact metadata and storage locations';
COMMENT ON TABLE deployments IS 'Stores CD deployment records';
