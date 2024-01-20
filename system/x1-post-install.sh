#!/bin/sh
set -e

# create an empty config if it doesn't exist
if [ -d /etc ]; then
	if [ ! -d /etc/x1 ]; then
		mkdir -p /etc/x1 2>/dev/null || true
	fi

	if [ ! -f /etc/x1/config.toml ]; then
  	cp system/etc/x1/config.toml /etc/x1/config.toml
  fi

  if [ -d /etc/default ] && [ ! -f /etc/default/x1 ]; then
		cp system/etc/default/x1 /etc/default/x1
	fi
fi

if [ -d /run/systemd/system ] && [ -d /etc/systemd/system/ ]; then
	systemctl daemon-reload
	systemctl try-restart x1
fi
