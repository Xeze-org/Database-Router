#!/usr/bin/env python3
r"""
pg-query.py — Interactive PostgreSQL REPL via the DB Router API
===============================================================
No psycopg2 needed. All queries go through the HTTPS API.

Usage:
    python pg-query.py                   # connect to default DB (unified_db)
    python pg-query.py mydbname          # connect to a different DB

Built-in commands:
    .tables              list all tables
    .schema <table>      show column names and types
    .count  <table>      row count
    .db <name>           switch active DB
    .help  / \?         show this help
    quit / exit / \q     exit
"""

import sys
import os
import json
import requests

# ── Load .env file (if present) ───────────────────────────────────────
def _load_dotenv(path=".env"):
    """Minimal .env loader — no external deps required."""
    try:
        with open(path) as f:
            for line in f:
                line = line.strip()
                if not line or line.startswith("#") or "=" not in line:
                    continue
                key, _, val = line.partition("=")
                key = key.strip()
                val = val.strip().strip('"').strip("'")
                if key and key not in os.environ:  # don't overwrite real env vars
                    os.environ[key] = val
    except FileNotFoundError:
        pass

_load_dotenv()

# ── Config ────────────────────────────────────────────────────────────────────
# Load from environment variables. Copy .env.example to .env and fill in values,
# then either `export` them or use: set -a && source .env && set +a
API_KEY = os.getenv("DB_API_KEY")
PG_BASE = os.getenv("DB_PG_BASE")

if not all([API_KEY, PG_BASE]):
    print("Missing required environment variables. Set:")
    print("  DB_API_KEY  — your API key")
    print("  DB_PG_BASE  — e.g. https://pg.yourdomain.com")
    sys.exit(1)

HEADERS  = {"X-API-Key": API_KEY, "Content-Type": "application/json"}
DB_NAME  = sys.argv[1] if len(sys.argv) > 1 else "unified_db"

# History file for readline (optional)
HISTORY  = os.path.expanduser("~/.pg_repl_history")

# ── Terminal colours (disabled on Windows if not supported) ───────────────────
def _ansi(code): return f"\033[{code}m"
RESET  = _ansi(0)
BOLD   = _ansi(1)
CYAN   = _ansi(36)
GREEN  = _ansi(32)
YELLOW = _ansi(33)
RED    = _ansi(31)
DIM    = _ansi(2)

# Disable colours if stdout isn't a tty (redirect to file, etc)
if not sys.stdout.isatty():
    RESET = BOLD = CYAN = GREEN = YELLOW = RED = DIM = ""

# ── Readline history (best-effort) ────────────────────────────────────────────
try:
    import readline
    try:
        readline.read_history_file(HISTORY)
    except FileNotFoundError:
        pass
    readline.set_history_length(1000)
    _rl_available = True
except ImportError:
    _rl_available = False   # Windows without pyreadline — still works, no history


def _save_history():
    if _rl_available:
        try:
            readline.write_history_file(HISTORY)
        except Exception:
            pass


# ── HTTP helpers ──────────────────────────────────────────────────────────────
def _get(url, **params):
    return requests.get(url, headers=HEADERS, params=params, timeout=15)

def _post(url, body):
    return requests.post(url, headers=HEADERS, json=body, timeout=15)

# ── Pretty table printer ──────────────────────────────────────────────────────
def print_table(rows, columns=None):
    if not rows:
        print(f"{DIM}(0 rows){RESET}")
        return

    if columns is None:
        columns = list(rows[0].keys())

    # Measure widths
    widths = {col: len(str(col)) for col in columns}
    for row in rows:
        for col in columns:
            widths[col] = max(widths[col], len(str(row.get(col) if row.get(col) is not None else "")))

    sep    = "─┼─".join("─" * widths[c] for c in columns)
    header = " │ ".join(f"{BOLD}{str(c).ljust(widths[c])}{RESET}" for c in columns)

    print()
    print(" │ ".join(str(c).ljust(widths[c]) for c in columns))
    print("─┼─".join("─" * widths[c] for c in columns))
    for row in rows:
        line = " │ ".join(
            str(row.get(col) if row.get(col) is not None else "").ljust(widths[col])
            for col in columns
        )
        print(line)

    count = len(rows)
    print(f"\n{GREEN}({count} row{'s' if count != 1 else ''}){RESET}\n")


