# tailscale-clipboard

Copy and paste across devices on the same [Tailscale](https://tailscale.com) network. Run a lightweight HTTP server on Linux; access it from any device (iPhone, tablet, another PC) via a browser.

## How it works

All Tailscale devices share a private IP network (`100.x.x.x`). The server binds to your machine's Tailscale IP and serves a small web UI. On any other device in your tailnet, open the URL in a browser to push or pull clipboard text.

## Build

Requires Go 1.18+.

```bash
go build -o clipboard-server .
```

## Usage

```bash
# Run bound to your Tailscale IP (default port 8765)
./clipboard-server -addr $(tailscale ip -4):8765

# Or listen on all interfaces
./clipboard-server -addr :8765
```

Open `http://<tailscale-ip>:8765` in any browser on your tailnet.

### Web UI

| Button | Action |
|--------|--------|
| **Pull** | Fetch current clipboard from server into the text area |
| **Push** | Send text area content to server |
| **Copy** | Write text area content to the local device clipboard |

## API

### GET `/api/clipboard`

Returns the current clipboard content.

```json
{
  "text": "hello from PC",
  "updatedAt": "2024-01-15T12:34:56Z"
}
```

### POST `/api/clipboard`

Updates the clipboard content. Returns `204 No Content`.

```json
{ "text": "some text to share" }
```

Example with curl:

```bash
# Push
curl -X POST http://100.x.x.x:8765/api/clipboard \
  -H 'Content-Type: application/json' \
  -d '{"text":"hello from PC"}'

# Pull
curl http://100.x.x.x:8765/api/clipboard
```

## Deployment

### Install binary

```bash
cp clipboard-server ~/.local/bin/
```

### systemd user service

The service runs as your user (`~/.config/systemd/user/`), so no `sudo` is needed and it has access to your Tailscale credentials.

Install and start the service:

```bash
mkdir -p ~/.config/systemd/user
cp clipboard-server.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user enable --now clipboard-server
```

The service resolves your Tailscale IP at startup via `tailscale ip -4` and binds to it on port 8765.

Check status and logs:

```bash
systemctl --user status clipboard-server
journalctl --user -u clipboard-server -f
```

### HTTPS via `tailscale serve`

The service uses `tailscale serve` to expose the server over HTTPS. The browser Clipboard API requires HTTPS to write to the device clipboard on iOS.

The systemd service handles this automatically:
- `ExecStartPre` registers the route with `tailscale serve`
- The Go server binds to `127.0.0.1:8765` (localhost only)
- `ExecStopPost` removes the route when the service stops

The server is reachable at `https://your-machine.your-tailnet.ts.net` from all devices in your tailnet, with a valid TLS certificate — no configuration needed.
