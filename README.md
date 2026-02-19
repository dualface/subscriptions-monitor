# Subscriptions Monitor (sub-mon)

AI service subscription usage monitor supporting Kimi, MiniMax, and ZenMux.

```text
AI Subscriptions Usage
+-----------+-----------------+----------------------------------+
| NAME      | PLAN            | USAGE                            |
+-----------+-----------------+----------------------------------+
| ZenMux    | Ultra Plan      | 7d Flows [######----] 69%        |
|           |                 |   resets in 2d19h                |
|           |                 | 5h Flows [#####-----] 51%        |
|           |                 |   resets in 27m                  |
|           |                 | Total Tokens: 504.9M tokens      |
|           |                 | API Requests: 6.9K requests      |
+-----------+-----------------+----------------------------------+
| Kimi Code | Moderato        | Daily Requests [#---------] 13%  |
|           |                 |   resets in 2d14h                |
|           |                 | Window (300m) [----------] 0%    |
|           |                 |   resets in 3h                   |
+-----------+-----------------+----------------------------------+
| MiniMax   | CodePlanStarter | MiniMax-M2 Usage [----------] 0% |
|           |                 |   resets in 1m                   |
+-----------+-----------------+----------------------------------+
Updated: 2026-02-17 19:58:01 (just now)
```

## Installation

```bash
make build
```

## Quick Start

1. Copy the example configuration:
   ```bash
   mkdir -p ~/.config/sub-mon
   cp config.example.yaml ~/.config/sub-mon/config.yaml
   ```

2. Edit `~/.config/sub-mon/config.yaml` with your credentials (see [How to Get Credentials](#how-to-get-credentials) below)

3. Run the monitor:
   ```bash
   ./bin/sub-mon
   ```

## Commands

- `sub-mon` - Query and display subscription usage (default)
- `sub-mon serve` - Start HTTP API server with 90s cache
- `sub-mon --help` - Show help

## How to Get Credentials

### Kimi Code

1. Open Chrome/Edge and go to https://www.kimi.com/code
2. Open Developer Tools (F12 or Ctrl+Shift+I)
3. Go to **Network** tab
4. Refresh the page (F5)
5. Find any request to `www.kimi.com` in the list
6. Click on the request and look at **Headers**:
   - **Authorization**: Copy the value after `Bearer ` → this is `auth_token`
   - **Cookie**: Copy the entire cookie string → this is `cookie`

Required fields:
- `auth_token`: Bearer token from Authorization header
- `cookie`: Full cookie string from the request

### MiniMax

1. Open Chrome/Edge and go to https://www.minimaxi.com
2. Log in to your account
3. Open Developer Tools (F12)
4. Go to **Network** tab
5. Refresh the page
6. Look for API requests to `minimaxi.com`
7. Click on any request and check **Headers**:
   - **Cookie**: Copy the full cookie string
8. For `group_id`:
   - Look at the URL of API requests, usually contains `?GroupId=xxxx`
   - Or check **Application** → **Local Storage** → `group_id`

Required fields:
- `cookie`: Session cookie from browser
- `group_id`: Your organization ID (usually in URL parameters)

### ZenMux

1. Open Chrome/Edge and go to https://zenmux.com
2. Log in to your account
3. Open Developer Tools (F12)
4. Go to **Application** tab (or **Storage** in Firefox)
5. On the left, expand **Cookies** → `https://zenmux.com`
6. Find and copy these values:
   - `ctoken` → `ctoken`
   - `session_id` → `session_id`
   - `session_id_sig` → `session_id_sig`

Alternative method using Network tab:
1. Go to **Network** tab
2. Refresh the page
3. Find any request to `zenmux.com`
4. Check **Request Headers** for **Cookie** field
5. Extract the three values from the cookie string

Required fields:
- `ctoken`: API token from cookies
- `session_id`: Session ID from cookies
- `session_id_sig`: Session signature from cookies

## Configuration

Configuration file location: `~/.config/sub-mon/config.yaml`

See [config.example.yaml](config.example.yaml) for a complete configuration template.

### Configuration Structure

```yaml
subscriptions:
  - name: <subscription-name>
    provider: <kimi|minimax|zenmux>
    auth:
      type: cookie
      extra:
        # provider-specific credentials

settings:
  timeout: 30s
  api_port: 3456
```

### Provider Configuration

#### Kimi Code

```yaml
- name: my-kimi
  provider: kimi
  auth:
    type: cookie
    extra:
      auth_token: "${KIMI_AUTH_TOKEN}"
      cookie: "${KIMI_COOKIE}"
```

#### MiniMax

```yaml
- name: my-minimax
  provider: minimax
  auth:
    type: cookie
    extra:
      cookie: "${MINIMAX_COOKIE}"
      group_id: "${MINIMAX_GROUP_ID}"
```

#### ZenMux

```yaml
- name: my-zenmux
  provider: zenmux
  auth:
    type: cookie
    extra:
      ctoken: "${ZENMUX_CTOKEN}"
      session_id: "${ZENMUX_SESSION_ID}"
      session_id_sig: "${ZENMUX_SESSION_ID_SIG}"
```

### Environment Variables

You can use environment variables in the config file using `${VAR_NAME}` syntax:

```yaml
auth:
  extra:
    auth_token: "${KIMI_AUTH_TOKEN}"
```

Set the variable before running:
```bash
export KIMI_AUTH_TOKEN="your-token-here"
./bin/sub-mon
```

### Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `timeout` | `10s` | Maximum time to wait for API responses |
| `api_port` | `3456` | HTTP server port for `serve` command |

### Security Notes

- Keep your config file secure: `chmod 600 ~/.config/sub-mon/config.yaml`
- Use environment variables for sensitive credentials
- Never commit credentials to version control
- Cookies may expire and need to be refreshed periodically

## Features

- **Kimi Code**: Monitor daily request quotas and rate limits
- **MiniMax**: Track per-model usage with window reset times
- **ZenMux**: Monitor 5h and 7d flow usage with cost breakdown

## API Server

When running `sub-mon serve`:

- **Cache TTL**: 90 seconds
- **Background Refresh**: Every 60 seconds
- **Endpoints**:
  - `GET /api/v1/health` - Health check
  - `GET /api/v1/usage` - Get usage data (cached)
  - `GET /api/v1/providers` - List available providers

Response headers:
- `X-Cache: HIT` - Returned cached data
- `X-Cache: MISS` - Fetched fresh data

## License

MIT
