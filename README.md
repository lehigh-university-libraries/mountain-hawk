# mountain-hawk

## Prerequisites

- **Docker & Docker Compose**: For running the GitHub MCP server
- **Go 1.24.1+**: For building and running the reviewer application
- **Ollama**: For running the AI model locally
- **GitHub App or PAT**: For GitHub API access

## Usage

```
docker compose run --build --remove-orphans mountain-hawk review \
  --owner lehigh-university-libraries \
  --repo mountain-hawk \
  --pr 7
```

## Setup

### 1. Clone and Build

```bash
git clone https://github.com/lehigh-university-libraries/mountain-hawk
cd mountain-hawk

go mod tidy
go build -o mountain-hawk ./cmd/cli
```

### 2. GitHub MCP Server Setup

Create the required files and directories:

```bash
# Create secrets directory
mkdir -p secrets

# Add your GitHub App private key
# from Lehigh's Mountain Hawk GitHub app
# https://github.com/organizations/lehigh-university-libraries/settings/apps/mountain-hawk
cp /path/to/your/github-app.pem secrets/github.pem

cp .env.example .env
```

Update your `.env` file:
```bash
GITHUB_APP_ID=123456
GITHUB_INSTALLATION_ID=789012
GITHUB_TOKEN=ghp_your_token_here
OLLAMA_HOST=http://localhost:11434
OLLAMA_MODEL=gpt-oss:20b
```

## Usage

### Command Line PR Review

You can review any public PR by passing the repository and PR number:

```bash
docker compose run mountain-hawk review \
  --owner=microsoft \
  --repo=vscode \
  --pr=123456
```

#### Examples

**Review a specific PR:**
```bash
# Review PR #1234 in facebook/react
docker compose run mountain-hawk review --owner=facebook --repo=react --pr=1234

# Review with verbose output
docker compose run mountain-hawk review --owner=golang --repo=go --pr=5678 --verbose

# Review a PR in your own organization
docker compose run mountain-hawk review --owner=myorg --repo=myproject --pr=42
```

## Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GITHUB_TOKEN` | ✅ | - | GitHub personal access token or app token |
| `OLLAMA_HOST` | ❌ | `http://localhost:11434` | Ollama API URL |
| `OLLAMA_MODEL` | ❌ | `gpt-oss:20b` | Ollama model to use |

### GitHub Token Permissions

Your GitHub token needs these permissions:
- **Repository permissions:**
  - Contents: Read
  - Metadata: Read  
  - Pull requests: Write
  - Issues: Read (for labels)

