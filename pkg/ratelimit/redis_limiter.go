package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// RedisRateLimiter Redis 기반 Rate Limiter
type RedisRateLimiter struct {
	client    *redis.Client
	keyPrefix string
	logger    *logrus.Logger
}

// LimitConfig Rate Limiting 설정
type LimitConfig struct {
	MaxRequests int           // 허용된 최대 요청 수
	Window      time.Duration // 시간 윈도우
	BurstSize   int           // 버스트 허용 크기
}

// LimitResult Rate Limiting 결과
type LimitResult struct {
	Allowed     bool          // 허용 여부
	Remaining   int           // 남은 요청 수
	ResetTime   time.Time     // 리셋 시간
	RetryAfter  time.Duration // 재시도 대기 시간
}

// NewRedisRateLimiter Redis Rate Limiter 생성
func NewRedisRateLimiter(redisURL, keyPrefix string) (*RedisRateLimiter, error) {
	// Redis 연결 설정 파싱
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("Redis URL 파싱 실패: %v", err)
	}

	// Redis 클라이언트 생성
	client := redis.NewClient(opt)

	// 연결 테스트
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis 연결 실패: %v", err)
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	return &RedisRateLimiter{
		client:    client,
		keyPrefix: keyPrefix,
		logger:    logger,
	}, nil
}

// CheckLimit IP별 Rate Limiting 체크
func (r *RedisRateLimiter) CheckLimit(ctx context.Context, identifier string, config LimitConfig) (*LimitResult, error) {
	key := r.getKey(identifier)
	now := time.Now()
	windowStart := now.Truncate(config.Window)
	
	// Lua 스크립트로 원자적 처리
	luaScript := `
		local key = KEYS[1]
		local window_start = ARGV[1]
		local max_requests = tonumber(ARGV[2])
		local current_time = tonumber(ARGV[3])
		local window_seconds = tonumber(ARGV[4])
		
		-- 현재 윈도우의 카운터 키
		local window_key = key .. ":" .. window_start
		
		-- 현재 요청 수 조회
		local current_count = redis.call('GET', window_key)
		if not current_count then
			current_count = 0
		else
			current_count = tonumber(current_count)
		end
		
		-- Rate limit 체크
		if current_count >= max_requests then
			-- 제한 초과
			local reset_time = window_start + window_seconds
			return {0, current_count, reset_time}
		else
			-- 허용 - 카운터 증가
			local new_count = redis.call('INCR', window_key)
			
			-- TTL 설정 (윈도우 크기 + 여유시간)
			redis.call('EXPIRE', window_key, window_seconds + 60)
			
			local reset_time = window_start + window_seconds
			return {1, max_requests - new_count, reset_time}
		end
	`

	result, err := r.client.Eval(ctx, luaScript, []string{key}, 
		windowStart.Unix(),
		config.MaxRequests,
		now.Unix(),
		int(config.Window.Seconds()),
	).Result()

	if err != nil {
		r.logger.WithError(err).Error("Redis Lua 스크립트 실행 실패")
		return nil, err
	}

	// 결과 파싱
	results := result.([]interface{})
	allowed := results[0].(int64) == 1
	remaining := int(results[1].(int64))
	resetTime := time.Unix(results[2].(int64), 0)

	var retryAfter time.Duration
	if !allowed {
		retryAfter = time.Until(resetTime)
	}

	// 로깅
	r.logger.WithFields(logrus.Fields{
		"identifier": identifier,
		"allowed":    allowed,
		"remaining":  remaining,
		"reset_time": resetTime,
		"window":     config.Window,
	}).Info("Rate limit 체크 완료")

	return &LimitResult{
		Allowed:    allowed,
		Remaining:  remaining,
		ResetTime:  resetTime,
		RetryAfter: retryAfter,
	}, nil
}

