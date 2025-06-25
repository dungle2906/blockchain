# 🐳 Dùng image chính thức của Golang từ DockerHub
FROM golang:1.24.4

# 📂 Tạo thư mục làm việc bên trong container
WORKDIR /app

# 📥 Copy toàn bộ mã nguồn vào container
COPY . .

# 📦 Tải dependency go (go.mod, go.sum)
RUN go mod tidy

# 🔨 Build file chính (cmd/main.go) thành binary ./main
RUN go build -o main cmd/main.go

# 🚀 Khi container chạy → thực thi lệnh này
CMD ["./main"]
