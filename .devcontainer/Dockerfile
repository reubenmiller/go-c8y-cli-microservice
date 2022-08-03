#-------------------------------------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation. All rights reserved.
# Licensed under the MIT License. See https://go.microsoft.com/fwlink/?linkid=2090316 for license information.
#-------------------------------------------------------------------------------------------------------------

FROM golang:1.18

# Avoid warnings by switching to noninteractive
ENV DEBIAN_FRONTEND=noninteractive

# This Dockerfile adds a non-root user with sudo access. Use the "remoteUser"
# property in devcontainer.json to use it. On Linux, the container user's GID/UIDs
# will be updated to match your local UID/GID (when using the dockerFile property).
# See https://aka.ms/vscode-remote/containers/non-root-user for details.
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

# Configure apt, install packages and tools
#wget https://packages.microsoft.com/config/debian/10/packages-microsoft-prod.deb \
#&& dpkg -i packages-microsoft-prod.deb \
#&& curl -sL https://deb.nodesource.com/setup_14.x | bash - \
RUN echo "Installing dependencies" \
    #
    # Update and install packages
    #
    && apt-get update \
    #
    && apt-get -y install --no-install-recommends apt-utils dialog 2>&1 \
    #
    # Verify git, process tools, lsb-release (common in install instructions for CLIs) installed
    && apt-get -y install git openssh-client less iproute2 procps lsb-release gnupg \
    #
    # Build Go tools w/module support
    && mkdir -p /tmp/gotools \
    && cd /tmp/gotools \
    && GOPATH=/tmp/gotools go install -v golang.org/x/tools/gopls@latest 2>&1 \
    && GOPATH=/tmp/gotools go install -v honnef.co/go/tools/cmd/staticcheck@latest \
    && GOPATH=/tmp/gotools go install -v golang.org/x/tools/cmd/gorename@latest \
    && GOPATH=/tmp/gotools go install -v golang.org/x/tools/cmd/goimports@latest \
    && GOPATH=/tmp/gotools go install -v golang.org/x/tools/cmd/guru@latest \
    && GOPATH=/tmp/gotools go install -v golang.org/x/lint/golint@latest \
    && GOPATH=/tmp/gotools go install -v github.com/mdempsky/gocode@latest \
    && GOPATH=/tmp/gotools go install -v github.com/haya14busa/goplay/cmd/goplay@latest \
    && GOPATH=/tmp/gotools go install -v github.com/sqs/goreturns@latest \
    && GOPATH=/tmp/gotools go install -v github.com/josharian/impl@latest \
    && GOPATH=/tmp/gotools go install -v github.com/davidrjenni/reftools/cmd/fillstruct@latest \
    && GOPATH=/tmp/gotools go install -v github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest  \
    && GOPATH=/tmp/gotools go install -v github.com/ramya-rao-a/go-outline@latest  \
    && GOPATH=/tmp/gotools go install -v github.com/acroca/go-symbols@latest  \
    && GOPATH=/tmp/gotools go install -v github.com/godoctor/godoctor@latest  \
    && GOPATH=/tmp/gotools go install -v github.com/rogpeppe/godef@latest  \
    && GOPATH=/tmp/gotools go install -v github.com/zmb3/gogetdoc@latest \
    && GOPATH=/tmp/gotools go install -v github.com/fatih/gomodifytags@latest  \
    && GOPATH=/tmp/gotools go install -v github.com/mgechev/revive@latest  \
    && GOPATH=/tmp/gotools go install -v github.com/go-delve/delve/cmd/dlv@latest 2>&1 \
    #
    # Install Go tools
    && mv /tmp/gotools/bin/* /usr/local/bin/ \
    #
    # Install golangci-lint
    && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin 2>&1 \
    #
    # Install go task runner
    && go install github.com/go-task/task/v3/cmd/task@latest \
    #
    # Create a non-root user to use if preferred - see https://aka.ms/vscode-remote/containers/non-root-user.
    && groupadd --gid $USER_GID $USERNAME \
    && useradd -s /bin/bash --uid $USER_UID --gid $USER_GID -m $USERNAME \
    # [Optional] Add sudo support
    && apt-get install -y sudo \
    && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME \
    #
    # Install Docker CE cli
    && apt-get install -y apt-transport-https ca-certificates curl gnupg-agent software-properties-common lsb-release \
    && curl -fsSL https://download.docker.com/linux/$(lsb_release -is | tr '[:upper:]' '[:lower:]')/gpg | apt-key add - 2>/dev/null \
    && add-apt-repository "deb [arch=$(dpkg --print-architecture)] https://download.docker.com/linux/$(lsb_release -is | tr '[:upper:]' '[:lower:]') $(lsb_release -cs) stable" \
    && apt-get update \
    && apt-get install -y docker-ce-cli \
    #
    # Install Docker Compose
    && curl -sSL "https://github.com/docker/compose/releases/download/1.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose \
    && chmod +x /usr/local/bin/docker-compose \
    #
    # Install shell tools
    && apt-get -y install --no-install-recommends \
        zsh \
        jq \
        zip \
        bash-completion \
    #
    # Clean up
    && apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/* /tmp/gotools

# Install other tooling (as default user)
USER $USERNAME
RUN curl -sSfL https://raw.githubusercontent.com/robbyrussell/oh-my-zsh/master/tools/install.sh | sh -s -- "" --unattended 2>&1

USER root
RUN curl https://reubenmiller.github.io/go-c8y-cli-repo/debian/PUBLIC.KEY | gpg --dearmor | tee /usr/share/keyrings/go-c8y-cli-archive-keyring.gpg >/dev/null \
    && sh -c "echo 'deb [signed-by=/usr/share/keyrings/go-c8y-cli-archive-keyring.gpg] http://reubenmiller.github.io/go-c8y-cli-repo/debian stable main' >> /etc/apt/sources.list" \
    #
    # task completions
    && curl https://raw.githubusercontent.com/sawadashota/go-task-completions/master/_task -o /usr/local/share/zsh/site-functions/_task \
    #
    # go-c8y-cli
    && apt-get update \
    && apt-get -y install --no-install-recommends  \
        go-c8y-cli \
    #
    # Load go-c8y-cli plugin
    && echo 'source "/etc/go-c8y-cli/shell/c8y.plugin.zsh"' >> /home/$USERNAME/.zshrc \
    #
    # Change default shell
    && usermod -s /usr/bin/zsh $USERNAME \
    #
    # Fix go permissions
    && chown -R ${USERNAME}:${USERNAME} /go \
    #
    # Add user to docker group
    && groupadd docker \
    && usermod -aG docker $USERNAME \
    #
    # Clean up
    && apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*


# Update this to "on" or "off" as appropriate
ENV GO111MODULE=auto
ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8

# Switch back to dialog for any ad-hoc use of apt-get
ENV DEBIAN_FRONTEND=dialog

USER $USERNAME
