FROM mcr.microsoft.com/devcontainers/go:dev-1.21

# Tools like sonar need java :(
RUN apt-get update \
    && apt-get -y install default-jre netcat-openbsd \
    && apt-get clean

# Need to set user to vscode before doing go install
USER vscode

# When running locally in dev environment can default to human readable output
# ENV ET_LOGGER=human_readable

# Install go tools
RUN go install golang.org/x/pkgsite/cmd/pkgsite@latest