#!/bin/sh

set -e

OS_NAME=$(uname -s)
BASE_URL="http://127.0.0.1:8181"

KDE_ENV_DIR="${HOME}/.config/plasma-workspace/env"
LAUNCH_AGENTS_DIR="${HOME}/Library/LaunchAgents"

PROFILE_PATH="${HOME}/.profile"
ZSH_PROFILE_PATH="${HOME}/.zshrc"
PLIST_PATH="${LAUNCH_AGENTS_DIR}/copilot.override.plist"

if [ "$OS_NAME" = "Darwin" ]; then
  BASH_PROFILE_PATH="${HOME}/.bash_profile"

  mkdir -p "${LAUNCH_AGENTS_DIR}"
  echo '<?xml version="1.0" encoding="UTF-8"?><!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd"><plist version="1.0"><dict><key>Label</key><string>copilot.override</string><key>ProgramArguments</key><array><string>sh</string><string>-c</string><string>' >"${PLIST_PATH}"
else
  BASH_PROFILE_PATH="${HOME}/.bashrc"
  mkdir -p "${KDE_ENV_DIR}"
fi

touch "${PROFILE_PATH}"
touch "${BASH_PROFILE_PATH}"
touch "${ZSH_PROFILE_PATH}"

GH_OVERRIDE_SHELL_NAME="copilot.override.sh"
GH_OVERRIDE_SHELL_FILE="${HOME}/.${GH_OVERRIDE_SHELL_NAME}"
echo '#!/bin/sh' >"${GH_OVERRIDE_SHELL_FILE}"

# shellcheck disable=SC2016
EXEC_LINE='___GH_OVERRIDE_SHELL_FILE="${HOME}/.copilot.override.sh"; if [ -f "${___GH_OVERRIDE_SHELL_FILE}" ]; then . "${___GH_OVERRIDE_SHELL_FILE}"; fi'

# shellcheck disable=SC2129
echo "export AGENT_DEBUG_OVERRIDE_PROXY_URL=\"${BASE_URL}\"" >>"${GH_OVERRIDE_SHELL_FILE}"
echo "export GITHUB_COPILOT_OVERRIDE_PROXY_URL=\"${BASE_URL}\"" >>"${GH_OVERRIDE_SHELL_FILE}"
echo "export AGENT_DEBUG_OVERRIDE_CAPI_URL=\"${BASE_URL}/v1\"" >>"${GH_OVERRIDE_SHELL_FILE}"
echo "export GITHUB_COPILOT_OVERRIDE_CAPI_URL=\"${BASE_URL}/v1\"" >>"${GH_OVERRIDE_SHELL_FILE}"

if [ "$OS_NAME" = "Darwin" ]; then
  launchctl setenv "AGENT_DEBUG_OVERRIDE_PROXY_URL" "${BASE_URL}"
  launchctl setenv "GITHUB_COPILOT_OVERRIDE_PROXY_URL" "${BASE_URL}"
  launchctl setenv "AGENT_DEBUG_OVERRIDE_CAPI_URL" "${BASE_URL}/v1"
  launchctl setenv "GITHUB_COPILOT_OVERRIDE_CAPI_URL" "${BASE_URL}/v1"

  # shellcheck disable=SC2129
  echo "launchctl setenv \"AGENT_DEBUG_OVERRIDE_PROXY_URL\" \"${BASE_URL}\"" >>"${PLIST_PATH}"
  echo "launchctl setenv \"GITHUB_COPILOT_OVERRIDE_PROXY_URL\" \"${BASE_URL}\"" >>"${PLIST_PATH}"
  echo "launchctl setenv \"AGENT_DEBUG_OVERRIDE_CAPI_URL\" \"${BASE_URL}/v1\"" >>"${PLIST_PATH}"
  echo "launchctl setenv \"GITHUB_COPILOT_OVERRIDE_CAPI_URL\" \"${BASE_URL}/v1\"" >>"${PLIST_PATH}"
fi

if [ "$OS_NAME" = "Darwin" ]; then
  # shellcheck disable=SC2016
  sed -i '' '/___GH_OVERRIDE_SHELL_FILE="${HOME}\/\.copilot\.override\.sh"; if /d' "${PROFILE_PATH}" >/dev/null 2>&1
  # shellcheck disable=SC2016
  sed -i '' '/___GH_OVERRIDE_SHELL_FILE="${HOME}\/\.copilot\.override\.sh"; if /d' "${BASH_PROFILE_PATH}" >/dev/null 2>&1
  # shellcheck disable=SC2016
  sed -i '' '/___GH_OVERRIDE_SHELL_FILE="${HOME}\/\.copilot\.override\.sh"; if /d' "${ZSH_PROFILE_PATH}" >/dev/null 2>&1
  
  echo '</string></array><key>RunAtLoad</key><true/></dict></plist>' >>"${PLIST_PATH}"
else
  # shellcheck disable=SC2016
  sed -i '/___GH_OVERRIDE_SHELL_FILE="${HOME}\/\.copilot\.override\.sh"; if /d' "${PROFILE_PATH}" >/dev/null 2>&1
  # shellcheck disable=SC2016
  sed -i '/___GH_OVERRIDE_SHELL_FILE="${HOME}\/\.copilot\.override\.sh"; if /d' "${BASH_PROFILE_PATH}" >/dev/null 2>&1
  # shellcheck disable=SC2016
  sed -i '/___GH_OVERRIDE_SHELL_FILE="${HOME}\/\.copilot\.override\.sh"; if /d' "${ZSH_PROFILE_PATH}" >/dev/null 2>&1
fi

echo "${EXEC_LINE}" >>"${PROFILE_PATH}"
echo "${EXEC_LINE}" >>"${BASH_PROFILE_PATH}"
echo "${EXEC_LINE}" >>"${ZSH_PROFILE_PATH}"

if [ "$OS_NAME" = "Darwin" ]; then
  echo 'done. the "kill Dock" command can fix the crash issue.'
else
  ln -sf "${GH_OVERRIDE_SHELL_FILE}" "${KDE_ENV_DIR}/${GH_OVERRIDE_SHELL_NAME}"
  echo "done. you'd better log off first!"
fi
