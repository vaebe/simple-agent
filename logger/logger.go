package logger

import (
    "os"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "gopkg.in/natefinch/lumberjack.v2"
)

// 全局日志变量
var Logger *zap.Logger

// Init 初始化日志系统
func Init() {
    // 确保 logs 目录存在
    logDir := "logs"
    if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
        panic("无法创建日志目录: " + err.Error())
    }

    // 使用 lumberjack 做日志轮转
    logWriter := &lumberjack.Logger{
        Filename:   "./logs/simple-agent.log", // 日志文件路径
        MaxSize:    100,                       // 每个日志文件最大尺寸（MB）
        MaxBackups: 3,                         // 保留旧文件的最大个数
        MaxAge:     7,                         // 保留旧文件的最大天数
        Compress:   true,                      // 是否压缩旧文件
        LocalTime:  true,                      // 使用本地时间命名日志文件
    }

    // 配置编码器
    encoderConfig := zap.NewProductionEncoderConfig()
    encoderConfig.TimeKey = "timestamp"
    encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

    // 创建文件核心配置
    fileCore := zapcore.NewCore(
        zapcore.NewJSONEncoder(encoderConfig),
        zapcore.AddSync(logWriter),
        zap.NewAtomicLevelAt(zap.InfoLevel),
    )

    // 控制台核心保持不变
    consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
    consoleCore := zapcore.NewCore(
        consoleEncoder,
        zapcore.AddSync(os.Stdout),
        zap.NewAtomicLevelAt(zap.InfoLevel),
    )

    // 组合多个核心
    teeCore := zapcore.NewTee(fileCore, consoleCore)
    Logger = zap.New(teeCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

    Logger.Debug("日志系统初始化完成")
}

// Sync 同步日志
func Sync() {
    if Logger != nil {
        Logger.Sync()
    }
}

// Info 记录信息日志
func Info(msg string, fields ...zap.Field) {
    if Logger != nil {
        Logger.Info(msg, fields...)
    }
}

// Error 记录错误日志
func Error(msg string, fields ...zap.Field) {
    if Logger != nil {
        Logger.Error(msg, fields...)
    }
}

// Debug 记录调试日志
func Debug(msg string, fields ...zap.Field) {
    if Logger != nil {
        Logger.Debug(msg, fields...)
    }
}

// Warn 记录警告日志
func Warn(msg string, fields ...zap.Field) {
    if Logger != nil {
        Logger.Warn(msg, fields...)
    }
}

// Fatal 记录致命错误日志并退出
func Fatal(msg string, fields ...zap.Field) {
    if Logger != nil {
        Logger.Fatal(msg, fields...)
    }
}