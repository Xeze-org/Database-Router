"""
db-router Python Example — minimal Flask test app.

Starts an HTTP server on :5000 that connects to the db-router web UI proxy
(:8080) and renders a simple dashboard showing live data from all three
database backends.

Run:
    pip install -r requirements.txt
    python app.py
    open http://localhost:5000
"""

import os
import json
from flask import Flask, render_template_string, jsonify, request
from client import DbRouterClient, DbRouterError

ROUTER_URL = os.getenv("ROUTER_URL", "http://localhost:8080")
PORT       = int(os.getenv("APP_PORT", 5000))

app    = Flask(__name__)
client = DbRouterClient(ROUTER_URL)

# ── HTML template ─────────────────────────────────────────────────────────────

HTML = """<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>db-router · Python Example</title>
<style>
  :root{--bg:#0d1117;--bg2:#161b22;--border:#30363d;--text:#e6edf3;
        --muted:#7d8590;--blue:#388bfd;--green:#3fb950;--red:#f85149;
        --amber:#d29922;--mono:'Cascadia Code',Consolas,monospace}
  *{box-sizing:border-box;margin:0;padding:0}
  body{background:var(--bg);color:var(--text);font-family:'Segoe UI',system-ui,sans-serif;
       font-size:13px;min-height:100vh}
  header{background:var(--bg2);border-bottom:1px solid var(--border);
         padding:12px 24px;display:flex;align-items:center;gap:12px}
  .logo{font-size:15px;font-weight:700}.logo span{color:var(--blue)}
  .badge{font-size:10px;padding:2px 8px;border-radius:20px;border:1px solid var(--border);
         background:var(--bg);color:var(--muted);font-weight:600}
  .badge.py{border-color:#3572A5;color:#3572A5}
  main{padding:24px;max-width:1100px;margin:0 auto;display:flex;flex-direction:column;gap:20px}
  h2{font-size:13px;font-weight:700;color:var(--muted);text-transform:uppercase;
     letter-spacing:.6px;margin-bottom:12px}
  .status-row{display:flex;gap:10px;flex-wrap:wrap}
  .chip{display:flex;align-items:center;gap:6px;padding:6px 14px;border-radius:20px;
        border:1px solid var(--border);background:var(--bg2);font-size:12px;font-weight:600}
  .chip.connected{border-color:var(--green);color:var(--green)}
  .chip.error{border-color:var(--red);color:var(--red)}
  .chip.disabled{color:var(--muted)}
  .dot{width:7px;height:7px;border-radius:50%;background:currentColor}
  .card{background:var(--bg2);border:1px solid var(--border);border-radius:8px;padding:16px}
  .card-header{display:flex;align-items:center;gap:8px;margin-bottom:14px}
  .card-title{font-size:13px;font-weight:700}
  .db-icon{font-size:16px}
  table{width:100%;border-collapse:collapse;font-size:12px}
  th{text-align:left;padding:6px 10px;background:#21262d;border:1px solid var(--border);
     color:var(--blue);font-weight:700;font-size:11px;text-transform:uppercase}
  td{padding:5px 10px;border:1px solid var(--border);vertical-align:top;
     max-width:220px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
  tr:hover td{background:#21262d}
  .pill{display:inline-block;padding:2px 8px;border-radius:20px;font-size:10px;
        font-weight:700;background:#0d1f12;color:var(--green)}
  .key-list{display:flex;flex-wrap:wrap;gap:6px}
  .key-chip{padding:3px 10px;background:#21262d;border:1px solid var(--border);
            border-radius:12px;font-family:var(--mono);font-size:11px;color:var(--text)}
  .info-box{padding:10px 14px;border-radius:6px;border-left:3px solid var(--green);
            background:#0d1f12;color:var(--green);font-size:12px;margin-bottom:10px}
  .err-box{border-color:var(--red);background:#1f0d0d;color:var(--red)}
  .form-row{display:flex;gap:8px;margin-bottom:12px;flex-wrap:wrap}
  input,select,textarea{background:var(--bg);border:1px solid var(--border);border-radius:5px;
    color:var(--text);padding:6px 10px;font-family:inherit;font-size:12px;outline:none}
  input:focus,textarea:focus{border-color:var(--blue)}
  textarea{resize:vertical;min-height:64px;font-family:var(--mono)}
  button{padding:6px 14px;border-radius:5px;border:none;background:var(--blue);color:#fff;
         font-size:12px;font-weight:600;cursor:pointer;font-family:inherit}
  button:hover{opacity:.85}
  pre{font-family:var(--mono);font-size:11.5px;white-space:pre-wrap;word-break:break-word;
      line-height:1.6;padding:10px;background:var(--bg);border:1px solid var(--border);
      border-radius:5px;max-height:260px;overflow-y:auto}
  .footer{text-align:center;color:var(--muted);font-size:11px;padding:20px;margin-top:8px}
  .sep{border:none;border-top:1px solid var(--border);margin:8px 0}
</style>
</head>
<body>
<header>
  <div class="logo">db<span>-</span>router</div>
  <div class="badge py">Python Example</div>
  <div class="badge" style="margin-left:auto;font-size:10px">→ {{ router_url }}</div>
</header>
<main>

  <!-- Status -->
  <div>
    <h2>Connection Status</h2>
    <div class="status-row">
      {% for name, icon, s in status %}
      <div class="chip {{ s.status if s else 'error' }}">
        <div class="dot"></div>
        {{ icon }} {{ name }}
        {% if s and s.get('host') %} · {{ s.host }}{% endif %}
        {% if s and s.get('database') %} / {{ s.database }}{% endif %}
        {% if s and s.status == 'error' %} · {{ s.get('error','?') }}{% endif %}
      </div>
      {% endfor %}
    </div>
  </div>

  <!-- PostgreSQL -->
  <div class="card">
    <div class="card-header">
      <span class="db-icon">🐘</span>
      <span class="card-title">PostgreSQL</span>
    </div>

    <h2>Databases</h2>
    <div class="key-list" style="margin-bottom:14px">
      {% for db in pg_dbs %}<div class="key-chip">{{ db }}</div>{% endfor %}
    </div>

    {% if pg_products %}
    <h2>shop · products <span class="pill">{{ pg_products|length }} rows</span></h2>
    <div style="overflow-x:auto;margin-bottom:14px">
    <table>
      <thead><tr>{% for col in pg_products[0].keys() %}<th>{{ col }}</th>{% endfor %}</tr></thead>
      <tbody>{% for row in pg_products %}<tr>{% for v in row.values() %}<td title="{{ v }}">{{ v }}</td>{% endfor %}</tr>{% endfor %}</tbody>
    </table>
    </div>
    {% endif %}

    {% if pg_orders %}
    <h2>shop · orders (joined) <span class="pill">{{ pg_orders|length }} rows</span></h2>
    <div style="overflow-x:auto">
    <table>
      <thead><tr>{% for col in pg_orders[0].keys() %}<th>{{ col }}</th>{% endfor %}</tr></thead>
      <tbody>{% for row in pg_orders %}<tr>{% for v in row.values() %}<td title="{{ v }}">{{ v }}</td>{% endfor %}</tr>{% endfor %}</tbody>
    </table>
    </div>
    {% endif %}

    <hr class="sep" style="margin-top:14px"/>
    <h2 style="margin-top:10px">Run SQL Query</h2>
    <form id="sql-form" style="margin-top:8px">
      <div class="form-row">
        <input id="sql-db" placeholder="database (e.g. shop)" style="width:180px"/>
        <button type="submit">Execute</button>
      </div>
      <textarea id="sql-query" style="width:100%" placeholder="SELECT * FROM products LIMIT 5">SELECT * FROM products LIMIT 5</textarea>
    </form>
    <div id="sql-result" style="margin-top:10px"></div>
  </div>

  <!-- MongoDB -->
  <div class="card">
    <div class="card-header">
      <span class="db-icon">🍃</span>
      <span class="card-title">MongoDB</span>
    </div>

    <h2>Databases</h2>
    <div class="key-list" style="margin-bottom:14px">
      {% for db in mongo_dbs %}<div class="key-chip">{{ db }}</div>{% endfor %}
    </div>

    {% if mongo_docs %}
    <h2>{{ mongo_col }} documents <span class="pill">{{ mongo_docs|length }}</span></h2>
    <div style="overflow-x:auto">
    <table>
      <thead><tr>{% for col in mongo_docs[0].keys() %}<th>{{ col }}</th>{% endfor %}</tr></thead>
      <tbody>{% for doc in mongo_docs %}<tr>{% for v in doc.values() %}<td title="{{ v }}">{{ v }}</td>{% endfor %}</tr>{% endfor %}</tbody>
    </table>
    </div>
    {% endif %}

    <hr class="sep" style="margin-top:14px"/>
    <h2 style="margin-top:10px">Find Documents</h2>
    <form id="mg-form" style="margin-top:8px">
      <div class="form-row">
        <input id="mg-db"  placeholder="database" style="width:150px" value="xeze_test"/>
        <input id="mg-col" placeholder="collection" style="width:150px"/>
        <button type="submit">Find</button>
      </div>
    </form>
    <div id="mg-result" style="margin-top:10px"></div>
  </div>

  <!-- Redis -->
  <div class="card">
    <div class="card-header">
      <span class="db-icon">🔴</span>
      <span class="card-title">Redis</span>
      <span style="margin-left:auto;color:var(--muted);font-size:11px">{{ redis_size }} keys total</span>
    </div>

    <h2>Keys <span class="pill">{{ redis_keys|length }} shown</span></h2>
    <div class="key-list" style="margin-bottom:14px">
      {% for k in redis_keys %}<div class="key-chip">{{ k }}</div>{% else %}<span style="color:var(--muted)">no keys</span>{% endfor %}
    </div>

    <hr class="sep"/>
    <h2 style="margin-top:10px">Set Key</h2>
    <form id="rd-set-form" style="margin-top:8px">
      <div class="form-row">
        <input id="rd-key"   placeholder="key" style="width:160px"/>
        <input id="rd-value" placeholder="value" style="width:200px"/>
        <input id="rd-ttl"   placeholder="ttl (sec, 0=∞)" style="width:120px" type="number" value="0"/>
        <button type="submit">Set</button>
      </div>
    </form>
    <div id="rd-set-result" style="margin-top:6px"></div>

    <hr class="sep"/>
    <h2 style="margin-top:10px">Get Key</h2>
    <form id="rd-get-form" style="margin-top:8px">
      <div class="form-row">
        <input id="rd-get-key" placeholder="key" style="width:200px"/>
        <button type="submit">Get</button>
      </div>
    </form>
    <div id="rd-get-result" style="margin-top:6px"></div>
  </div>

</main>
<div class="footer">db-router Python example · connected to {{ router_url }}</div>

<script>
const api = (path, opts={}) => fetch(path, opts).then(r=>r.json());

function renderJSON(data) {
  const pre = document.createElement('pre');
  pre.textContent = JSON.stringify(data, null, 2);
  return pre;
}
function renderTable(rows) {
  if (!rows || !rows.length) return Object.assign(document.createElement('div'), {
    className:'info-box', textContent:'No rows returned.'
  });
  const t = document.createElement('table');
  const keys = [...new Set(rows.flatMap(r => Object.keys(r)))];
  t.innerHTML = '<thead><tr>' + keys.map(k=>`<th>${k}</th>`).join('') + '</tr></thead>';
  const tb = document.createElement('tbody');
  rows.forEach(row => {
    const tr = document.createElement('tr');
    keys.forEach(k => { const td = document.createElement('td'); td.textContent = row[k] ?? ''; td.title = String(row[k] ?? ''); tr.appendChild(td); });
    tb.appendChild(tr);
  });
  t.appendChild(tb);
  const w = document.createElement('div');
  w.style.overflowX = 'auto';
  w.appendChild(t);
  return w;
}
function showResult(el, data) {
  el.innerHTML = '';
  if (data.error) {
    el.innerHTML = `<div class="info-box err-box">${data.error}</div>`;
    return;
  }
  const rows = data.rows || data.data || data.documents;
  if (rows) el.appendChild(renderTable(rows));
  else el.appendChild(renderJSON(data));
}

document.getElementById('sql-form').addEventListener('submit', async e => {
  e.preventDefault();
  const res = document.getElementById('sql-result');
  res.textContent = 'Running…';
  const data = await api('/proxy/pg/query', {
    method:'POST', headers:{'Content-Type':'application/json'},
    body: JSON.stringify({query: document.getElementById('sql-query').value,
                          database: document.getElementById('sql-db').value})
  });
  showResult(res, data);
});

document.getElementById('mg-form').addEventListener('submit', async e => {
  e.preventDefault();
  const res = document.getElementById('mg-result');
  res.textContent = 'Querying…';
  const data = await api(`/proxy/mongo/find?db=${encodeURIComponent(document.getElementById('mg-db').value)}&col=${encodeURIComponent(document.getElementById('mg-col').value)}`);
  showResult(res, data);
});

document.getElementById('rd-set-form').addEventListener('submit', async e => {
  e.preventDefault();
  const res = document.getElementById('rd-set-result');
  const data = await api('/proxy/redis/set', {
    method:'POST', headers:{'Content-Type':'application/json'},
    body: JSON.stringify({key: document.getElementById('rd-key').value,
                          value: document.getElementById('rd-value').value,
                          ttl: parseInt(document.getElementById('rd-ttl').value)||0})
  });
  res.innerHTML = '';
  res.appendChild(renderJSON(data));
});

document.getElementById('rd-get-form').addEventListener('submit', async e => {
  e.preventDefault();
  const res = document.getElementById('rd-get-result');
  const data = await api(`/proxy/redis/get?key=${encodeURIComponent(document.getElementById('rd-get-key').value)}`);
  res.innerHTML = '';
  res.appendChild(renderJSON(data));
});
</script>
</body>
</html>"""

