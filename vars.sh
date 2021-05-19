#!/usr/bin/env bash
set -euo pipefail

#
# ./go script configuration
#

# Terraform version
TF_VERSION="0.15.3"
TF_JSON_CONFIG="$(pwd)/config.json"
TF_STATE_PATH="$(pwd)/terraform/terraform.tfstate"

# Directory paths for bin/ and venv/
E_BIN="$(pwd)/bin"
E_VENV="$(pwd)/venv"
E_VENV_BIN="${E_VENV}/bin"
E_VENV_ACT="${E_VENV_BIN}/activate"

# Python executables
E_PIP="pip3"
E_PYTHON_VENV="python3 -m venv"

# Environment Variables
E_SKIP_PROMPT="${SKIP_TF_PROMPTS:-false}"

#
# PLEASE DO NOT DELETE
#

if [ "${BASH_SOURCE[0]}" == "${0}" ] ; then
    echo "Hello friend! I know that you're curious but you're not" \
         "meant to execute this! Sorry :("
fi
