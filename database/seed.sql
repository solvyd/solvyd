-- Sample data for testing and development

-- Insert a test user
INSERT INTO users (username, email, password_hash, full_name, roles, active)
VALUES 
    ('admin', 'admin@ritmo.dev', '$2a$10$dummyhashfordevonly', 'Admin User', ARRAY['admin'], true),
    ('developer', 'dev@ritmo.dev', '$2a$10$dummyhashfordevonly', 'Developer User', ARRAY['developer'], true),
    ('viewer', 'viewer@ritmo.dev', '$2a$10$dummyhashfordevonly', 'Viewer User', ARRAY['viewer'], true);

-- Insert sample plugins
INSERT INTO plugins (name, type, version, description, enabled, binary_path)
VALUES 
    ('git-scm', 'scm', '1.0.0', 'Git source control management plugin', true, '/plugins/git-scm.so'),
    ('github-scm', 'scm', '1.0.0', 'GitHub integration plugin', true, '/plugins/github-scm.so'),
    ('docker-build', 'build', '1.0.0', 'Docker build plugin', true, '/plugins/docker-build.so'),
    ('maven-build', 'build', '1.0.0', 'Maven build plugin', true, '/plugins/maven-build.so'),
    ('npm-build', 'build', '1.0.0', 'NPM build plugin', true, '/plugins/npm-build.so'),
    ('s3-artifacts', 'artifact', '1.0.0', 'AWS S3 artifact storage', true, '/plugins/s3-artifacts.so'),
    ('slack-notification', 'notification', '1.0.0', 'Slack notifications', true, '/plugins/slack-notification.so'),
    ('k8s-deployment', 'deployment', '1.0.0', 'Kubernetes deployment plugin', true, '/plugins/k8s-deployment.so');

-- Insert sample jobs
INSERT INTO jobs (name, description, scm_type, scm_url, scm_branch, build_config, triggers, enabled)
VALUES 
    (
        'backend-service',
        'Backend API service build',
        'github',
        'https://github.com/example/backend-service',
        'main',
        '{"type": "maven", "goals": ["clean", "package"], "jdk_version": "17"}'::jsonb,
        '[{"type": "webhook", "events": ["push"]}, {"type": "cron", "schedule": "0 2 * * *"}]'::jsonb,
        true
    ),
    (
        'frontend-app',
        'Frontend React application',
        'github',
        'https://github.com/example/frontend-app',
        'main',
        '{"type": "npm", "commands": ["install", "build"], "node_version": "18"}'::jsonb,
        '[{"type": "webhook", "events": ["push", "pull_request"]}]'::jsonb,
        true
    ),
    (
        'data-pipeline',
        'ETL data pipeline',
        'git',
        'https://gitlab.com/example/data-pipeline',
        'develop',
        '{"type": "python", "commands": ["pip install -r requirements.txt", "python -m pytest", "python setup.py sdist"]}'::jsonb,
        '[{"type": "cron", "schedule": "0 */6 * * *"}]'::jsonb,
        true
    );

-- Insert sample workers
INSERT INTO workers (name, hostname, ip_address, max_concurrent_builds, labels, status, agent_version, capabilities)
VALUES 
    (
        'worker-01',
        'worker-01.ritmo.local',
        '10.0.1.10',
        4,
        '{"type": "linux", "arch": "amd64", "zone": "us-west-1a"}'::jsonb,
        'online',
        '1.0.0',
        '{"docker": true, "kubernetes": false}'::jsonb
    ),
    (
        'worker-02',
        'worker-02.ritmo.local',
        '10.0.1.11',
        4,
        '{"type": "linux", "arch": "amd64", "zone": "us-west-1b"}'::jsonb,
        'online',
        '1.0.0',
        '{"docker": true, "kubernetes": true}'::jsonb
    ),
    (
        'worker-mac-01',
        'worker-mac-01.ritmo.local',
        '10.0.1.20',
        2,
        '{"type": "darwin", "arch": "arm64", "zone": "us-west-1a"}'::jsonb,
        'online',
        '1.0.0',
        '{"docker": true, "xcode": true}'::jsonb
    );

-- Update worker heartbeats to be recent
UPDATE workers SET last_heartbeat = CURRENT_TIMESTAMP;

-- Insert some sample builds
DO $$
DECLARE
    job_backend UUID;
    job_frontend UUID;
    worker_1 UUID;
    worker_2 UUID;
    build_id UUID;
