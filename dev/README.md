# Development Guide (Documentation)

## Requirements

- [Task](https://taskfile.dev/)
- [bun](https://bun.sh/)
- [Go](https://go.dev/)
- [golangci-lint](https://golangci-lint.run/)

## Workspace Initialization

```sh
sh init.sh
```

## Run Development Server

Start a local development server with live reloading:

```sh
task dev
```

Access the documentation preview at [http://localhost:5173/](http://localhost:5173/).

## Build Static Site

Generate a static build of the documentation:

```sh
task build
```

The output will be in the `.vitepress/dist` directory and ready for deployment.
