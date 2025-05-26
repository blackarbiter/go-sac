package handlers

import (
	"github.com/blackarbiter/go-sac/internal/task/service"
)

// 全局任务处理器实例
var taskHandler *TaskHandler

// InitHandlers 初始化所有处理程序
func InitHandlers(taskService service.TaskService) {
	taskHandler = NewTaskHandler(taskService)
}

// GetTaskHandler 获取任务处理器实例
func GetTaskHandler() *TaskHandler {
	return taskHandler
}
