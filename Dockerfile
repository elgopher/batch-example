FROM golang:1.18.2-alpine
WORKDIR /app
COPY . ./
RUN go build -o /batch-example
EXPOSE 8080
CMD [ "/batch-example" ]
