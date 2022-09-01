
## Usage
The following assumes you have the plugin installed via

```shell
kubectl krew install {{ .PluginName }}
```

### Scan images in your current kubecontext

```shell
kubectl {{ .PluginName }}
```

### Scan images in another kubecontext

```shell
kubectl {{ .PluginName }} --context=context-name
```

## How it works
Write a brief description of your plugin here.