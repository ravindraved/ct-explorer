# CT Explorer

AWS Control Tower Explorer — a single-binary Go application with an embedded React frontend that provides a visual interface for browsing your AWS Organization structure, Control Tower controls, Service Control Policies, the Control Catalog ontology, and security posture.

## Prerequisites

- **Go 1.25+** — [install guide](https://go.dev/doc/install)
- **Node.js 20+** and **npm** — only needed if you want to modify the frontend
- **AWS credentials** — via EC2 instance profile (recommended) or CLI profile
- **EC2 instance** — the app runs on port 8000 and should be accessed via SSH tunnel, not exposed to the internet

## IAM Permissions

The EC2 instance profile (or IAM user/role) needs the following permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "OrganizationsReadOnly",
      "Effect": "Allow",
      "Action": [
        "organizations:ListRoots",
        "organizations:ListOrganizationalUnitsForParent",
        "organizations:ListAccountsForParent",
        "organizations:DescribeOrganizationalUnit",
        "organizations:DescribeAccount",
        "organizations:ListPolicies",
        "organizations:DescribePolicy",
        "organizations:ListTargetsForPolicy",
        "organizations:ListPoliciesForTarget"
      ],
      "Resource": "*"
    },
    {
      "Sid": "ControlTowerReadOnly",
      "Effect": "Allow",
      "Action": [
        "controltower:ListEnabledControls",
        "controltower:GetEnabledControl",
        "controltower:ListLandingZones",
        "controltower:GetLandingZone"
      ],
      "Resource": "*"
    },
    {
      "Sid": "ControlCatalogReadOnly",
      "Effect": "Allow",
      "Action": [
        "controlcatalog:ListControls",
        "controlcatalog:GetControl",
        "controlcatalog:ListDomains",
        "controlcatalog:ListObjectives"
      ],
      "Resource": "*"
    },
    {
      "Sid": "STSCallerIdentity",
      "Effect": "Allow",
      "Action": "sts:GetCallerIdentity",
      "Resource": "*"
    }
  ]
}
```

All permissions are read-only. The app never modifies your AWS resources.

## Quick Start (EC2)

### Option A: Pre-built Binary (no build tools needed)

A pre-built binary for Amazon Linux (x86_64) is included in the repo. No Go, Node.js, or npm required.

```bash
# Clone the repo
git clone https://github.com/ravindraved/aws-ct-controls-mapping.git
cd aws-ct-controls-mapping

# Run directly
CT_AUTH_ENABLED=false ./bin/ct-explorer
```

The binary is statically linked and includes the embedded React frontend — just download and run.

### Option B: Build from Source

If you want to modify the code or rebuild the binary yourself, you'll need Go 1.25+ and Node.js 20+.

#### 1. Build

```bash
# Clone the repo
git clone https://github.com/ravindraved/aws-ct-controls-mapping.git
cd aws-ct-controls-mapping

# Build the frontend
cd web
npm ci
npm run build
cd ..