# ── Flask routes ──────────────────────────────────────────────────────────────

def safe(fn, default=None):
    """Run fn, return default on any error."""
    try:
        return fn()
    except Exception:
        return default


@app.route("/")
def index():
    health = safe(lambda: client.health(), {})

    def chip(key):
        return health.get(key, {"status": "error", "error": "unreachable"})

    status = [
        ("PostgreSQL", "🐘", chip("postgres")),
        ("MongoDB",    "🍃", chip("mongo")),
        ("Redis",      "🔴", chip("redis")),
    ]

    pg_dbs    = safe(lambda: client.pg_databases(), [])
    pg_products = safe(lambda: client.pg_select("shop", "products", 10), [])

    pg_orders = []
    if pg_products:
        res = safe(lambda: client.pg_query(
            "SELECT o.id, p.name, o.quantity, o.total, o.status "
            "FROM orders o JOIN products p ON p.id = o.product_id "
            "ORDER BY o.ordered_at DESC",
            "shop"
        ))
        if res and "rows" in res:
            pg_orders = res["rows"]

    mongo_dbs = safe(lambda: client.mongo_databases(), [])
    mongo_docs, mongo_col = [], ""
    for db in mongo_dbs:
        cols = safe(lambda: client.mongo_collections(db), [])
        if cols:
            mongo_col = f"{db} · {cols[0]}"
            mongo_docs = safe(lambda: client.mongo_find(db, cols[0])[:5], [])
            break

    redis_keys = safe(lambda: client.redis_keys("*")[:20], [])
    redis_info = safe(lambda: client.redis_info(), {})
    redis_size = redis_info.get("db_size", "?")

    return render_template_string(
        HTML,
        router_url=ROUTER_URL,
        status=status,
        pg_dbs=pg_dbs,
        pg_products=pg_products,
        pg_orders=pg_orders,
        mongo_dbs=mongo_dbs,
        mongo_docs=mongo_docs,
        mongo_col=mongo_col,
        redis_keys=redis_keys,
        redis_size=redis_size,
    )


