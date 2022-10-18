// Licensed to the LF AI & Data foundation under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/milvus-io/milvus/internal/log"
	"github.com/milvus-io/milvus/internal/proto/internalpb"
	"github.com/milvus-io/milvus/internal/util/ratelimitutil"
	"github.com/milvus-io/milvus/internal/util/typeutil"
)

var (
	// ProxySearchVectors record the number of vectors search successfully.
	ProxySearchVectors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "search_vectors_count",
			Help:      "counter of vectors successfully searched",
		}, []string{nodeIDLabelName})

	// ProxyInsertVectors record the number of vectors insert successfully.
	ProxyInsertVectors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "insert_vectors_count",
			Help:      "counter of vectors successfully inserted",
		}, []string{nodeIDLabelName})

	// ProxySQLatency record the latency of search successfully.
	ProxySQLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "sq_latency",
			Help:      "latency of search or query successfully",
			Buckets:   buckets,
		}, []string{nodeIDLabelName, queryTypeLabelName})

	// ProxyCollectionSQLatency record the latency of search successfully, per collection
	ProxyCollectionSQLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "collection_sq_latency",
			Help:      "latency of search or query successfully, per collection",
			Buckets:   buckets,
		}, []string{nodeIDLabelName, queryTypeLabelName, collectionName})

	// ProxyMutationLatency record the latency that mutate successfully.
	ProxyMutationLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "mutation_latency",
			Help:      "latency of insert or delete successfully",
			Buckets:   buckets, // unit: ms
		}, []string{nodeIDLabelName, msgTypeLabelName})

	// ProxyMutationLatency record the latency that mutate successfully, per collection
	ProxyCollectionMutationLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "collection_mutation_latency",
			Help:      "latency of insert or delete successfully, per collection",
			Buckets:   buckets,
		}, []string{nodeIDLabelName, msgTypeLabelName, collectionName})
	// ProxyWaitForSearchResultLatency record the time that the proxy waits for the search result.
	ProxyWaitForSearchResultLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "sq_wait_result_latency",
			Help:      "latency that proxy waits for the result",
			Buckets:   buckets, // unit: ms
		}, []string{nodeIDLabelName, queryTypeLabelName})
	// ProxyReduceResultLatency record the time that the proxy reduces search result.
	ProxyReduceResultLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "sq_reduce_result_latency",
			Help:      "latency that proxy reduces search result",
			Buckets:   buckets, // unit: ms
		}, []string{nodeIDLabelName, queryTypeLabelName})

	// ProxyDecodeResultLatency record the time that the proxy decodes the search result.
	ProxyDecodeResultLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "sq_decode_result_latency",
			Help:      "latency that proxy decodes the search result",
			Buckets:   buckets, // unit: ms
		}, []string{nodeIDLabelName, queryTypeLabelName})

	// ProxyMsgStreamObjectsForPChan record the number of MsgStream objects per PChannel on each collection_id on Proxy.
	ProxyMsgStreamObjectsForPChan = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "msgstream_obj_num",
			Help:      "number of MsgStream objects per physical channel",
		}, []string{nodeIDLabelName, channelNameLabelName})

	// ProxySendMutationReqLatency record the latency that Proxy send insert request to MsgStream.
	ProxySendMutationReqLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "mutation_send_latency",
			Help:      "latency that proxy send insert request to MsgStream",
			Buckets:   buckets, // unit: ms
		}, []string{nodeIDLabelName, msgTypeLabelName})

	// ProxyCacheHitCounter record the number of Proxy cache hits or miss.
	ProxyCacheStatsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "cache_hit_count",
			Help:      "count of cache hits/miss",
		}, []string{nodeIDLabelName, cacheNameLabelName, cacheStateLabelName})

	// ProxyUpdateCacheLatency record the time that proxy update cache when cache miss.
	ProxyUpdateCacheLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "cache_update_latency",
			Help:      "latency that proxy update cache when cache miss",
			Buckets:   buckets, // unit: ms
		}, []string{nodeIDLabelName})

	// ProxySyncTimeTick record Proxy synchronization timestamp statistics, differentiated by Channel.
	ProxySyncTimeTick = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "sync_epoch_time",
			Help:      "synchronized unix epoch per physical channel and default channel",
		}, []string{nodeIDLabelName, channelNameLabelName})

	// ProxyApplyPrimaryKeyLatency record the latency that apply primary key.
	ProxyApplyPrimaryKeyLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "apply_pk_latency",
			Help:      "latency that apply primary key",
			Buckets:   buckets, // unit: ms
		}, []string{nodeIDLabelName})

	// ProxyApplyTimestampLatency record the latency that proxy apply timestamp.
	ProxyApplyTimestampLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "apply_timestamp_latency",
			Help:      "latency that proxy apply timestamp",
			Buckets:   buckets, // unit: ms
		}, []string{nodeIDLabelName})

	// ProxyFunctionCall records the number of times the function of the DDL operation was executed, like `CreateCollection`.
	ProxyFunctionCall = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "req_count",
			Help:      "count of operation executed",
		}, []string{nodeIDLabelName, functionLabelName, statusLabelName})

	// ProxyReqLatency records the latency that for all requests, like "CreateCollection".
	ProxyReqLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "req_latency",
			Help:      "latency of each request",
			Buckets:   buckets, // unit: ms
		}, []string{nodeIDLabelName, functionLabelName})

	// ProxyReceiveBytes record the received bytes of messages in Proxy
	ProxyReceiveBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "receive_bytes_count",
			Help:      "count of bytes received  from sdk",
		}, []string{nodeIDLabelName, msgTypeLabelName})

	// ProxyReadReqSendBytes record the bytes sent back to client by Proxy
	ProxyReadReqSendBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "send_bytes_count",
			Help:      "count of bytes sent back to sdk",
		}, []string{nodeIDLabelName})

	// ProxyLimiterRate records rates of rateLimiter in Proxy.
	ProxyLimiterRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: milvusNamespace,
			Subsystem: typeutil.ProxyRole,
			Name:      "limiter_rate",
			Help:      "",
		}, []string{nodeIDLabelName, msgTypeLabelName})
)

