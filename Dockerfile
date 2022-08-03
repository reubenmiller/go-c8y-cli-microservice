FROM golang:1.18 as builder
WORKDIR /app
# Add go modules files
ADD ./go.mod .

# Cache go dependencies (only re-download if go.mod changes)
RUN go mod download

# Install build dependencies
RUN apt-get update \
 && apt-get install -y jq \
 && rm -rf /var/lib/apt/lists/*

# Add all project files
ADD . .


# Build the application (injecting build env information using the -X go flags)
RUN \
  REPO_PATH="$(git config --get remote.origin.url | cut -d@ -f2 | sed 's/\.git$//g' )/pkg/app" && \
  VERSION=$(cat ./cumulocity.json | jq -r '.version') && \
  BRANCH_NAME=$(git branch | grep \* | cut -d ' ' -f2) && \
  COMMIT_HASH=$(git rev-parse --short HEAD) && \
  BUILD_TIME=$(date +%Y-%m-%dT%H:%M:%S%z) && \
  GO_LINKER_OPTS="-w -s -X $REPO_PATH.Version=$VERSION -X $REPO_PATH.Commit=$COMMIT_HASH -X $REPO_PATH.Branch=$BRANCH_NAME -X $REPO_PATH.BuildTime=$BUILD_TIME" && \
  echo "$GO_LINKER_OPTS" && \
  CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="$GO_LINKER_OPTS" -o /go/bin/app ./cmd/main/main.go


#
# Application image
#
FROM alpine:3.11
RUN apk --no-cache add ca-certificates git wget jq bash
RUN apk update \
    && wget -O /etc/apk/keys/reuben.d.miller\@gmail.com-61e3680b.rsa.pub https://reubenmiller.github.io/go-c8y-cli-repo/alpine/PUBLIC.KEY \
    && sh -c "echo 'https://reubenmiller.github.io/go-c8y-cli-repo/alpine/stable/main'" >> /etc/apk/repositories \
    && apk --no-cache add go-c8y-cli

WORKDIR /go/bin
COPY --from=builder /go/bin/app .
COPY config/application.production.properties ./application.properties
CMD ["./app"]
