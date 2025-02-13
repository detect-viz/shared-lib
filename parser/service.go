package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"shared-lib/interfaces"
	"shared-lib/models"
	"shared-lib/parser/factory"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Service Parser 服務
type Service struct {
	db      interfaces.Database
	config  *models.ParserConfig
	logger  interfaces.Logger
	factory *factory.ParserFactory
}

// NewService 創建 Parser 服務
func NewService(db interfaces.Database, logger interfaces.Logger) *Service {
	return &Service{
		db:      db,
		logger:  logger,
		factory: factory.NewParserFactory(),
	}
}

// SetConfig 設置配置
func (s *Service) SetConfig(config *models.ParserConfig) {
	s.config = config
}

// Run 啟動服務
func (s *Service) Run() error {
	if s.config == nil {
		return fmt.Errorf("配置未初始化")
	}

	ticker := time.NewTicker(s.config.CheckPeriod)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.ProcessFiles(); err != nil {
			s.logger.Error("處理檔案失敗", zap.Error(err))
			continue
		}
	}
	return nil
}

// Close 關閉服務
func (s *Service) Close() error {
	// TODO: 實現優雅關閉邏輯
	return nil
}

// DetectSourceType 檢測數據源類型
func (s *Service) DetectSourceType(content []byte) (string, error) {
	if len(content) == 0 {
		return "", fmt.Errorf("empty content")
	}

	// 讀取前幾行來判斷類型
	scanner := bufio.NewScanner(bytes.NewReader(content))
	var firstLines []string
	for i := 0; scanner.Scan() && i < 5; i++ {
		firstLines = append(firstLines, scanner.Text())
	}

	// 檢查文件特徵
	for _, line := range firstLines {
		switch {
		case strings.HasPrefix(line, "AAA,"):
			return "nmon", nil
		case strings.Contains(line, "LOGMAN_START"):
			return "logman", nil
		case strings.Contains(line, "NJMON_START"):
			return "njmon", nil
		case strings.Contains(line, "AWR Report"):
			return "awr", nil
		case strings.Contains(line, "Tablespace Report"):
			return "tablespace", nil
		case strings.Contains(line, "Connection Report"):
			return "connection", nil
		}
	}

	return "", fmt.Errorf("unknown source type")
}

// ProcessFiles 處理檔案
func (s *Service) ProcessFiles() error {
	// 獲取待處理的檔案列表
	files, err := s.db.GetPendingFiles()
	if err != nil {
		s.logger.Error("獲取待處理檔案失敗", zap.Error(err))
		return err
	}

	for _, file := range files {
		// 讀取檔案內容
		content, err := os.ReadFile(file.Path)
		if err != nil {
			s.logger.Error("讀取檔案失敗",
				zap.String("path", file.Path),
				zap.Error(err))
			continue
		}

		// 解析數據
		metrics, err := s.Parse(content)
		if err != nil {
			s.logger.Error("解析檔案失敗",
				zap.String("path", file.Path),
				zap.Error(err))
			continue
		}

		// 保存解析結果
		if err := s.db.SaveMetrics(file.ID, metrics); err != nil {
			s.logger.Error("保存指標數據失敗",
				zap.String("file_id", file.ID),
				zap.Error(err))
			continue
		}

		// 更新檔案狀態
		file.Status = "processed"
		if err := s.db.UpdateFile(file); err != nil {
			s.logger.Error("更新檔案狀態失敗",
				zap.String("file_id", file.ID),
				zap.Error(err))
		}
	}

	return nil
}

// Parse 解析數據
func (s *Service) Parse(content []byte) (map[string]interface{}, error) {
	// 1. 檢測數據源類型
	sourceType, err := s.DetectSourceType(content)
	if err != nil {
		return nil, fmt.Errorf("detect source type failed: %w", err)
	}

	// 2. 創建對應的解析器
	parser, err := s.factory.Create(sourceType)
	if err != nil {
		return nil, fmt.Errorf("create parser failed: %w", err)
	}

	// 3. 解析數據
	return parser.Parse(bytes.NewBuffer(content))
}
