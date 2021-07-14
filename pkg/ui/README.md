# Morphling User Interface

This is the source code for the Morphling UI.

## Folder structure


2. Go backend you can find under [pkg/ui](./).

3. React frontend you can find under [pkg/ui/frontend](./frontend).

## Development

While development you have different ways to run Morphling UI.

### Serve UI frontend and backend

1. Run `npm run build` under `/frontend` folder. It will create `/frontend/build` directory with optimized production build.

2. Go to `cmd/ui`.

3. Run `main.go` file with appropriate flags. 
```
go run main.go --build-dir=../../pkg/ui/frontend/build/ --port=8082
```

After that, you can access the UI using this URL: `http://localhost:8082/morphling/`.


## Code style

To make frontend code consistent and easy to review we use [Prettier](https://prettier.io/). 

### IDE integration

For VSCode you can install plugin: "Prettier - Code formatter" and it will pick Prettier config automatically.

You can edit [settings.json](https://code.visualstudio.com/docs/getstarted/settings#_settings-file-locations) file for VSCode to autoformat on save.

```json
  "settings": {
    "editor.formatOnSave": true
  }
```

For others IDE see [this](https://prettier.io/docs/en/editors.html).

### Check and format code

Before submitting PR check and format your code. To check your code run `npm run format:check` under `/frontend` folder. To format your code run `npm run format:write` under `/frontend` folder.
If all files formatted you can submit the PR.