// CheckBurstLimit 버스트 요청 체크 (더 짧은 시간 윈도우)
func (r *RedisRateLimiter) CheckBurstLimit(ctx context.Context, identifier string, config LimitConfig) (*LimitResult, error) {
	// 1초 윈도우로 버스트 체크
	burstConfig := LimitConfig{
		MaxRequests: config.BurstSize,
		Window:      time.Second,
	}
	
	key := r.getKey(identifier + ":burst")
	
	// 슬라이딩 윈도우 방식으로 버스트 체크
	now := time.Now()
	oneSecondAgo := now.Add(-time.Second)
	
	pipe := r.client.Pipeline()
	
	// 1초 이전 데이터 삭제
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(oneSecondAgo.UnixNano(), 10))
	
	// 현재 요청 수 카운트
	countCmd := pipe.ZCard(ctx, key)
	
	// 현재 시간을 스코어로 하는 멤버 추가
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: fmt.Sprintf("%d-%d", now.UnixNano(), time.Now().Nanosecond()),
	})
	
	// TTL 설정
	pipe.Expire(ctx, key, time.Second*2)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	
	currentCount := int(countCmd.Val())
	
	allowed := currentCount < config.BurstSize
	remaining := config.BurstSize - currentCount - 1
	if remaining < 0 {
		remaining = 0
	}
	
	return &LimitResult{
		Allowed:   allowed,
		Remaining: remaining,
		ResetTime: now.Add(time.Second),
	}, nil
}

// GetStats Rate Limiting 통계 조회
func (r *RedisRateLimiter) GetStats(ctx context.Context, identifier string) (map[string]interface{}, error) {
	key := r.getKey(identifier)
	
	// 패턴으로 모든 윈도우 키 조회
	pattern := key + ":*"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	
	stats := make(map[string]interface{})
	stats["identifier"] = identifier
	stats["total_windows"] = len(keys)
	
	if len(keys) > 0 {
		// 가장 최근 윈도우의 카운트 조회
		latestKey := keys[len(keys)-1]
		count, err := r.client.Get(ctx, latestKey).Int()
		if err != nil && err != redis.Nil {
			return nil, err
		}
		stats["current_count"] = count
		
		// TTL 조회
		ttl, err := r.client.TTL(ctx, latestKey).Result()
		if err != nil {
			return nil, err
		}
		stats["window_ttl"] = ttl.Seconds()
	}
	
	return stats, nil
}

// BlockIP IP 차단 (일정 시간)
func (r *RedisRateLimiter) BlockIP(ctx context.Context, ip string, duration time.Duration, reason string) error {
	key := r.getKey("blocked:" + ip)
	
	blockInfo := map[string]interface{}{
		"blocked_at": time.Now().Unix(),
		"duration":   duration.Seconds(),
		"reason":     reason,
	}
	
	err := r.client.HMSet(ctx, key, blockInfo).Err()
	if err != nil {
		return err
	}
	
	// TTL 설정
	err = r.client.Expire(ctx, key, duration).Err()
	if err != nil {
		return err
	}
	
	r.logger.WithFields(logrus.Fields{
		"ip":       ip,
		"duration": duration,
		"reason":   reason,
	}).Warn("IP 차단됨")
	
	return nil
}

// IsBlocked IP 차단 상태 확인
func (r *RedisRateLimiter) IsBlocked(ctx context.Context, ip string) (bool, string, error) {
	key := r.getKey("blocked:" + ip)
	
	result, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return false, "", err
	}
	
	if len(result) == 0 {
		return false, "", nil
	}
	
	reason, exists := result["reason"]
	if !exists {
		reason = "알 수 없는 이유"
	}
	
	return true, reason, nil
}

// getKey Redis 키 생성
func (r *RedisRateLimiter) getKey(identifier string) string {
	return fmt.Sprintf("%s:%s", r.keyPrefix, identifier)
}

// Close Redis 연결 종료
func (r *RedisRateLimiter) Close() error {
	return r.client.Close()
}