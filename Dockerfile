FROM golang:1.26-alpine

WORKDIR /app

# 開発中にコードが変わったら自動で反映させるツール（air）
RUN go install github.com/air-verse/air@latest

CMD ["air"]