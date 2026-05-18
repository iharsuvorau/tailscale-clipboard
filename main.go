package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	mu        sync.RWMutex
	clipboard string
	updatedAt time.Time
)

func main() {
	addr := flag.String("addr", ":8765", "listen address")
	prefix := flag.String("prefix", "/clipboard", "URL path prefix (no trailing slash)")
	flag.Parse()

	p := strings.TrimRight(*prefix, "/")

	http.HandleFunc("/", handleIndex(p))
	http.HandleFunc("/api/clipboard", handleClipboard)

	log.Printf("Listening on %s with prefix %s", *addr, p)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func handleClipboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		mu.RLock()
		text := clipboard
		ts := updatedAt
		mu.RUnlock()
		json.NewEncoder(w).Encode(map[string]any{
			"text":      text,
			"updatedAt": ts,
		})

	case http.MethodPost:
		var body struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		mu.Lock()
		clipboard = body.Text
		updatedAt = time.Now()
		mu.Unlock()
		log.Printf("Clipboard updated: %d chars", len(body.Text))
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleIndex(prefix string) http.HandlerFunc {
	html := strings.Replace(indexHTML, "{{base}}", prefix+"/", 1)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	}
}

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0">
<base href="{{base}}">
<title>Clipboard</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }

  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    background: #0f0f10;
    color: #e8e8ea;
    min-height: 100dvh;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 24px 16px;
    gap: 16px;
  }

  h1 {
    font-size: 1.1rem;
    font-weight: 600;
    letter-spacing: 0.04em;
    color: #a0a0a8;
    text-transform: uppercase;
  }

  #status {
    font-size: 0.78rem;
    color: #606068;
    height: 1em;
  }
  #status.ok  { color: #34d399; }
  #status.err { color: #f87171; }

  textarea {
    width: 100%;
    max-width: 600px;
    min-height: 220px;
    resize: vertical;
    background: #1a1a1e;
    color: #e8e8ea;
    border: 1px solid #2e2e36;
    border-radius: 12px;
    padding: 14px;
    font-size: 1rem;
    line-height: 1.5;
    outline: none;
    transition: border-color 0.15s;
  }
  textarea:focus { border-color: #5865f2; }

  .actions {
    display: flex;
    gap: 10px;
    width: 100%;
    max-width: 600px;
  }

  button {
    flex: 1;
    padding: 13px 0;
    border: none;
    border-radius: 10px;
    font-size: 0.95rem;
    font-weight: 600;
    cursor: pointer;
    transition: opacity 0.15s, transform 0.1s;
  }
  button:active { transform: scale(0.97); opacity: 0.85; }

  #btn-pull {
    background: #2e2e3a;
    color: #c8c8d4;
  }
  #btn-push {
    background: #5865f2;
    color: #fff;
  }

  #btn-copy {
    background: #1a1a1e;
    color: #a0a0a8;
    border: 1px solid #2e2e36;
    flex: 0 0 auto;
    padding: 13px 18px;
  }

  #updated {
    font-size: 0.75rem;
    color: #505058;
    max-width: 600px;
    width: 100%;
  }
</style>
</head>
<body>
<h1>&#x1F4CB; Clipboard</h1>
<span id="status"></span>

<textarea id="txt" placeholder="Paste text here, or pull from server…"></textarea>
<div id="updated"></div>

<div class="actions">
  <button id="btn-pull" onclick="pull()">&#x2193; Pull</button>
  <button id="btn-push" onclick="push()">&#x2191; Push</button>
  <button id="btn-copy" onclick="copyLocal()" title="Copy to device clipboard">Copy</button>
</div>

<script>
const txt = document.getElementById('txt');
const status = document.getElementById('status');
const updatedEl = document.getElementById('updated');

function setStatus(msg, cls) {
  status.textContent = msg;
  status.className = cls || '';
  if (cls) setTimeout(() => { status.textContent = ''; status.className = ''; }, 2500);
}

async function pull() {
  try {
    const r = await fetch('api/clipboard');
    if (!r.ok) throw new Error(r.status);
    const d = await r.json();
    txt.value = d.text;
    const ts = d.updatedAt ? new Date(d.updatedAt).toLocaleTimeString() : '—';
    updatedEl.textContent = 'Last updated: ' + ts;
    setStatus('Pulled', 'ok');
  } catch(e) {
    setStatus('Pull failed', 'err');
  }
}

async function push() {
  try {
    const r = await fetch('api/clipboard', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({ text: txt.value }),
    });
    if (!r.ok) throw new Error(r.status);
    updatedEl.textContent = 'Last updated: ' + new Date().toLocaleTimeString();
    setStatus('Pushed', 'ok');
  } catch(e) {
    setStatus('Push failed', 'err');
  }
}

async function copyLocal() {
  try {
    await navigator.clipboard.writeText(txt.value);
    setStatus('Copied to clipboard', 'ok');
  } catch(e) {
    txt.select();
    setStatus('Select all and copy manually', '');
  }
}

// Pull on load
pull();
</script>
</body>
</html>
`
