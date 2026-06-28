## dev
```bash
swag init -g cmd/main.go

docker build -t firego .
docker run -p 54321:54321 firego
```

## prod
```bash
docker run -p 54321:54321 -e ENV=production firego
```