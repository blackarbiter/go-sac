package rabbitmq

// Exchange names
const (
	TaskDispatchExchange  = "task_dispatch_exchange"
	ResultProcessExchange = "result_process_exchange"
	NotificationExchange  = "notification_exchange"
	RetryExchange         = "retry_exchange"
)

// Queue names
const (
	// Scan priority queues
	ScanHighPriorityQueue   = "scan.high_priority"
	ScanMediumPriorityQueue = "scan.medium_priority"
	ScanLowPriorityQueue    = "scan.low_priority"

	// Asset queue
	AssetTaskQueue = "asset_task_queue"

	// Result queue
	ResultStorageQueue = "result_storage_queue"

	// Notification queues
	NotificationEmailQueue  = "notification_email_queue"
	NotificationSMSQueue    = "notification_sms_queue"
	NotificationSystemQueue = "notification_system_queue"

	// Retry queues
	RetryQueue5Min          = "retry_queue_5min"
	ManualInterventionQueue = "manual_intervention_queue"
)

// Routing key patterns
const (
	// Scan task routing patterns
	ScanHighPattern   = "scan.*.high"
	ScanMediumPattern = "scan.*.medium"
	ScanLowPattern    = "scan.*.low"

	// Asset task routing pattern
	AssetPattern = "asset.*"

	// Result routing pattern
	ResultStoragePattern = "result.storage"

	// Retry patterns
	RetryPattern  = "retry.#"
	ManualPattern = "manual.#"
)

// Dead letter exchange headers
const (
	DeadLetterExchange   = "x-dead-letter-exchange"
	DeadLetterRoutingKey = "x-dead-letter-routing-key"
	MessageTTL           = "x-message-ttl"
	MaxPriority          = "x-max-priority"
)
