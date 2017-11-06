# Configuration

## Create UAA client

1. Get endpoint

   ```bash
   $ cf curl /info | jq -r .token_endpoint

   https://uaa.local.pcfdev.io
   ```

2. Configure cli

   This requires [UAA Cli](https://github.com/cloudfoundry/cf-uaac). On ubuntu 16.04, installation from
   gem requires ```ruby-dev``` apt package.

   ```
   $ uaac target https://uaa.local.pcfdev.io
   ```

3. Log as admin

   ```
   $ uaac token client get

   Client ID:  <my-admin-user>
   Client secret:  <my-admin-secret>

   Successfully fetched token via client credentials grant.
   Target: https://uaa.local.pcfdev.io
   Context: <my-admin-secret>, from client <my-admin-secret>
   ```

4. Create cf-wall client in UAA

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

## Create config service

First, create a local configuration file. You can start from ```./config/cf-wall.json.sample```

```
$ cp ./config/cf-wall.json.sample ./config/cf-wall.json
```

Then adapt values for your environment.
```
{
  // url to cloud-foundry api endpoint
  "cc-url"            : "https://api.local.pcfdev.io",

  // TLS certificate and key, if any.
  "http-cert"         : "",
  "http-key"          : "",

  // HTTP server port. Local use only, it will be overriden by cloudfoundry PORT env variable
  "http-port"         : 80,

  // Logger serverity level
  "log-level"         : "debug",

  // UAA client name and secret
  "uaa-client"        : "cf-wall",
  "uaa-secret"        : "cf-wall",

  // url to UAA api endpoint, same used in uaac
  "uaa-url"           : "https://uaa.local.pcfdev.io",

  // email origin for all mails sent by cf-wall
  "mail-from"         : "root@localhost"
}
```

Then create cloudfoundry user provided service with theses variables
```
$ cf cups cf-wall-config -p ./config/cf-wall.json
```

## Create smtp service

Create service from broker if your cloudfoundry instance provides one. Make sure the service
is tagged with **email** or **smtp**.

Otherwise you can create another user-provided service as follow :
```
$ cf cups cf-wall-smtp -p '{"host" : "<my-smtp-host>", "port" : <my-smtp-port>}'
```

Because we can't provide tags to user provided service ([issue 5-11-2017](https://github.com/cloudfoundry/cli/issues/1110))
the service name **must** contains the word *smtp* in order for cf-wall to detect it and populate its configuration
correctly.

## Deploy application

```
$ cf push cf-wall
```



<!-- Local Variables: -->
<!-- ispell-local-dictionary: "american" -->
<!-- End: -->
