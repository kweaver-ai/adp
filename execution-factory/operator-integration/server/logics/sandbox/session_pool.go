package sandbox

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/drivenadapters"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/telemetry"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

const (
	sessionIDPrefix           = "sess_aoi_"
	defaultMaxSessions        = 3
	defaultMaxConcurrentTasks = 100
	defaultActiveSessions     = 1
	defaultContextTimeout     = 30 * time.Second
	// 最大重试次数
	maxRetryCount = 3
	// 会话运行状态检查间隔
	sessionStatusRunningCheckInterval = time.Second
	// 等待会话运行超时
	waitSessionRunningTimeout = 30 * time.Second
	// 后台工作器间隔
	backgroundWorkerInterval = time.Minute
)

// SessionPool 会话池接口
type SessionPool interface {
	ExecuteCode(ctx context.Context, req *interfaces.ExecuteCodeReq) (*interfaces.ExecuteCodeResp, error)
}

type sessionItem struct {
	ID           string
	RunningTasks int
	LastUsedAt   time.Time
}

// 添加会话到池
func (p *sessionPoolImpl) addSession(sessionID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sessions[sessionID] = &sessionItem{
		ID:           sessionID,
		RunningTasks: 0,
		LastUsedAt:   time.Now(),
	}
}

// getSessionItem 获取会话项
func (p *sessionPoolImpl) getSessionItem(sessionID string) (sessionItem *sessionItem, ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	sessionItem, ok = p.sessions[sessionID]
	return
}

// 删除会话
func (p *sessionPoolImpl) removeSession(sessionID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.sessions, sessionID)
}

// 更新运行任务数
// updateRunningTasks 更新会话运行任务数
func (p *sessionPoolImpl) updateRunningTasks(sessionID string, delta int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if item, exists := p.sessions[sessionID]; exists {
		item.RunningTasks += delta
		item.LastUsedAt = time.Now()
	}
}

// findBestSession 寻找最佳会话: 堆叠分配策略：寻找负载最高但未满的会话
func (p *sessionPoolImpl) findBestSession() (bestSession *sessionItem) {
	p.mu.Lock()
	// 1. 堆叠分配策略：寻找负载最高但未满的会话
	sessionIDs := []string{}
	for _, item := range p.sessions {
		p.logger.Infof("Session %s: RunningTasks=%d, LastUsedAt=%v\n", item.ID, item.RunningTasks, item.LastUsedAt)
		if item.RunningTasks < p.maxConcurrentTasks {
			if bestSession == nil || item.RunningTasks > bestSession.RunningTasks {
				// 检查任务是否可用
				exists, _, _ := p.client.QuerySession(context.Background(), item.ID)
				if !exists {
					sessionIDs = append(sessionIDs, item.ID)
					continue
				}
				bestSession = item
			}
		}
	}
	p.mu.Unlock()
	// 删除所有无效会话
	for _, sessionID := range sessionIDs {
		p.removeSession(sessionID)
	}
	return bestSession
}

type sessionPoolImpl struct {
	client             interfaces.SandBoxControlPlane
	sessions           map[string]*sessionItem // key: sessionID
	mu                 sync.Mutex
	maxSessions        int
	maxConcurrentTasks int
	activeSessions     int
	logger             interfaces.Logger
	stopCh             chan struct{}
	templateID         string
	reqConfig          config.SessionResourcesConfig
}

var (
	poolInstance *sessionPoolImpl
	poolOnce     sync.Once
)

