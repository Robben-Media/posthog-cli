# posthog-cli

PostHog CLI — Product analytics from the command line.

## Installation

### Download Binary

Download the latest release from [GitHub Releases](https://github.com/Robben-Media/posthog-cli/releases).

### Build from Source

```bash
git clone https://github.com/Robben-Media/posthog-cli.git
cd posthog-cli
go build ./cmd/posthog
```

## Configuration

posthog-cli requires a PostHog API key. Credentials are stored securely in your system keyring.

**Store credentials:**

```bash
posthog-cli auth set-key --project-id <id> --region us
```

The CLI will prompt for your API key interactively. You can also pipe it:

```bash
echo "phx_..." | posthog-cli auth set-key --project-id <id>
```

**Environment variable override:**

```bash
export POSTHOG_API_KEY="your-api-key"
```

**Check status:**

```bash
posthog-cli auth status
```

**Remove credentials:**

```bash
posthog-cli auth remove
```

## Commands

### auth

Manage API credentials.

| Command | Description |
|---------|-------------|
| `auth set-key` | Store API key, project ID, and region in keyring |
| `auth status` | Show authentication status |
| `auth remove` | Remove stored credentials |

### project

| Command | Description |
|---------|-------------|
| `project info` | Show current project name, timezone, and settings |
| `project list` | List all projects in the organization |

### events

| Command | Description |
|---------|-------------|
| `events list` | List events with optional filters |

**Flags:** `--event` (filter by event name), `--property` (filter by property, e.g. `$current_url=/roofing/`), `--days` (default 30), `--limit` (default 50)

### actions

| Command | Description |
|---------|-------------|
| `actions list` | List all actions |
| `actions get <id>` | Get action details by ID |

### sessions

| Command | Description |
|---------|-------------|
| `sessions list` | List session recordings |
| `sessions get <id>` | Get session recording details |

**Flags (list):** `--days` (default 7), `--limit` (default 20)

### dashboard

| Command | Description |
|---------|-------------|
| `dashboard list` | List all dashboards |
| `dashboard get <id>` | Get dashboard with insight summaries |

**Flags (list):** `--limit` (default 50)

### query

| Command | Description |
|---------|-------------|
| `query <hogql>` | Execute a HogQL query |

### heatmap

| Command | Description |
|---------|-------------|
| `heatmap` | Get heatmap data (click/scroll coordinates) for a URL |

**Flags:** `--url` (required, URL path), `--days` (default 7), `--type` (click or scroll, default click)

### web

Web analytics convenience commands.

| Command | Description |
|---------|-------------|
| `web top-pages` | Top pages by pageviews |
| `web scroll-depth` | Average scroll depth per URL |
| `web engagement` | Rage clicks, dead clicks, and quick backs |
| `web sources` | Traffic sources by referring domain |

**Flags:** `--days` (default 30), `--limit` (default 20), `--url` (required for scroll-depth)

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Output JSON to stdout (best for scripting) |
| `--plain` | Output stable, parseable text to stdout (TSV; no colors) |
| `--verbose` | Enable verbose logging |
| `--force` | Skip confirmations for destructive commands |
| `--no-input` | Never prompt; fail instead (useful for CI) |
| `--color` | Color output: auto, always, or never (default auto) |

## License

MIT
