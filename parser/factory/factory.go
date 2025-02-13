package factory

import (
	"fmt"
	"shared-lib/parser/base"
	"shared-lib/parser/nmon"
)

// ParserFactory 解析器工廠
type ParserFactory struct {
	creators map[string]func() base.MetricParser
}

// NewParserFactory 創建解析器工廠
func NewParserFactory() *ParserFactory {
	f := &ParserFactory{
		creators: make(map[string]func() base.MetricParser),
	}

	// 註冊默認解析器
	f.Register("nmon", func() base.MetricParser { return nmon.NewParser() })

	return f
}

// Register 註冊解析器創建函數
func (f *ParserFactory) Register(typ string, creator func() base.MetricParser) {
	f.creators[typ] = creator
}

// Create 創建指定類型的解析器
func (f *ParserFactory) Create(typ string) (base.MetricParser, error) {
	creator, ok := f.creators[typ]
	if !ok {
		return nil, fmt.Errorf("unsupported parser type: %s", typ)
	}
	return creator(), nil
}