// GetSessionPool 获取会话池实例
func GetSessionPool() SessionPool {
	poolOnce.Do(func() {
		conf := config.NewConfigLoader()
		client := drivenadapters.NewSandBoxControlPlaneClient()
		maxConcurrentTasks := conf.SandboxControlPlane.MaxConcurrentTasks
		if maxConcurrentTasks <= 0 {
			maxConcurrentTasks = defaultMaxConcurrentTasks
		}
		maxSessions := conf.SandboxControlPlane.MaxSessions
		if maxSessions <= 0 {
			maxSessions = defaultMaxSessions
		}
		activeSessions := conf.SandboxControlPlane.ActiveSessions
		if activeSessions <= 0 {
			activeSessions = defaultActiveSessions
		} else if activeSessions > maxSessions {
			activeSessions = maxSessions
		}

		poolInstance = &sessionPoolImpl{
			client:             client,
			sessions:           make(map[string]*sessionItem),
			maxSessions:        maxSessions,
			maxConcurrentTasks: maxConcurrentTasks,
			activeSessions:     activeSessions,
			logger:             conf.GetLogger(),
			stopCh:             make(chan struct{}),
			templateID:         conf.SandboxControlPlane.TemplateID,
			reqConfig:          conf.SandboxControlPlane.SessionResources,
		}
		// 打印配置信息
		poolInstance.logger.Infof("SessionPool initialized with maxSessions: %d, maxConcurrentTasks: %d, activeSessions: %d, templateID: %s, sessionResources: %v",
			poolInstance.maxSessions, poolInstance.maxConcurrentTasks, poolInstance.activeSessions, poolInstance.templateID, poolInstance.reqConfig)

		// 初始化：从控制平面同步已存在的确定性会话，并补足到 activeSessions 数量
		poolInstance.initSessions()

		// 启动后台管理任务：健康检查与空闲缩容/预热
		go poolInstance.backgroundWorker()
	})
	return poolInstance
}

func (p *sessionPoolImpl) initSessions() {
	ctx := context.Background()
	recoveredCount := 0
	for i := 0; i < p.maxSessions; i++ {
		id := fmt.Sprintf("%s%d", sessionIDPrefix, i)
		// 检查会话是否存在且状态为 Running
		exists, detail, err := p.client.QuerySession(ctx, id)
		if err == nil && exists && detail != nil && detail.Status == interfaces.SessionStatusRunning {
			poolInstance.addSession(id)
			recoveredCount++
		}
	}
	p.logger.Infof("Recovered %d sessions during initialization", recoveredCount)

	// 初始预热，补足到 activeSessions
	p.prewarmSessions()
}

func (p *sessionPoolImpl) ExecuteCode(ctx context.Context, req *interfaces.ExecuteCodeReq) (resp *interfaces.ExecuteCodeResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"language": req.Language,
		"timeout":  req.Timeout,
		"code":     req.Code,
		"event":    req.Event,
	})
	sessionID, err := p.acquireSession(ctx, maxRetryCount)
	if err != nil {
		return nil, err
	}
	defer p.releaseSession(sessionID)
	resp, err = p.client.ExecuteCodeSync(ctx, sessionID, req)
	if err != nil {
		p.logger.WithContext(ctx).Errorf("ExecuteCodeSync failed for session %s: %v", sessionID, err)
		return nil, err
	}
	return resp, nil
}

// acquireSession 从会话池中获取一个会话
func (p *sessionPoolImpl) acquireSession(ctx context.Context, retryCount int) (sessionID string, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"retryCount": retryCount,
	})
	// 是否需要重试
	var needRetry bool
	defer func(count int) {
		if !needRetry { // 不需要重试
			return
		}
		// 重试次数达到上限
		if count < 0 {
			err = fmt.Errorf("[acquireSession] retryCount %d exceeds maxRetryCount %d", count, maxRetryCount)
			return
		}
		// 暂停时间: 每次重试间隔增加 1 秒
		time.Sleep(time.Duration(count) * time.Second)
		sessionID, err = p.acquireSession(ctx, count-1)
	}(retryCount)
	// 1. 堆叠分配策略：寻找负载最高但未满的会话
	bestSession := p.findBestSession()
	if bestSession != nil {
		p.updateRunningTasks(bestSession.ID, 1)
		sessionID = bestSession.ID
		return
	}

	// 2. 尝试寻找可创建的槽位
	var targetID string
	for i := 0; i < p.maxSessions; i++ {
		id := fmt.Sprintf("%s%d", sessionIDPrefix, i)
		if _, ok := p.getSessionItem(id); !ok {
			targetID = id
			break
		}
	}

	// 3. 如果所有槽位都有 Session，但都满了（因为步骤1没找到），则报错
	if targetID == "" {
		if retryCount == 0 {
			return "", fmt.Errorf("all %d sessions are at max concurrency (%d)", p.maxSessions, p.maxConcurrentTasks)
		}
		// 递归重试：如果当前 ID 创建失败，递归尝试下一个可用 ID
		needRetry = true
		return
	}

	// 5. 执行远程创建
	p.logger.Infof("Creating new session slot: %s", targetID)
	if err = p.ensureRemoteSession(ctx, targetID); err != nil {
		p.logger.Errorf("Failed to create session %s: %v", targetID, err)
		// 创建失败，移除占位符
		// 容错重试：如果当前 ID 创建失败，递归尝试下一个可用 ID
		// 注意：需要先清理当前失败的占位
		p.removeSession(targetID) // 清理占位符（兜底）
		// 尝试重试
		needRetry = true
		return
	}
	return targetID, nil
}

