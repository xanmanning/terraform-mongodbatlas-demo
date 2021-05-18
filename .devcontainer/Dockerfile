# [Choice] Debian version: jessie, buster
ARG VARIANT=buster
FROM debian:${VARIANT}


RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install curl git python3-dev python3-pip \
        python3-venv shellcheck sudo unzip docker.io jq \
    && curl -L \
        "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" \
        -o /usr/bin/kubectl \
    && chmod +x /usr/bin/kubectl
RUN useradd -m vscode && \
    echo 'vscode ALL=(ALL:ALL) NOPASSWD: ALL' > /etc/sudoers.d/vscode && \
    echo 'source /etc/bash_completion.d/git-prompt' >> /home/vscode/.bashrc && \
    echo 'export PS1="${PS1}\[\033[38;5;196m\]$(__git_ps1)\[$(tput sgr0)\] "' >> /home/vscode/.bashrc
