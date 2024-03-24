
# NPM private registry over the OCI protocol (POC)

The current project is a Proof-of-concept of using OCI as a backend for the NPM registry.


## What is it good for?

Store your private NPM packages safely as artefacts in your existing OCI compliant registry infrastructure.

### Usage

TBD

### Example

Please check the [testdata](testdata) folder

#### Supported API methods

* npm login with legacy auth method (any credentials)
* npm publish
* npm install

### Environment Variables

There are not many options in configure the application except the following.

* `PORT` - specifies port, default `9000`
* `DEBUG` - enabled debug if it's `TRUE`
* `USE_TLS` - enabled HTTP over TLS
* `OCI_URL` - OCI compliant registry e.g. `oci://admin:bitnami@localhost/library`
### TODO

* Add authorisation flows
* CI/CD Pipeline with GitHub Action
* Add tests
