# How to deploy

## create cert

```
sh tls.sh
```

## deploy yamls

`cd deploy && kuberctl apply -f secret.yaml -f deployment.yaml -f service.yaml -f mutatingwebhook.yaml`

# How to use

label namespace
`kuberctl label ns default pod-admission-webhook-injection=enabled`
