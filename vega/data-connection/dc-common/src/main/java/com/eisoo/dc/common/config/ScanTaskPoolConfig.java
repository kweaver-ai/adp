package com.eisoo.dc.common.config;

import com.google.common.util.concurrent.ListeningExecutorService;
import com.google.common.util.concurrent.MoreExecutors;
import lombok.Data;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.ThreadPoolExecutor;
import java.util.concurrent.TimeUnit;

/**
 * @author Tian.lan
 */
@Configuration
@Data
public class ScanTaskPoolConfig {
    @Value("${scan.numThread}")
    private Integer numThread;

    @Bean(name = "dsScheduleScanExecutor")
    public ListeningExecutorService dsScheduleScanExecutorPool() {
        return MoreExecutors.listeningDecorator(new ThreadPoolExecutor(
                20, // 核心线程数
                20, // 最大线程数
                0L,
                TimeUnit.MILLISECONDS, // 空闲线程等待新任务的最长时间
                new LinkedBlockingQueue<>() // 使用无界队列
        ));
    }

    @Bean(name = "dsScanExecutor")
    public ListeningExecutorService dsScanExecutorPool() {
        return MoreExecutors.listeningDecorator(new ThreadPoolExecutor(
                20, // 核心线程数
                20, // 最大线程数
                0L,
                TimeUnit.MILLISECONDS, // 空闲线程等待新任务的最长时间
                new LinkedBlockingQueue<>() // 使用无界队列
        ));
    }

    @Bean(name = "tableScanExecutor")
    public ListeningExecutorService tableScanExecutorPool() {
        return MoreExecutors.listeningDecorator(new ThreadPoolExecutor(
                20, // 核心线程数
                20, // 最大线程数
                0L,
                TimeUnit.MILLISECONDS, // 空闲线程等待新任务的最长时间
                new LinkedBlockingQueue<>() // 使用无界队列
        ));
    }

    @Bean(name = "scanTaskExecutor")
    public ListeningExecutorService scanTaskExecutorPool() {
        return MoreExecutors.listeningDecorator(new ThreadPoolExecutor(
                10,
                10,
                0L,
                TimeUnit.MILLISECONDS,
                new LinkedBlockingQueue<>()
        ));
    }
}
