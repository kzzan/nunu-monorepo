package v1

// MetaResponse 是元信息接口响应。
type MetaResponse struct {
	Response
	Data MetaData `json:"data"`
}

// HealthResponse 是健康检查接口响应。
type HealthResponse struct {
	Response
	Data HealthData `json:"data"`
}

// ManifestResponse 是站点清单接口响应。
type ManifestResponse struct {
	Response
	Data ManifestData `json:"data"`
}
