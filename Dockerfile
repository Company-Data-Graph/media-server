FROM golang:latest

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build media-server
RUN chmod +x -R .


# ENV variables example :
# ENV MEDIA_SERVER_HOST="localhost"
# ENV MEDIA_SERVER_PORT=8082
# ENV MEDIA_SERVER_ADMIN_PASS="test"
# ENV MEDIA_SERVER_DATA_ROUTE_NAME="/data/"
# ENV MEDIA_SERVER_DATA_ROUTE_STORAGE_ROUTE="data/"

EXPOSE 8082

CMD [ "./media-server", "--mode=env" ]