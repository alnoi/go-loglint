package main

import (
	"golang.org/x/tools/go/analysis"

	"github.com/alnoi/go-loglint/internal/analyzer/loglint"
)

func New(conf any) ([]*analysis.Analyzer, error) {
	// пока игнорим conf (бонусом подключим позже)
	return []*analysis.Analyzer{loglint.Analyzer}, nil
}
