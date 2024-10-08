package tmp

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/elastic/elastic-agent-libs/logp"
	"go.uber.org/zap"
)

func Debug(msg string, args ...any) {
	dbg(msg, false, args...)
}

func DebugWithStack(msg string, args ...any) {
	dbg(msg, true, args...)
}

func dbg(msg string, printStack bool, args ...any) {
	newArgs := make([]any, 0, 2*len(args))
	name := ""
	for i, arg := range args {
		if i%2 == 0 {
			name = arg.(string)
			continue
		}
		newArgs = append(newArgs, name+".json", j(arg), name+".type", t(arg))
	}
	newArgs = append(newArgs, "proc", procInfo(printStack))
	l := logp.NewLogger("storage-poc", zap.AddCallerSkip(1))
	l.Infow(msg, newArgs...)
	_ = l.Sync()
}

func procInfo(printStack bool) map[string]any {
	m := map[string]any{
		"pid":  os.Getpid(),
		"goid": goID(),
	}
	if printStack {
		m["stack"] = debug.Stack()
	}
	return m
}

func goID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func j(x any) string {
	b, _ := json.Marshal(x)
	return string(b)
}

func t(x any) string {
	return fmt.Sprintf("%T", x)
}
