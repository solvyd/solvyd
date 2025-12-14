package sdk

// Plugin is the base interface all plugins must implement
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// Version returns the plugin version
	Version() string

	// Type returns the plugin type (scm, build, artifact, notification, deployment)
	Type() string

	// Initialize initializes the plugin with configuration
	Initialize(config map[string]interface{}) error

	// Execute executes the plugin
	Execute(context *ExecutionContext) (*Result, error)

	// Cleanup performs cleanup after execution
	Cleanup() error
}

// ExecutionContext provides context for plugin execution
type ExecutionContext struct {
	BuildID    string
	JobID      string
	WorkDir    string
	EnvVars    map[string]string
	Parameters map[string]interface{}
	Secrets    map[string]string
	Logger     Logger
}

// Result contains the result of plugin execution
type Result struct {
	Success      bool
	ExitCode     int
	ErrorMessage string
	Output       string
	Artifacts    []Artifact
	Metadata     map[string]interface{}
}

// Artifact represents a build artifact
type Artifact struct {
	Name           string
	Path           string
	SizeBytes      int64
	ChecksumSHA256 string
	Metadata       map[string]string
}

// Logger interface for plugin logging
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

// SCMPlugin interface for source control plugins
type SCMPlugin interface {
	Plugin
	Clone(url, branch, commitSHA string, dest string) error
	GetCommitInfo(commitSHA string) (*CommitInfo, error)
}

// CommitInfo contains commit metadata
type CommitInfo struct {
	SHA       string
	Message   string
	Author    string
	Email     string
	Timestamp string
}

// BuildPlugin interface for build tool plugins
type BuildPlugin interface {
	Plugin
	Build() error
	Test() error
}

// ArtifactPlugin interface for artifact storage plugins
type ArtifactPlugin interface {
	Plugin
	Upload(artifact *Artifact) (string, error)
	Download(url string, dest string) error
	Promote(artifactID, fromEnv, toEnv string) error
}

// NotificationPlugin interface for notification plugins
type NotificationPlugin interface {
	Plugin
	Notify(message *NotificationMessage) error
}

// NotificationMessage contains notification details
type NotificationMessage struct {
	Title       string
	Body        string
	Level       string // info, success, warning, error
	BuildID     string
	JobName     string
	Status      string
	URL         string
	Attachments []Attachment
}

// Attachment for notification messages
type Attachment struct {
	Title  string
	Text   string
	Color  string
	Fields []Field
}

// Field for attachment
type Field struct {
	Title string
	Value string
	Short bool
}

// DeploymentPlugin interface for deployment plugins
type DeploymentPlugin interface {
	Plugin
	Deploy(deployment *DeploymentRequest) (*DeploymentResult, error)
	Rollback(deploymentID string) error
	GetStatus(deploymentID string) (*DeploymentStatus, error)
}

// DeploymentRequest contains deployment details
type DeploymentRequest struct {
	Environment string
	ArtifactURL string
	TargetURL   string
	Config      map[string]interface{}
	Secrets     map[string]string
}

// DeploymentResult contains deployment result
type DeploymentResult struct {
	DeploymentID  string
	Status        string
	DeploymentURL string
	Metadata      map[string]interface{}
}

// DeploymentStatus contains deployment status
type DeploymentStatus struct {
	DeploymentID string
	Status       string
	Message      string
	UpdatedAt    string
}
