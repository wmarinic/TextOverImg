# TextOverImg
Naive implementation of an app that users can submit text and an image URL to.
The app returns the image with the text placed over it. Users can login to remove the image watermark.

The login is hard coded to username: test, password: test for now.

## Getting Started
### Dependencies, Docker and Postgres (for windows users)
Install dependencies and golang-migrate.
```
go get ./...
go get -u -tags 'postgres' github.com/golang-migrate/migrate/cli
```
Running db (WSL) and migrations (cmd).
```
docker run -e POSTGRES_USER=local -e POSTGRES_PASSWORD=pass -e POSTGRES_DB=inspirationifierdb -p 5432:5432 postgres:11.10-alpine
migrate -path ./migrations -database "postgres://local:pass@localhost:5432/inspirationifierdb?sslmode=disable" up
```
In WSL, pull the latest version of sqlc and re-generate  sqlc bindings
```
docker pull kjconroy/sqlc
docker run --rm -v $(pwd):/src -w /src
```

### Build and Run
Can now build and run the code.
```
go build
TextOverImg
```
Alternatively.
```
go run main.go
```

### Creating Migrations
To create extra migrations, use the following command.
```
migrate create -ext sql -dir ./migrations -seq create_users_table
```

### Example queries to check endpoints:
```
curl -X POST -d "{\"url\": \"image-url-here.jpg\", \"text\": \"Inpsiration Quote Here!\"}" http://localhost:3000/image

curl -X POST -d "{\"username\": \"test\", \"password\": \"test\"}" http://localhost:3000/login

curl -X POST -d "{\"username\": \"test1\", \"password\": \"test\"}" http://localhost:3000/register
```

### TODO / improvements:
- better looking frontend
- store and serve images from a AWS S3 bucket / azure blob ?
