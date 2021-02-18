# Render Template

GitHub Action to render file based on template and passed variables.

## Inputs

* `template` – path to Go template file
* `vars` – template variables in YAML format

## Output

`result` – rendered template

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
      ...
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
