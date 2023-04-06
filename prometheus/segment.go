// 代码片段
package main

import (
	"net/http"
	"runtime/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	http.Handle("/metrics", promHttp.Handle())  // promHttp.Handle 是prometheus包中自己定义的
	http.ListenAndServe(server.MetricsBindAddress,nil) 

}

// 注册指标
func RegisterMetrics(){
	RegisterMetricsOnce.Do(
		func(){
			prometheus.MustRegister(APIServerRequest)
			prometheus.MustRegister(WorkQueueSize)
		},
	)
	// 代码中输出指标
	//metrics.AddAPIServerRequest(controllerName,constants.CoreAPIGroup,constants.SecretResource,constants.Get,cn.Namespace)
}
