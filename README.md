# Blogo

Blogo is a Go package that provides a simple solution for creating a simple blog. It allows you to manage posts (create, edit, view and delete) with authentication to protect destructive functions.

## Features

- **Post Management:** Create, edit, view and delete posts.
- **Authentication:** Simple login system to secure administrative routes.
- **Markdown Rendering:** Converto Markdown content to HTML using Goldmark.
- **PostgreSQL DB:** Utilizes SQLC for query generation and pgx for database connectivity.
- **Database Migrations:** Supports database migrations using [golang-migrate](https://github.com/golang-migrate/migrate).

## Technologies Used

- **GO**
- **PGX for PostgreSQL Connection**
- **SQLC for auto-generated SQL queries**
- **Goldmark for markdown parsing**

## Design and Configuration

The current design is fixed, allowing configuration only for the blog name, the page title and a single administrator.

## Usage Example

An example of how to use the Blogo package is provided in [`/cmd/blog/main.go`](cmd/blog/main.go) file. In this file, you can see how to:

- Set up environment variables.
- Initialize the database connection.
- Run migrations.
- Configure Blogo with the necessary dependencies ( uth, log, location and queries )
- Start the server.

This repository is also deployed as my personal Blog at [Luiz Gustavo](https://luizgustavojunqueira.up.railway.app/)

## Running the project locally

1. Set up the necessary environment variables defined in [example](.env.example)
2. Run the following command to start the database and the server:

```bash
docker compose up database && docker compose up blog_backend
```

3. Or run just the database container and live relaod the code with air:

```bash
docker compose up database && air
```

## Deploying

For deploying your blog, there is a dockerfile provided.

# Roadmap

- [x] Auto generate table of contents
- [x] Add posts descriptions
- [x] Add read time
- [x] Add posts tags
- [x] Group posts by year and/or month
- [ ] Redesign the post editor
- [ ] Search by name and tag
- [ ] Tests
