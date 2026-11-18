# Sub2API Kiro Docker Image

Sub2API Kiro is the Kiro-enhanced distribution of Sub2API. The published image supports GPT-5.6, Claude Opus 5, Claude Opus 4.8, and Claude Sonnet 5 through the Kiro integration.

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

- `latest` - Latest stable release
- `x.y.z` - Specific version
- `x.y` - Latest patch of minor version
- `x` - Latest minor of major version

## Links

- [GitHub Repository](https://github.com/tamakiramimy/sub2api-kiro)
- [Deployment Guide](./README.md)
