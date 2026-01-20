package summary

// AIClient AI 客户端接口
type AIClient interface {
	GenerateSummary(prompt string) (string, error)
}
