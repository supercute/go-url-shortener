### Simple URL Shortener written in Go

#### Features

- Create short links
- Redirect to original link
- JWT authentication
- Admin functions (delete users, get users list)

#### Installation

1. Clone the repository
2. Copy `.env.example` to `.env` and fill in the required variables
3. Change the `admin_token` in `configs/config.json` ( It's just an pet project, so I can store it in the config.json :D )
4. Run `go mod download`
5. Run `go build`
6. Run `./go-url-shoterer`
7. Read the documentation in `docs/docs.md`