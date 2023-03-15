# 控制器需要完成的功能

1. 观察，通过监控kubernetes资源对象的变化来获取对象的状态，需要注入eventHandler让client-go将变化的事件对象信息放到workQueue中
2. 分析，确定当前状态和期望状态的不同，由worker完成
3. 执行，对当前对象进行修改，使其状态和期望相同，由worker完成
4. 更新，更新对象当前的状态

# client-go

1. 需要使用和k8s对应的client-go版本