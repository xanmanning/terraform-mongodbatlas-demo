#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

##
# HELP TEXT
##
#/
#/ Terraform MongoDB Atlas Demo
#/ ----------------------------
#/
#/ Usage:
#/
#/   ./go [action] [option]
#/
#/ Source:
#/
#/   https://github.com/xanmanning/terraform-mongodbatlas-demo
#/
#/ Description:
#/
#/   Bash script to aid in testing and building a MongoDB cluster in Atlas
#/   using Terraform.
#/
#/   For more information about each action run `./go [action]` (no option)
#/
#/   Script inspired by:
#/
#/     - https://www.thoughtworks.com/insights/blog/praise-go-script-part-i
#/     - https://www.thoughtworks.com/insights/blog/praise-go-script-part-ii
#/
#/ Actions:
#/
#/   - test         Test the code and local environment
#/   - build        Build the local environment and infrastructure
#/   - destroy      Destroy the deployed infrastructure
#/   - cleanup      Tidy up your local environment
#/
##/ Examples:
#/

##
# GLOBAL VARIABLES
##
source vars.sh

##
# List of tools, python modules and downloads to check for
##
A_TOOLS=(
    curl
    unzip
    python3
    pip3
    shellcheck
    docker
    jq
)
A_PYTHON_MODULES=(
    venv
)
A_TOOLS_DOWNLOAD=(
    terraform
    bashate
)

##
# Load controller environment
##
# shellcheck source=/dev/null
test -f "${E_VENV_ACT}" && source "${E_VENV_ACT}"
test -d "${E_BIN}" && PATH="${E_BIN}:${PATH}"

##
# HELPER FUNCTIONS
##

#
# usage:
#   Prints the above help text, replaces `./go` with the name of the
#   script.
#
function usage {
    local F
    F=$(basename "${0}")
    grep '^#/' "${0}" | sed -e "s#\\./go#${F}#g" | cut -c4-
    exit 0
}

#
# Catch the --help|-h flag, print usage.
#
expr "$*" : ".*--help$" > /dev/null && usage
expr "$*" : ".*-h$" > /dev/null && usage

##
# MAIN FUNCTIONS
##

#
# check_environment:
#   Check the environment is able to run this project.
#
function check_environment {
    local A_MISSING_PKG
    local TOOL_TEST

    A_MISSING_PKG="false"

    for TOOL in "${A_TOOLS[@]}" ; do
        TOOL_TEST="$(command -v "${TOOL}" 2>/dev/null || true)"
        if [ "${TOOL_TEST}" == "" ] ; then
            echo " - ${TOOL} not found, Please refer to your package manager" \
                 "for instructions for installation."
            A_MISSING_PKG="true"
        fi
    done

    for MODULE in "${A_PYTHON_MODULES[@]}" ; do
        MODULE_TEST="$(python3 -c "help(\"modules ${MODULE}\")" 2>/dev/null | \
            grep -E "^${MODULE} " || true)"
        if [ "${MODULE_TEST}" == "" ] ; then
            echo " - python3-${MODULE} not found, Please refer to your" \
                 "package manager or pip3 for instructions for installation."
            A_MISSING_PKG="true"
        fi
    done

    for TOOL in "${A_TOOLS_DOWNLOAD[@]}" ; do
        TOOL_TEST="$(command -v "${TOOL}" 2>/dev/null || true)"
        if [ "${TOOL_TEST}" == "" ] ; then
            echo " - ${TOOL} not found, Please run: ${0} build controller"
            A_MISSING_PKG="true"
        fi
    done

    if [ "${A_MISSING_PKG}" != "true" ] ; then
        echo "Controller environment OK"
    else
        exit 1
    fi
}

#
# check_go
#   Self checks for go script
#
function check_go {
    set +eu

    echo "Running bashate tests..."
    bashate "${0}"

    echo "Running shellcheck tests..."
    shellcheck -x -e SC1117 "${0}"

    set -eu
}

