# Development Guide

## Requirements

- [bun](https://bun.sh/)

## Workspace Initialization

```sh
bun install
```

## Run Development Server

Start a local development server with live reloading:

```sh
bun run dev
```

Access the documentation preview at [http://localhost:5173/](http://localhost:5173/).

## Build Static Site

Generate a static build of the documentation:

```sh
bun run build
```

The output will be in the `.vitepress/dist` directory and ready for deployment.
