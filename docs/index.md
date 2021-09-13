# qbec Provider

The qbec provider exposes a single data source that allows for evaluation of jsonnet code.
It uses an opinionated jsonnet VM as exposed by [qbec](https://qbec.io/) for this purpose.

The VM has all the features of a standard jsonnet VM and the following additions:

* [Additional native functions](https://qbec.io/reference/jsonnet-native-funcs/)
* [A glob importer](https://qbec.io/reference/jsonnet-glob-importer/) for importing a bag of files as an object
* [A data source importer](https://qbec.io/reference/jsonnet-external-data/) to import data from external sources.

-> Note: If you care about compatibility with the standard VM you should refrain from using these features.

## Example usage

```hcl
provider "qbec" {}
```
