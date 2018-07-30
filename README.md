# genspec

[![Build Status](https://travis-ci.org/kubeflow-incubator/genspec.svg?branch=master)](https://travis-ci.org/kubeflow-incubator/genspec)

This tool is used for generating OpenAPI specification `swagger.json` for [Kubeflow/tf-operator](https://github.com/kubeflow/tf-operator).
The specification includes model definitions and routing information, which is necessary to generate client libraries.

## Installation

```
$ go install github.com/kubeflow-incubator/genspec
```

## Generate Swagger

```
$ genspec --output swagger.json
```

Complete usage:

```
$ genspec --help
Generate OpenAPI specification for TFJob

Usage:
  genspec [flags]

Flags:
  -h, --help            help for genspec
      --output string   Path to write OpenAPI spec file (default "swagger.json")
```

## Acknowledgements

This work is inspired by [tamalsaha/kube-openapi-generator](https://github.com/tamalsaha/kube-openapi-generator).