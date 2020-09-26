FROM golangci/golangci-lint:v1.31.0-alpine

ENTRYPOINT ["/usr/bin/golangci-lint", "run",\
            "--modules-download-mode=vendor",\
            "--deadline=15m"]
