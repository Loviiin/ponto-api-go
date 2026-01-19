FROM mcr.microsoft.com/devcontainers/go:1.25-bookworm

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt,sharing=locked \
	apt-get update && apt-get install -y curl jq
