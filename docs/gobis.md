This document is step-by-step tutorial of the deployment and configuration of a gobis server
on cloud foundry.

All examples are given when working with [PCFDev](https://pivotal.io/pcf-dev),
a local cloud foundry instance.

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-refresh-toc -->
**Table of Contents**

- [1 Create UAA client for gobis](#1-create-uaa-client-for-gobis)
    - [I. Get endpoint](#i-get-endpoint)
    - [II. Configure cli](#ii-configure-cli)
    - [III. Log as admin](#iii-log-as-admin)
    - [IV. Create client](#iv-create-client)
- [2 Configure gobis](#2-configure-gobis)
    - [I. Get gobis-server code](#i-get-gobis-server-code)
    - [II. Create configuration file](#ii-create-configuration-file)
    - [III. Push configuration to cloud foundry](#iii-push-configuration-to-cloud-foundry)
- [3 Deploy gobis](#3-deploy-gobis)
    - [I. Push gobis application cloud foundry](#i-push-gobis-application-cloud-foundry)
    - [II. Declare a route going through gobis](#ii-declare-a-route-going-through-gobis)
- [4 What now ?](#4-what-now-)

<!-- markdown-toc end -->


# 1 Create UAA client for gobis

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

We create the UAA client used by gobis. This client requires *openid*,
*cloud_controller.admin* and *scim.read* from user and also
needs *openid* has its own authority.

[see doc](https://docs.cloudfoundry.org/concepts/architecture/uaa.html)

For grant types, we'll need *authorization_code* to get user's token from uaa
[(doc)](https://aaronparecki.com/oauth-2-simplified/#authorization), server
as well as *refresh_token* to automatically renew user's token
[(doc)](https://docs.cloudfoundry.org/api/uaa/version/4.7.0/index.html#refresh-token)

Redirect URIs is a list of trusted callbacks to send user after authorization is complete.
These callbacks are implemented by gobis to handle authorization code for us.
Note that we specify wildcard hostname, this allows us to use the same gobis server
to authenticate user on multiple application.
```
$ uaac client add -i

Client ID:  gobis-admin
New client secret:  *******
Verify new client secret:  *******
scope (list):  openid scim.read cloud_controller.admin
authorized grant types (list):  refresh_token authorization_code
authorities (list):  openid
access token validity (seconds):
refresh token validity (seconds):
redirect uri (list): https://*.local.pcfdev.io/login http://*.local.pcfdev.io/login
autoapprove (list):  true
signup redirect url (url):
```

# 2 Configure gobis

## I. Get gobis-server code

```bash
go get github/orange-cloudfoundry/gobis-server
cd $GOPATH/src/github/orange-cloudfoundry/gobis-server
```

## II. Create configuration file

In the following:
 - ```insecure_skip_verify```: set to true if working with self-signing ssl certificates (like in pcfdev)
 - ```forwarded_header```: we tell gobis that is should redirect user to url given is this header instead
   of redirecting always to same url. This header will be added by cloud foundry as we'll bind our app
   to a route-service that points to gobis
 - ```middleware_params.oauth2.client_id```: client name we just created in the UAA
 - ```middleware_params.oauth2.client_secret```: password you chose for this new client
 - ```middleware_params.oauth2.*_uri```: endpoints to our oauth2 server, here cloud foundry'a UAA,
 - ```middleware_params.oauth2.auth_key```: a strong secret used to encode user session's cookie
 - ```middleware_params.oauth2.scopes```: scopes that gobis will require at user login, this heavily
   depends on what proxifed applications will need.
 - ```middleware_params.jwt.secret```: public signing key of our oauth2 jwt tokens. You can get
   this information with the following command ```uaac signing key```. Get the returned *value* section
   and put that into the json string (add **\n** for newlines).


```bash
cat - > config.json <<EOF
{
  "host" : "0.0.0.0",
  "log_level": "error",
  "protected_headers": [],
  "routes": [
    {
      "name": "app",
      "path": "/**",
      "insecure_skip_verify" : true,
      "show_error": true,
      "no_buffer": false,
      "remove_proxy_headers" : false,
      "forwarded_header" : "X-CF-Forwarded-Url",
      "middleware_params": {
        "oauth2" : {
          "enabled"              : true,
          "client_id"            : "gobis-admin",
          "client_secret"        : "<gobis-admin-secret>",
          "authorization_uri"    : "https://uaa.local.pcfdev.io/oauth/authorize",
          "access_token_uri"     : "https://uaa.local.pcfdev.io/oauth/token",
          "user_info_uri"        : "https://uaa.local.pcfdev.io/userinfo",
          "auth_key"             : "<put-here-a-strong-key>",
          "login_path"           : "/login",
          "logout_path"          : "/logout",
          "use_redirect_url"     : true,
          "use_route_transport"  : true,
          "insecure_skip_verify" : true,
          "pass_token"           : true,
          "scopes"               : [ "openid", "cloud_controller.admin", "scim.read" ]
        },
        "jwt" : {
          "enabled" : true,
          "alg": "RS256",
          "secret" : "<uaa-public-signing-key>"
        },
        "info_page" : {
          "enabled": true,
          "path": "/auth/user_info",
          "show_authorization_header": true,
          "authorization_key_name": "token"
        }
      }
    }
  ]
}
EOF
```

In the end, it should look something like this:
```json
{
  "host" : "0.0.0.0",
  "log_level": "debug",
  "protected_headers": [],
  "routes": [
    {
      "name": "app",
      "path": "/**",
      "insecure_skip_verify" : true,
      "show_error": true,
      "no_buffer": false,
      "remove_proxy_headers" : false,
      "forwarded_header" : "X-CF-Forwarded-Url",
      "middleware_params": {
        "oauth2" : {
          "enabled"              : true,
          "client_id"            : "gobis-admin",
          "client_secret"        : "my-super-strong-secret",
          "authorization_uri"    : "https://uaa.local.pcfdev.io/oauth/authorize",
          "access_token_uri"     : "https://uaa.local.pcfdev.io/oauth/token",
          "user_info_uri"        : "https://uaa.local.pcfdev.io/userinfo",
          "auth_key"             : "my-super-cookie-secret",
          "login_path"           : "/login",
          "logout_path"          : "/logout",
          "use_redirect_url"     : true,
          "use_route_transport"  : true,
          "insecure_skip_verify" : true,
          "pass_token"           : true,
          "scopes"               : [ "openid", "cloud_controller.admin", "scim.read" ]
        },
        "jwt" : {
          "enabled" : true,
          "alg": "RS256",
          "secret" : "-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDHFr+KICms+tuT1OXJwhCUmR2d\nKVy7psa8xzElSyzqx7oJyfJ1JZyOzToj9T5SfTIq396agbHJWVfYphNahvZ/7uMX\nqHxf+ZH9BL1gk9Y6kCnbM5R60gfwjyW1/dQPjOzn9N394zd2FJoFHwdq9Qs0wBug\nspULZVNRxq7veq/fzwIDAQAB\n-----END PUBLIC KEY-----"
        },
        "info_page" : {
          "enabled": true,
          "path": "/auth/user_info",
          "show_authorization_header": true,
          "authorization_key_name": "token"
        }
      }
    }
  ]
}
```

## III. Push configuration to cloud foundry

Here we create a user-provided service with our configuration.
Gobis will read its configuration automatically using [Gautocloud](https://github.com/cloudfoundry-community/gautocloud).

Note: It is important that the service name contains **config** keyword.

```bash
$ cf cups gobis-config -p ./config.json
```

# 3 Deploy gobis

## I. Push gobis application cloud foundry

First, we create a little deploy directory with a manifest:
```bash
$ mkdir deploy
$ cp ${GOPATH}/bin/gobis-server deploy/
$ cd deploy
$ cat - > manifest.yml <<EOF
---
applications:
  - name: gobis
    memory: 256M
    instances: 1
    buildpack: binary_buildpack
    command: ./gobis-server
    services:
      - gobis-config
EOF
```

Then push the application to cloudfoundry.
```bash
$ cf push
```

## II. Declare a route going through gobis

Finally, we declare a user-provided route service to cloud foundry that goes through
out freshly deployed gobis server

```bash
$ cf cups gobis-auth-service -r https://gobis.local.pcfdev.io
```


# 4 What now ?

Now we can bind this route-service to any application using the command ```cf bind-route-service```.

You can try out gobis with [cf-wall](https://github.com/orange-cloudfoundry/cf-wall),
a mailing web application for cloud foundry that is natively compatible with gobis.

Step-by-step tutorial [here](https://github.com/orange-cloudfoundry/cf-wall/blob/mater/docs/configuration.md)


<!-- Local Variables: -->
<!-- ispell-local-dictionary: "american" -->
<!-- End: -->
