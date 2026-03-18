# ─── Stage 1: Builder ────────────────────────────────────────────────────────
FROM node:22-slim AS builder

# node-pty 컴파일에 필요한 빌드 도구
RUN apt-get update && apt-get install -y --no-install-recommends \
    python3 make gcc g++ libc6-dev \
    && rm -rf /var/lib/apt/lists/*

# pnpm 활성화
RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /app

# 의존성 먼저 복사 (레이어 캐시 최적화)
COPY package.json pnpm-lock.yaml ./
COPY scripts/ ./scripts/

# 의존성 설치 (node-pty 네이티브 빌드 포함)
RUN pnpm install --frozen-lockfile

# 소스 복사 후 Next.js 빌드
COPY . .
RUN pnpm build


# ─── Stage 2: Runner ─────────────────────────────────────────────────────────
FROM node:22-slim AS runner

# 런타임 필수 패키지 + Claude CLI 설치 도구
RUN apt-get update && apt-get install -y --no-install-recommends \
    python3 make gcc g++ libc6-dev \
    curl ca-certificates git \
    zsh \
    && rm -rf /var/lib/apt/lists/*

# pnpm 활성화
RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /app

# 빌드 결과물 복사
COPY --from=builder /app/.next       ./.next
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/public      ./public
COPY package.json pnpm-lock.yaml tsconfig.json next.config.ts ./
COPY server.ts ./
COPY scripts/ ./scripts/
COPY src/ ./src/

# Claude CLI 설치 (npm 글로벌)
RUN npm install -g @anthropic-ai/claude-code 2>/dev/null || true

# 환경변수 기본값
ENV NODE_ENV=production \
    PORT=4000 \
    HOSTNAME=0.0.0.0

EXPOSE 4000

# tsx 로 서버 실행 (Next.js + WebSocket 통합)
CMD ["node_modules/.bin/tsx", "server.ts"]
