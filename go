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

    JSON_FILE_LIST="$(find . -iname '*.json' -type f)"

    for JSON_FILE in ${JSON_FILE_LIST} ; do
        echo "Testing ${JSON_FILE}..."
        jq --slurp . "${JSON_FILE}" -Mc 1>/dev/null 2> >(tee -a .fail >&2) && \
            echo "OK"
        echo ""
    done

    if [ -f .fail ] && [ "$(<.fail)" != "" ] ; then
        echo "JSON syntax checking failed"
        test -f .fail && rm -f .fail
        exit 1
    else
        echo "Tests pass"
    fi

    test -f .fail && rm -f .fail

    set -eu
}

#
# check_terraform_validity
#   Checks Terraform is valid
#
function check_terraform_validity {
    TF_ENV_LIST="$(find terraform/env/ -maxdepth 1 \
        -type d -not -path terraform/env/)"

    for TF_ENV in ${TF_ENV_LIST} ; do
        TF_ENV_NAME="${TF_ENV/terraform\/env\//}"
        echo "Testing environment/module validity: ${TF_ENV_NAME}..."
        terraform_validate "${TF_ENV_NAME}"
    done
}

#
# check_terraform_style
#   Checks Terraform is valid
#
function check_terraform_style {
    TF_ENV_LIST="$(find terraform/env/ -maxdepth 1 \
        -type d -not -path terraform/env/)"

    TF_MODULE_LIST="$(find terraform/modules/ -maxdepth 1 \
        -type d -not -path terraform/modules/)"

    for TF_ENV in ${TF_ENV_LIST} ; do
        TF_ENV_NAME="${TF_ENV/terraform\/env\//}"
        echo "Testing environment style: ${TF_ENV_NAME}..."
        terraform_fmt "${TF_ENV}" | tee -a .fail
    done

    for TF_MOD in ${TF_MODULE_LIST} ; do
        TF_MOD_NAME="${TF_MOD/terraform\/modules\//}"
        echo "Testing modules style: ${TF_MOD_NAME}..."
        terraform_fmt "${TF_MOD}" | tee -a .fail
    done

    FAILCOUNT="$(grep ".tf" .fail || true)"
    test -f .fail && rm -f .fail

    if [ "${FAILCOUNT}" != "" ] ; then
        echo "Failed"
        exit 1
    else
        echo "Passed"
        exit 0
    fi
}

#
# check_idempotence
#   Checks Terraform is idempotent
#
function check_idempotence {
    TF_ENV_LIST="$(find terraform/env/ -maxdepth 1 \
        -type d -not -path terraform/env/)"

    E_SKIP_PROMPT="true"

    for TF_ENV in ${TF_ENV_LIST} ; do
        TF_ENV_NAME="${TF_ENV/terraform\/env\//}"
        echo "Testing idempotence: ${TF_ENV_NAME}..."
        terraform_get "${TF_ENV_NAME}"
        terraform_init "${TF_ENV_NAME}"
        terraform_plan "${TF_ENV_NAME}"
        terraform_apply "${TF_ENV_NAME}"
        terraform_plan "${TF_ENV_NAME}"
        terraform_apply "${TF_ENV_NAME}" | tee -a .fail
        terraform_destroy "${TF_ENV_NAME}"
    done

    FAILI="$(grep "Resources: 0 added, 0 changed, 0 destroyed" .fail || true)"
    test -f .fail && rm -f .fail

    if [ "${FAILI}" == "" ] ; then
        echo "Failed"
        exit 1
    else
        echo "Passed"
        exit 0
    fi
}