# ── Proxy endpoints (called by JS) ────────────────────────────────────────────

@app.route("/proxy/pg/query", methods=["POST"])
def proxy_pg_query():
    body = request.get_json(force=True)
    try:
        return jsonify(client.pg_query(body.get("query", ""), body.get("database", "")))
    except DbRouterError as e:
        return jsonify({"error": str(e)}), 400


@app.route("/proxy/mongo/find")
def proxy_mongo_find():
    db  = request.args.get("db", "")
    col = request.args.get("col", "")
    try:
        docs = client.mongo_find(db, col)
        return jsonify({"documents": docs, "count": len(docs)})
    except DbRouterError as e:
        return jsonify({"error": str(e)}), 400


@app.route("/proxy/redis/set", methods=["POST"])
def proxy_redis_set():
    body = request.get_json(force=True)
    try:
        return jsonify(client.redis_set(body["key"], body["value"], body.get("ttl", 0)))
    except DbRouterError as e:
        return jsonify({"error": str(e)}), 400


@app.route("/proxy/redis/get")
def proxy_redis_get():
    key = request.args.get("key", "")
    try:
        return jsonify(client.redis_get(key))
    except DbRouterError as e:
        return jsonify({"error": str(e)}), 400


# ── Main ──────────────────────────────────────────────────────────────────────

if __name__ == "__main__":
    print(f"  db-router Python example")
    print(f"  Router proxy : {ROUTER_URL}")
    print(f"  App          : http://localhost:{PORT}")
    app.run(host="0.0.0.0", port=PORT, debug=False)
