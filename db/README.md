# Setup database

## Prerequisites

Create your Google Cloud project and configure it for gcloud.

```sh
gcloud config set project YOUR-PROJECT-ID
```


## Configure Cloud Spanner

Enable Cloud Spanner API.

```sh
gcloud services enable spanner.googleapis.com
```

Create a Spanner instance.

```sh
gcloud spanner instances create demo-instance \
  --config regional-asia-northeast1 \
  --description "instance for demo" \
  --processing-units 100
```

Create a Spanner database.

```sh
gcloud spanner databases create api-database \
  --instance demo-instance
```

## Configure schema

Install [hammer](https://github.com/daichirata/hammer), a CLI to manage Cloud Spanner schema.

```sh
go install github.com/daichirata/hammer@latest
```

Configure application default credentials if you haven't done this.

```sh
gcloud auth application-default login
```

Confirm diff.

```sh
hammer diff \
  spanner://projects/$(
    gcloud config get-value project
  )/instances/demo-instance/databases/api-database \
  schema.sql
```

Apply the schema.

```sh
hammer apply \
  spanner://projects/$(
    gcloud config get-value project
  )/instances/demo-instance/databases/api-database \
  schema.sql
```
