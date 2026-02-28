package zap

// mock for tests with analysistest

type Logger struct{}
type SugaredLogger struct{}

func (l *Logger) Debug(msg string, fields ...Field)  {}
func (l *Logger) Info(msg string, fields ...Field)   {}
func (l *Logger) Warn(msg string, fields ...Field)   {}
func (l *Logger) Error(msg string, fields ...Field)  {}
func (l *Logger) DPanic(msg string, fields ...Field) {}
func (l *Logger) Panic(msg string, fields ...Field)  {}
func (l *Logger) Fatal(msg string, fields ...Field)  {}

func (l *Logger) Sugar() *SugaredLogger { return &SugaredLogger{} }

func (s *SugaredLogger) Debug(args ...any)  {}
func (s *SugaredLogger) Info(args ...any)   {}
func (s *SugaredLogger) Warn(args ...any)   {}
func (s *SugaredLogger) Error(args ...any)  {}
func (s *SugaredLogger) DPanic(args ...any) {}
func (s *SugaredLogger) Panic(args ...any)  {}
func (s *SugaredLogger) Fatal(args ...any)  {}

func (s *SugaredLogger) Debugf(tpl string, args ...any)  {}
func (s *SugaredLogger) Infof(tpl string, args ...any)   {}
func (s *SugaredLogger) Warnf(tpl string, args ...any)   {}
func (s *SugaredLogger) Errorf(tpl string, args ...any)  {}
func (s *SugaredLogger) DPanicf(tpl string, args ...any) {}
func (s *SugaredLogger) Panicf(tpl string, args ...any)  {}
func (s *SugaredLogger) Fatalf(tpl string, args ...any)  {}

type Field struct{}

func String(key, val string) Field    { return Field{} }
func Any(key string, val any) Field   { return Field{} }
func Int(key string, val int) Field   { return Field{} }
func Bool(key string, val bool) Field { return Field{} }
func Error(err error) Field           { return Field{} }

func NewNop() *Logger { return &Logger{} }
