# Ritmo GitOps Configuration Example

This is an example GitOps configuration repository for Ritmo.

## Structure

```
.
├── jobs/               # Job configurations
├── credentials/        # Credential configurations (encrypted)
├── plugins/            # Plugin configurations
├── webhooks/           # Webhook configurations
└── README.md          # This file
```

## Usage

1. Fork this repository
2. Customize the configurations for your environment
3. Configure Ritmo to use this repository:

```bash
export RITMO_GITOPS_ENABLED=true
export RITMO_GITOPS_REPO_URL=https://github.com/yourorg/ritmo-config
export RITMO_GITOPS_REPO_BRANCH=main
export RITMO_GITOPS_TOKEN=ghp_your_token_here
```

4. Start Ritmo - it will automatically sync configuration from this repository

## Security

- Never commit plain text secrets
- Use encrypted credentials or external secret references
- Review all changes via Pull Requests
- Use branch protection on main branch