# ── Run SQL ───────────────────────────────────────────────────────────────────
def run_sql(sql):
    sql = sql.strip().rstrip(";").strip()
    if not sql:
        return
    try:
        r = _post(f"{PG_BASE}/query", {"query": sql})
        data = r.json()

        if "error" in data:
            print(f"\n{RED}ERROR:{RESET} {data['error']}\n")
            return

        # SELECT-like: has rows/columns
        if "rows" in data:
            rows = data["rows"] or []
            cols = data.get("columns") or (list(rows[0].keys()) if rows else [])
            print_table(rows, cols)
            return

        # INSERT
        if "inserted_id" in data:
            print(f"\n{GREEN}INSERT{RESET} id={data['inserted_id']}\n")
            return

        # UPDATE / DELETE
        if "rows_affected" in data:
            verb = "UPDATE/DELETE"
            print(f"\n{GREEN}{verb}{RESET} {data['rows_affected']} row(s) affected\n")
            return

        # Fallback
        print(f"\n{json.dumps(data, indent=2)}\n")

    except requests.ConnectionError:
        print(f"\n{RED}ERROR:{RESET} Cannot reach {PG_BASE}\n")
    except Exception as exc:
        print(f"\n{RED}ERROR:{RESET} {exc}\n")


# ── Input validation ──────────────────────────────────────────────────────────
def is_valid_identifier(name):
    """Check if a name is a valid SQL identifier (alphanumeric + underscore)."""
    if not name:
        return False
    # Allow letters, digits, underscores. Must start with letter or underscore.
    import re
    return bool(re.match(r'^[a-zA-Z_][a-zA-Z0-9_]*$', name))

def sanitize_identifier(name):
    """Validate and quote a SQL identifier to prevent injection."""
    if not is_valid_identifier(name):
        raise ValueError(f"Invalid identifier: {name}")
    return f'"{name}"'  # Use double quotes for PostgreSQL identifiers

# ── Special dot-commands ──────────────────────────────────────────────────────
HELP_TEXT = f"""
{BOLD}PostgreSQL commands:{RESET}
  .databases           list all databases on the server
  .tables              list all tables in the active database
  .schema <table>      column names, types and nullability
  .count  <table>      SELECT COUNT(*) shortcut
  .db     <name>       switch active database
  .createdb <name>     create a new database

{BOLD}SQL:{RESET}
  Type any SQL statement. End with ; OR press Enter on a blank line to execute.
  Multi-line: keep typing, end with ; or blank line.

{BOLD}Exit:{RESET}  quit | exit | \\q
"""


