#!/usr/bin/env bash
# Live smoke test against a real test server. Requires:
#   DSCRD_SMOKE_GUILD   — guild ID the bot is in
#   DSCRD_SMOKE_CHANNEL — channel ID the bot can read/write
# plus working credentials (active profile or DSCRD_TOKEN).
#
# Exercises the full read -> send -> react -> search -> delete flow and prints
# PASS/FAIL per step. Run before cutting a release.
set -u

BIN="${DSCRD_BIN:-./dscrd}"
GUILD="${DSCRD_SMOKE_GUILD:?set DSCRD_SMOKE_GUILD}"
CHANNEL="${DSCRD_SMOKE_CHANNEL:?set DSCRD_SMOKE_CHANNEL}"
STAMP="smoke-$$-$(date +%s)"
FAILS=0

step() {
  local name="$1"
  shift
  if out="$("$@" 2>&1)"; then
    echo "PASS  $name"
    [ -n "$out" ] && echo "$out" | sed 's/^/      /'
  else
    echo "FAIL  $name"
    echo "$out" | sed 's/^/      /'
    FAILS=$((FAILS + 1))
  fi
}

step "auth test"      "$BIN" auth test
step "guild list"     "$BIN" guild list
step "guild info"     "$BIN" guild info --guild "$GUILD"
step "channel list"   "$BIN" channel list --guild "$GUILD"
step "msg read"       "$BIN" msg read --channel "$CHANNEL" --limit 3

echo "--- write flow ---"
MSG_ID="$("$BIN" msg send --channel "$CHANNEL" --text "$STAMP" | awk '{print $2}')"
if [ -n "$MSG_ID" ]; then
  echo "PASS  msg send ($MSG_ID)"
else
  echo "FAIL  msg send"
  FAILS=$((FAILS + 1))
fi

if [ -n "$MSG_ID" ]; then
  step "msg react"    "$BIN" msg react --channel "$CHANNEL" --message "$MSG_ID" --emoji "👀"
  step "msg edit"     "$BIN" msg edit --channel "$CHANNEL" --message "$MSG_ID" --text "$STAMP-edited"
  # Search indexing lags; tolerate a miss but report it.
  step "search"       "$BIN" search messages --guild "$GUILD" --content "$STAMP"
  step "msg delete"   "$BIN" msg delete --channel "$CHANNEL" --message "$MSG_ID"
fi

echo "--- done: $FAILS failure(s) ---"
exit "$((FAILS > 0))"
