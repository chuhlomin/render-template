# Render Template

[![main](https://github.com/chuhlomin/render-template/actions/workflows/main.yml/badge.svg)](https://github.com/chuhlomin/render-template/actions/workflows/main.yml) [![release](https://github.com/chuhlomin/render-template/actions/workflows/release.yml/badge.svg)](https://github.com/chuhlomin/render-template/actions/workflows/release.yml)

GitHub Action to render file based on template and passed variables.

## Inputs

| Name        | Description                                   | Required |
|-------------|-----------------------------------------------|----------|
| template    | Path to template                              | true     |
| vars        | Variables to use in template (in YAML format) | false    |
| vars_path   | Path to YAML file with variables              | false    |
| result_path | Desired path to result file                   | false    |

You must set at least `vars` or `vars_path`.  
You may set both of them (`vars` values will precede over `vars_path`).

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
        uses: actions/checkout@v2

      <...>

      - name: Render template
        id: render_template
        uses: chuhlomin/render-template@v1.5
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