//RegisterProxy registers Proxy metrics
func RegisterProxy(registry *prometheus.Registry) {
	registry.MustRegister(ProxySearchVectors)
	registry.MustRegister(ProxyInsertVectors)

	registry.MustRegister(ProxySQLatency)
	registry.MustRegister(ProxyCollectionSQLatency)
	registry.MustRegister(ProxyMutationLatency)
	registry.MustRegister(ProxyCollectionMutationLatency)

	registry.MustRegister(ProxyWaitForSearchResultLatency)
	registry.MustRegister(ProxyReduceResultLatency)
	registry.MustRegister(ProxyDecodeResultLatency)

	registry.MustRegister(ProxyMsgStreamObjectsForPChan)

	registry.MustRegister(ProxySendMutationReqLatency)

	registry.MustRegister(ProxyCacheStatsCounter)
	registry.MustRegister(ProxyUpdateCacheLatency)

	registry.MustRegister(ProxySyncTimeTick)
	registry.MustRegister(ProxyApplyPrimaryKeyLatency)
	registry.MustRegister(ProxyApplyTimestampLatency)

	registry.MustRegister(ProxyFunctionCall)
	registry.MustRegister(ProxyReqLatency)

	registry.MustRegister(ProxyReceiveBytes)
	registry.MustRegister(ProxyReadReqSendBytes)

	registry.MustRegister(ProxyLimiterRate)
}

// SetRateGaugeByRateType sets ProxyLimiterRate metrics.
func SetRateGaugeByRateType(rateType internalpb.RateType, nodeID int64, rate float64) {
	if ratelimitutil.Limit(rate) == ratelimitutil.Inf {
		return
	}
	nodeIDStr := strconv.FormatInt(nodeID, 10)
	log.Debug("set rates", zap.Int64("nodeID", nodeID), zap.String("rateType", rateType.String()), zap.Float64("rate", rate))
	switch rateType {
	case internalpb.RateType_DMLInsert:
		ProxyLimiterRate.WithLabelValues(nodeIDStr, InsertLabel).Set(rate)
	case internalpb.RateType_DMLDelete:
		ProxyLimiterRate.WithLabelValues(nodeIDStr, DeleteLabel).Set(rate)
	case internalpb.RateType_DQLSearch:
		ProxyLimiterRate.WithLabelValues(nodeIDStr, SearchLabel).Set(rate)
	case internalpb.RateType_DQLQuery:
		ProxyLimiterRate.WithLabelValues(nodeIDStr, QueryLabel).Set(rate)
	}
}

func CleanupCollectionMetrics(nodeID int64, collection string) {
	ProxyCollectionSQLatency.Delete(prometheus.Labels{nodeIDLabelName: strconv.FormatInt(nodeID, 10),
		queryTypeLabelName: SearchLabel, collectionName: collection})
	ProxyCollectionSQLatency.Delete(prometheus.Labels{nodeIDLabelName: strconv.FormatInt(nodeID, 10),
		queryTypeLabelName: QueryLabel, collectionName: collection})
	ProxyCollectionMutationLatency.Delete(prometheus.Labels{nodeIDLabelName: strconv.FormatInt(nodeID, 10),
		msgTypeLabelName: InsertLabel, collectionName: collection})
	ProxyCollectionMutationLatency.Delete(prometheus.Labels{nodeIDLabelName: strconv.FormatInt(nodeID, 10),
		msgTypeLabelName: DeleteLabel, collectionName: collection})
}
