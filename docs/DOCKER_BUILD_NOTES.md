# Docker 镜像构建注意事项

本文记录本仓库构建和发布 Docker 镜像时的网络、多架构与缓存注意事项。不要在此文档、Shell 历史或 Dockerfile 中记录 Registry 密码或访问令牌。

## 镜像与标签约定

以内部 Quay 为例：

```sh
export IMAGE=quay.globalapp.mindray.com/prd/sub2api-kiro
export VERSION="$(tr -d '\n' < backend/cmd/server/VERSION)"
```

发布时应同时保留架构专用标签和多架构标签：

- `${VERSION}-amd64`、`amd64`
- `${VERSION}-arm64`、`arm64`
- `${VERSION}`、`latest`（包含 AMD64 与 ARM64 的 OCI manifest）

使用密码标准输入登录，避免密码进入命令行参数：

```sh
printf '%s' "$REGISTRY_PASSWORD" | \
  docker login --username "$REGISTRY_USERNAME" --password-stdin quay.globalapp.mindray.com
unset REGISTRY_PASSWORD
```

## 代理与镜像源

构建依赖 Docker daemon 的代理配置，而不是 Git 的代理配置。开始构建前确认：

```sh
docker info --format 'HTTPProxy={{.HTTPProxy}}\nHTTPSProxy={{.HTTPSProxy}}\nNoProxy={{.NoProxy}}'
```

公司网络中，Docker daemon 可使用 `http://hkproxy.mindray.com:8080` 访问公共镜像与依赖源。内部 Registry 应位于 `NO_PROXY`，例如 `quay.globalapp.mindray.com`，使推送走内部直连。

公共 Alpine 源出现超时时，根 `Dockerfile` 支持可选的 `ALPINE_REPOSITORY` 参数。该参数默认为空，CI 和外部网络仍使用 Alpine 官方源；受限网络可在构建时传入：

```sh
--build-arg ALPINE_REPOSITORY=https://mirrors.aliyun.com/alpine
```

推荐同时使用可达的 npm 和 Go 模块镜像：

```sh
--build-arg NPM_CONFIG_REGISTRY=https://registry.npmmirror.com \
--build-arg GOPROXY=https://mirrors.aliyun.com/goproxy/,direct \
--build-arg GOSUMDB=off
```

`GOSUMDB=off` 仅适用于受限网络下的受控构建环境；可访问公共校验服务时应保留默认值。

## 多架构构建

首次构建会下载前端依赖、Go 模块和基础镜像，通常需要数分钟。使用 `--progress=plain` 可以看到实际停留的阶段。

Go 编译缓存必须按目标架构分开。根 `Dockerfile` 已使用：

```dockerfile
--mount=type=cache,id=sub2api-gobuild-${TARGETARCH},target=/root/.cache/go-build
```

不要把 AMD64 和 ARM64 的 Go 构建产物写入同一个 BuildKit cache ID；并发构建时可能长期停在 `go build` 阶段。模块下载缓存 `sub2api-gomod` 可以共享。

为便于排障和稳定发布，推荐按架构顺序构建，再合并远端 manifest：

```sh
docker buildx build --progress=plain --platform linux/amd64 --push \
  --build-arg VERSION="$VERSION" \
  --build-arg COMMIT="$(git rev-parse --short HEAD)" \
  --build-arg ALPINE_REPOSITORY=https://mirrors.aliyun.com/alpine \
  --build-arg NPM_CONFIG_REGISTRY=https://registry.npmmirror.com \
  --build-arg GOPROXY=https://mirrors.aliyun.com/goproxy/,direct \
  --build-arg GOSUMDB=off \
  --tag "${IMAGE}:${VERSION}-amd64" --tag "${IMAGE}:amd64" .

docker buildx build --progress=plain --platform linux/arm64 --push \
  --build-arg VERSION="$VERSION" \
  --build-arg COMMIT="$(git rev-parse --short HEAD)" \
  --build-arg ALPINE_REPOSITORY=https://mirrors.aliyun.com/alpine \
  --build-arg NPM_CONFIG_REGISTRY=https://registry.npmmirror.com \
  --build-arg GOPROXY=https://mirrors.aliyun.com/goproxy/,direct \
  --build-arg GOSUMDB=off \
  --tag "${IMAGE}:${VERSION}-arm64" --tag "${IMAGE}:arm64" .

docker buildx imagetools create \
  --tag "${IMAGE}:${VERSION}" --tag "${IMAGE}:latest" \
  "${IMAGE}:${VERSION}-amd64" "${IMAGE}:${VERSION}-arm64"
```

## 发布后验证

验证多架构标签必须显示 `linux/amd64` 和 `linux/arm64`：

```sh
docker buildx imagetools inspect "${IMAGE}:${VERSION}"
docker buildx imagetools inspect "${IMAGE}:latest"
```

输出中的 `unknown/unknown` 条目通常是 BuildKit 生成的 provenance attestation，不是可拉取的运行时平台镜像。

## 常见问题

| 现象 | 原因 | 处理方式 |
| --- | --- | --- |
| `apk add` 找不到所有包，前面有 Alpine `APKINDEX` 超时 | 公共 Alpine 源不可达 | 传入 `ALPINE_REPOSITORY`，并确认镜像 URL 可访问 |
| 长时间停在 `go build` | 多架构任务争用 Go 编译缓存，或首次编译仍在进行 | 使用架构隔离的缓存 ID；先用 `--progress=plain` 确认，再按架构顺序构建 |
| Docker 登录成功但推送失败 | Registry 地址、权限或 `NO_PROXY` 配置错误 | 确认登录和镜像标签使用同一个 Registry 域名，并检查 daemon 的 `NoProxy` |
| 标签格式报 `invalid reference format` | zsh 解析 `$IMAGE:$VERSION` 时产生歧义 | 始终写成 `${IMAGE}:${VERSION}` |
