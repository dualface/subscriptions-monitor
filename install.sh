#!/usr/bin/env bash

set -euo pipefail

SERVICE_NAME="sub-mon"
INSTALL_BIN="/usr/local/bin/sub-mon"
INSTALL_SERVICE="/etc/systemd/system/sub-mon.service"
INSTALL_CONFIG_DIR="/etc/sub-mon"
INSTALL_CONFIG="${INSTALL_CONFIG_DIR}/config.yaml"
SERVICE_USER="sub-mon"
SERVICE_GROUP="sub-mon"
SERVICE_WORKDIR="/var/lib/sub-mon"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_BIN="${1:-${SCRIPT_DIR}/bin/sub-mon}"
SOURCE_SERVICE="${SCRIPT_DIR}/sub-mon.service"
SOURCE_CONFIG_EXAMPLE="${SCRIPT_DIR}/config.example.yaml"

require_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    echo "Please run as root: sudo ./install.sh [path/to/sub-mon]" >&2
    exit 1
  fi
}

check_sources() {
  if [[ ! -f "${SOURCE_BIN}" ]]; then
    echo "Executable not found: ${SOURCE_BIN}" >&2
    echo "Build it first: make build" >&2
    exit 1
  fi

  if [[ ! -f "${SOURCE_SERVICE}" ]]; then
    echo "Service file not found: ${SOURCE_SERVICE}" >&2
    exit 1
  fi
}

ensure_service_account() {
  if ! getent group "${SERVICE_GROUP}" >/dev/null; then
    groupadd --system "${SERVICE_GROUP}"
  fi

  if ! id -u "${SERVICE_USER}" >/dev/null 2>&1; then
    useradd \
      --system \
      --gid "${SERVICE_GROUP}" \
      --home-dir "${SERVICE_WORKDIR}" \
      --create-home \
      --shell /usr/sbin/nologin \
      "${SERVICE_USER}"
  fi
}

install_binary() {
  install -D -m 0755 "${SOURCE_BIN}" "${INSTALL_BIN}"
}

install_service_file() {
  install -D -m 0644 "${SOURCE_SERVICE}" "${INSTALL_SERVICE}"
}

install_config_if_needed() {
  install -d -m 0755 "${INSTALL_CONFIG_DIR}"

  if [[ ! -f "${INSTALL_CONFIG}" ]]; then
    if [[ -f "${SOURCE_CONFIG_EXAMPLE}" ]]; then
      install -m 0640 "${SOURCE_CONFIG_EXAMPLE}" "${INSTALL_CONFIG}"
      echo "Installed default config to ${INSTALL_CONFIG}"
      echo "Please edit it with your credentials before production use."
    else
      echo "Warning: ${INSTALL_CONFIG} does not exist and config.example.yaml was not found." >&2
    fi
  fi

  chown root:"${SERVICE_GROUP}" "${INSTALL_CONFIG}" 2>/dev/null || true
}

prepare_runtime_dirs() {
  install -d -m 0755 -o "${SERVICE_USER}" -g "${SERVICE_GROUP}" "${SERVICE_WORKDIR}"
}

reload_and_enable_service() {
  systemctl daemon-reload
  systemctl enable --now "${SERVICE_NAME}.service"
}

show_status() {
  systemctl --no-pager --full status "${SERVICE_NAME}.service" || true
}

main() {
  require_root
  check_sources
  ensure_service_account
  install_binary
  install_service_file
  install_config_if_needed
  prepare_runtime_dirs
  reload_and_enable_service
  show_status

  echo
  echo "Installation complete."
  echo "- Binary : ${INSTALL_BIN}"
  echo "- Service: ${INSTALL_SERVICE}"
  echo "- Config : ${INSTALL_CONFIG}"
}

main "$@"
