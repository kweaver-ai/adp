package com.eisoo.dc.common.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.scheduling.concurrent.ThreadPoolTaskScheduler;

import java.util.concurrent.ThreadPoolExecutor;

/**
 * @author Tian.lan
 */
@Configuration
public class ScanTaskSchedulerConfig {
    /**
     * 配置线程池任务调度器，用于执行CronTask
     */
    @Bean
    public ThreadPoolTaskScheduler threadPoolTaskScheduler() {
        ThreadPoolTaskScheduler scheduler = new ThreadPoolTaskScheduler();
        // 核心线程数
        scheduler.setPoolSize(10);
        // 线程名前缀
        scheduler.setThreadNamePrefix("scheduler-cron-task-");
        scheduler.setAwaitTerminationSeconds(60); // 停机等待时间
        scheduler.setWaitForTasksToCompleteOnShutdown(true); // 等待所有任务执行完成再停机
        // 等待所有任务执行完再关闭
        scheduler.setWaitForTasksToCompleteOnShutdown(true);
        //CallerRunsPolicy（生产推荐，核心任务不丢失，由提交线程执行）
        scheduler.setRejectedExecutionHandler(new ThreadPoolExecutor.CallerRunsPolicy());
        scheduler.initialize();
        return scheduler;
    }
}
