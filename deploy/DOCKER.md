# Sub2API Kiro Docker Image

Sub2API Kiro is the Kiro-enhanced distribution of Sub2API. The published image supports GPT-5.6, Claude Opus 5, Claude Opus 4.8, and Claude Sonnet 5 through the Kiro integration, on a Wei-Shaw 0.2.25 baseline.

Kiro Claude protocol support includes implicit `<thinking>` tag stripping for Opus 4.7+/5+ and Sonnet 5+, forced `tool_choice`, structured output (`response_format`), `stop_sequences`, and automatic history trimming when a request exceeds the Kiro input limit. Kiro accounts are reachable through the Anthropic Messages, OpenAI Responses, and Chat Completions gateways.

## Image

```text
tamakiramimy/sub2api-kiro:latest
```

The `latest` tag is a multi-architecture image for `linux/amd64` and `linux/arm64`.

## Docker Compose

Docker Compose is the supported deployment path because it pulls the published Kiro image and configures the application, PostgreSQL, Redis, persistent storage, and required environment variables together.

```bash
git clone https://github.com/tamakiramimy/sub2api-kiro.git
cd sub2api-kiro/deploy
cp .env.example .env
chmod 600 .env
docker compose -f docker-compose.local.yml up -d
```

Use `docker-compose.local.yml` for local data directories that are easier to back up and migrate. To use an image from a private registry, set `SUB2API_IMAGE` before running Docker Compose; it overrides `tamakiramimy/sub2api-kiro:latest`. See [README.md](./README.md) for environment variables, upgrades, and operational commands.

## Supported Architectures

- `linux/amd64`
- `linux/arm64`

## Tags

- `latest` - Latest release, a multi-architecture manifest
- `YYYYMMDD_HHMMSS` - A specific build, for example `20261126_151231`, a multi-architecture manifest
- `amd64-YYYYMMDD_HHMMSS`, `arm64-YYYYMMDD_HHMMSS` - The single-architecture images of that build
- `amd64-latest`, `arm64-latest` - The single-architecture images of the latest build

## Links

- [GitHub Repository](https://github.com/tamakiramimy/sub2api-kiro)
- [Deployment Guide](./README.md)
