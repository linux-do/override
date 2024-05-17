#!/bin/sh

set -e

OS_NAME=$(uname -s)

KDE_ENV_DIR="${HOME}/.config/plasma-workspace/env"

PROFILE_PATH="${HOME}/.profile"
ZSH_PROFILE_PATH="${HOME}/.zshrc"
PLIST_PATH="${HOME}/Library/LaunchAgents/copilot.override.plist"

if [ "$OS_NAME" = "Darwin" ]; then
  BASH_PROFILE_PATH="${HOME}/.bash_profile"
else
  BASH_PROFILE_PATH="${HOME}/.bashrc"
fi

touch "${PROFILE_PATH}"
touch "${BASH_PROFILE_PATH}"
touch "${ZSH_PROFILE_PATH}"

GH_OVERRIDE_SHELL_NAME="copilot.override.sh"
GH_OVERRIDE_SHELL_FILE="${HOME}/.${GH_OVERRIDE_SHELL_NAME}"

rm -rf "${GH_OVERRIDE_SHELL_FILE}"

if [ "$OS_NAME" = "Darwin" ]; then
  launchctl unsetenv "AGENT_DEBUG_OVERRIDE_PROXY_URL"
  launchctl unsetenv "GITHUB_COPILOT_OVERRIDE_PROXY_URL"
  launchctl unsetenv "AGENT_DEBUG_OVERRIDE_CAPI_URL"
  launchctl unsetenv "GITHUB_COPILOT_OVERRIDE_CAPI_URL"

  rm -rf "${PLIST_PATH}"

  # shellcheck disable=SC2016
  sed -i '' '/___GH_OVERRIDE_SHELL_FILE="${HOME}\/\.copilot\.override\.sh"; if /d' "${PROFILE_PATH}" >/dev/null 2>&1
  # shellcheck disable=SC2016
  sed -i '' '/___GH_OVERRIDE_SHELL_FILE="${HOME}\/\.copilot\.override\.sh"; if /d' "${BASH_PROFILE_PATH}" >/dev/null 2>&1
  # shellcheck disable=SC2016
  sed -i '' '/___GH_OVERRIDE_SHELL_FILE="${HOME}\/\.copilot\.override\.sh"; if /d' "${ZSH_PROFILE_PATH}" >/dev/null 2>&1

  echo 'done.'
else
  # shellcheck disable=SC2016
  sed -i '/___GH_OVERRIDE_SHELL_FILE="${HOME}\/\.copilot\.override\.sh"; if /d' "${PROFILE_PATH}" >/dev/null 2>&1
  # shellcheck disable=SC2016
  sed -i '/___GH_OVERRIDE_SHELL_FILE="${HOME}\/\.copilot\.override\.sh"; if /d' "${BASH_PROFILE_PATH}" >/dev/null 2>&1
  # shellcheck disable=SC2016
  sed -i '/___GH_OVERRIDE_SHELL_FILE="${HOME}\/\.copilot\.override\.sh"; if /d' "${ZSH_PROFILE_PATH}" >/dev/null 2>&1

  # shellcheck disable=SC2115
  rm -rf "${KDE_ENV_DIR}/${GH_OVERRIDE_SHELL_NAME}"
  echo "done. you'd better log off first!"
fi
