# Multi-stage build for Go application
FROM golang:1.21-alpine AS builder

# 작업 디렉터리 설정
WORKDIR /app

# Go 모듈 파일 복사 및 의존성 다운로드
COPY go.mod go.sum ./
RUN go mod download

# 소스 코드 복사
COPY . .

# 바이너리 빌드 (CGO 비활성화로 Alpine에서 실행 가능)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o waf-gateway ./cmd/waf-gateway

# 최종 실행 이미지
FROM alpine:latest

# 보안을 위해 root가 아닌 사용자로 실행
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# 타임존 및 CA 인증서 설치
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# 빌드된 바이너리 복사
COPY --from=builder /app/waf-gateway .

# 사용자 변경
USER appuser

# 포트 노출
EXPOSE 8080

# 헬스체크 설정
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 실행 명령
CMD ["./waf-gateway"]