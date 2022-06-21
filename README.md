# CMS Service

---
- Before starting, you can visit the [alpha](https://hub.alpha.kidsloop.net/) to experience.
- Visit the [API documentation](https://swagger-ui.kidsloop.net/) to view the interface provided by the cms service

## Preparation

---
- Golang Installation  `version v1.18+`
- configure git `.gitconfig`
```text
[url "ssh://git@github.com/"]
        insteadOf = https://github.com/
```
- mysql `version 5.7`
- redis `version 5.0.6`
- [flyway](https://flywaydb.org/) 
- [gin-swagger](https://github.com/swaggo/gin-swagger)

## Environment variable

---

| Env Var                               | Description                                              |
|---------------------------------------|----------------------------------------------------------|
| storage_protocol                      | storage protocol                                         |
| storage_bucket                        | storage bucket                                           |
| storage_region                        | storage region                                           |
| storage_endpoint                      | storage endpoint                                         |
| storage_download_mode                 | storage download mode native/cloudfront                  |
| storage_sig_mode                      | true/false                                               |
| storage_accelerate                    | true/false                                               |
| cdn_path                              | cdn path                                                 |
| cdn_key_id                            | cdn key id                                               |
| cdn_private_key_path                  | cdn private key path                                     |
| -                                     |                                                          |
| connection_string                     | mysql connection string                                  |
| max_open_conns                        | mysql max open connect                                   |
| max_idle_conns                        | mysql max idle connect                                   |
| conn_max_life_time                    | mysql connect max life time                              |
| conn_max_idle_time                    | mysql connect max idle time                              |
| slow_threshold                        | mysql slow threshold                                     |
| -                                     |                                                          |
| open_cache                            | whether to enable caching                                |
| redis_host                            | redis host                                               |
| redis_port                            | redis port                                               |
| redis_password                        | redis password                                           |
| -                                     |                                                          |
| schedule_review_type_enabled          | whether to enable the review type schedule               |
| max_repeat_year                       | The maximum age limit for repeatedly creating a schedule |
| cache_expiration                      | cache expiration                                         |
| -                                     |                                                          |
| h5p_private_key_path                  | h5p private key path                                     |
| -                                     |                                                          |
| live_token_private_key_path           | live token private key path                              |
| live_assets_url_prefix                | live assets url prefix                                   |
| -                                     |                                                          |
| ams_endpoint                          | ams endpoint                                             |
| jwt_public_key_path                   | jwt public key path                                      |
| user_service_api_key                  | user service api key                                     |
| -                                     |                                                          |
| h5p_endpoint                          | assessment service endpoint                              |
| -                                     |                                                          |
| data_service_endpoint                 | data service endpoint                                    |
| data_service_api_key                  | data service api key                                     |
| data_service_public_key               | data service public key                                  |
| -                                     |                                                          |
| cors_domain_list                      | cors domain list                                         |
| cors_allow_file_protocol              | cors allow file protocol                                 |
|                                       |                                                          |
| user_cache_expiration                 | set user cache expiration                                |
| user_permission_cache_expiration      | set user permission cache expiration                     |
| -                                     |                                                          |
| NEW_RELIC_APP_NAME                    | newRelic app name                                        |
| NEW_RELIC_LICENSE_KEY                 | newRelic license key                                     |
| NEW_RELIC_DISTRIBUTED_TRACING_ENABLED | newRelic distributed tracing enabled                     |
| NEW_RELIC_LABELS                      | newRelic labels                                          |
|                                       |                                                          |
| LOG_LEVEL                             | set log level                                            |


## Development Specification

---
- git commit specification [refer to this specification](https://www.conventionalcommits.org/en/v1.0.0/)
```text
^(feat|fix|docs|refactor|test|ci|chore)\(\w+\-\d+(\s\w+\-\d+)*?\):\s.+$
example：
git commit -m "feat(NKL-1): add new api xxxxxx"
git commit -m "feat(NKL-1 CNCD-2 TEST-12345): add new api xxxxxx"
```

## Development Flow

---
- Database Design  [DB Schema](https://calmisland.atlassian.net/wiki/spaces/NKL/pages/991363132/DB+Schema)
  - The migration tool used is flyway, [more commands](https://flywaydb.org/documentation/command/migrate)
      ```bash
      flyway info # View flyway information, and database connection status
      flyway migrate # execute sql operation
      ```
- API interface design and provide interface documentation to the front end
   ```bash
      swag init # Generate api documents locally to verify whether there are any problems with the documents
    ```
- Business development
- Unit test functionality
    ```bash
      go test -v example_test.go
    ```
- Push the code to `github`, wait for review
  ```bash
    go build  # Check for compilation problems before committing
    git add .
    git commit -m "feat(NKL-1): add new api xxxxxx" # Reference code submission specification
    git push
  ```
- After deploying to the alpha environment, interface with the front-end for joint debugging and testing
  - [Alpha front-end address](https://auth.alpha.kidsloop.net/)
  - [Alpha backend address](https://cms.alpha.kidsloop.net/v1/ping)

## Project structure

---
   ```Plain Text
   ├── api              // API layer, interface definition
   ├── cmd              // Executable tools are defined here
   ├── config           // Project related configuration items
   ├── constant
   ├── da               // Data access layer for operating the database
   ├── deploy
   ├── entity           // Define business entities and database entities in the project
   ├── external         // Get data provided by external services
   ├── model            // business logic layer
   ├── mq
   ├── mutex
   ├── schema           // Database Migration Script
   ├── test
   ├── utils    
   └── main.go            
   ```


## External KidsLoop dependencies

---
- [User Service](https://github.com/KL-Engineering/user-service) GraphQL API to get basic user information
- [Assessment Service](https://github.com/KL-Engineering/kidsloop-assessment-service) GraphQL API to get student assessment information (scores, etc.)


## Useful Tools

---

- [VSCode](https://code.visualstudio.com/)
- [Goland](https://www.jetbrains.com/go/promo/)
- [Postman](https://www.postman.com/)  for testing API requests
- [gin-swagger](https://github.com/swaggo/gin-swagger) for generating api documentation
- [flyway](https://flywaydb.org/) for database Migration Tool