BEGIN
    SELECT id INTO job_backend FROM jobs WHERE name = 'backend-service';
    SELECT id INTO job_frontend FROM jobs WHERE name = 'frontend-app';
    SELECT id INTO worker_1 FROM workers WHERE name = 'worker-01';
    SELECT id INTO worker_2 FROM workers WHERE name = 'worker-02';
    
    -- Successful backend build
    INSERT INTO builds (job_id, status, queued_at, started_at, completed_at, worker_id, scm_commit_sha, scm_commit_message, scm_author, branch, triggered_by, exit_code)
    VALUES (
        job_backend, 'success', 
        CURRENT_TIMESTAMP - INTERVAL '2 hours',
        CURRENT_TIMESTAMP - INTERVAL '2 hours' + INTERVAL '30 seconds',
        CURRENT_TIMESTAMP - INTERVAL '2 hours' + INTERVAL '5 minutes',
        worker_1,
        'abc123def456',
        'Fix: Resolve database connection pool issue',
        'John Doe',
        'main',
        'webhook',
        0
    ) RETURNING id INTO build_id;
    
    -- Add some logs for this build
    INSERT INTO build_logs (build_id, sequence_number, log_line, stream)
    VALUES 
        (build_id, 1, '[INFO] Scanning for projects...', 'stdout'),
        (build_id, 2, '[INFO] Building backend-service 1.0.0', 'stdout'),
        (build_id, 3, '[INFO] Compiling source files...', 'stdout'),
        (build_id, 4, '[INFO] Running tests...', 'stdout'),
        (build_id, 5, '[INFO] Tests passed: 145/145', 'stdout'),
        (build_id, 6, '[INFO] Building JAR...', 'stdout'),
        (build_id, 7, '[INFO] BUILD SUCCESS', 'stdout');
    
    -- Add artifact for this build
    INSERT INTO artifacts (build_id, name, path, size_bytes, checksum_sha256, storage_plugin, storage_url, promotion_status)
    VALUES (
        build_id,
        'backend-service-1.0.0.jar',
        'builds/' || build_id::text || '/backend-service-1.0.0.jar',
        15728640,
        'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
        's3-artifacts',
        's3://ritmo-artifacts/builds/' || build_id::text || '/backend-service-1.0.0.jar',
        'staging'
    );
    
    -- Failed frontend build
    INSERT INTO builds (job_id, status, queued_at, started_at, completed_at, worker_id, scm_commit_sha, scm_commit_message, scm_author, branch, triggered_by, exit_code, error_message)
    VALUES (
        job_frontend, 'failed',
        CURRENT_TIMESTAMP - INTERVAL '1 hour',
        CURRENT_TIMESTAMP - INTERVAL '1 hour' + INTERVAL '15 seconds',
        CURRENT_TIMESTAMP - INTERVAL '1 hour' + INTERVAL '3 minutes',
        worker_2,
        'xyz789abc012',
        'Update: Upgrade React to v18',
        'Jane Smith',
        'feature/react-18',
        'webhook',
        1,
        'Test suite failed: 3 tests failed'
    );
    
    -- Running build
    INSERT INTO builds (job_id, status, queued_at, started_at, worker_id, scm_commit_sha, scm_commit_message, scm_author, branch, triggered_by)
    VALUES (
        job_backend, 'running',
        CURRENT_TIMESTAMP - INTERVAL '5 minutes',
        CURRENT_TIMESTAMP - INTERVAL '4 minutes',
        worker_1,
        'def456ghi789',
        'Feature: Add new API endpoints',
        'Bob Johnson',
        'main',
        'manual'
    );
    
    -- Queued build
    INSERT INTO builds (job_id, status, queued_at, scm_commit_sha, scm_commit_message, scm_author, branch, triggered_by)
    VALUES (
        job_frontend, 'queued',
        CURRENT_TIMESTAMP - INTERVAL '1 minute',
        'mno345pqr678',
        'Fix: Correct TypeScript types',
        'Alice Williams',
        'main',
        'webhook'
    );
END $$;

-- Insert sample credentials (encrypted with dummy data)
INSERT INTO credentials (name, type, encrypted_data, description, created_by)
VALUES 
    ('github-pat', 'token', '\xdeadbeef'::bytea, 'GitHub personal access token', 'admin'),
    ('docker-registry', 'username_password', '\xdeadbeef'::bytea, 'Docker registry credentials', 'admin'),
    ('aws-s3', 'token', '\xdeadbeef'::bytea, 'AWS S3 access credentials', 'admin');

-- Insert sample audit logs
INSERT INTO audit_logs (user_id, action, resource_type, resource_id, details, ip_address)
SELECT 
    u.id,
    'create_job',
    'job',
    j.id,
    '{"job_name": "' || j.name || '"}'::jsonb,
    '192.168.1.100'
FROM users u, jobs j
WHERE u.username = 'admin'
LIMIT 3;

-- Insert sample metrics
INSERT INTO metrics (metric_name, metric_type, value, labels, timestamp)
VALUES 
    ('build_queue_depth', 'gauge', 5, '{"status": "queued"}'::jsonb, CURRENT_TIMESTAMP),
    ('build_duration_seconds', 'histogram', 245, '{"job": "backend-service", "status": "success"}'::jsonb, CURRENT_TIMESTAMP - INTERVAL '30 minutes'),
    ('worker_utilization', 'gauge', 0.75, '{"worker": "worker-01"}'::jsonb, CURRENT_TIMESTAMP),
    ('builds_total', 'counter', 1523, '{"status": "success"}'::jsonb, CURRENT_TIMESTAMP),
    ('builds_total', 'counter', 187, '{"status": "failed"}'::jsonb, CURRENT_TIMESTAMP);

COMMENT ON TABLE users IS 'Sample users: admin/admin, developer/developer, viewer/viewer (passwords are "password" - NOT FOR PRODUCTION)';
