### Test data

* npm-library folder is a simple NPM package which we want to store in private registry.
* npm-project folder is the project which utilises the library.


### Steps 

Build registry
```shell
./do.sh build
```

Update location and credentials for your OCI compliant storage and run the registry.
```shell
export OCI_URL=oci://admin:bitnami@localhost/library
./do.sh run
```

```shell
cd npm-library
npm login --scope=@corp
```
Use arbitrary data as credentials. After successful login run the following command:

```shell
npm publish
```

You should see `Publishing to https://localhost:9000 with tag latest and default access`

Now let's try installing our private package we just published
```shell
cd npm-project
npm i @corp/npm-library
```
You should see successful message like `added 1 package, and audited 2 packages in 2s`
