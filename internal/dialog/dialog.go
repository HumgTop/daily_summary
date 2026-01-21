package dialog

// Dialog 对话框接口
type Dialog interface {
	// ShowInput 显示文本输入对话框
	// title: 对话框标题
	// message: 提示信息
	// defaultText: 默认文本
	// 返回: (用户输入的文本, 是否点击了确定, 错误)
	ShowInput(title, message, defaultText string) (string, bool, error)

	// ShowNotification 显示系统通知
	// title: 通知标题
	// message: 通知内容
	// 返回: 错误
	ShowNotification(title, message string) error
}
