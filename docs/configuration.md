<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-refresh-toc -->
**Table of Contents**

- [1. Create UAA client for cf-wall](#1-create-uaa-client-for-cf-wall)
    - [I. Get endpoint](#i-get-endpoint)
    - [II. Configure cli](#ii-configure-cli)
    - [III. Log as admin](#iii-log-as-admin)
    - [IV. Create client](#iv-create-client)
- [2. Configure cf-wall](#2-configure-cf-wall)
    - [I. Create configuration file](#i-create-configuration-file)
    - [II. Push configuration to cloud foundry](#ii-push-configuration-to-cloud-foundry)
    - [III. Create smtp service](#iii-create-smtp-service)
- [3. Deploy application](#3-deploy-application)

<!-- markdown-toc end -->

# 1. Create UAA client for cf-wall

## I. Get endpoint

```bash
$ cf curl /info | jq -r .token_endpoint

https://uaa.local.pcfdev.io
```

## II. Configure cli

This requires [UAA Cli](https://github.com/cloudfoundry/cf-uaac). On ubuntu 16.04, installation from
gem requires ```ruby-dev``` apt package.

```
$ uaac target https://uaa.local.pcfdev.io
```

## III. Log as admin

```
$ uaac token client get

Client ID:  <my-admin-user>
Client secret:  <my-admin-secret>

Successfully fetched token via client credentials grant.
Target: https://uaa.local.pcfdev.io
Context: <my-admin-secret>, from client <my-admin-secret>
```

**Hint**: on pcfdev, the default secret for **admin** client is **admin-client-secret**


## IV. Create client

We create a client used by cf-wall to fetch user mails from UAA.

This client is given *scim.read* and *clients.read* authorities
[see doc](https://docs.cloudfoundry.org/concepts/architecture/uaa.html)

```
$ uaac client add -i

Client ID:  cf-wall
New client secret:  *******
Verify new client secret:  *******
scope (list):  uaa.none
authorized grant types (list):  client_credentials
authorities (list):  scim.read,clients.read
access token validity (seconds):
refresh token validity (seconds):
redirect uri (list):
autoapprove (list):  true
signup redirect url (url):
```


# 2. Configure cf-wall

## I. Create configuration file

First, create a local configuration file.
You can start from ```./config/cf-wall.json.sample```

```
$ cp ./config/cf-wall.json.sample ./config/cf-wall.json
```

Then adapt values for your environment.
```
{
  // UAA client name and secret
  "uaa-client"        : "cf-wall",
  "uaa-secret"        : "cf-wall",

  // url to UAA api endpoint, same used in uaac
  "uaa-url"           : "https://uaa.local.pcfdev.io",

  // don't check ssl certificates when working with pcfdev
  "uaa-skip-verify"    : true,

  // url to cloud-foundry api endpoint
  "cc-url"            : "https://api.local.pcfdev.io",

  // don't check ssl certificates when working with pcfdev
  "cc-skip-verify"    : true,

  // TLS certificate and key, if any.
  "http-cert"         : "",
  "http-key"          : "",

  // HTTP server port. Local use only, it will be overriden by cloudfoundry PORT env variable
  "http-port"         : 80,

  // Logger serverity level
  "log-level"         : "debug",

  // email origin for all mails sent by cf-wall
  "mail-from"         : "root@localhost",

  // when true, don't actually send mails, test only
  "mail-dry": false,

  // static list of carbon copy recipients
  "mail-cc" : "root@localhost",

  // tag to automatically prepend to mail subjects
  "mail-tag": "[cf-wall]",

  // prase html template at each requests (test only)
  "reload-templates" : false
}
```

## II. Push configuration to cloud foundry

Create cloud foundry user provided service with theses variables.

```
$ cf cups cf-wall-config -p ./config/cf-wall.json
```

## III. Create smtp service

Create service from broker if your cloud foundry instance provides one. Make sure the service
is tagged with **email** or **smtp**.

Otherwise you can create another user-provided service as follow :
```
$ cf cups cf-wall-smtp -p '{"host" : "<my-smtp-host>", "port" : <my-smtp-port>}'
```

Because we can't provide tags to user provided service ([issue 5-11-2017](https://github.com/cloudfoundry/cli/issues/1110))
the service name **must** contains the word *smtp* in order for cf-wall to detect it and populate its configuration
correctly.

# 3. Deploy cf-wall

## I. Push cf-wall application to cloud foundry

First, we create a little deploy directory with a manifest:
```bash
$ cat - > manifest.yml <<EOF
---
applications:
  - name: cf-wall
    memory: 256M
    instances: 1
    buildpack: go_buildpack
    services:
      - cf-wall-config
      - cf-wall-smtp
EOF
```

Then push the application to cloudfoundry.
```bash
$ cf push
```

# 4. Deploy application

In the section, we suppose a running gobis route service that will act as
a reverse proxy to authenticate cf-wall's users.

You can install and configure gobis by following this
[tutorial](./gobis.md).



<!-- Local Variables: -->
<!-- ispell-local-dictionary: "american" -->
<!-- End: -->
