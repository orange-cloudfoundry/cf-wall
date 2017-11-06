<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-refresh-toc -->
**Table of Contents**

- [Errors](#errors)
- [Endpoints](#endpoints)
    - [/orgs](#orgs)
    - [/spaces](#spaces)
    - [/orgs/{{org_guid}}/spaces](#orgsorgguidspaces)
    - [/services](#services)
    - [/buildpacks](#buildpacks)
    - [/users](#users)
    - [/message](#message)

<!-- markdown-toc end -->

# Errors

On error, cf-wall API replies with HTTP 400 or 500 with the following payload:

```
{
  // user error code
  "code": 10,
  // error description
  "error": "unable to read organizations from CC api"
}
```

**Fields**
- *code* User error code, anything below 50 is user related
- *description* Detailed description of the problem

| Code | Meaning                                              |
|------|------------------------------------------------------|
| 10   | Invalid or missing authorization header              |
| 50   | Could not communicate with Cloudfoundry API          |
| 51   | Invalid UAA credentials                              |
| 52   | Gautocloud error, could not fetch  SMTP credentials  |
| 53   | Could not communicate with SMTP server               |


# Endpoints

## /orgs

Get all available organizations.

* Method : GET
* Headers : Authorization (bearer)
* Reponse 200 :
  ```
  [
       {
           // organization name
           "name": "org-1",
           // organization guid
           "guid": "fc95a4c6-b07f-4b9b-871e-1f5d67b06071"
       },
       {
           "name": "org-2",
           "guid": "4ca50e06-2a3c-49d5-a1e6-dffa3cf6a7b4"
       },
       ...
  ]
  ```


## /spaces

Get all available spaces.

* Method : GET
* Headers : Authorization (bearer)
* Reponse 200 :

  ```
  [
       {
           // space name
           "name": "space-1",
           // space guid
           "guid": "fc95a4c6-b07f-4b9b-871e-1f5d67b06071",
           // organization guid that holds given space
           "org_guid" : "fc95a4c6-b07f-4b9b-871e-1f5d67b06071"
       },
       {
           "name": "space-2",
           "guid": "4ca50e06-2a3c-49d5-a1e6-dffa3cf6a7b4",
           "org_guid" : "ddddb760-ef75-40f4-9d52-1bb557b61af8"
       },
       ...
  ]
  ```

## /orgs/{{org_guid}}/spaces

Get spaces belonging to given **{{org_guid}}** organization.

* Method : GET
* Headers : Authorization (bearer)
* Reponse 200 :

  ```
  [
       {
           // space name
           "name": "space-1",
           // space guid
           "guid": "fc95a4c6-b07f-4b9b-871e-1f5d67b06071",
           // organization guid that holds given space
           "org_guid" : "fc95a4c6-b07f-4b9b-871e-1f5d67b06071"
       },
       {
           "name": "space-5",
           "guid": "4ca50e05-2a3c-49d5-a1e4-dffa3cf6a7b4",
           "org_guid" : "fc95a4c6-b07f-4b9b-871e-1f5d67b06071"
       },
       ...
  ]
  ```

## /services

Get all available services.

* Method : GET
* Headers : Authorization (bearer)
* Reponse 200 :

  ```
  [
       {
           // service name
           "name": "service-1",
           // service guid
           "guid": "fc95a4c6-b07f-4b9b-871e-1f5d67b06071"
       },
       {
           "name": "service-2",
           "guid": "4ca50e06-2a3c-49d5-a1e6-dffa3cf6a7b4"
       },
       ...
  ]
  ```

## /buildpacks

Get all available buildpacks.

* Method : GET
* Headers : Authorization (bearer)
* Reponse 200 :

  ```
  [
       {
           // service name
           "name": "service-1",
           // service guid
           "guid": "fc95a4c6-b07f-4b9b-871e-1f5d67b06071"
       },
       {
           "name": "service-2",
           "guid": "4ca50e06-2a3c-49d5-a1e6-dffa3cf6a7b4"
       },
       ...
  ]
  ```

## /users

Get all available users.

* Method : GET
* Headers : Authorization (bearer)
* Reponse 200 :

  ```
  [
       {
           // user name
           "name": "user-1",
           // user guid
           "guid": "fc95a4c6-b07f-4b9b-871e-1f5d67b06071"
       },
       {
           "name": "user-2",
           "guid": "4ca50e06-2a3c-49d5-a1e6-dffa3cf6a7b4"
       },
       ...
  ]
  ```


## /message

Send mail to given targets

* Method: POST

* Headers: Authorization (bearer)

* Request payload (all fields are mandatory):
  ```
  {
    // list of targeted organizations guids
    "orgs"       : [ "f3a76849-3324-4448-b36b-0f0c9392fc91", "61ecb1cb-47fa-4286-8d1b-8c279df65de7" ],

    // list of targeted spaces guids
    "spaces"     : [ "4ca50e06-2a3c-49d5-a1e6-dffa3cf6a7b4" ],

    // list of targeted services guids
    "services"   : [ "0f89a6da-5388-47f5-a20b-16a3f31522b1" ],

    // list of targeted buildpacks guids
    "buildpacks" : [ "0a01ace3-4a0f-458a-a78d-4a6ef6deeac8" ],

    // list of targeted users guids
    "users"      : [ "0a01ace3-4a0f-458a-a78d-4a6ef6deeac8", "61ecb1cb-47fa-4286-8d1b-8c279df65de7" ],

    // mail subject
    "subject" : "My Pretty Subject",

    // mail body (markdown syntax)
    "message" : "# Title 1\n - list1\n"
  }
  ```

* Response 204 (No content)
