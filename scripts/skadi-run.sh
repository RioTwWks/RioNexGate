#!/bin/sh
# Run skadicore and restart when skadi.toml changes (RioNexGate core.Reload()).
# Full restart is used because SIGHUP only reloads [protocol.*]; transport
# sections (reality/xhttp) require a process restart.
set -eu

CONFIG="/config/skadi.toml"
BIN="${SKADICORE_BIN:-/skadicore}"

while [ $# -gt 0 ]; do
	case "$1" in
	--config|-c)
		CONFIG="${2:-$CONFIG}"
		shift 2
		;;
	*)
		# Forward unknown args to skadicore after --
		break
		;;
	esac
done

mtime() {
	if stat -c %Y "$CONFIG" >/dev/null 2>&1; then
		stat -c %Y "$CONFIG"
	elif stat -f %m "$CONFIG" >/dev/null 2>&1; then
		stat -f %m "$CONFIG"
	else
		echo 0
	fi
}

echo "skadi-run: watching ${CONFIG} (bin=${BIN})"

# Wait until RioNexGate writes the first config.
i=0
while [ ! -f "$CONFIG" ]; do
	i=$((i + 1))
	if [ "$i" -gt 60 ]; then
		echo "skadi-run: config not found after wait: ${CONFIG}" >&2
		exit 1
	fi
	echo "skadi-run: waiting for ${CONFIG}..."
	sleep 2
done

while true; do
	echo "skadi-run: starting skadicore"
	"$BIN" --config "$CONFIG" "$@" &
	pid=$!
	last_mtime=$(mtime)

	while kill -0 "$pid" 2>/dev/null; do
		sleep 2
		current_mtime=$(mtime)
		if [ "$current_mtime" != "$last_mtime" ]; then
			echo "skadi-run: config changed, restarting skadicore (pid ${pid})"
			kill "$pid" 2>/dev/null || true
			wait "$pid" 2>/dev/null || true
			break
		fi
	done

	if kill -0 "$pid" 2>/dev/null; then
		continue
	fi

	wait "$pid" 2>/dev/null || true
	echo "skadi-run: skadicore exited, restarting in 3s"
	sleep 3
done
