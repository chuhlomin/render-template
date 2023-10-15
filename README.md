# Render Template

[![main](https://github.com/chuhlomin/render-template/actions/workflows/main.yml/badge.svg)](https://github.com/chuhlomin/render-template/actions/workflows/main.yml)
[![release](https://github.com/chuhlomin/render-template/actions/workflows/release.yml/badge.svg)](https://github.com/chuhlomin/render-template/actions/workflows/release.yml)
[![DockerHub](https://img.shields.io/badge/docker-hub-4988CC)](https://hub.docker.com/repository/docker/chuhlomin/render-template)

GitHub Action to render file based on template and passed variables.

## Inputs

| Name        | Description                                   | Required |
|-------------|-----------------------------------------------|----------|
| template    | Path to template                              | true     |
| vars        | Variables to use in template (in YAML format) | false    |
| vars_path   | Path to YAML file with variables              | false    |
| result_path | Desired path to result file                   | false    |
| timezone    | Timezone to use in `date` template function   | false    |

You must set at least `vars` or `vars_path`.  
You may set both of them (`vars` values will precede over `vars_path`).

Variables names must be alphanumeric strings (must not contain any hyphens).

There are few template functions available:

- `date` – formats timestamp using Go's [time layout](https://golang.org/pkg/time/#pkg-constants).  
  Example: `{{ "2023-05-11T01:42:04Z" | date "2006-01-02" }}` will be rendered as `2023-05-11`.  
  You may use `timezone` input to set timezone for `date` function (e.g. `timezone: "America/New_York"`).

- `mdlink` – creates markdown link.  
  Example: `{{ "https://github.com" | mdlink "GitHub" }}` will be rendered as `[GitHub](https://github.com)`.

- `number` – formats number in English locale.  
  Example: `{{ 1234567890 | number }}` will be rendered as `1,234,567,890`.

- `base64` – encodes string to base64.  
  Example: `{{ "hello" | base64 }}` will be rendered as `aGVsbG8=`.

## Outputs

| Name   | Description           |
|--------|-----------------------|
| result | Rendered file content |

## Example

`kube.template.yml`

```yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .deployment }}
  labels:
    app: {{ .app }}
spec:
  replicas: 3
  selector:
    matchLabels:
      app: {{ .app }}
  template:
    metadata:
      labels:
        app: {{ .app }}
    spec:
      containers:
      - name: {{ .app }}
        image: {{ .image }}
        ports:
        - containerPort: 80
```

`.github/workflows/main.yml`

```yml
name: main
on:
  push:
    branches:
      - main
env:
  DOCKER_IMAGE: username/image
  DEPLOYMENT_NAME: nginx-deployment
jobs:
  main:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3

      <...>

      - name: Render template
        id: render_template
        uses: chuhlomin/render-template@v1
        with:
          template: kube.template.yml
          vars: |
            image: ${{ env.DOCKER_IMAGE }}:${{ github.sha }}
            deployment: ${{ env.DEPLOYMENT_NAME }}
            app: nginx

      - name: Deploy
        timeout-minutes: 4
        run: |-
          echo '${{ steps.render_template.outputs.result }}' | kubectl apply -f -
          kubectl rollout status deployment/$DEPLOYMENT_NAME
```