#
# confirm_plan
#   Asks the user if it is OK to continue with the plan
#
function confirm_plan {
    [ "${E_SKIP_PROMPT}" == "true" ] && return

    while true; do
        echo ""
        echo "WARNING: The above will incur cost."
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
# check_atlas_login
#   Determines if we have access to the API
#
function check_atlas_login {
    local VARTEST
    local VTESTFAIL
    local CURLTEST
    local ATLAS_VARS=(
        MONGODB_ATLAS_PUBLIC_KEY
        MONGODB_ATLAS_PRIVATE_KEY
        MONGODB_ATLAS_PROJECT_ID
    )

    VTESTFAIL="false"

    for AV in "${ATLAS_VARS[@]}" ; do
        VARTEST="${AV}"
        if [ -z "${!VARTEST:-}" ] ; then
            echo "${VARTEST} is not set!"
            VTESTFAIL="true"
        fi
    done

    if [ "${VTESTFAIL}" == "true" ] ; then
        echo "Authentication failure."
        exit 1
    fi

    CURLTEST="$(curl -s \
        --user "${MONGODB_ATLAS_PUBLIC_KEY}:${MONGODB_ATLAS_PRIVATE_KEY}" \
        --digest \
        --header "Accept: application/json" \
        --request GET \
        --include \
        https://cloud.mongodb.com/api/atlas/v1.0 | grep -E "HTTP.*200" || \
        echo "Failed to connect to MongoDB Atlas API!")"

    echo "${CURLTEST}"
    if [[ "${CURLTEST}" =~ ^Failed ]] ; then
        exit 1
    fi
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
# list_terraform_envs
#   Generates a list of Environments
#
function list_terraform_envs {
    local TF_ACTION
    local TF_ENV_LIST
    local TF_ENV_NAME

    TF_ACTION="${1:-build}"
    TF_ENV_LIST="$(find terraform/env/ -maxdepth 1 -type d)"

    for TF_ENV in ${TF_ENV_LIST} ; do
        TF_ENV_NAME="${TF_ENV/terraform\/env\//}"
        if [ "${TF_ENV_NAME}" != "" ] ; then
            echo -e "\t${TF_ENV_NAME}\t\t\t${TF_ACTION^}s" \
                "(${TF_ENV_NAME^^}) MongoDB cluster"
        fi
    done
}

#
# run_terraform_command
#   Executes a terraform command
#
function run_terraform_command {
    local TF_CMD
    TF_CMD="${1:-false}"

    if [ "${TF_CMD}" == "false" ] ; then
        echo "Unknown Terraform command!"
        exit 1
    fi

    if [[ "${TF_CMD}" =~ terraform*$ ]] ; then
        echo "Unknown Terraform command: ${TF_CMD}"
        exit 1
    fi

    echo "Running Terraform command:"
    echo ""
    echo -e "\t${TF_CMD}"
    echo ""
    echo ""

    /usr/bin/env bash -c "${TF_CMD}"
}

#
# terraform_get
#   Gets Terraform modules
#
function terraform_get {
    local TF_DIR
    local TF_CMD
    local TF_ENV
    TF_ENV="${1:-unknown}"
    TF_DIR="terraform/env/${TF_ENV}"
    TF_CMD="terraform get"

    if [ ! -d "${TF_DIR}" ] ; then
        echo "${TF_ENV} environment not found!"
        exit 1
    fi

    cd "${TF_DIR}"

    run_terraform_command "${TF_CMD}"

    cd - >/dev/null 2>&1
}

#
# terraform_validate
#   Validates Terraform modules
#
function terraform_validate {
    local TF_DIR
    local TF_CMD
    local TF_ENV
    TF_ENV="${1:-unknown}"
    TF_DIR="terraform/env/${TF_ENV}"
    TF_CMD="terraform validate"

    if [ ! -d "${TF_DIR}" ] ; then
        echo "${TF_ENV} environment not found!"
        exit 1
    fi

    terraform_init "${TF_ENV}"

    cd "${TF_DIR}"

    run_terraform_command "${TF_CMD}"

    cd - >/dev/null 2>&1
}

#
# terraform_fmt
#   Checks Terraform format
#
function terraform_fmt {
    local TF_DIR
    local TF_CMD
    TF_DIR="${1:-unknown}"
    TF_CMD="terraform fmt"

    if [ ! -d "${TF_DIR}" ] ; then
        echo "${TF_DIR} not found!"
        exit 1
    fi

    cd "${TF_DIR}"

    run_terraform_command "${TF_CMD}"

    cd - >/dev/null 2>&1
}

#
# terraform_init
#   Initializes a Terraform project
#
function terraform_init {
    local TF_DIR
    local TF_CMD
    local TF_ENV
    TF_ENV="${1:-unknown}"
    TF_DIR="terraform/env/${TF_ENV}"
    TF_CMD="terraform init"

    if [ ! -d "${TF_DIR}" ] ; then
        echo "${TF_ENV} environment not found!"
        exit 1
    fi

    cd "${TF_DIR}"

    run_terraform_command "${TF_CMD}"

    cd - >/dev/null 2>&1
}

#
# terraform_plan
#   Plans your terraform project
#
function terraform_plan {
    local TF_DIR
    local TF_CMD
    local TF_ENV

    TF_ENV="${1:-unknown}"
    TF_DIR="terraform/env/${TF_ENV}"
    TF_CMD="terraform plan -out project.tfplan -var 'env_id=${TF_ENV}'" \
    TF_CMD+=" -var 'json_config=${TF_JSON_CONFIG}'"
    TF_CMD+=" -var 'project_id=${MONGODB_ATLAS_PROJECT_ID}'"

    if [ ! -d "${TF_DIR}" ] ; then
        echo "${TF_ENV} environment not found!"
        exit 1
    fi

    cd "${TF_DIR}"

    run_terraform_command "${TF_CMD}"

    cd - >/dev/null 2>&1
}

#
# terraform_apply
#   Runs terraform apply against your project
#
function terraform_apply {
    local TF_DIR
    local TF_CMD
    local TF_ENV

    TF_ENV="${1:-unknown}"
    TF_DIR="terraform/env/${TF_ENV}"
    TF_CMD="terraform apply project.tfplan"

    if [ ! -d "${TF_DIR}" ] ; then
        echo "${TF_ENV} environment not found!"
        exit 1
    fi

    cd "${TF_DIR}"

    run_terraform_command "${TF_CMD}"

    cd - >/dev/null 2>&1
}

#
# terraform_destroy
#   Destroys your terraform project
#
function terraform_destroy {
    local TF_DIR
    local TF_CMD
    local TF_ENV
    local TF_APPROVE

    TF_APPROVE=""

    [ "${E_SKIP_PROMPT}" == "true" ] && TF_APPROVE=" -auto-approve"

    TF_ENV="${1:-unknown}"
    TF_DIR="terraform/env/${TF_ENV}"
    TF_CMD="terraform apply -destroy -var 'env_id=${TF_ENV}'" \
    TF_CMD+=" -var 'json_config=${TF_JSON_CONFIG}'"
    TF_CMD+=" -var 'project_id=${MONGODB_ATLAS_PROJECT_ID}'"
    TF_CMD+="${TF_APPROVE}"

    if [ ! -d "${TF_DIR}" ] ; then
        echo "${TF_ENV} environment not found!"
        exit 1
    fi

    cd "${TF_DIR}"

    run_terraform_command "${TF_CMD}"

    cd - >/dev/null 2>&1
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
        atlas_login)
            check_atlas_login
            ;;
        terraform_validity)
            check_terraform_validity
            ;;
        terraform_style)
            check_terraform_style
            ;;
        idempotence)
            check_idempotence
            ;;
        *)
            echo ""
            echo "Available options:"
            echo ""
            echo -e "\tgo                   Run tests against the ./go script"
            echo -e "\tcontroller           Check your local controller"
            echo -e "\tconfig               Run tests against JSON files"
            echo -e "\tatlas_login          Check you can log in to Atlas"
            echo -e "\tterraform_validity   Run validity checks"
            echo -e "\tterraform_style      Run format checks"
            echo -e "\tidempotence          Check that the code is idempotent"
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
        false)
            echo ""
            echo "Available options:"
            echo ""
            echo -e "\tcontroller\t\tBuild the Terraform Controller"
            list_terraform_envs "build"
            echo ""
            exit 1
            ;;
        *)
            terraform_get "${OPTIONS}"
            terraform_init "${OPTIONS}"
            terraform_plan "${OPTIONS}"
            confirm_plan
            terraform_apply "${OPTIONS}"
            ;;
    esac
}

#
# Function to control the destroy action
#
function action_destroy_command {
    local OPTIONS
    OPTIONS="${1:-false}"

    case "${OPTIONS}" in
        false)
            echo ""
            echo "Available options:"
            echo ""
            list_terraform_envs "destroy"
            echo ""
            exit 1
            ;;
        *)
            terraform_destroy "${OPTIONS}"
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
            for TFTYPE in tfplan tfstate tfstate.backup lock.hcl ; do
                find terraform/ -type f -name "*.${TFTYPE}" -delete
            done
            find terraform/ \
                -maxdepth 3 \
                -type d \
                -name ".terraform" \
                -exec rm -rf {} \;
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