#
# check_config
#   Checks JSON syntax is valid
#
function check_config {
    set +eu

    for JSON_FILE in $(find . -iname '*.json' -type f) ; do
        echo "Testing ${JSON_FILE}..."
        cat "${JSON_FILE}" | jq -Mc '.' 1>/dev/null 2> >(tee -a .fail >&2) && \
            echo "OK"
        echo ""
    done

    if [ "$(<.fail)" != "" ] ; then
        echo "JSON syntax checking failed"
        test -f .fail && rm -f .fail
        exit 1
    fi

    test -f .fail && rm -f .fail

    set -eu
}

#
# build_environment:
#   Builds an environment that can run Terraform code
#
function build_environment {
    E_PYTHON_VENV="python3 -m venv"
    test -d "${E_BIN}" || mkdir "${E_BIN}"
    test -d "${E_VENV}" || /usr/bin/env bash -c "${E_PYTHON_VENV} ${E_VENV}"

    if [ ! -f "${E_VENV_ACT}" ] ; then
        echo "Expected virtualenv ${E_VENV_ACT} was not found."
        exit 1
    fi

    # shellcheck source=/dev/null
    source "${E_VENV_ACT}"
    PATH="${E_BIN}:${PATH}"

    ${E_PIP} install pip --upgrade
    ${E_PIP} install -r requirements.txt

    download_terraform
}

#
# download_terraform
#   Downloads and extracts a specified version of Terraform
#
function download_terraform {
    local OS_PLATFORM
    local OS_PLATFORM_RAW
    local TF_FILE
    local TF_URL
    local THIS_DIR

    OS_PLATFORM_RAW="$(uname -s)"
    OS_PLATFORM="${OS_PLATFORM_RAW,,}"

    TF_FILE="terraform_${TF_VERSION}_${OS_PLATFORM}_amd64.zip"
    TF_URL="https://releases.hashicorp.com/terraform"

    if [ "$(command -v terraform || true)" == "" ] ; then
        curl -Ssl \
            "${TF_URL}/${TF_VERSION}/${TF_FILE}" \
            -o "${E_BIN}/${TF_FILE}"

        THIS_DIR="$(pwd)"
        cd "${E_BIN}"
        unzip "${TF_FILE}"
        cd "${THIS_DIR}"
    fi
}

#
# terraform_plan
#
function terraform_plan {
    local TFPLAN
    TFPLAN="${E_ANSIBLE_PLAYBOOK} -vv playbooks/deploy_infra.yml --check"

    echo ""
    echo "Running Terraform Plan (via Ansible Playbook)"
    echo -e "\t${TFPLAN}"
    echo ""

    bash -c "${TFPLAN}"

    while true; do
        echo ""
        echo "WARNING: The above will incur cost in Azure."
        read -r -p "Does the above plan look acceptable [y/n]: " ASK_ACCEPTABLE
        case ${ASK_ACCEPTABLE} in
            [Yy]*)
                echo "Continue..."
                break
                ;;
            [Nn]*)
                echo "Aborting."
                exit 0
                ;;
            *)
                echo "Please answer yes or no."
                ;;
        esac
    done
}

#
# terraform_apply
#
function terraform_apply {
    run_ansible_playbook "deploy_infra"
}

#
# terraform_destroy
#
function terraform_destroy {
    while true; do
        echo ""
        echo "WARNING: This will destroy all project resources."
        read -r -p "Do you want to continue? [y/n]: " ASK_ACCEPTABLE
        case ${ASK_ACCEPTABLE} in
            [Yy]*)
                echo "Continue..."
                break
                ;;
            [Nn]*)
                echo "Aborting."
                exit 0
                ;;
            *)
                echo "Please answer yes or no."
                ;;
        esac
    done

    PRC=1
    while [ ${PRC} -gt 0 ] ; do
        set +euo pipefail
        run_ansible_playbook "destroy_infra"
        PRC="${?}"
    done
}