func (p *sessionPoolImpl) ensureRemoteSession(ctx context.Context, sessionID string) error {
	// 创建前检查是否存在
	exists, _, err := p.client.QuerySession(ctx, sessionID)
	if err != nil {
		p.logger.Errorf("QuerySession failed for session %s: %v", sessionID, err)
		return err
	}
	if !exists {
		// 执行创建
		req := &interfaces.CreateSessionReq{
			ID:         sessionID,
			TemplateID: p.templateID,
			Timeout:    p.reqConfig.Timeout,
			CPU:        p.reqConfig.CPU,
			Memory:     p.reqConfig.Memory,
			Disk:       p.reqConfig.Disk,
		}

		_, err := p.client.CreateSession(ctx, req)
		if err != nil {
			p.logger.Warnf("[ensureRemoteSession] Failed to create session %s: %v", sessionID, err)
			return err
		}
	}

	// 等待 Running 状态
	err = p.waitForSessionRunning(ctx, sessionID)
	if err != nil {
		return err
	}
	p.addSession(sessionID)
	return nil
}

func (p *sessionPoolImpl) waitForSessionRunning(ctx context.Context, sessionID string) error {
	ticker := time.NewTicker(sessionStatusRunningCheckInterval)
	defer ticker.Stop()
	timeout := time.After(waitSessionRunningTimeout)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for session %s to be running", sessionID)
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			exists, detail, err := p.client.QuerySession(ctx, sessionID)
			if err != nil {
				p.logger.Errorf("QuerySession failed for session %s: %v", sessionID, err)
				return err
			}
			if !exists {
				// 会话创建失败
				return fmt.Errorf("session %s failed to create, not found", sessionID)
			}
			switch detail.Status {
			case interfaces.SessionStatusRunning:
				return nil // 会话已运行，成功
			case interfaces.SessionStatusFailed, interfaces.SessionStatusTerminated:
				err := p.client.DeleteSession(ctx, sessionID)
				if err != nil {
					p.logger.Warnf("Failed to delete session %s before creation: %v", sessionID, err)
					return err
				}
				return fmt.Errorf("session %s failed to create, status: %s", sessionID, detail.Status)
			case interfaces.SessionStatusCreating:
				// 继续等待
			}
		}
	}
}

// releaseSession 释放会话槽位，允许其他任务使用
func (p *sessionPoolImpl) releaseSession(sessionID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if item, ok := p.sessions[sessionID]; ok {
		item.RunningTasks--
		if item.RunningTasks < 0 {
			item.RunningTasks = 0
		}
		item.LastUsedAt = time.Now()
	}
}

// invalidateSession 从会话池移除会话槽位，同时异步删除远程资源
func (p *sessionPoolImpl) invalidateSession(sessionID string) {
	// 异步删除远程资源
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
		defer cancel()
		_ = p.client.DeleteSession(ctx, sessionID)
	}()
}

