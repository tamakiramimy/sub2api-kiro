# Sub2API Kiro Docker Image

Sub2API Kiro is the Kiro-enhanced distribution of Sub2API. It tracks stable upstream updates while retaining Kiro account support, including OAuth, AWS Builder ID, token import, and OpenAI Responses / Chat Completions protocol bridging.

Recent built-in model support includes:

- GPT-6 Astra through `gpt-6` and `gpt-6-astra`.
- GPT Image 2.5 through `gpt-image-2.5-flare` and `gpt-image-2.5-sunburst`.
- Claude Fable 5 and `claude-fable-5-1` (Fable 5.1), with adaptive reasoning effort support.
- Kiro compatibility aliases for GPT-5.6, Claude Opus 4.8, and Claude Sonnet 5.

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

Use `docker-compose.local.yml` for local data directories that are easier to back up and migrate. To use an image from a private registry, set `SUB2API_IMAGE` before running Docker Compose; it overrides `tamakiramimy/sub2api-kiro:latest`. See [README.md](https://github.com/tamakiramimy/sub2api-kiro#readme) for environment variables, upgrades, and operational commands.

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
