# external-modules-transfer
This tool helps to transfer deckhouse external modules images from one container registry to another.

# Usage
```text
Usage of external-modules-transfer:

  This tool helps to transfer deckhouse external modules images
  from one container registry to another.

  -module string
        external module name
  -pull-ca string
        ca certificate for pull registry
  -pull-disable-auth
        disable auth for pull registry
  -pull-insecure
        use http protocol for pull registry
  -pull-registry string
        registry address, that contains external modules
        (you should be logged in to registry via docker login)
  -push-ca string
        ca certificate for push registry
  -push-disable-auth
        disable auth for push registry
  -push-insecure
        use http protocol for push registry
  -push-registry string
        registry address to push external module from pull repo
        (you should be logged in to registry via docker login)
  -release string
        release channel to use (default "alpha")
```
# Installation
Requirements: Go >= 1.18
```
git clone https://github.com/alex123012/external-modules-transfer.git
cd external-modules-transfer
make build # or use `make build-macos` if you are on mac
```
# Example
```bash
docker login registry.exmaple-pull.com/external-modules
...
docker login registry.example-push.com/deckhouse-external-modules
....
external-modules-transfer --module external-module-name --pull-registry "registry.exmaple-pull.com/external-modules" --push-registry "registry.example-push.com/deckhouse-external-modules" --release alpha
```