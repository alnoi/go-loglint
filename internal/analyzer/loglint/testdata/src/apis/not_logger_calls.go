package apis

type NotLogger struct{}

func (NotLogger) Info(string) {}

func NotLoggerCalls() {
	var l NotLogger
	l.Info("Hello")
	l.Info("ok...")
	l.Info("привет")

	Print("Hello")
	Print("ok…")
	Print("why?!")

	makeFake().Info("Hello")
	makeFake().Info("ok...")
}

func Print(string) {}

type fake struct{}

func (fake) Info(string) {}

func makeFake() fake { return fake{} }
