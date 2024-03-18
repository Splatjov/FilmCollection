## FilmCollection test task for VK
### Running
1. Create .env file
```shell
cp .env.example .env
```
2. Run docker-compose
```shell
docker-compose up -d
```

### API
Documentation on Swagger can be found at oas.yaml

### Testing
```shell
go test -v ./... 
```

### About realization
- Used pure golang http (new 1.22 router), without any frameworks.
- For database used PostgreSQL (pgx driver).
- Documentation on Swagger 3.0
- Docker & docker-compose for running
- With outside logging (can be found in logs folder)
- With outside ports, hosts, etc (specified in .env file)
