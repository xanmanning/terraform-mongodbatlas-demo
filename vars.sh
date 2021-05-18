#!/usr/bin/env bash
set -euo pipefail

#
# ./go script configuration
#

# Terraform version
TF_VERSION="0.15.3"

# Directory paths for bin/ and venv/
E_BIN="$(pwd)/bin"
E_VENV="$(pwd)/venv"
E_VENV_BIN="${E_VENV}/bin"
E_VENV_ACT="${E_VENV_BIN}/activate"

# Python executables
E_PIP="pip3"
E_PYTHON_VENV="python3 -m venv"

#
# PLEASE DO NOT DELETE
#

if [ "${BASH_SOURCE[0]}" == "${0}" ] ; then
    echo "Hello friend! I know that you're curious but you're not" \
         "meant to execute this! Sorry :("
fi
