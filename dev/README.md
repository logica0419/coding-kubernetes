# Development Guide

## Requirements

- [Task](https://taskfile.dev/)
- [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/)

## Workspace Initialization

```sh
sh init.sh
```

## Run Development Server

Start a local development server with live reloading:

```sh
task dev
```

Access the documentation preview at [http://localhost:8000](http://localhost:8000).

## Build Static Site

Generate a static build of the documentation:

```sh
task build
```

The output will be in the `site/` directory and ready for deployment.
