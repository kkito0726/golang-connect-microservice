FROM mcr.microsoft.com/devcontainers/go:1

# Docker CLI (docker-outside-of-docker)
RUN apt-get update && apt-get install -y --no-install-recommends \
      docker.io \
    && rm -rf /var/lib/apt/lists/*

# Node.js LTS
RUN curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

# claude インストーラーが ~/.local/bin に配置することがあるため PATH に追加
ENV PATH="/root/.local/bin:${PATH}"