def handle_special(cmd, db_ref):
    """
    Handle dot-commands. Returns True if handled, False if it should be
    treated as SQL.  db_ref is a mutable list [db_name] so we can reassign.
    """
    parts = cmd.strip().split(None, 2)
    op = parts[0].lower()

    # ── .databases ────────────────────────────────────────────────
    if op == ".databases":
        r = _get(f"{PG_BASE}/databases")
        databases = r.json().get("databases") or []
        if not databases:
            print(f"\n{DIM}No databases found{RESET}\n")
        else:
            print(f"\n{BOLD}Databases:{RESET}")
            for d in databases:
                marker = f"  {GREEN}*{RESET}" if d == db_ref[0] else "   "
                print(f"{marker} {CYAN}{d}{RESET}")
            print()
        return True

    # ── .tables ───────────────────────────────────────────────────────────────
    if op == ".tables":
        r = _get(f"{PG_BASE}/tables/{db_ref[0]}")
        tables = r.json().get("tables") or []
        if not tables:
            print(f"\n{DIM}No tables found in {db_ref[0]}{RESET}\n")
        else:
            print(f"\n{BOLD}Tables in {db_ref[0]}:{RESET}")
            for t in tables:
                print(f"  {CYAN}{t}{RESET}")
            print()
        return True

    # ── .schema <table> ───────────────────────────────────────────────────────
    if op == ".schema":
        if len(parts) < 2:
            print(f"\n{YELLOW}Usage:{RESET} .schema <table>\n")
            return True
        table = parts[1]
        try:
            # Validate identifier to prevent SQL injection
            if not is_valid_identifier(table):
                print(f"\n{RED}ERROR:{RESET} Invalid table name. Use only letters, numbers, and underscores.\n")
                return True
            # Use parameterized query with proper escaping
            run_sql(f"""
                SELECT column_name, data_type, is_nullable, column_default
                FROM information_schema.columns
                WHERE table_name = $${table}$$
                ORDER BY ordinal_position
            """)
        except ValueError as e:
            print(f"\n{RED}ERROR:{RESET} {e}\n")
        return True

    # ── .count <table> ────────────────────────────────────────────────────────
    if op == ".count":
        if len(parts) < 2:
            print(f"\n{YELLOW}Usage:{RESET} .count <table>\n")
            return True
        try:
            # Sanitize table name to prevent SQL injection
            safe_table = sanitize_identifier(parts[1])
            run_sql(f"SELECT COUNT(*) AS total FROM {safe_table}")
        except ValueError as e:
            print(f"\n{RED}ERROR:{RESET} {e}\n")
        return True

    # ── .createdb <name> ─────────────────────────────────────────────
    if op == ".createdb":
        if len(parts) < 2:
            print(f"\n{YELLOW}Usage:{RESET} .createdb <name>\n")
            return True
        name = parts[1]
        try:
            # Sanitize database name to prevent SQL injection
            safe_name = sanitize_identifier(name)
            r = _post(f"{PG_BASE}/query", {"query": f"CREATE DATABASE {safe_name}"})
            data = r.json()
            if "error" in data and data["error"]:
                print(f"\n{RED}ERROR:{RESET} {data['error']}\n")
            else:
                print(f"\n{GREEN}CREATE DATABASE{RESET} {name}\n")
                print(f"{DIM}Tip: switch to it with  .db {name}{RESET}\n")
        except Exception as exc:
            print(f"\n{RED}ERROR:{RESET} {exc}\n")
        return True

    # ── .db <name> ────────────────────────────────────────────────────────────
    if op == ".db":
        if len(parts) < 2:
            print(f"\n{YELLOW}Current DB:{RESET} {db_ref[0]}\n")
            return True
        db_ref[0] = parts[1]
        print(f"\n{GREEN}Switched to database:{RESET} {db_ref[0]}\n")
        return True

    # ── .help / \? ────────────────────────────────────────────────────────────
    if op in (".help", r"\?"):
        print(HELP_TEXT)
        return True

    return False   # not a special command — treat as SQL


# ── Main loop ─────────────────────────────────────────────────────────────────
def main():
    db_ref = [DB_NAME]   # mutable so handle_special can reassign

    # Test connection
    try:
        r = _get(f"{PG_BASE}/test")
        info = r.json()
        status = info.get("status", "?")
        host   = info.get("host", PG_BASE)
        db_srv = info.get("database", "?")
        print(f"\n{GREEN}Connected{RESET} — PostgreSQL @ {host}/{db_srv}")
        print(f"Active DB : {CYAN}{db_ref[0]}{RESET}")
        print(f"Type {BOLD}.help{RESET} for commands\n")
    except Exception as e:
        print(f"\n{YELLOW}Warning:{RESET} Could not reach {PG_BASE}: {e}\n")

    buffer = []

    try:
        while True:
            try:
                prompt = f"  {DIM}->{RESET} " if buffer else f"{CYAN}{db_ref[0]}{RESET}=# "
                line = input(prompt)
            except EOFError:
                print(f"\n{DIM}Bye!{RESET}")
                break

            stripped = line.strip()

            # Blank line — execute buffered SQL (without requiring ;)
            if not stripped:
                if buffer:
                    sql = " ".join(buffer)
                    if _rl_available:
                        readline.add_history(sql)
                    run_sql(sql)
                    buffer.clear()
                continue

            # Exit commands
            if stripped.lower() in ("quit", "exit", r"\q"):
                print(f"{DIM}Bye!{RESET}")
                break

            # Dot-commands (only when no partial buffer)
            if not buffer and stripped.startswith(".") or stripped.startswith("\\"):
                if handle_special(stripped, db_ref):
                    if _rl_available:
                        readline.add_history(stripped)
                    continue
                # Falls through to SQL

            buffer.append(stripped)

            # Execute immediately if line ends with ;
            if stripped.endswith(";"):
                sql = " ".join(buffer)
                if _rl_available:
                    readline.add_history(sql)
                run_sql(sql)
                buffer.clear()

    except KeyboardInterrupt:
        print(f"\n{DIM}Bye!{RESET}")
    finally:
        _save_history()


if __name__ == "__main__":
    main()
