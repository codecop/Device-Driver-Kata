go vet

%home%\go\bin\golint

%home%\go\bin\golangci-lint run --enable-all

@setlocal
@set PATH=%home%\go\bin;%PATH%
gocritic check-project -color -enable=all -withExperimental -withOpinionated -disable=unexportedCall .
@endlocal