func (p *sessionPoolImpl) prewarmSessions() {
	p.mu.Lock()
	currentCount := len(p.sessions)
	needed := p.activeSessions - currentCount
	p.mu.Unlock()

	if needed <= 0 {
		return
	}

	p.logger.Infof("Pre-warming %d sessions to reach activeSessions limit (%d)", needed, p.activeSessions)

	for i := 0; i < needed; i++ {
		// 使用 acquireSession 逻辑来查找可用 ID 并创建
		// 这里我们直接调用内部逻辑或者复用部分逻辑
		// 简单起见，我们直接尝试寻找空闲槽位并创建
		p.mu.Lock()
		var targetID string
		for j := 0; j < p.maxSessions; j++ {
			id := fmt.Sprintf("%s%d", sessionIDPrefix, j)
			if _, ok := p.sessions[id]; !ok {
				targetID = id
				break
			}
		}
		p.mu.Unlock()

		if targetID == "" {
			break
		}

		go func(sid string) {
			prewarmCtx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
			defer cancel()

			if err := p.ensureRemoteSession(prewarmCtx, sid); err != nil {
				p.logger.Errorf("Failed to pre-warm session %s: %v", sid, err)
				p.removeSession(sid)
				return
			}
			p.logger.Infof("Successfully pre-warmed session: %s", sid)
		}(targetID)
	}
}

func (p *sessionPoolImpl) backgroundWorker() {
	ticker := time.NewTicker(backgroundWorkerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			return
		case <-ticker.C:
			p.maintainPool()
		}
	}
}

func (p *sessionPoolImpl) maintainPool() {
	ctx := context.Background()
	p.mu.Lock()
	// 复制一份当前会话列表进行检查，避免长时间持有锁
	currentSessions := make([]string, 0, len(p.sessions))
	for id := range p.sessions {
		currentSessions = append(currentSessions, id)
	}
	p.mu.Unlock()

	// 1. 健康检查与修复
	for _, id := range currentSessions {
		exists, detail, err := p.client.QuerySession(ctx, id)
		if err != nil || !exists || (detail.Status != interfaces.SessionStatusRunning && detail.Status != interfaces.SessionStatusCreating) {
			p.logger.Warnf("Session %s is unhealthy or missing, removing from pool", id)
			p.removeSession(id)
			p.invalidateSession(id)
		}
	}

	// 2. 预热管理：补足到 activeSessions
	p.prewarmSessions()

	// 3. 空闲管理：根据 activeSessions 配置保留活跃的空闲 session
	p.mu.Lock()
	var idleItems []*sessionItem
	for _, item := range p.sessions {
		if item.RunningTasks == 0 {
			idleItems = append(idleItems, item)
		}
	}
	if len(idleItems) > p.activeSessions {
		// 按最后使用时间排序，保留最新的
		// 简单的做法：除了第一个，其他的都删掉（或者找到最晚使用的保留）
		latestIdx := 0
		for i := 1; i < len(idleItems); i++ {
			if idleItems[i].LastUsedAt.After(idleItems[latestIdx].LastUsedAt) {
				latestIdx = i
			}
		}

		for i, item := range idleItems {
			if i == latestIdx {
				continue
			}
			p.logger.Infof("Scaling down idle session: %s", item.ID)
			// 从会话池移除会话槽位
			delete(p.sessions, item.ID)
			p.invalidateSession(item.ID)
		}
	}
	p.mu.Unlock()
}

// Close 关闭全局会话池
func Close() {
	if poolInstance == nil {
		return
	}
	close(poolInstance.stopCh)
	// 并发关闭会话池
	waitGroup := sync.WaitGroup{}
	for _, pool := range poolInstance.sessions {
		waitGroup.Add(1)
		poolInstance.removeSession(pool.ID)
		go func(sessionID string) {
			_ = poolInstance.client.DeleteSession(context.Background(), sessionID)
			waitGroup.Done()
		}(pool.ID)
	}
	waitGroup.Wait()
}
