FROM golang:1.11
ADD . /go/src/pill-recommender/
RUN go get github.com/gorilla/handlers
Run go get github.com/gorilla/mux
RUN go install /go/src/pill-recommender
ENTRYPOINT /go/bin/pill-recommender
EXPOSE 8082
