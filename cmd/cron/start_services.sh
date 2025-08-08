#!/bin/bash

# 启动分布式定时任务服务脚本

echo "=== 启动分布式定时任务服务 ==="

# 编译项目
echo "1. 编译项目..."
go build -o cron-service cmd/cron/main.go
go build -o cron-web cmd/cron/web/main.go

if [ $? -eq 0 ]; then
    echo "✅ 编译成功"
else
    echo "❌ 编译失败"
    exit 1
fi

# 启动任务执行实例
echo "2. 启动任务执行实例..."
./cron-service > logs/instance.log 2>&1 &
INSTANCE_PID=$!
sleep 3

# 检查实例是否启动成功
if ps -p $INSTANCE_PID > /dev/null; then
    echo "✅ 任务执行实例启动成功 (PID: $INSTANCE_PID)"
else
    echo "❌ 任务执行实例启动失败"
    exit 1
fi

# 启动Web管理界面
echo "3. 启动Web管理界面..."
./cron-web > logs/web.log 2>&1 &
WEB_PID=$!
sleep 3

# 检查Web界面是否启动成功
if ps -p $WEB_PID > /dev/null; then
    echo "✅ Web管理界面启动成功 (PID: $WEB_PID)"
else
    echo "❌ Web管理界面启动失败"
    kill $INSTANCE_PID 2>/dev/null
    exit 1
fi

echo ""
echo "=== 服务启动完成 ==="
echo "📊 任务执行实例: PID $INSTANCE_PID"
echo "🌐 Web管理界面: PID $WEB_PID"
echo "🔗 访问地址: http://localhost:8081"
echo ""
echo "📋 服务信息:"
echo "   - Web服务端口: 8081"
echo "   - 任务执行实例: 运行中"
echo "   - Web管理界面: 运行中"
echo ""
echo "🛑 停止服务: kill $INSTANCE_PID $WEB_PID"
echo "📋 查看日志: tail -f logs/instance.log logs/web.log"
echo ""
echo "=== 服务已启动，按 Ctrl+C 停止 ==="

# 等待用户中断
trap 'echo ""; echo "正在停止服务..."; kill $INSTANCE_PID $WEB_PID 2>/dev/null; echo "服务已停止"; exit 0' INT

# 保持脚本运行
while true; do
    sleep 1
done 