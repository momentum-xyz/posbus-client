package pbc

//var (
//	logger *zap.SugaredLogger
//	dlevel = zap.NewAtomicLevelAt(zapcore.DebugLevel)
//)

//func init() {
//cfg := zap.Config{
//	Level:            dlevel,
//	Encoding:         "console",
//	EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
//	OutputPaths:      []string{"stdout"},
//	ErrorOutputPaths: []string{"stdout"},
//	// NOTE: set this false to enable stack trace
//	DisableStacktrace: true,
//}

//l, err := cfg.Build()
//if err != nil {
//	panic(err)
//}
//logger = l.Sugar()

//logger.L().Info("Logger initialized")
//}

//func L() *zap.SugaredLogger {
//	if logger == nil {
//		panic("Logger is not initialized")
//	}
//	return logger
//}
