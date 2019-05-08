@rem Watch and run tests.
@rem see [simple filesystem watcher and executor](https://github.com/tmc/watcher)
@rem install with `go get github.com/tmc/watcher`
%home%\go\bin\watcher -depth 0 %home%\go\bin\gotest %*