#
# Function to control the test action
#
function action_test_command {
    local OPTIONS
    OPTIONS="${1:-false}"

    case "${OPTIONS}" in
        controller)
            check_environment
            ;;
        go)
            check_go
            ;;
        config)
            check_config
            ;;
        *)
            echo ""
            echo "Available options:"
            echo ""
            echo -e "\tgo                   Run tests against the ./go script"
            echo -e "\tcontroller           Check your local controller"
            echo -e "\tconfig               Run tests against JSON files"
            echo ""
            exit 1
            ;;
    esac
}

#
# Function to control the build action
#
function action_build_command {
    local OPTIONS
    OPTIONS="${1:-false}"

    case "${OPTIONS}" in
        controller)
            build_environment
            ;;
        infra)
            terraform_plan
            terraform_apply
            run_ansible_playbook "inventory_setup"
            ;;
        k3s_cluster)
            run_ansible_playbook "deploy_k3s"
            ;;
        *)
            echo ""
            echo "Available options:"
            echo ""
            echo -e "\tcontroller             Build Terraform Controller"
            echo ""
            exit 1
            ;;
    esac
}

#
# Function to control the destroy action
#
function action_destroy_command {
    local OPTIONS
    OPTIONS="${1:-false}"

    check_environment

    case "${OPTIONS}" in
        infra)
            terraform_destroy
            ;;
        k3s_cluster)
            run_ansible_playbook "destroy_k3s"
            ;;
        *)
            echo ""
            echo "Available options:"
            echo ""
            echo -e "\tk3s_cluster            Destroy K3S cluster"
            echo -e "\tinfra                  Destroy infrastructure in Azure"
            echo ""
            exit 1
            ;;
    esac
}

#
# Function to control the cleanup action
#
function action_cleanup_command {
    local OPTIONS
    OPTIONS="${1:-false}"

    case "${OPTIONS}" in
        controller)
            while true; do
                echo ""
                echo "WARNING: This will reset the working directory."
                echo "         Please ensure you have destroyed your infra"
                echo "         before running this else you may lose"
                echo "         Terraform state and cluster config."
                echo ""
                read -r -p "Do you want to continue? [y/n]: " ASK_ACCEPTABLE
                case ${ASK_ACCEPTABLE} in
                    [Yy]*)
                        echo "Continue..."
                        break
                        ;;
                    [Nn]*)
                        echo "Aborting."
                        exit 0
                        ;;
                    *)
                        echo "Please answer yes or no."
                        ;;
                esac
            done
            # shellcheck source=/dev/null
            test -d "${E_VENV}" && rm -rf "${E_VENV}"
            test -d "${E_BIN}" && rm -rf "${E_BIN}"
            git reset HEAD --hard
            ;;
        *)
            echo ""
            echo "Available options:"
            echo ""
            echo -e "\tcontroller          Reset working directory to default"
            echo ""
            exit 1
            ;;
    esac
}

#
# Action selector function, used to call Ansible/Terraform
#
function select_action {
    local ACTION
    local OPTIONS
    ACTION="${1}"
    OPTIONS="${2:-false}"

    case "${ACTION}" in
        test)
            action_test_command "${OPTIONS}"
            ;;
        build)
            action_build_command "${OPTIONS}"
            ;;
        destroy)
            action_destroy_command "${OPTIONS}"
            ;;
        cleanup)
            action_cleanup_command "${OPTIONS}"
            ;;
        *)
            usage
            exit 1
            ;;
    esac
}

#
# Main function
#
function main {
    local ACTION
    ACTION="${1:-false}"

    if [ "${ACTION}" == "false" ] ; then
        usage
        exit 0
    fi

    select_action "${@}"
}

##
# INVOCATION
##

if [[ "${BASH_SOURCE[0]}" = "${0}" ]] ; then
    main "${@}"
fi