# Copy frontend build into Go embed directory
cp -r web/dist/* go_api_server/internal/static/web_dist/

# Build the Go binary
cd go_api_server
make build
```

The binary is at `go_api_server/bin/ct-explorer-go`.

#### 2. Run

```bash
# Auth disabled (local/EC2 use — no Cognito needed)
CT_AUTH_ENABLED=false ./go_api_server/bin/ct-explorer-go
```

The app starts on port 8000 by default.

### 3. Access via SSH tunnel

Do NOT open port 8000 in your EC2 security group. Instead, use an SSH tunnel to access the app securely from your local machine:

```bash
# From your local machine
ssh -L 8000:localhost:8000 ec2-user@<your-ec2-ip>
```

Then open `http://localhost:8000` in your browser. The tunnel encrypts all traffic and avoids exposing the app to the internet.

If port 8000 is already in use locally, pick a different local port:

```bash
ssh -L 9000:localhost:8000 ec2-user@<your-ec2-ip>
# Then open http://localhost:9000
```

## Environment Variables

| Variable | Description | Default |
|---|---|---|
| `PORT` | HTTP listen port | `8000` |
| `LOG_LEVEL` | Log level (DEBUG, INFO, WARN, ERROR) | `INFO` |
| `CT_AUTH_ENABLED` | Enable Cognito auth (`true`/`false`) | `true` |
| `CT_COGNITO_DOMAIN` | Cognito Hosted UI URL | (empty) |
| `CT_COGNITO_CLIENT_ID` | Cognito app client ID | (empty) |
| `CT_COGNITO_POOL_ID` | Cognito User Pool ID | (empty) |
| `CT_COGNITO_REGION` | Cognito pool region | `ap-south-1` |
| `CT_CACHE_DIR` | Directory for cache files | OS temp dir |

For EC2 local use, only `CT_AUTH_ENABLED=false` is needed. The Cognito variables are only required for Fargate/CloudFront deployments with authentication enabled.

## Security Notes

- The app binds to `0.0.0.0:8000` by default — always use SSH tunneling, never expose the port directly
- When `CT_AUTH_ENABLED=false`, there is no authentication — anyone who can reach port 8000 has full access to the read-only data
- All AWS API calls are read-only; the app cannot modify your organization or controls
- No AWS credentials are stored or logged by the app

## Project Structure

```
ct-explorer/
├── go_api_server/          # Go backend
│   ├── cmd/server/         # Entry point
│   ├── internal/
│   │   ├── api/            # HTTP handlers, middleware, router
│   │   ├── fetchers/       # AWS API calls
│   │   ├── models/         # Domain structs
│   │   ├── store/          # In-memory data store + cache
│   │   ├── search/         # Search engine
│   │   ├── posture/        # Ontology posture aggregation
│   │   ├── logging/        # Structured logging + WebSocket broadcast
│   │   └── static/         # Embedded frontend (via Go embed)
│   ├── go.mod
│   └── Makefile
└── web/                    # React frontend (Vite + Tailwind)
    ├── src/
    │   ├── features/       # View components (org, controls, catalog, etc.)
    │   ├── hooks/           # React hooks for API consumption
    │   ├── components/      # Shared UI components
    │   └── api/             # API client
    └── package.json
```

## Cognito Authentication

CT Explorer supports optional authentication via Amazon Cognito User Pools. This is designed for production deployments (e.g., ECS Fargate behind an ALB/CloudFront) where you need to restrict access to authorized users.

### How It Works

The app uses the **OAuth 2.0 Authorization Code flow with PKCE** (Proof Key for Code Exchange), which is the recommended approach for single-page applications:

1. The React SPA fetches `/api/auth/config` to check if auth is enabled
2. If enabled, the SPA generates a PKCE code verifier/challenge pair
3. The user is redirected to the Cognito Hosted UI to log in
4. After login, Cognito redirects back with an authorization code
5. The SPA exchanges the code + PKCE verifier for JWT tokens via Cognito's `/oauth2/token` endpoint
6. The `id_token` (JWT) is stored in `sessionStorage` and sent as a `Bearer` token on every API request
7. The Go backend validates the JWT on each request — checking signature (via JWKS), expiry, and issuer

### Auth Modes

| Mode | When | How |
|---|---|---|
| **Disabled** (`CT_AUTH_ENABLED=false`) | Local / EC2 use | No login required. All API endpoints are open. |
| **Enabled** (`CT_AUTH_ENABLED=true`) | ECS / Fargate / production | Cognito login required. JWT validated on every `/api/*` request. |

When auth is disabled, the JWT middleware is a no-op and the SPA skips the login flow entirely.

### Cognito Setup

To enable authentication, you need a Cognito User Pool with:

1. **User Pool** — create one in the AWS Console or via IaC (CloudFormation/CDK/Terraform)
2. **App Client** — create an app client with:
   - No client secret (required for PKCE with public SPA clients)
   - Allowed OAuth flows: Authorization code grant
   - Allowed OAuth scopes: `openid`, `email`, `profile`
   - Callback URL: `https://<your-domain>/` (your CloudFront or ALB URL)
   - Sign-out URL: `https://<your-domain>/`
3. **Hosted UI domain** — configure a Cognito domain (e.g., `https://your-app.auth.ap-south-1.amazoncognito.com`)

Then set the environment variables on your ECS task:

```bash
CT_AUTH_ENABLED=true
CT_COGNITO_DOMAIN=https://your-app.auth.ap-south-1.amazoncognito.com
CT_COGNITO_CLIENT_ID=your-cognito-app-client-id
CT_COGNITO_POOL_ID=ap-south-1_XXXXXXXXX
CT_COGNITO_REGION=ap-south-1
```

### Endpoints Excluded from Auth

These endpoints are accessible without a JWT even when auth is enabled:

| Endpoint | Reason |
|---|---|
| `GET /api/health` | ALB health checks |
| `GET /api/auth/config` | SPA needs Cognito config before login |
| Static files (`/*`) | SPA HTML/JS/CSS assets |
| `GET /ws/logs` | WebSocket (separate auth model) |


