package rest

// Этот файл содержит именованные типы-обёртки над инлайн-анонимными структурами
// тел запросов/ответов, чтобы swaggo дал им осмысленные имена в swagger.json
// вместо auto-generated inline_response_200_N. Логики здесь нет.

// ----- Health -----

// healthResponse — ответ GET /health. Определён в health.go, продублирован
// swag-комментарием через @Success … {object} healthResponse.

// ----- Events -----

// pollResponse — ответ POST /events/poll. См. event_poll.go.

// ----- Orders: обёртки для list/refund, объявленных инлайн -----

// OrdersListResponse оборачивает список заказов для GET /orders.
//
// Пример:
//
//	{ "orders": [ { "id": "WMBY8JNK", "status": "active", ... } ] }
type OrdersListResponse struct {
	Orders []OrderListItem `json:"orders" example:""`
}

// OrderRefundResponse — ответ POST /orders/{id}/refund.
type OrderRefundResponse struct {
	Ok      bool   `json:"ok" example:"true"`
	OrderID string `json:"orderId" example:"WMBY8JNK"`
}

// ----- Chats -----

// ChatMessageRequest — тело POST /chats/{id}/messages.
type ChatMessageRequest struct {
	Text string `json:"text" example:"Ваш заказ готов: ABC-KEY-123"`
}

// ChatMessageResponse — ответ POST /chats/{id}/messages.
type ChatMessageResponse struct {
	Ok        bool   `json:"ok" example:"true"`
	MessageID string `json:"messageId,omitempty" example:"88123456"`
}

// ----- Control -----

// ControlResumeResponse — ответ POST /control/resume.
type ControlResumeResponse struct {
	Status  string `json:"status" example:"accepted"`
	Message string `json:"message" example:"resume signal sent; main will re-read .env and re-init runner"`
}

// ----- Ошибки (единый конверт, см. errors.go) -----

// EngineError — общий формат тела ошибки для всех эндпоинтов.
// retryable: true означает внутреннюю ошибку (500), которую стоит повторить.
//
// Пример:
//
//	{
//	  "error": {
//	    "code": "auth_lost",
//	    "message": "auth lost: golden_seal expired or missing",
//	    "retryable": false
//	  }
//	}
type EngineError struct {
	Error EngineErrorBody `json:"error"`
}

// EngineErrorBody — внутреннее тело ошибки.
type EngineErrorBody struct {
	Code      string `json:"code" example:"auth_lost"`
	Message   string `json:"message" example:"auth lost: golden_seal expired or missing"`
	Retryable bool   `json:"retryable" example:"false"`
}
