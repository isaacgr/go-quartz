# loggir

A simple logging library, matching the format I like for my projects.

# Usage

```go

import "github.com/isaacgr/loggir"

debug := false

// Create the logger instance
// All logs are prepended with the timestamp, then info level
log := loggir.GetLogger("main", debug)

log.Info("hello world", "Key1", "Value1", "Key2", "Value2")
// INFO:main: hello world. Key1 [Value1]. Key2 [Value2].


// Create a copy of the handler for a subcomponent or submodule
clientLog := log.With("module", "submodule", "component", "config")
log.Warn("hello world", "Key1", "Value1", "Key2", "Value2")
// WARN:main.submodule.config: hello world. Key1 [Value1]. Key2 [Value2].
```
