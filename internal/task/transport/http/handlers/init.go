package handlers

import (
	"github.com/blackarbiter/go-sac/internal/task/service"
)

// InitHandlers 初始化所有处理程序
func InitHandlers(taskService service.TaskService) {
	taskHandler = NewTaskHandler(taskService)
}
