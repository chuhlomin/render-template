FROM --platform=$BUILDPLATFORM golang:1.17 as build-env

WORKDIR /go/src/app
ADD . /go/src/app

RUN go test -mod=vendor -cover ./...
RUN go build -mod=vendor -o /go/bin/app


FROM gcr.io/distroless/static:966f4bd97f611354c4ad829f1ed298df9386c2ec
# latest-amd64 -> 966f4bd97f611354c4ad829f1ed298df9386c2ec
# https://github.com/GoogleContainerTools/distroless/tree/master/base

LABEL name="render-template"
LABEL repository="http://github.com/chuhlomin/render-template"
LABEL homepage="http://github.com/chuhlomin/render-template"

LABEL maintainer="Konstantin Chukhlomin <mail@chuhlomin.com>"
LABEL com.github.actions.name="Render template"
LABEL com.github.actions.description="Renders file based on template and passed variables"
LABEL com.github.actions.icon="file-text"
LABEL com.github.actions.color="purple"

COPY --from=build-env /go/bin/app /app

CMD ["/app"]
