# CMS Service

---
- Before starting, you can visit the [alpha](https://hub.alpha.kidsloop.net/) to experience.
- Visit the [API documentation](https://api.alpha.kidsloop.net/user/) to view the interface provided by the cms service

## Preparation

---
- Golang Installation  `go v1.18`
- configure git `.gitconfig`
```text
[url "ssh://git@github.com/"]
        insteadOf = https://github.com/
```

## Development Flow

---

- Database Design
- API interface design and provide interface documentation to the front end
- Business development
- Unit test functionality
- Push the code to `github` , wait for review
- After deploying to the alpha environment, interface with the front-end for joint debugging and testing

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