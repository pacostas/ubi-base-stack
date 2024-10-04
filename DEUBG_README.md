## Test locally

### Push image worfklow

```bash
act workflow_dispatch  -W .github/workflows/push-image.yml -P ubuntu-latest=-self-hosted -e ./push-image.dispatch.event.json -j push --env GITHUB_REPOSITORY="paketo-community/ubi-base-stack" --secret-file="./my.secrets"
```

the my.secrets must have below content

```env
PAKETO_BOT_GITHUB_TOKEN=
PAKETO_BUILDPACKS_DOCKERHUB_USERNAME=
PAKETO_BUILDPACKS_DOCKERHUB_PASSWORD=
```
