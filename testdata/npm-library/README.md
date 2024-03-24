### Private NPM repo setup

Disable SSL verify (for dev purposes only)
```shell
npm config set strict-ssl false
```

Create a new scoped package
```shell
npm init --scope=@corp
```

Add `.npmrc` file
```shell
@corp:registry=https://localhost:9000/npm
```
OR

```shell
npm config set @corp:registry https://localhost:9000/npm
```
Login to the registry
```shell
npm  login --scope=@corp
```

User any credentials for username and password

Finally, publish your package

```shell
npm publish
```
